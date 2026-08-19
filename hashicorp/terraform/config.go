package terraform

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/pkg/errors"
)

type Config struct {
	// Directory holding one subdirectory per workspace. Workspace names are
	// discovered from it rather than configured: any child directory containing
	// `.tf` files and not prefixed with `.` or `_` becomes a valid `workspace`
	// argument, whether or not it appears under `subscriptions`.
	Path string `json:"path" yaml:"path"`
	// Azure tenant ID — sets ARM_TENANT_ID for all workspaces; required for CLI auth via az login --allow-no-subscriptions
	TenantID string `json:"tenantId,omitempty" yaml:"tenantId,omitempty"`
	// Per-workspace Azure subscription and backend settings, keyed by workspace
	// directory name. Optional: a workspace with no entry still runs, but gets no
	// `ARM_SUBSCRIPTION_ID` and no `-backend-config` flags on `init`.
	Subscriptions map[string]Subscription `json:"subscriptions" yaml:"subscriptions"`
	// Service principals selectable with `--service-principal`, keyed by the name
	// offered in completion. Selecting one overrides the default Azure auth chain
	// for that invocation, including the subscription.
	ServicePrincipals map[string]ServicePrincipal `json:"servicePrincipals" yaml:"servicePrincipals"`
}

type Subscription struct {
	// Azure subscription ID
	ID string `json:"id" yaml:"id"`
	// Backend state storage configuration
	Backend Backend `json:"backend" yaml:"backend"`
	// Named SSH proxy for this workspace. Unused — nothing in this provider reads
	// it, so setting it has no effect.
	Proxy string `json:"proxy" yaml:"proxy"`
}

type Backend struct {
	// Resource group containing the storage account
	ResourceGroupName string `json:"resourceGroupName" yaml:"resourceGroupName"`
	// Storage account name
	StorageAccountName string `json:"storageAccountName" yaml:"storageAccountName"`
	// Blob container name
	ContainerName string `json:"containerName" yaml:"containerName"`
	// State file key (defaults to <workspace>.tfstate)
	Key string `json:"key" yaml:"key"`
}

type ServicePrincipal struct {
	// Tenant ID
	TenantID string `json:"tenantId" yaml:"tenantId"`
	// Application client ID
	ClientID string `json:"clientId" yaml:"clientId"`
	// Application client secret, in plaintext. Held in the project config and
	// exported as ARM_CLIENT_SECRET, so keep it out of version control.
	ClientSecret string `json:"clientSecret" yaml:"clientSecret"`
	// Azure subscription ID
	SubscriptionID string `json:"subscriptionId" yaml:"subscriptionId"`
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (c Config) WorkspaceNames() []string {
	var ret []string

	entries, err := os.ReadDir(c.Path)
	if err != nil {
		return ret
	}

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") || strings.HasPrefix(entry.Name(), "_") {
			continue
		}

		// only include dirs that contain .tf files
		if c.hasTFFiles(filepath.Join(c.Path, entry.Name())) {
			ret = append(ret, entry.Name())
		}
	}

	return ret
}

func (c Config) WorkspacePath(workspace string) string {
	return filepath.Join(c.Path, workspace)
}

// BackendInitArgs returns -backend-config flags for `terraform init`.
// Terraform's azurerm backend does not read storage config from env vars;
// these values must be supplied as CLI flags.
func (c Config) BackendInitArgs(workspace string) []string {
	sub, ok := c.Subscriptions[workspace]
	if !ok {
		return nil
	}

	b := sub.Backend
	if b.ResourceGroupName == "" || b.StorageAccountName == "" || b.ContainerName == "" {
		return nil
	}

	key := b.Key
	if key == "" {
		key = workspace + ".tfstate"
	}

	return []string{
		"-backend-config=resource_group_name=" + b.ResourceGroupName,
		"-backend-config=storage_account_name=" + b.StorageAccountName,
		"-backend-config=container_name=" + b.ContainerName,
		"-backend-config=key=" + key,
	}
}

// WorkspaceTargets parses all .tf files in the workspace directory and returns
// a sorted list of addressable targets: "module.NAME" and "TYPE.NAME" for resources.
func (c Config) WorkspaceTargets(workspace string) []string {
	dir := c.WorkspacePath(workspace)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	moduleRe := regexp.MustCompile(`module\s+"([^"]+)"`)
	resourceRe := regexp.MustCompile(`resource\s+"([^"]+)"\s+"([^"]+)"`)

	seen := map[string]struct{}{}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".tf") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}

		for _, m := range moduleRe.FindAllSubmatch(data, -1) {
			seen["module."+string(m[1])] = struct{}{}
		}

		for _, m := range resourceRe.FindAllSubmatch(data, -1) {
			seen[string(m[1])+"."+string(m[2])] = struct{}{}
		}
	}

	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}

	sort.Strings(out)

	return out
}

func (c Config) Subscription(name string) (Subscription, error) {
	value, ok := c.Subscriptions[name]
	if !ok {
		return Subscription{}, errors.Errorf("subscription not found: %s", name)
	}

	return value, nil
}

func (c Config) SubscriptionNames() []string {
	keys := make([]string, 0, len(c.Subscriptions))
	for k := range c.Subscriptions {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func (c Config) ServicePrincipal(name string) (ServicePrincipal, error) {
	value, ok := c.ServicePrincipals[name]
	if !ok {
		return ServicePrincipal{}, errors.Errorf("service principal not found: %s", name)
	}

	return value, nil
}

func (c Config) ServicePrincipalNames() []string {
	keys := make([]string, 0, len(c.ServicePrincipals))
	for k := range c.ServicePrincipals {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

// ------------------------------------------------------------------------------------------------
// ~ Private methods
// ------------------------------------------------------------------------------------------------

func (c Config) hasTFFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tf") {
			return true
		}
	}

	return false
}

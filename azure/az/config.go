package az

import (
	"sort"

	"github.com/pkg/errors"
	"github.com/samber/lo"
)

// Config holds the Azure CLI's local config directory, tenant and the
// subscriptions and service principals available to the az command.
type Config struct {
	// Config path
	ConfigPath string `json:"configPath" yaml:"configPath"`
	// Tenant id
	TenantID string `json:"tenantId" yaml:"tenantId"`
	// Named SSH proxy for az login (tenant-level commands)
	Proxy string `json:"proxy,omitempty" yaml:"proxy,omitempty"`
	// Subscription configurations
	Subscriptions map[string]Subscription `json:"subscriptions" yaml:"subscriptions"`
	// Authentication service principals
	ServicePrincipals map[string]ServicePrincipal `json:"servicePrincipals" yaml:"servicePrincipals"`
}

// ServicePrincipal holds the credentials for a non-interactive `az login`.
type ServicePrincipal struct {
	// Tenant id
	TenantID string `json:"tenantId" yaml:"tenantId"`
	// Application client id
	ClientID string `json:"clientId" yaml:"clientId"`
	// Application password
	ClientSecret string `json:"clientSecret" yaml:"clientSecret"`
}

// Subscription returns the named subscription configuration. It returns an
// error if the name is not configured.
func (c Config) Subscription(name string) (Subscription, error) {
	value, ok := c.Subscriptions[name]
	if !ok {
		return Subscription{}, errors.Errorf("resource group not found: %s", name)
	}

	return value, nil
}

// SubscriptionNames returns the configured subscription names, sorted.
func (c Config) SubscriptionNames() []string {
	keys := lo.Keys(c.Subscriptions)
	sort.Strings(keys)

	return keys
}

// ServicePrincipal returns the named service principal configuration. It
// returns an error if the name is not configured.
func (c Config) ServicePrincipal(name string) (ServicePrincipal, error) {
	value, ok := c.ServicePrincipals[name]
	if !ok {
		return ServicePrincipal{}, errors.Errorf("service principal not found: %s", name)
	}

	return value, nil
}

// ServicePrincipalNames returns the configured service principal names, sorted.
func (c Config) ServicePrincipalNames() []string {
	keys := lo.Keys(c.ServicePrincipals)
	sort.Strings(keys)

	return keys
}

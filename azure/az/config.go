package az

import (
	"sort"

	"github.com/pkg/errors"
	"github.com/samber/lo"
)

type Config struct {
	// Config path
	ConfigPath string `json:"configPath" yaml:"configPath"`
	// Tenant id
	TenantID string `json:"tenantId" yaml:"tenantId"`
	// Subscription configurations
	Subscriptions map[string]Subscription `json:"subscriptions" yaml:"subscriptions"`
	// Authentication service principals
	ServicePrincipals map[string]ServicePrincipal `json:"servicePrincipals" yaml:"servicePrincipals"`
}

type ServicePrincipal struct {
	// Tenant id
	TenantID string `json:"tenantId" yaml:"tenantId"`
	// Application client id
	ClientID string `json:"clientId" yaml:"clientId"`
	// Application password
	ClientSecret string `json:"clientSecret" yaml:"clientSecret"`
}

func (c Config) Subscription(name string) (Subscription, error) {
	value, ok := c.Subscriptions[name]
	if !ok {
		return Subscription{}, errors.Errorf("resource group not found: %s", name)
	}

	return value, nil
}

func (c Config) SubscriptionNames() []string {
	keys := lo.Keys(c.Subscriptions)
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
	keys := lo.Keys(c.ServicePrincipals)
	sort.Strings(keys)

	return keys
}

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

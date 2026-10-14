package k3d

import (
	"context"
	"encoding/json"
	"os"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/shell"
	"github.com/spf13/viper"
)

type (
	// K3d wraps the k3d CLI, exposing the configured clusters and registry and
	// the running clusters k3d itself reports.
	K3d struct {
		l         log.Logger
		cfg       *Config
		configKey string
	}
	// Option configures a K3d during construction.
	Option func(*K3d) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// WithConfigKey sets the viper key the instance reads its Config from,
// overriding the default "k3d".
func WithConfigKey(v string) Option {
	return func(o *K3d) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// New builds a K3d, reading its Config from the configKey (default "k3d")
// after applying opts, and sets K3D_FIX_DNS=1 in the process environment to
// work around github.com/k3d-io/k3d/issues/209. It returns an error if any
// Option, the config unmarshal, or setting the environment variable fails.
func New(l log.Logger, opts ...Option) (*K3d, error) {
	inst := &K3d{
		l:         l.Named("k3d"),
		configKey: "k3d",
	}

	for _, opt := range opts {
		if opt != nil {
			if err := opt(inst); err != nil {
				return nil, err
			}
		}
	}

	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	// https://github.com/k3d-io/k3d/issues/209
	if err := os.Setenv("K3D_FIX_DNS", "1"); err != nil {
		return nil, err
	}

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Getter
// ------------------------------------------------------------------------------------------------

// Config returns the instance's configuration.
func (i *K3d) Config() *Config {
	return i.cfg
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

// Registry returns the k3d registry named "k3d-"+name, or nil if no such
// registry exists.
func (i *K3d) Registry(ctx context.Context, name string) (*Registry, error) {
	out, err := shell.New(ctx, i.l, "k3d", "registry", "list", "--output", "json").Output()
	if err != nil {
		return nil, err
	}

	var registries []*Registry
	if err := json.Unmarshal(out, &registries); err != nil {
		return nil, err
	}

	for _, registry := range registries {
		if registry.Name == "k3d-"+name {
			return registry, nil
		}
	}

	return nil, nil //nolint: nilnil
}

// Clusters returns every k3d cluster currently running, whether or not it is
// one this project configures.
func (i *K3d) Clusters(ctx context.Context) ([]*Cluster, error) {
	out, err := shell.New(ctx, i.l, "k3d", "cluster", "list", "--output", "json").Output()
	if err != nil {
		return nil, err
	}

	var clusters []*Cluster
	if err := json.Unmarshal(out, &clusters); err != nil {
		return nil, err
	}

	return clusters, nil
}

// Cluster returns the running k3d cluster named name, or nil if it is not
// currently running.
func (i *K3d) Cluster(ctx context.Context, name string) (*Cluster, error) {
	clusters, err := i.Clusters(ctx)
	if err != nil {
		return nil, err
	}

	for _, cluster := range clusters {
		if cluster.Name == name {
			return cluster, nil
		}
	}

	return nil, nil //nolint: nilnil
}

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
	K3d struct {
		l         log.Logger
		cfg       *Config
		configKey string
	}
	Option func(*K3d) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func WithConfigKey(v string) Option {
	return func(o *K3d) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

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

func (i *K3d) Config() *Config {
	return i.cfg
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

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

func (i *K3d) Cluster(ctx context.Context, name string) (*Cluster, error) {
	out, err := shell.New(ctx, i.l, "k3d", "cluster", "list", "--output", "json").Output()
	if err != nil {
		return nil, err
	}

	var clusters []*Cluster
	if err := json.Unmarshal(out, &clusters); err != nil {
		return nil, err
	}

	for _, cluster := range clusters {
		if cluster.Name == name {
			return cluster, nil
		}
	}

	return nil, nil //nolint: nilnil
}

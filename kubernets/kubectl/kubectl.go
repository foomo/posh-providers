package kubectl

import (
	"context"
	"os"
	"path"
	"strings"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type (
	Kubectl struct {
		l                 log.Logger
		cfg               Config
		cache             cache.Namespace
		configKey         string
		authTokenProvider AuthTokenProvider
	}
	Option            func(*Kubectl) error
	AuthTokenProvider func(ctx context.Context, kubeContext string) (token string, err error)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithConfigKey(v string) Option {
	return func(o *Kubectl) error {
		o.configKey = v
		return nil
	}
}

func CommandWithAuthTokenProvider(provider AuthTokenProvider) Option {
	return func(o *Kubectl) error {
		o.authTokenProvider = provider
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func New(l log.Logger, cache cache.Cache, opts ...Option) (*Kubectl, error) {
	inst := &Kubectl{
		l:         l,
		cache:     cache.Get("kubectl"),
		configKey: "kubectl",
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

	if err := files.MkdirAll(inst.cfg.ConfigPath); err != nil {
		return nil, errors.Wrapf(err, "failed to create config path: %s", inst.cfg.ConfigPath)
	}

	if err := os.Setenv("KUBECONFIG", env.Path(inst.cfg.ConfigPath, "kubeconfig.yaml")); err != nil {
		return nil, errors.Wrapf(err, "failed to set KUBECONFIG")
	}

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (k *Kubectl) Config() Config {
	return k.cfg
}

func (k *Kubectl) Cluster(name string) *Cluster {
	return NewCluster(k, name)
}

//nolint:forcetypeassert
func (k *Kubectl) Clusters() Clusters {
	return k.cache.Get("clusters", func() any {
		// list files
		entries, err := os.ReadDir(k.cfg.ConfigPath)
		if err != nil {
			k.l.Debugf("failed to read config path: %s", err.Error())
			return Clusters{}
		}

		// get clusters
		clusters := Clusters{}
		for _, entry := range entries {
			cluster := strings.TrimSuffix(entry.Name(), ".yaml")
			if !entry.IsDir() && path.Ext(entry.Name()) == ".yaml" {
				if err := os.Chmod(path.Join(k.cfg.ConfigPath, entry.Name()), os.FileMode(0o600)); err != nil {
					k.l.Debugf("failed to chmod file: %s", err.Error())
				}
				clusters = append(clusters, NewCluster(k, cluster))
			}
		}

		return clusters
	}).(Clusters)
}

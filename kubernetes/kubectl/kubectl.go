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
	// Kubectl manages the kubeconfig files this project's clusters are stored
	// under, keyed by cluster name.
	Kubectl struct {
		l                 log.Logger
		cfg               Config
		cache             cache.Namespace
		configKey         string
		authTokenProvider AuthTokenProvider
	}
	// Option configures a Kubectl during construction.
	Option func(*Kubectl) error
	// AuthTokenProvider resolves a bearer token for kubeContext, used to
	// authenticate kubectl commands run against a Cluster.
	AuthTokenProvider func(ctx context.Context, kubeContext string) (token string, err error)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// CommandWithConfigKey sets the viper key the instance reads its Config from,
// overriding the default "kubectl".
func CommandWithConfigKey(v string) Option {
	return func(o *Kubectl) error {
		o.configKey = v
		return nil
	}
}

// CommandWithAuthTokenProvider sets the provider used to fetch a bearer token
// for `kubectl` invocations against a Cluster.
func CommandWithAuthTokenProvider(provider AuthTokenProvider) Option {
	return func(o *Kubectl) error {
		o.authTokenProvider = provider
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// New builds a Kubectl, reading its Config from the configKey (default
// "kubectl") after applying opts, creates the configured config path if
// missing, and sets KUBECONFIG to "<configPath>/kubeconfig.yaml" in the
// process environment. It returns an error if any Option, the config
// unmarshal, or either filesystem/environment step fails.
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

// Config returns the instance's configuration.
func (k *Kubectl) Config() Config {
	return k.cfg
}

// Cluster returns a Cluster handle for name. It does not check that a
// kubeconfig for that name exists.
func (k *Kubectl) Cluster(name string) *Cluster {
	return NewCluster(k, name)
}

// Clusters returns a Cluster for every "<name>.yaml" kubeconfig file found
// directly under the configured config path, chmod-ing each to 0600 along the
// way.
//
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

package doctl

import (
	"context"
	"os"
	"path"

	"github.com/digitalocean/godo"
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type (
	Doctl struct {
		l                 log.Logger
		cfg               Config
		cache             cache.Namespace
		configKey         string
		authTokenProvider AuthTokenProvider
	}
	Option            func(*Doctl) error
	AuthTokenProvider func(ctx context.Context, kubeContext string) (token string, err error)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithConfigKey(v string) Option {
	return func(o *Doctl) error {
		o.configKey = v
		return nil
	}
}

func CommandWithAuthTokenProvider(provider AuthTokenProvider) Option {
	return func(o *Doctl) error {
		o.authTokenProvider = provider
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func New(l log.Logger, cache cache.Cache, opts ...Option) (*Doctl, error) {
	inst := &Doctl{
		l:         l,
		cache:     cache.Get("doctl"),
		configKey: "doctl",
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

	if err := files.MkdirAll(path.Dir(inst.cfg.ConfigPath)); err != nil {
		return nil, errors.Wrapf(err, "failed to create config path: %s", path.Dir(inst.cfg.ConfigPath))
	}

	if err := os.Setenv("DIGITALOCEAN_CONFIG", env.Path(inst.cfg.ConfigPath)); err != nil {
		return nil, errors.Wrapf(err, "failed to set DIGITALOCEAN_CONFIG")
	}

	if value, err := inst.AccessToken(); err != nil && value != "" {
		if err := os.Setenv("DIGITALOCEAN_ACCESS_TOKEN", value); err != nil {
			return nil, errors.Wrapf(err, "failed to set DIGITALOCEAN_ACCESS_TOKEN")
		}
	}

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (k *Doctl) Client() (*godo.Client, error) {
	accessToken, err := k.AccessToken()
	if err != nil {
		return nil, err
	}
	return godo.NewFromToken(accessToken), nil
}

func (k *Doctl) AccessToken() (string, error) {
	if _, err := os.Stat(k.cfg.ConfigPath); err != nil {
		return "", err
	}

	yamlFile, err := os.ReadFile(k.cfg.ConfigPath)
	if err != nil {
		return "", err
	}

	var c struct {
		AccessToken string `yaml:"access-token"`
	}
	if err = yaml.Unmarshal(yamlFile, &c); err != nil {
		return "", err
	}

	return c.AccessToken, nil
}

func (k *Doctl) Cluster(ctx context.Context, name string) (*godo.KubernetesCluster, error) {
	c, err := k.Client()
	if err != nil {
		return nil, err
	}
	cluster, _, err := c.Kubernetes.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

//nolint:forcetypeassert
func (k *Doctl) Clusters(ctx context.Context) []*godo.KubernetesCluster {
	return k.cache.Get("clusters", func() any {
		c, err := k.Client()
		if err != nil {
			k.l.Info(err.Error())
			return []*godo.KubernetesCluster{}
		}
		clusters, _, err := c.Kubernetes.List(ctx, nil)
		if err != nil {
			k.l.Info(err.Error())
			return []*godo.KubernetesCluster{}
		}
		return clusters
	}).([]*godo.KubernetesCluster)
}

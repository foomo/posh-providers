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
	// Doctl reads the provider's configuration and exposes the DigitalOcean API
	// client and the clusters it is allowed to reach.
	Doctl struct {
		l                 log.Logger
		cfg               Config
		cache             cache.Namespace
		configKey         string
		authTokenProvider AuthTokenProvider
	}
	// Option configures a Doctl built by New.
	Option func(*Doctl) error
	// AuthTokenProvider resolves a DigitalOcean API token for the given kube
	// context.
	AuthTokenProvider func(ctx context.Context, kubeContext string) (token string, err error)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// CommandWithConfigKey overrides the viper key the Config is unmarshalled
// from, which defaults to "doctl".
func CommandWithConfigKey(v string) Option {
	return func(o *Doctl) error {
		o.configKey = v
		return nil
	}
}

// CommandWithAuthTokenProvider sets the AuthTokenProvider used to resolve the
// DigitalOcean API token.
func CommandWithAuthTokenProvider(provider AuthTokenProvider) Option {
	return func(o *Doctl) error {
		o.authTokenProvider = provider
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// New reads the doctl Config from viper, creates the config directory,
// exports DIGITALOCEAN_CONFIG so the doctl binary reads the configured file,
// and exports DIGITALOCEAN_ACCESS_TOKEN when an access token is already
// present. It returns an error if the config cannot be unmarshalled or the
// config directory cannot be created.
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

// Client returns a DigitalOcean API client authenticated with the access
// token read from the doctl config file.
func (k *Doctl) Client() (*godo.Client, error) {
	accessToken, err := k.AccessToken()
	if err != nil {
		return nil, err
	}

	return godo.NewFromToken(accessToken), nil
}

// AccessToken reads and returns the access token stored in the doctl config
// file, written there by `auth init`. It returns an error if the file does
// not exist or cannot be parsed.
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

// Cluster fetches the DigitalOcean Kubernetes cluster with the given name or
// ID from the API.
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

// Clusters returns every DigitalOcean Kubernetes cluster reachable with the
// configured access token, caching the result. It returns an empty slice
// rather than an error if the API call fails.
//
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

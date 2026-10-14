package gcloud

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"regexp"

	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/shell"
	"github.com/pkg/errors"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/spf13/viper"
)

type (
	// GCloud holds the resolved gcloud configuration and sets up the local
	// gcloud SDK state directory (CLOUDSDK_CONFIG) for the project.
	GCloud struct {
		l                    log.Logger
		cfg                  *Config
		cache                cache.Namespace
		configKey            string
		accountFileNameRegex *regexp.Regexp
	}
	// Option configures a GCloud in New.
	Option func(*GCloud)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// WithConfig sets the Config directly, bypassing viper unmarshalling.
func WithConfig(v *Config) Option {
	return func(o *GCloud) {
		o.cfg = v
	}
}

// WithConfigKey sets the viper key New reads its Config from, overriding the
// default "gcloud". Ignored if WithConfig is also given.
func WithConfigKey(v string) Option {
	return func(o *GCloud) {
		o.configKey = v
	}
}

// WithAccountFileNameRegex overrides the regular expression used to parse
// service account key filenames, replacing the default
// `(\w+)@(\w+)-(\w+)\.json`.
func WithAccountFileNameRegex(v *regexp.Regexp) Option {
	return func(o *GCloud) {
		o.accountFileNameRegex = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// New creates a GCloud, unmarshalling its Config from viper under configKey
// ("gcloud" unless overridden by WithConfigKey or WithConfig), creating the
// configured ConfigPath, and setting CLOUDSDK_CONFIG to it. It returns an
// error if the config cannot be unmarshalled, the directory cannot be
// created, or the environment variable cannot be set.
func New(l log.Logger, cache cache.Cache, opts ...Option) (*GCloud, error) {
	inst := &GCloud{
		l:                    l,
		cache:                cache.Get("gcloud"),
		configKey:            "gcloud",
		accountFileNameRegex: regexp.MustCompile(`(\w+)@(\w+)-(\w+)\.json`),
	}

	for _, opt := range opts {
		if opt != nil {
			opt(inst)
		}
	}

	if inst.cfg == nil && inst.configKey != "" {
		if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
			return nil, err
		}
	}

	// ensure config path
	if err := files.MkdirAll(inst.cfg.ConfigPath); err != nil {
		return nil, errors.Wrapf(err, "failed to create directory %q", inst.cfg.ConfigPath)
	}

	if err := os.Setenv("CLOUDSDK_CONFIG", env.Path(inst.cfg.ConfigPath)); err != nil {
		return nil, err
	}

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

// Config returns the resolved gcloud configuration.
func (p *GCloud) Config() *Config {
	return p.cfg
}

// ServiceAccountKeysPath returns the directory where retrieved service
// account key files are cached, under the configured ConfigPath.
func (p *GCloud) ServiceAccountKeysPath() string {
	return path.Join(p.cfg.ConfigPath, "service_account_keys")
}

// ActiveAccount returns the active account email
func (p *GCloud) ActiveAccount(ctx context.Context, l log.Logger) (string, error) {
	out, err := shell.New(ctx, l, "gcloud", "auth", "list", "--format", "json").Quiet().Output()
	if err != nil {
		return "", err
	}

	var list []struct {
		Account string `json:"account"`
		Status  string `json:"status"`
	}
	if err := json.Unmarshal(out, &list); err != nil {
		return "", err
	}

	for _, item := range list {
		if item.Status == "ACTIVE" {
			return item.Account, nil
		}
	}

	return "", errors.New("no active account")
}

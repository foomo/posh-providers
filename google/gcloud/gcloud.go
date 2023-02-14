package gcloud

import (
	"os"
	"path"
	"regexp"

	"github.com/foomo/posh/pkg/env"
	"github.com/pkg/errors"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/spf13/viper"
)

type (
	GCloud struct {
		l                    log.Logger
		cfg                  *Config
		cache                cache.Namespace
		configKey            string
		accountFileNameRegex *regexp.Regexp
	}
	Option func(*GCloud)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func WithConfig(v *Config) Option {
	return func(o *GCloud) {
		o.cfg = v
	}
}

func WithConfigKey(v string) Option {
	return func(o *GCloud) {
		o.configKey = v
	}
}

func WithAccountFileNameRegex(v *regexp.Regexp) Option {
	return func(o *GCloud) {
		o.accountFileNameRegex = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

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

	if err := os.Setenv("CLOUDSDK_CONFIG", path.Join(os.Getenv(env.ProjectRoot), inst.cfg.ConfigPath)); err != nil {
		return nil, err
	}

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (p GCloud) ServiceAccountKeysPath() string {
	return path.Join(p.cfg.ConfigPath, "service_account_keys")
}

package gcloud

import (
	"os"
	"path"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/spf13/viper"
)

type (
	GCloud struct {
		l         log.Logger
		cfg       Config
		cache     cache.Namespace
		configKey string
	}
	Option func(*GCloud) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithConfigKey(v string) Option {
	return func(o *GCloud) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func New(l log.Logger, cache cache.Cache, opts ...Option) (*GCloud, error) {
	inst := &GCloud{
		l:         l,
		cache:     cache.Get("gcloud"),
		configKey: "gcloud",
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
		return nil, err
	} else if err := os.Setenv("CLOUDSDK_CONFIG", path.Join(os.Getenv(env.ProjectRoot), inst.cfg.ConfigPath)); err != nil {
		return nil, err
	}

	if err := os.Setenv("CLOUDSDK_CORE_PROJECT", inst.cfg.Project); err != nil {
		return nil, err
	}

	return inst, nil
}

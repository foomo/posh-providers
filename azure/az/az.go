package az

import (
	"os"
	"path"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/env"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/util/files"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type (
	AZ struct {
		l         log.Logger
		cfg       Config
		cache     cache.Namespace
		configKey string
	}
	Option func(*AZ) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithConfigKey(v string) Option {
	return func(o *AZ) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func New(l log.Logger, cache cache.Cache, opts ...Option) (*AZ, error) {
	inst := &AZ{
		l:         l,
		cache:     cache.Get("az"),
		configKey: "az",
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
		return nil, errors.Wrapf(err, "failed to create config path: %s", inst.cfg.ConfigPath)
	}

	if err := os.Setenv("AZURE_CONFIG_DIR", env.Path(inst.cfg.ConfigPath)); err != nil {
		return nil, errors.Wrapf(err, "failed to set AZURE_CONFIG_DIR")
	}

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Getter
// ------------------------------------------------------------------------------------------------

func (a *AZ) Config() Config {
	return a.cfg
}

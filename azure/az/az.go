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
	// AZ wraps the Azure CLI, loading its Config and pointing AZURE_CONFIG_DIR
	// at a project-local directory.
	AZ struct {
		l         log.Logger
		cfg       Config
		cache     cache.Namespace
		configKey string
	}
	// Option configures an AZ in New.
	Option func(*AZ) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// WithConfigKey sets the viper key the Config is unmarshalled from,
// overriding the default "az".
func WithConfigKey(v string) Option {
	return func(o *AZ) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// New builds an AZ, unmarshalling its Config from the configured viper key,
// creating the configured ConfigPath directory, and pointing the az CLI at it
// via AZURE_CONFIG_DIR. It returns an error if any of those steps fail.
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

// Config returns the loaded configuration.
func (a *AZ) Config() Config {
	return a.cfg
}

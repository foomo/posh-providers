package gnat

import (
	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
)

type (
	GNAT struct {
		l         log.Logger
		cfg       Config
		cache     cache.Namespace
		configKey string
	}
	Option func(*GNAT) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func WithConfigKey(v string) Option {
	return func(o *GNAT) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func New(l log.Logger, cache cache.Cache, opts ...Option) (*GNAT, error) {
	inst := &GNAT{
		l:         l,
		cache:     cache.Get("gnat"),
		configKey: "gnat",
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

	return inst, nil
}

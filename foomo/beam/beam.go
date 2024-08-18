package beam

import (
	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
)

type (
	Beam struct {
		l         log.Logger
		cfg       Config
		op        *onepassword.OnePassword
		configKey string
	}
	Option func(*Beam) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func WithConfigKey(v string) Option {
	return func(o *Beam) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// New command
func New(l log.Logger, op *onepassword.OnePassword, opts ...Option) (*Beam, error) {
	inst := &Beam{
		l:         l,
		op:        op,
		configKey: "beam",
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

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (t *Beam) Config() Config {
	return t.cfg
}

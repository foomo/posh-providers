package stackit

import (
	"context"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
)

type (
	Stackit struct {
		l         log.Logger
		cfg       Config
		cache     cache.Namespace
		configKey string
	}
	Option            func(*Stackit) error
	AuthTokenProvider func(ctx context.Context, kubeContext string) (token string, err error)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithConfigKey(v string) Option {
	return func(o *Stackit) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

func New(l log.Logger, cache cache.Cache, opts ...Option) (*Stackit, error) {
	inst := &Stackit{
		l:         l,
		cache:     cache.Get("stackit"),
		configKey: "stackit",
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
// ~ Getter
// ------------------------------------------------------------------------------------------------

func (s *Stackit) Config() Config {
	return s.cfg
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (s *Stackit) ClusterNames(projectName string) ([]string, error) {
	project, err := s.cfg.Project(projectName)
	if err != nil {
		return nil, err
	}
	return project.ClusterNames(), nil
}

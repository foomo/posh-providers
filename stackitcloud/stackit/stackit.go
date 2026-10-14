package stackit

import (
	"context"

	"github.com/foomo/posh/pkg/cache"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
)

type (
	// Stackit holds the project/cluster configuration shared by stackit commands.
	Stackit struct {
		l         log.Logger
		cfg       Config
		cache     cache.Namespace
		configKey string
	}
	// Option configures a Stackit built by New.
	Option func(*Stackit) error
	// AuthTokenProvider resolves an auth token for kubeContext.
	AuthTokenProvider func(ctx context.Context, kubeContext string) (token string, err error)
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// CommandWithConfigKey overrides the viper key the Stackit config is read
// from, which defaults to "stackit".
func CommandWithConfigKey(v string) Option {
	return func(o *Stackit) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// New builds a Stackit, unmarshalling its config from viper under configKey.
// It returns an error if the unmarshalling or an option fails.
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

// Config returns the Stackit's configuration.
func (s *Stackit) Config() Config {
	return s.cfg
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

// ClusterNames returns the configured cluster names for the project keyed as
// projectName. It returns an error if projectName is not configured.
func (s *Stackit) ClusterNames(projectName string) ([]string, error) {
	project, err := s.cfg.Project(projectName)
	if err != nil {
		return nil, err
	}

	return project.ClusterNames(), nil
}

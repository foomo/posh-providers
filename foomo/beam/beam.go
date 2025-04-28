package beam

import (
	"fmt"

	gokaziconfig "github.com/foomo/gokazi/pkg/config"
	"github.com/foomo/gokazi/pkg/gokazi"
	"github.com/foomo/posh-providers/onepassword"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
)

type (
	Beam struct {
		l         log.Logger
		cfg       Config
		gk        *gokazi.Gokazi
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
func New(l log.Logger, op *onepassword.OnePassword, gk *gokazi.Gokazi, opts ...Option) (*Beam, error) {
	inst := &Beam{
		l:         l,
		op:        op,
		gk:        gk,
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

	for key, value := range inst.cfg.Clusters {
		inst.gk.Add("beam.cluster."+key, gokaziconfig.Task{
			Name:        "cloudflared",
			Description: fmt.Sprintf("Cloudflare tunnel to cluster: '%s' [:%d]", key, value.Port),
			Args: []string{
				"access", "tcp",
				"--hostname", value.Hostname,
				"--url", fmt.Sprintf("127.0.0.1:%d", value.Port),
			},
		})
	}
	for key, value := range inst.cfg.Databases {
		inst.gk.Add("beam.database."+key, gokaziconfig.Task{
			Name:        "cloudflared",
			Description: fmt.Sprintf("Cloudflare tunnel to database: '%s' [:%d]", key, value.Port),
			Args: []string{
				"access", "tcp",
				"--hostname", value.Hostname,
				"--url", fmt.Sprintf("127.0.0.1:%d", value.Port),
			},
		})
	}
	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (t *Beam) Config() Config {
	return t.cfg
}

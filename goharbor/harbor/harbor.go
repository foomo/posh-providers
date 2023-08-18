package harbor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/shell"
	"github.com/spf13/viper"
)

// Harbor command
type (
	Harbor struct {
		l            log.Logger
		cfg          Config
		configKey    string
		signedIn     bool
		signedInTime time.Time
	}
	Option func(*Harbor) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func CommandWithConfigKey(v string) Option {
	return func(o *Harbor) error {
		o.configKey = v
		return nil
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// New command
func New(l log.Logger, opts ...Option) (*Harbor, error) {
	inst := &Harbor{
		l:         l,
		configKey: "harbor",
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

func (t *Harbor) Config() Config {
	return t.cfg
}

func (t *Harbor) IsAuthenticated(ctx context.Context) bool {
	if t.signedIn && time.Since(t.signedInTime) < time.Hour {
		return true
	}
	out, _ := shell.New(ctx, t.l,
		"docker",
		"pull",
		fmt.Sprintf("%s/%s/null:null", t.cfg.DockerRegistry(), t.cfg.Project),
	).Quiet().CombinedOutput()
	if strings.HasPrefix(string(out), "Error response from daemon: unknown") {
		t.signedIn = true
		t.signedInTime = time.Now()
		return true
	}
	t.signedIn = false
	return false
}

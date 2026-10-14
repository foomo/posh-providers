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

type (
	// Harbor tracks the configured Harbor registry and a short-lived cache of
	// whether the local docker daemon is authenticated against it.
	Harbor struct {
		l            log.Logger
		cfg          Config
		configKey    string
		signedIn     bool
		signedInTime time.Time
	}
	// Option configures a Harbor in New.
	Option func(*Harbor) error
)

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// CommandWithConfigKey sets the viper key Harbor reads its Config from,
// overriding the default "harbor".
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

// Config returns the Harbor registry configuration.
func (t *Harbor) Config() Config {
	return t.cfg
}

// IsAuthenticated reports whether the local docker daemon is currently
// logged in to the configured registry. It caches a positive result for an
// hour; otherwise it probes by pulling a nonexistent image from the
// configured project and checking the daemon's error response.
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

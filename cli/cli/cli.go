package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/foomo/go/options"
	"github.com/foomo/posh/pkg/exec"
	"github.com/foomo/posh/pkg/log"
	"github.com/spf13/viper"
)

// CLI wraps the GitHub CLI (gh), loading its Config and ensuring
// GITHUB_TOKEN is set from `gh auth token` for other commands to use.
type CLI struct {
	l         log.Logger
	cfg       Config
	configKey string
}

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

// WithConfigKey sets the viper key the Config is unmarshalled from,
// overriding the default "gh".
func WithConfigKey(v string) options.Option[*CLI] {
	return func(o *CLI) {
		o.configKey = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

// New builds a CLI, unmarshalling its Config from the configured viper key.
// It returns an error if that unmarshal fails.
func New(l log.Logger, opts ...options.Option[*CLI]) (*CLI, error) {
	inst := &CLI{
		l:         l,
		configKey: "gh",
	}

	options.Apply(inst, opts...)

	if err := viper.UnmarshalKey(inst.configKey, &inst.cfg); err != nil {
		return nil, err
	}

	return inst, nil
}

// ------------------------------------------------------------------------------------------------
// ~ Getter
// ------------------------------------------------------------------------------------------------

// Config returns the loaded configuration.
func (c *CLI) Config() Config {
	return c.cfg
}

// LoadToken sets GITHUB_TOKEN from `gh auth token` unless it is already set.
// It returns an error if that call fails or the environment variable cannot
// be set.
func (c *CLI) LoadToken(ctx context.Context) error {
	if os.Getenv("GITHUB_TOKEN") != "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()

	var s strings.Builder
	if err := exec.NewCommand(ctx, "gh", "auth", "token").Stdout(&s).Run(); err != nil {
		return fmt.Errorf("failed to get github token: %w", err)
	}

	if err := os.Setenv("GITHUB_TOKEN", strings.Trim(s.String(), "\n")); err != nil {
		return fmt.Errorf("failed to set github token: %w", err)
	}

	return nil
}

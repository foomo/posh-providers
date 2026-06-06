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

type CLI struct {
	l         log.Logger
	cfg       Config
	configKey string
}

// ------------------------------------------------------------------------------------------------
// ~ Options
// ------------------------------------------------------------------------------------------------

func WithConfigKey(v string) options.Option[*CLI] {
	return func(o *CLI) {
		o.configKey = v
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Constructor
// ------------------------------------------------------------------------------------------------

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

func (c *CLI) Config() Config {
	return c.cfg
}

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

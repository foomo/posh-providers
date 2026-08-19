package rclone

import (
	"context"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type Config struct {
	// Path of the generated rclone configuration file, relative to the project
	// root. Exported as RCLONE_CONFIG for every command in the shell session, so
	// rclone reads this file instead of the user's own ~/.config/rclone.
	Path string `json:"path" yaml:"path"`
	// Config is the rclone configuration file template in rclone's INI format,
	// written verbatim to Path by `rclone init`. Values may contain 1Password
	// secret references (op://vault/item/field), which are resolved by piping
	// the template through `op inject`.
	Config string `json:"config" yaml:"config"`
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (c Config) RenderConfig(ctx context.Context) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "op", "inject")
	cmd.Stdin = strings.NewReader(c.Config)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to inject config: %s", out)
	}

	return out, nil
}

package rclone

import (
	"context"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

type Config struct {
	Path   string `json:"path" yaml:"path"`
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

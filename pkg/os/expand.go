package os

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/allex/envsubst"
)

// ExpandEnv expands environment variables and injects supported secret refs
func ExpandEnv(ctx context.Context, v string) (string, error) {
	v, err := envsubst.String(v)
	if err != nil {
		return "", fmt.Errorf("failed to expand environment variables: %w", err)
	}

	switch {
	case strings.HasPrefix(v, "op://"):
		out, err := exec.CommandContext(ctx, "op", "get", v).Output()
		if err != nil {
			return "", fmt.Errorf("failed to inject 1Password secret: %w", err)
		}

		v = string(out)
	}

	return v, nil
}

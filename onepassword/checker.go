package onepassword

import (
	"context"
	"os/exec"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
)

func AuthChecker(op *OnePassword) check.Checker {
	return func(ctx context.Context, l log.Logger) []check.Info {
		name := "1Password"

		err := exec.CommandContext(ctx, "op", "whoami", "--account", op.cfg.Account).Run()
		if err != nil {
			return []check.Info{check.NewNoteInfo("⛹", name, "Disconnected")}
		}

		return []check.Info{check.NewSuccessInfo("⛹", name, op.cfg.Account)}
	}
}

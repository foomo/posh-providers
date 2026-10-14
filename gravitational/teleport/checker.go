package teleport

import (
	"context"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
)

// AuthChecker returns a check.Checker that reports whether p has an
// authenticated tsh session, for display in the posh prompt.
func AuthChecker(p *Teleport) check.Checker {
	return func(ctx context.Context, l log.Logger) []check.Info {
		name := "Teleport"

		if p.IsAuthenticated(ctx) {
			return []check.Info{check.NewSuccessInfo("⛹", name, p.Config().Hostname)}
		} else {
			return []check.Info{check.NewNoteInfo("⛹", name, "Unauthenticated")}
		}
	}
}

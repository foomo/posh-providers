package harbor

import (
	"context"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
)

func AuthChecker(h *Harbor) check.Checker {
	return func(ctx context.Context, l log.Logger) []check.Info {
		name := "Harbor"
		if h.IsAuthenticated(ctx) {
			return []check.Info{check.NewSuccessInfo(name, "Authenticated")}
		} else {
			return []check.Info{check.NewFailureInfo(name, "Run `harbor auth` to sign into docker")}
		}
	}
}

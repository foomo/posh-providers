package teleport

import (
	"context"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
)

func AuthChecker(p *Teleport) check.Checker {
	return func(ctx context.Context, l log.Logger) check.Info {
		name := "Teleport"
		if p.IsAuthenticated(ctx) {
			return check.NewSuccessInfo(name, "Authenticated")
		} else {
			return check.NewFailureInfo(name, "Run `teleport auth` to sign into teleport")
		}
	}
}

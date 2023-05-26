package onepassword

import (
	"context"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
)

func AuthChecker(p *OnePassword) check.Checker {
	return func(ctx context.Context, l log.Logger) check.Info {
		name := "1Password"
		if ok, _ := p.IsAuthenticated(); ok {
			return check.NewSuccessInfo(name, "Authenticated")
		} else {
			return check.NewFailureInfo(name, "Run `op auth` to sign into 1password")
		}
	}
}

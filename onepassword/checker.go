package onepassword

import (
	"context"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
)

func SessionChecker(p *OnePassword) check.Checker {
	return func(ctx context.Context, l log.Logger) check.Info {
		name := "1Password: Session"
		if ok, _ := p.Session(); ok {
			return check.NewSuccessInfo(name, "Signed in")
		} else {
			return check.NewFailureInfo(name, "Run `op signin` to sign into 1password")
		}
	}
}

package gcloud

import (
	"context"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
)

func AuthChecker(p *GCloud) check.Checker {
	return func(ctx context.Context, l log.Logger) []check.Info {
		name := "GCloud"
		if account, err := p.ActiveAccount(ctx, l); err != nil {
			return []check.Info{check.NewSuccessInfo("\uF084", name, "Unauthenticated")}
		} else {
			return []check.Info{check.NewSuccessInfo("\uF084", name, account)}
		}
	}
}

package gcloud

import (
	"context"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
)

func AccountChecker(p *GCloud) check.Checker {
	return func(ctx context.Context, l log.Logger) check.Info {
		name := "GCloud: Account"
		if account, err := p.ActiveAccount(ctx, l); err != nil {
			return check.NewFailureInfo(name, "Error: "+err.Error())
		} else {
			return check.NewSuccessInfo(name, account)
		}
	}
}

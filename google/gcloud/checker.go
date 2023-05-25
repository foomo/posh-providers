package gcloud

import (
	"context"
	"fmt"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
)

func AuthChecker(p *GCloud) check.Checker {
	return func(ctx context.Context, l log.Logger) check.Info {
		name := "GCloud"
		if account, err := p.ActiveAccount(ctx, l); err != nil {
			return check.NewFailureInfo(name, "Error: "+err.Error())
		} else {
			return check.NewSuccessInfo(name, fmt.Sprintf("Authenticated (%s)", account))
		}
	}
}

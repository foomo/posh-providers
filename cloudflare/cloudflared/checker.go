package cloudflared

import (
	"context"
	"fmt"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
)

func AccessChecker(cf *Cloudflared, access Access) check.Checker {
	return func(ctx context.Context, l log.Logger) []check.Info {
		name := "Cloudflared"

		note := fmt.Sprintf("%s:%d", access.Hostname, access.Port)
		if cf.IsConnected(ctx, access) {
			return []check.Info{check.NewSuccessInfo("\uF084", name, note)}
		}

		return []check.Info{check.NewNoteInfo("\uEB11", name, note)}
	}
}

package beam

import (
	"context"
	"fmt"

	"github.com/foomo/posh-providers/cloudflare/cloudflared"
	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
)

func ClusterChecker(cf *cloudflared.Cloudflared, cluster Cluster) check.Checker {
	return func(ctx context.Context, l log.Logger) check.Info {
		name := "Beam Cluster"
		title := fmt.Sprintf("%s => :%d", cluster.Hostname, cluster.Port)

		if cf.IsConnected(ctx, cloudflared.Access{
			Type:     "tcp",
			Hostname: cluster.Hostname,
			Port:     cluster.Port,
		}) {
			return check.NewSuccessInfo(name, title)
		}

		return check.NewNoteInfo(name, title)
	}
}

func DatabaseChecker(cf *cloudflared.Cloudflared, database Database) check.Checker {
	return func(ctx context.Context, l log.Logger) check.Info {
		name := "Beam Database"
		title := fmt.Sprintf("%s => :%d", database.Hostname, database.Port)

		if cf.IsConnected(ctx, cloudflared.Access{
			Type:     "tcp",
			Hostname: database.Hostname,
			Port:     database.Port,
		}) {
			return check.NewSuccessInfo(name, title)
		}

		return check.NewNoteInfo(name, title)
	}
}

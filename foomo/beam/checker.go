package beam

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
)

func TunnelChecker(p *Beam, tunnel, cluster string) check.Checker {
	return func(ctx context.Context, l log.Logger) check.Info {
		name := "Beam"
		c := p.Config().GetTunnel(tunnel).GetCluster(cluster)
		addr := fmt.Sprintf("127.0.0.1:%d", c.Port)
		if _, err := net.DialTimeout("tcp", addr, time.Second); err != nil {
			return check.NewNoteInfo(name, fmt.Sprintf("Tunnel `%s` to cluster `%s` is closed", tunnel, cluster))
		} else {
			return check.NewSuccessInfo(name, fmt.Sprintf("Tunnel `%s` to cluster `%s` is open", tunnel, cluster))
		}
	}
}

package proxy

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
)

// Checker returns a posh check.Checker that dials each proxy's SSH host:port
// and reports reachable / unreachable.
func (c Config) Checker() check.Checker {
	return func(ctx context.Context, l log.Logger) []check.Info {
		infos := make([]check.Info, 0, len(c))

		for name, p := range c {
			port := p.SSHPort
			if port == 0 {
				port = 22
			}

			addr := fmt.Sprintf("%s:%d", p.SSHHost, port)
			dialCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			conn, err := (&net.Dialer{}).DialContext(dialCtx, "tcp", addr)

			cancel()

			if err != nil {
				infos = append(infos, check.NewFailureInfo("proxy:"+name, addr))
			} else {
				conn.Close()

				infos = append(infos, check.NewSuccessInfo("proxy:"+name, addr))
			}
		}

		return infos
	}
}

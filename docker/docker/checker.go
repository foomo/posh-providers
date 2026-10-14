package docker

import (
	"context"
	"net"
	"time"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
	"github.com/moby/moby/client"
)

// Socket is the path to the Docker daemon's Unix socket checked by
// SocketChecker.
var Socket = "/var/run/docker.sock"

// APIChecker reports whether the Docker daemon answers a Ping over the
// client's default connection.
func APIChecker(ctx context.Context, l log.Logger) []check.Info {
	title := "Docker API"

	cli, err := client.New()
	if err != nil {
		return []check.Info{check.NewNoteInfo("⚓︎", title, "Stopped")}
	}
	defer cli.Close()

	_, err = cli.Ping(ctx, client.PingOptions{})
	if err != nil {
		return []check.Info{check.NewNoteInfo("⚓︎", title, "Stopped")}
	}

	return []check.Info{check.NewSuccessInfo("⚓︎", title, "Running")}
}

// SocketChecker reports whether Socket can be dialed within 500ms.
func SocketChecker(ctx context.Context, l log.Logger) []check.Info {
	title := "Docker"

	d := &net.Dialer{Timeout: 500 * time.Millisecond}

	conn, err := d.DialContext(ctx, "unix", Socket)
	if err != nil {
		return []check.Info{check.NewNoteInfo("⚓︎", title, Socket)}
	}
	defer conn.Close()

	return []check.Info{check.NewSuccessInfo("⚓︎", title, Socket)}
}

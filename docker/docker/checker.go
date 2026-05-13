package docker

import (
	"context"
	"net"
	"time"

	"github.com/foomo/posh/pkg/log"
	"github.com/foomo/posh/pkg/prompt/check"
	"github.com/moby/moby/client"
)

var Socket = "/var/run/docker.sock"

func APIChecker(ctx context.Context, l log.Logger) []check.Info {
	title := "Docker"

	cli, err := client.New()
	if err != nil {
		return []check.Info{check.NewNoteInfo("⚓︎", title, "Not running (api)")}
	}
	defer cli.Close()

	_, err = cli.Ping(ctx, client.PingOptions{})
	if err != nil {
		return []check.Info{check.NewNoteInfo("⚓︎", title, "Not running")}
	}

	return []check.Info{check.NewSuccessInfo("⚓︎", title, "Running")}
}

func SocketChecker(ctx context.Context, l log.Logger) []check.Info {
	title := "Docker"

	d := &net.Dialer{Timeout: 500 * time.Millisecond}

	conn, err := d.DialContext(ctx, "unix", Socket)
	if err != nil {
		return []check.Info{check.NewNoteInfo("⚓︎", title, "Not running ("+Socket+")")}
	}
	defer conn.Close()

	return []check.Info{check.NewSuccessInfo("⚓︎", title, "Running")}
}

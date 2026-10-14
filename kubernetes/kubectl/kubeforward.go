package kubectl

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	gonet "github.com/foomo/go/net"
	"github.com/pterm/pterm"
)

// Kubeforward sets up a `kubectl port-forward` session to forward a local port to a target in a Kubernetes cluster.
// It returns a new context, the local address for the port forward, and a function to cancel the port-forward session.
func Kubeforward(ctx context.Context, cluster *Cluster, profile, target, namespace string, port int) (context.Context, string, context.CancelFunc) {
	ctx, cancel := context.WithCancelCause(ctx)

	freePort, err := gonet.FreePort(ctx)
	if err != nil {
		cancel(err)
	}

	portMapping := fmt.Sprintf("%d:%d", freePort, port)

	pf := exec.CommandContext(ctx, "kubectl",
		"--namespace", namespace,
		"port-forward", target, portMapping,
	)

	pf.Env = append(os.Environ(), cluster.Env(profile))

	pterm.Debug.Printfln("Starting kubectl port-forward %s.%s [%s]", namespace, target, portMapping)

	if err := pf.Start(); err != nil {
		cancel(err)
	}

	addr := fmt.Sprintf("127.0.0.1:%d", freePort)
	if err := gonet.WaitFor(ctx, "tcp", addr, 10*time.Second); err != nil {
		cancel(err)
	}

	return ctx, addr, func() {
		cancel(nil)
	}
}

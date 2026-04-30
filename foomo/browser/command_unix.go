//go:build !windows

package browser

import (
	"fmt"
	"os/exec"
	"syscall"
)

func newWatcherCmd(chromePID, sshPID int) *exec.Cmd {
	cmd := exec.Command("/bin/sh", "-c", fmt.Sprintf( //nolint:noctx
		"while kill -0 %d 2>/dev/null; do sleep 0.5; done; kill -9 %d 2>/dev/null",
		chromePID, sshPID,
	))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return cmd
}

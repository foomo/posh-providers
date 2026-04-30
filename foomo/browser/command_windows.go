//go:build windows

package browser

import (
	"fmt"
	"os/exec"
	"syscall"
)

func newWatcherCmd(chromePID, sshPID int) *exec.Cmd {
	cmd := exec.Command("powershell", "-Command", fmt.Sprintf( //nolint:noctx
		"while (Get-Process -Id %d -ErrorAction SilentlyContinue) { Start-Sleep -Milliseconds 500 }; Stop-Process -Id %d -Force -ErrorAction SilentlyContinue",
		chromePID, sshPID,
	))
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}

	return cmd
}

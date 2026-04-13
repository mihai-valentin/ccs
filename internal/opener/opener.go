package opener

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/mihai/ccs/internal/model"
)

// OpenSession opens a Claude Code session by resuming it in the session's
// original working directory. In same-terminal mode (newTerminal=false), the
// process replaces the current terminal. In new-terminal mode, the process is
// spawned detached with its own session ID (Setsid).
func OpenSession(session model.Session, newTerminal bool) error {
	if _, err := os.Stat(session.Cwd); os.IsNotExist(err) {
		return fmt.Errorf(
			"session directory %q no longer exists; use --cwd to open in the current directory instead",
			session.Cwd,
		)
	}

	shellCmd := fmt.Sprintf("cd %q && claude --resume %q", session.Cwd, session.ID)
	cmd := exec.Command("bash", "-c", shellCmd)

	if newTerminal {
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start detached session: %w", err)
		}
		// Detach — don't wait for the child process.
		return nil
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("session exited with error: %w", err)
	}
	return nil
}

package player

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mohsinkaleem/ytui-go/internal/types"
)

// Play launches mpv with the given URL and format, handing terminal control to mpv
func Play(mpvPath, url, formatID string) tea.Cmd {
	if mpvPath == "" {
		mpvPath = "mpv"
	}

	args := []string{url}
	if formatID != "" {
		args = append(args, "--ytdl-format="+formatID)
	}

	cmd := exec.Command(mpvPath, args...)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return types.MPVExitedMsg{Err: err}
	})
}

// IsAvailable checks if mpv binary is available
func IsAvailable(mpvPath string) bool {
	if mpvPath == "" {
		mpvPath = "mpv"
	}
	_, err := exec.LookPath(mpvPath)
	return err == nil
}

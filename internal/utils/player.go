package utils

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mohsinkaleem/ytui-go/internal/types"
)

// PlayerManager wraps mpv process
type PlayerManager struct {
	mpvPath string
}

// NewPlayerManager creates a player manager
func NewPlayerManager(mpvPath string) *PlayerManager {
	if mpvPath == "" {
		mpvPath = "mpv"
	}
	return &PlayerManager{mpvPath: mpvPath}
}

// Play returns a tea.Cmd that hands terminal to mpv
func (pm *PlayerManager) Play(url, formatID string) tea.Cmd {
	args := []string{url}
	if formatID != "" {
		args = append(args, "--ytdl-format="+formatID)
	}

	cmd := exec.Command(pm.mpvPath, args...)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return types.MPVExitedMsg{Err: err}
	})
}

// IsAvailable checks if mpv is in PATH
func (pm *PlayerManager) IsAvailable() bool {
	_, err := exec.LookPath(pm.mpvPath)
	return err == nil
}

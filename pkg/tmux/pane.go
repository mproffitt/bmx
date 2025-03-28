package tmux

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// CreatePane creates a new pane by splitting the current
// active pane
//
// This is simply a wrapper to SplitWindow
func CreatePane(target, startPath, startCommand string, respawn bool) error {
	if !respawn {
		return SplitWindow(target, startPath, startCommand, false)
	}

	args := []string{
		"respawn-pane", "-k", "-t", target,
	}
	if startPath != "" {
		args = append(args, "-c", startPath)
	}
	if startCommand != "" {
		args = append(args, startCommand)
	}
	return ExecSilent(args)
}

// KillPane kills the target pane
func KillPane(target string) (err error) {
	err = ExecSilent([]string{
		"kill-pane", "-t", target,
	})
	return
}

// HasPane checks if the target pane index exists in the target window
func HasPane(target string, pane uint) bool {
	out, _, _ := Exec([]string{
		"list-panes", "-t", target, "-F", "#{pane_index}",
	})
	panes := strings.Split(out, "\n")
	return slices.Contains(panes, fmt.Sprintf("%d", pane))
}

// GetPaneIndex gets a pane index for a given paneId
func GetPaneIndex(id uint) uint {
	out, _, _ := Exec([]string{
		"display-message", "-t", fmt.Sprintf("%%%d", id),
		"-p", "-F", "#{pane_index}",
	})
	uid, _ := strconv.ParseUint(out, 10, 64)
	return uint(uid)
}

// Gets the PID of the target pane
func GetPanePid(target string) int32 {
	out, _, err := Exec([]string{
		"display-message", "-t", target, "-p", "#{pane_pid}",
	})
	if err != nil {
		return -1
	}
	pid, _ := strconv.Atoi(out)
	return int32(pid)
}

// Get all panes across all sessions
func ListAllPanes() []string {
	stdout, _, err := Exec([]string{
		"list-panes", "-a", "-F", "#S:#I.#P",
	})
	if err != nil {
		return []string{}
	}

	return strings.Split(stdout, "\n")
}

// Maximize pane makes this the largest it can be given a current window layout
func MazimizeCurrentPane(target string) {
	_ = ExecSilent([]string{
		"resize-pane", "-t", target, "-U", "999",
	})
}

// Gets the current command for a given pane
func PaneCurrentCommand(sessionPane string) string {
	out, _, _ := Exec([]string{
		"display", "-p", "-t", sessionPane, "#{pane_current_command}",
	})

	return out
}

// Sets the title for the pane
func SetPaneTitle(paneId *uint, name string) error {
	pane := fmt.Sprintf("%%%d", *paneId)
	return ExecSilent([]string{
		"select-pane", "-t", pane, "-T", name,
	})
}

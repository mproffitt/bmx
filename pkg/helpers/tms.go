package helpers

import (
	"os"
	"os/exec"
	"path/filepath"
)

func ExecutableName() string {
	return filepath.Base(os.Args[0])
}

func ExecString() string {
	executable := ExecutableName()
	cwd, _ := os.Getwd()
	path, err := exec.LookPath(executable)
	// Popups should be executed in the same directory as the
	// user location. We do this via `cd` to `tmux run`
	if err != nil {
		return "cd " + cwd + "; " + filepath.Join(cwd, executable)
	}
	return "cd " + cwd + "; " + path
}

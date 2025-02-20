package tmux

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	bmx "github.com/mproffitt/bmx/pkg/exec"
)

func Exec(args []string) (string, string, error) {
	tmux, err := exec.LookPath("tmux")
	if err != nil {
		return "", "", errors.ErrUnsupported
	}

	return bmx.Exec(tmux, args)
}

func ExecSilent(args []string) error {
	_, _, err := Exec(args)
	return err
}

func GetTmuxEnvVar(session, name string) string {
	args := []string{
		"show-environment", "-t", session, name,
	}
	out, _, err := Exec(args)
	out = strings.TrimSpace(out)
	if err != nil || out == "unknown variable: "+name {
		return ""
	}

	// account for variables that include `=` signs
	// such as those representing commands or variables
	// that need to be re-exported from scripts
	return strings.Join(strings.Split(out, "=")[1:], "=")
}

func CapturePane(targetPane string) (string, error) {
	args := []string{
		"capture-pane", "-ep", "-t", targetPane,
	}
	output, _, err := Exec(args)
	if err != nil {
		return "", err
	}
	return output, nil
}

func DisplayPopup(w, h, t, b string, args []string) error {
	command := []string{
		"display-popup",
		"-h", h, "-w", w,
		"-S", "fg=" + b,
		"-b", "rounded",
		"-T", t,
		"-E", "bash", "-c",
	}

	cmd := strings.Join(args, " ")
	command = append(command, cmd)
	err := ExecSilent(command)
	return err
}

func DisplayMenu(title, border, fg, bg string, args [][]string) error {
	command := []string{
		"display-menu",
		"-b", "rounded",
		"-S", fmt.Sprintf("fg=%s", border),
		"-T", title,
		"-x", "C",
		"-y", "C",
	}

	for _, c := range args {
		command = append(command, c...)
	}
	err := ExecSilent(command)
	return err
}

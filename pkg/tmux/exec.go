// Copyright (c) 2025 Martin Proffitt <mprooffitt@choclab.net>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package tmux

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	bmx "github.com/mproffitt/bmx/pkg/exec"
	"github.com/muesli/reflow/truncate"
)

// Exec wraps the exec command with `tmux` as the
// command name. This means you can focus solely on
// the tmux commands
func Exec(args []string) (string, string, error) {
	tmux, err := exec.LookPath("tmux")
	if err != nil {
		return "", "", errors.ErrUnsupported
	}

	tmuxArgs := []string{
		"-S", GetSocketPath(),
	}
	tmuxArgs = append(tmuxArgs, args...)

	return bmx.Exec(tmux, tmuxArgs)
}

// ExecSilent supresses Standard out and Standard error
// but returns any error from  the command with
// stdout, stderr being available as fields in the error
// response where available
func ExecSilent(args []string) error {
	_, _, err := Exec(args)
	return err
}

// Captures the given pane.
//
// If the value of truncateWidth is not equal to 0,
// this method will attempt to truncate each line to the given width
// preserving ansi escape sequences where present
func CapturePane(target string, truncateWidth int) (string, error) {
	args := []string{
		"capture-pane", "-ep", "-t", target,
	}
	output, _, err := Exec(args)
	if err != nil {
		return "", err
	}

	if truncateWidth > 0 {
		builder := strings.Builder{}
		for _, line := range strings.Split(output, "\n") {
			line = truncate.String(line, uint(truncateWidth))
			builder.WriteString(line + "\n")
		}
		output = builder.String()
	}
	return output, nil
}

// Run the given command in a popup window
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

// Display a menu
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

// Get the base index
func GetBaseIndex() uint {
	out, _, err := Exec([]string{
		"show", "-gv", "base-index",
	})
	if err != nil {
		return 0
	}
	index, err := strconv.ParseUint(out, 10, 8)
	if err != nil {
		return 0
	}
	return uint(index)
}

// Get the current socket path
func GetSocketPath() string {
	tmux, err := exec.LookPath("tmux")
	if err != nil {
		return ""
	}

	args := []string{
		"display-message", "-p", "#{socket_path}",
	}

	path, _, _ := bmx.Exec(tmux, args)

	return path
}

// Get an environment variable from the TMUX env
func GetTmuxEnvVar(target, name string) string {
	args := []string{
		"show-environment", "-t", target, name,
	}
	out, _, err := Exec(args)
	if err != nil || out == "unknown variable: "+name {
		return ""
	}

	// account for variables that include `=` signs
	// such as those representing commands or variables
	// that need to be re-exported from scripts
	return strings.Join(strings.Split(out, "=")[1:], "=")
}

func IsRunning() bool {
	return GetSocketPath() != ""
}

func SendKeys(target string, keysToSend string) {
	if keysToSend == "" {
		return
	}
	_ = ExecSilent([]string{
		"send-keys", "-t", target,
		keysToSend, "C-m",
	})
}

func SwitchClient(target string) error {
	return ExecSilent([]string{
		"switch-client", "-t", target,
	})
}

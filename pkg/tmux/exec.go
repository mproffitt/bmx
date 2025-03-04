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

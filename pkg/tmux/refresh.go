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
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/mproffitt/bmx/pkg/helpers"
)

var EnvironmentVars = []string{}

var CommonShells = []string{
	"sh", "csh", "ksh", "zsh", "bash", "dash", "fish", "tcsh",
}

// Refresh the TMUX environment
//
// This causes tmux to reload all its configs, then optionally
// sends the KUBECONFIG environment variable,
//
// If sendVars is true, it will also trigger writing the export
// command to all shells
func Refresh(includeKubeconfig, sendVars bool) error {
	args := []string{
		"display-message", "-p", "#{config_files}",
	}
	stdout, _, err := Exec(args)
	if err != nil {
		return fmt.Errorf("failed to load list of config files %w", err)
	}

	for _, file := range strings.Split(stdout, ",") {
		args = []string{
			"source-file", file,
		}
		err := ExecSilent(args)
		if err != nil {
			return fmt.Errorf("failed to source file %q %w", file, err)
		}
	}
	return nil
}

func SetSessionEnvironment(session, variable, value string) error {
	args := []string{
		"set-environment", "-t", session, variable, value,
	}
	_, e, err := Exec(args)
	if err != nil {
		return fmt.Errorf("failed to set %q environment variable for session %q %q %w", variable, session, e, err)
	}
	return nil
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

func PaneCurrentCommand(sessionPane string) string {
	out, _, _ := Exec([]string{
		"display", "-p", "-t", sessionPane, "#{pane_current_command}",
	})

	return out
}

// Send tmux environment vars to all running panes
//
// This function uses the send-keys functionality to attempt
// to suspend any current job, write the given environment
// variables and then resum the jobs.
//
// All error messages from exec are ignored and the behaviour
// of this command may be unpredictable. Use with caution
func SendVars(varsToSend []string) {
	for _, sessionPane := range ListAllPanes() {
		out := PaneCurrentCommand(sessionPane)

		skipSuspend := false
		out = filepath.Base(out)
		if out != "" && (out == helpers.ExecutableName() || slices.Contains(CommonShells, out)) {
			skipSuspend = true
		}
		if !skipSuspend {
			_ = ExecSilent([]string{
				"send-keys", "-t", sessionPane, "C-z", "C-m",
			})
		}
		for _, v := range varsToSend {
			_ = ExecSilent([]string{
				"send-keys", "-t", sessionPane,
				fmt.Sprintf("export $(tmux show-env %s)", v), "C-m",
			})
		}
		if !skipSuspend {
			_ = ExecSilent([]string{
				"send-keys", "-t", sessionPane, "fg", "C-m",
			})
		}
		fmt.Fprintf(os.Stdout, "Refreshed pane '%q\n", sessionPane)
	}
}

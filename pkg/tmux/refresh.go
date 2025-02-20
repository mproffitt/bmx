package tmux

import (
	"fmt"
	"os"
	"strings"

	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/kubernetes"
)

var EnvironmentVars = []string{}

var CommonShells = []string{
	"sh", "csh", "ksh", "zsh", "bash", "dash", "fish", "tcsh",
}

func findShell(shell string) bool {
	for _, v := range CommonShells {
		if shell == v {
			return true
		}
	}
	return false
}

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
			"source-file", strings.TrimSpace(file),
		}
		err := ExecSilent(args)
		if err != nil {
			return fmt.Errorf("failed to source file %q %w", file, err)
		}
	}

	if includeKubeconfig {
		err = UpdateSessionEnvironment()
		if err != nil {
			return err
		}
		EnvironmentVars = append(EnvironmentVars, "KUBECONFIG")
	}
	if sendVars {
		SendVars(EnvironmentVars)
	}

	return nil
}

func UpdateSessionEnvironment() error {
	for _, session := range ListSessions() {
		configFile, err := kubernetes.CreateConfig(session.Name)
		if err != nil {
			return fmt.Errorf("failed to create kubeconfig for session %q %w", session.Name, err)
		}
		args := []string{
			"set-environment", "-t", session.Name, "KUBECONFIG", configFile,
		}
		var e string
		_, e, err = Exec(args)
		if err != nil {
			return fmt.Errorf("failed to set KUBECONFIG environment variable for session %q %q %w", session.Name, e, err)
		}
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

	stdout = strings.TrimSpace(stdout)
	return strings.Split(stdout, "\n")
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
		out, _, _ := Exec([]string{
			"display", "-p", "-t", sessionPane, "#{pane_current_command}",
		})
		skipSuspend := false
		if out != "" && (out == helpers.ExecutableName() || findShell(out)) {
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

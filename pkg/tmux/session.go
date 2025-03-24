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
	"strings"

	"github.com/mproffitt/bmx/pkg/kubernetes"
)

// AttachSession attaches to the given named session
func AttachSession(name string) error {
	args := []string{
		"switch-client", "-t", name,
	}
	_, _, err := Exec(args)
	if err != nil {
		args := []string{
			"attach-session", "-t", name,
		}
		_, _, err := Exec(args)
		if err != nil {
			return err
		}
	}
	return nil
}

// Create a session with the given name, path and optionally command.
func CreateSession(name, path, command string, includeKubeConfig, attach bool) error {
	args := []string{
		"new-session", "-d",
		"-s", name, "-c", path,
	}

	if includeKubeConfig {
		if config, err := kubernetes.CreateConfig(name); err == nil {
			kubeConfig := fmt.Sprintf("KUBECONFIG=%s", config)
			args = append(args, "-e", kubeConfig)
		}
	}

	if command != "" {
		envVar := fmt.Sprintf("COMMAND=%q", command)
		args = append(args, "-e", envVar, command)
	}

	_, _, err := Exec(args)
	if err != nil {
		return err
	}
	if attach {
		return AttachSession(name)
	}
	return nil
}

// Get the name of the current session
func CurrentSession() string {
	name, _, err := Exec([]string{
		"display-message", "-p", "#{session_name}",
	})
	if err == nil {
		return ""
	}
	return name
}

// If this server has a given session by name
func HasSession(name string) bool {
	if err := ExecSilent([]string{"has-session", "-t", name}); err != nil {
		return false
	}
	return true
}

// Kill the given session
//
// Note: This kills the session  by name without checking
// if it is first attached or is current or not.
//
// This results in the behaviour that TMUX will fall back to
// the previous session in the list.
//
// For more controlled behaviour, create or attach to a different
// session before killing the old one.
func KillSession(sessionName string) error {
	args := []string{
		"kill-session", "-t", sessionName,
	}
	_, _, err := Exec(args)
	return err
}

// ListSessions returns a list of sessions on the
// current tmux server
func ListSessions() []string {
	sessions, _, err := Exec([]string{
		"list-sessions", "-F",
		"#{session_name},#{session_windows},#{session_created},#{session_attached},#{session_group},#{session_path}",
	})
	if err != nil {
		return []string{}
	}
	return strings.Split(sessions, "\n")
}

// Creates a new session and attaches to it.
// If the session already exists, it is simply attached.
func NewSessionOrAttach(in map[string]any, includeKubeConfig bool) error {
	var (
		name, owner, path string
		command           string
		ok                bool
	)
	if _, ok = in["name"]; !ok {
		return fmt.Errorf("cannot create session. empty name")
	}
	name = in["name"].(string)

	if _, ok = in["owner"]; ok {
		owner = in["owner"].(string)
	}

	if _, ok = in["path"]; ok {
		path = in["path"].(string)
		if path == "" {
			path, _ = os.UserHomeDir()
		}
	}

	if _, ok = in["command"]; ok {
		command = in["command"].(string)
	}

	if HasSession(name) {
		if SessionPath(name) == path {
			return AttachSession(name)
		}
		// prevent sessions being named '-session'
		if owner != "" {
			sessionName := strings.Join([]string{owner, name}, "-")
			if HasSession(sessionName) && SessionPath(sessionName) == path {
				return AttachSession(sessionName)
			}
			return CreateSession(sessionName, path, command, includeKubeConfig, true)
		}
		// It should be almost impossible to get to this point
		// but it --is-- possible
		return fmt.Errorf("duplicate session without an owner %q", name)
	}
	return CreateSession(name, path, command, includeKubeConfig, true)
}

// Rename a tmux session
func RenameSession(target, name string) error {
	return ExecSilent([]string{
		"rename-session", "-t", target, name,
	})
}

// List all panes in a given session
func SessionPanes(session string) ([]string, error) {
	args := []string{
		"list-panes", "-t", session, "-F", "#{pane_id}",
	}
	output, _, err := Exec(args)
	if err != nil {
		return []string{}, err
	}
	return strings.Split(output, "\n"), nil
}

// Get the path for a given session
//
// This calls tmux display-message #{session_path} and if that returns
// empty, returns the user home directory instead
func SessionPath(name string) string {
	path, _, err := Exec([]string{
		"display-message", "-t",
		name, "-p", "#{session_path}",
	})
	if err != nil {
		path, _ = os.UserHomeDir()
	}
	return path
}

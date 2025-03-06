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
	"sort"
	"strings"

	"github.com/mproffitt/bmx/pkg/kubernetes"
)

type (
	GetBy  int
	SortBy int
)

const (
	Name SortBy = iota
	NameReverse
	Oldest
	Newest
)

const (
	First GetBy = iota
	Last
)

// Attach to the given named session
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

// Get the name of the current session
func CurrentSession() string {
	name, _, err := Exec([]string{
		"display-message", "-p", "#{session_name}",
	})
	if err == nil {
		return ""
	}
	return strings.TrimSpace(name)
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

// If a given session name exists
func HasSession(name string) bool {
	for _, session := range ListSessions() {
		if session.Name == name {
			return true
		}
	}
	return false
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

// Kills the named session and switches to the alternative
//
// If `new` doesn't exist, it switches to the oldest session
func KillSwitch(old, new string) error {
	sessions := SortedSessionlist(Oldest)
	var oldest, realnew string
	{
		for _, session := range sessions {
			if session.Name != old {
				oldest = session.Name
			}
			if session.Name == new {
				realnew = new
			}
		}
		if realnew == "" {
			realnew = oldest
		}
	}
	err := AttachSession(realnew)
	if err != nil {
		return err
	}
	return KillSession(old)
}

// List all sessions
//
// This lists all tmux sessions and returns them in the order
// returned by tmux
//
// If you want a different ordering, you should use
// SortedSessionlist instead
func ListSessions() []Session {
	sessions, _, err := Exec([]string{
		"list-sessions",
	})
	if err != nil {
		return []Session{}
	}

	details := make([]Session, 0)
	for _, session := range strings.Split(sessions, "\n") {
		if strings.TrimSpace(session) == "" {
			continue
		}
		d := NewSessionFromString(session)
		details = append(details, d)
	}
	return details
}

// Creates a new session and attaches to it.
// If the session already exists, it is simply attached.
func NewSessionOrAttach(in map[string]any, filter string, includeKubeConfig bool) error {
	var (
		repo, owner, path string
		command           string
		ok                bool
	)
	if repo, ok = in["name"].(string); !ok {
		if filter == "" {
			return fmt.Errorf("no repo specified")
		}
		repo = filter
	}
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

	if HasSession(repo) {
		if SessionPath(repo) == path {
			return AttachSession(repo)
		}
		sessionName := strings.Join([]string{owner, repo}, "-")
		if HasSession(sessionName) {
			if SessionPath(sessionName) == path {
				return AttachSession(sessionName)
			}
		}
		return CreateSession(sessionName, path, command, includeKubeConfig, true)
	}
	return CreateSession(repo, path, command, includeKubeConfig, true)
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
	return strings.TrimSpace(path)
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

func SessionWindows(session string) ([]string, error) {
	args := []string{
		"list-windows", "-t", session, "-F", "#S:#I",
	}
	out, _, err := Exec(args)
	if err != nil {
		return []string{}, err
	}

	return strings.Split(out, "\n"), nil
}

// List all tmux sessions and sort them by the order provided
func SortedSessionlist(by SortBy) []Session {
	sessions := ListSessions()
	sort.SliceStable(sessions, func(i, j int) bool {
		switch by {
		case Name: // default behaviour
		case NameReverse:
			return sessions[j].Name < sessions[i].Name
		case Newest:
			return sessions[i].Created.Unix() < sessions[j].Created.Unix()
		case Oldest:
			return sessions[j].Created.Unix() < sessions[i].Created.Unix()
		}
		return sessions[i].Name < sessions[j].Name
	})
	return sessions
}

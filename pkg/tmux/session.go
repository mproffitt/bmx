package tmux

import (
	"fmt"
	"os"
	"strings"

	"github.com/mproffitt/bmx/pkg/kubernetes"
)

func SessionPanes(session string) ([]string, error) {
	args := []string{
		"lsp", "-t", session, "-F", "#{pane_id}",
	}
	output, _, err := Exec(args)
	if err != nil {
		return []string{}, err
	}
	panes := strings.Split(strings.TrimSpace(output), "\n")
	return panes, nil
}

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

func CurrentSession() string {
	name, _, err := Exec([]string{
		"display-message", "-p", "#{session_name}",
	})
	if err == nil {
		return ""
	}
	return strings.TrimSpace(name)
}

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

func HasSession(name string) bool {
	for _, session := range ListSessions() {
		if session.Name == name {
			return true
		}
	}
	return false
}

func KillSession(sessionName string) error {
	args := []string{
		"kill-session", "-t", sessionName,
	}
	_, _, err := Exec(args)
	return err
}

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

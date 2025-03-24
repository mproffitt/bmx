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

package manager

import (
	"fmt"
	"sort"
	"sync"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/kubernetes"
	"github.com/mproffitt/bmx/pkg/tmux"
	"github.com/mproffitt/bmx/pkg/tmux/ui/session"
)

type ManagerReadyMsg struct {
	Ready bool
}

func ManagerReadyCmd(ready bool) tea.Cmd {
	return func() tea.Msg {
		return ManagerReadyMsg{Ready: ready}
	}
}

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

type Model struct {
	sync.Mutex
	sessions []*session.Session
	colours  *config.ColourStyles
	Ready    bool
}

type Iterator func(yield func(int, *session.Session) bool)

// Creates a new Session Manager
func New() (*Model, Iterator) {
	m := Model{}

	return &m, func(yield func(key int, val *session.Session) bool) {
		func(yield func(key int, val *session.Session) bool) bool {
			for k, v := range m.sessions {
				if !yield(k, v) {
					return false
				}
			}
			return false
		}(yield)
	}
}

func (m *Model) Init() tea.Cmd {
	return m.load()
}

// Get the session with the given name
func (m *Model) Session(name string) *session.Session {
	for _, session := range m.sessions {
		if session.Name == name {
			return session
		}
	}

	return nil
}

// Has returns true if the session exists
func (m *Model) Has(name string) bool {
	// If a given session name exists
	for _, session := range m.sessions {
		if session.Name == name {
			return true
		}
	}
	return false
}

// Get the session items
func (m *Model) Items() []*session.Session {
	sort.SliceStable(m.sessions, func(i, j int) bool {
		if m.sessions[i].Attached != m.sessions[j].Attached {
			return m.sessions[i].Attached
		}
		return m.sessions[i].Name < m.sessions[j].Name
	})
	return m.sessions
}

// Kills the named session and switches to the alternative
//
// If `new` doesn't exist, it switches to the oldest session
func (m *Model) KillSwitch(old, new string) error {
	current := tmux.CurrentSession()
	sessions := m.Sort(Oldest)
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

	var err error
	// Only switch session if we're deleting current
	if realnew != current && current == old {
		err = tmux.AttachSession(realnew)
		if err != nil {
			return err
		}
	}
	err = tmux.KillSession(old)
	m.load()
	return err
}

func (m *Model) Len() int {
	return len(m.sessions)
}

// Refresh sessions
func (m *Model) Refresh(includeKubeconfig, sendVars bool) error {
	envvars := make([]string, 0)
	if includeKubeconfig {
		err := m.UpdateEnvironment()
		if err != nil {
			return err
		}
		envvars = append(envvars, "KUBECONFIG")
	}

	err := tmux.Refresh(includeKubeconfig)
	if err != nil {
		return err
	}

	if sendVars {
		tmux.SendVars(envvars)
	}

	return nil
}

// List all tmux sessions and sort them by the order provided
func (m *Model) Sort(by SortBy) []*session.Session {
	sort.SliceStable(m.sessions, func(i, j int) bool {
		switch by {
		case Name: // default behaviour
		case NameReverse:
			return m.sessions[j].Name < m.sessions[i].Name
		case Newest:
			return m.sessions[i].Created.Unix() < m.sessions[j].Created.Unix()
		case Oldest:
			return m.sessions[j].Created.Unix() < m.sessions[i].Created.Unix()
		}
		return m.sessions[i].Name < m.sessions[j].Name
	})
	return m.sessions
}

// Send an update to TMUX for the KUBECONFIG session name
func (m *Model) UpdateEnvironment() error {
	for _, session := range m.sessions {
		configFile, err := kubernetes.CreateConfig(session.Name)
		if err != nil {
			return fmt.Errorf("failed to create kubeconfig for session %q %w", session.Name, err)
		}
		err = tmux.SetSessionEnvironment(session.Name, "KUBECONFIG", configFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Model) WithColours(c *config.ColourStyles) *Model {
	m.colours = c
	return m
}

func (m *Model) Reload() tea.Cmd {
	return m.load()
}

func (m *Model) load() tea.Cmd {
	m.sessions = make([]*session.Session, 0)
	var wg sync.WaitGroup
	for _, s := range tmux.ListSessions() {
		wg.Add(1)
		go func() {
			defer wg.Done()
			session := session.New(s, m.colours)
			m.Lock()
			m.sessions = append(m.sessions, session)
			m.Unlock()
		}()
	}
	wg.Wait()
	m.Ready = true
	return ManagerReadyCmd(m.Ready)
}

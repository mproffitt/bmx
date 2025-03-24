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

package session

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/tmux"
	"github.com/mproffitt/bmx/pkg/tmux/ui/window"
)

type Renamable interface {
	GetName() string
	Rename(newname string) error
}

type Session struct {
	Attached   bool
	Created    time.Time
	Index      uint
	Group      string // Future
	Name       string
	NumWindows int
	Path       string
	Windows    []*window.Window

	colours *config.ColourStyles
	command string
}

// Load a session details and return a new session object
func New(session string, c *config.ColourStyles) *Session {
	parts := strings.Split(session, ",")

	details := Session{
		Attached: parts[3] != "0",
		Name:     parts[0],
		Group:    parts[4],
		Path:     parts[5],
		colours:  c,
	}
	count, _ := strconv.Atoi(parts[1])
	details.NumWindows = count

	t, err := strconv.ParseInt(parts[2], 10, 64)
	{
		if err != nil {
			t = time.Now().Unix()
		}
		details.Created = time.Unix(t, 0)
	}
	details.Windows = window.ListWindows(details.Name)
	return &details
}

// Attach to the current session
func (s *Session) Attach() error {
	return tmux.AttachSession(s.Name)
}

// Get the description of this session
func (s *Session) Description() string {
	date := s.Created.Format(time.ANSIC)
	if s.Attached {
		return fmt.Sprintf("active\n%s", date)
	}
	return date
}

// Filter value for filterable lists
func (s *Session) FilterValue() string { return s.Name }

// Gets the name of the current session
func (s *Session) GetName() string { return s.Name }

// Kill the window with the given index
func (s *Session) KillWindow(index uint64) error {
	target := fmt.Sprintf("%s:%d", s.Name, index)
	return tmux.KillWindow(target)
}

// Rename session
func (s *Session) Rename(name string) error {
	return tmux.RenameSession(s.Name, name)
}

// Marshal an individual session
func (s *Session) ToHelperStruct() helpers.Session {
	session := helpers.Session{
		Name:    s.Name,
		Command: s.command,
		Path:    s.Path,
		Windows: make([]helpers.Window, 0),
	}
	for _, window := range s.Windows {
		session.Windows = append(session.Windows, window.ToHelperStruct())
	}
	return session
}

// Get the session title
func (s *Session) Title() string {
	return s.Name
}

func (s *Session) Window(index uint64) *window.Window {
	for i, w := range s.Windows {
		if w.Index == index {
			return s.Windows[i]
		}
	}
	return nil
}

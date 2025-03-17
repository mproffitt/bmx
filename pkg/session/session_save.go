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
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/tmux/ui/session"
)

func (m *model) save() tea.Cmd {
	log.Debug("triggering save")
	s := make([]session.Session, 0)
	for _, v := range m.manager.Items() {
		s = append(s, *v)
	}
	sessions := unpackArray(s)
	err := m.config.SetSessions(sessions)
	if err != nil {
		return helpers.NewErrorCmd(err)
	}
	return nil
}

func unpackArray[S ~[]E, E any](s S) []any {
	r := make([]any, len(s))
	for i, e := range s {
		r[i] = e
	}
	return r
}

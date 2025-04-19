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
	"github.com/charmbracelet/bubbles/list"
	tmuxui "github.com/mproffitt/bmx/pkg/tmux/ui/window"
)

func (m *model) setItems() {
	switch m.active {
	case sessionManager:
		m.setSessionItems()
	case windowManager:
		m.setWindowsItems(m.session.Name)
	}
}

func (m *model) setSessionItems() {
	sessions := m.manager.Items()
	items := make([]list.Item, len(sessions))
	for i, s := range sessions {
		s.Index = uint(i)
		items[i] = s
	}
	if len(items) > 0 {
		m.ready = true
	}
	_ = m.list.SetItems(items)
	index := 0
	if m.session != nil {
		index = int(m.session.Index)
	}
	m.list.Select(index)
}

func (m *model) setWindowsItems(session string) {
	windows := tmuxui.ListWindows(session)
	items := make([]list.Item, len(windows))
	for i, w := range windows {
		items[i] = w
	}
	_ = m.list.SetItems(items)
	m.list.Select(0)
}

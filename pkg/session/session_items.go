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

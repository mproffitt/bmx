package session

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	CtrlN    key.Binding
	Delete   key.Binding
	Enter    key.Binding
	Help     key.Binding
	Quit     key.Binding
	ShiftTab key.Binding
	Tab      key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.CtrlN, k.Delete, k.Enter, k.Help,
		},
		{
			k.Quit, k.ShiftTab, k.Tab,
		},
	}
}

func mapKeys() keyMap {
	return keyMap{
		CtrlN: key.NewBinding(key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "Create new session")),
		Delete: key.NewBinding(key.WithKeys("delete", "x"),
			key.WithHelp("del/x", "Delete current item")),
		Enter: key.NewBinding(key.WithKeys("enter"),
			key.WithHelp("↩", "Select current item")),
		Help: key.NewBinding(key.WithKeys("?", "f1"),
			key.WithHelp("?", "Help")),
		Quit: key.NewBinding(key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("esc", "Close overlays or Quit")),
		ShiftTab: key.NewBinding(key.WithKeys("shift+tab"),
			key.WithHelp("⇧ ↹", "Previous pane")),
		Tab: key.NewBinding(key.WithKeys("tab"),
			key.WithHelp("↹", "Next pane")),
	}
}

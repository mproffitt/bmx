package config

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
)

func keymap() *huh.KeyMap {
	h := huh.NewDefaultKeyMap()
	h.FilePicker = huh.FilePickerKeyMap{
		GoToTop:  key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "first"), key.WithDisabled()),
		GoToLast: key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "last"), key.WithDisabled()),
		PageUp:   key.NewBinding(key.WithKeys("K", "pgup"), key.WithHelp("pgup", "page up"), key.WithDisabled()),
		PageDown: key.NewBinding(key.WithKeys("J", "pgdown"), key.WithHelp("pgdown", "page down"), key.WithDisabled()),
		Back:     key.NewBinding(key.WithKeys("h", "left"), key.WithHelp("h", "back"), key.WithDisabled()),
		Select:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select"), key.WithDisabled()),
		Up:       key.NewBinding(key.WithKeys("up", "k", "ctrl+k", "ctrl+p"), key.WithHelp("↑", "up"), key.WithDisabled()),
		Down:     key.NewBinding(key.WithKeys("down", "j", "ctrl+j", "ctrl+n"), key.WithHelp("↓", "down"), key.WithDisabled()),

		Open:   key.NewBinding(key.WithKeys("l", "right", "enter"), key.WithHelp("right", "open")),
		Close:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close"), key.WithDisabled()),
		Prev:   key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "back")),
		Next:   key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next")),
		Submit: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
	}
	return h
}

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

package repos

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type keyMap struct {
	Down     key.Binding
	Enter    key.Binding
	Pageup   key.Binding
	Pagedown key.Binding
	Quit     key.Binding
	Up       key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	// Pageup and Pagedown currently do not work
	// with the panel pager. leaving them out for now
	return [][]key.Binding{
		{
			k.Enter, k.Quit,
		},
		{
			k.Up, k.Down, k.Pageup, k.Pagedown,
		},
	}
}

func mapKeys() keyMap {
	return keyMap{
		Down: key.NewBinding(key.WithKeys("down"),
			key.WithHelp("↓", "move down")),
		Enter: key.NewBinding(key.WithKeys("enter"),
			key.WithHelp("enter", "Set current context")),
		Pageup: key.NewBinding(key.WithKeys("pgup"),
			key.WithHelp("pgup", "previous page")),
		Pagedown: key.NewBinding(key.WithKeys("pgdown"),
			key.WithHelp("pgdn", "next page")),
		Up: key.NewBinding(key.WithKeys("up"),
			key.WithHelp("↑", "move up")),
	}
}

func (m *Model) Help() string {
	helpmsg := help.New()
	helpmsg.ShowAll = true
	help := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Style.FocusedColor)).Render("Context Pane")
	help = lipgloss.JoinVertical(lipgloss.Left, help, helpmsg.View(m.keymap))
	return help
}

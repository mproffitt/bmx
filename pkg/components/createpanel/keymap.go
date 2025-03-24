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

package createpanel

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/mproffitt/bmx/pkg/components/dialog"
	"github.com/mproffitt/bmx/pkg/components/icons"
)

type keyMap struct {
	Down     key.Binding
	Enter    key.Binding
	Left     key.Binding
	Right    key.Binding
	ShiftTab key.Binding
	Tab      key.Binding
	Up       key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter}
}

func (k keyMap) FullHelp() [][]key.Binding {
	// Pageup and Pagedown currently do not work
	// with the panel pager. leaving them out for now
	return [][]key.Binding{
		{
			k.Enter, k.Tab, k.ShiftTab,
		},
		{
			k.Up, k.Down, k.Left, k.Right,
		},
	}
}

func mapKeys() keyMap {
	return keyMap{
		Down: key.NewBinding(
			key.WithKeys("down"),
			key.WithHelp(icons.Down, "Move down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp(icons.Enter, "Set current context"),
		),
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp(icons.Left, "Previous character"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp(icons.Right, "Accept suggestion"),
		),
		ShiftTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp(icons.ShiftTab, "Previous field"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp(icons.Tab, "Next field"),
		),
		Up: key.NewBinding(
			key.WithKeys("up"),
			key.WithHelp(icons.Up, "Move up"),
		),
	}
}

func (m *Model) Help() dialog.HelpEntry {
	km := help.KeyMap(m.keymap)
	entry := dialog.HelpEntry{
		Keymap: &km,
		Title:  "New session",
	}
	return entry
}

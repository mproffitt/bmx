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
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/mproffitt/bmx/pkg/dialog"
)

type keyMap struct {
	CtrlN       key.Binding
	Delete      key.Binding
	Enter       key.Binding
	Help        key.Binding
	HideContext key.Binding
	Quit        key.Binding
	ShiftTab    key.Binding
	Tab         key.Binding
	ToggleZoom  key.Binding
}

func (k *keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k *keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			k.CtrlN, k.Delete, k.Enter, k.Help, k.HideContext,
		},
		{
			k.Quit, k.ShiftTab, k.Tab, k.ToggleZoom,
		},
	}
}

func mapKeys() *keyMap {
	return &keyMap{
		CtrlN: key.NewBinding(key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "Create new session")),
		Delete: key.NewBinding(key.WithKeys("delete", "x"),
			key.WithHelp("del/x", "Delete current item")),
		Enter: key.NewBinding(key.WithKeys("enter"),
			key.WithHelp("↩", "Select current item")),
		Help: key.NewBinding(key.WithKeys("?", "f1"),
			key.WithHelp("?", "Help")),

		HideContext: key.NewBinding(key.WithKeys("k"),
			key.WithHelp("k", "Hide context pane")),
		Quit: key.NewBinding(key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("esc", "Close overlays or Quit")),
		ShiftTab: key.NewBinding(key.WithKeys("shift+tab"),
			key.WithHelp("⇧ ↹", "Previous pane")),
		Tab: key.NewBinding(key.WithKeys("tab"),
			key.WithHelp("↹", "Next pane")),
		ToggleZoom: key.NewBinding(key.WithKeys("z"),
			key.WithHelp("z", "Toggle viewport zoom")),
	}
}

func (m *model) Help() dialog.HelpEntry {
	km := help.KeyMap(m.keymap)
	entry := dialog.HelpEntry{
		Keymap: &km,
		Title:  "Session manager",
	}
	return entry
}

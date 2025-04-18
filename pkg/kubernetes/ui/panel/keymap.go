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

package panel

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/mproffitt/bmx/pkg/components/dialog"
	"github.com/mproffitt/bmx/pkg/components/icons"
)

type keyMap struct {
	Delete    key.Binding
	Down      key.Binding
	Enter     key.Binding
	KillPanel key.Binding
	Left      key.Binding
	Move      key.Binding
	Pageup    key.Binding
	Pagedown  key.Binding
	Right     key.Binding
	Space     key.Binding
	ShiftDel  key.Binding
	Up        key.Binding
	Login     key.Binding
}

func (k *keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Delete, k.KillPanel}
}

func (k *keyMap) FullHelp() [][]key.Binding {
	// Pageup and Pagedown currently do not work
	// with the panel pager. leaving them out for now
	return [][]key.Binding{
		{
			k.Delete, k.Enter, k.KillPanel, k.Move, k.Space, k.Pageup,
		},
		{
			k.ShiftDel, k.Up, k.Down, k.Left, k.Right, k.Login, k.Pagedown,
		},
	}
}

func mapKeys() *keyMap {
	return &keyMap{
		Delete: key.NewBinding(key.WithKeys("delete", "x"),
			key.WithHelp("del/x", "Delete the current item")),
		Down: key.NewBinding(key.WithKeys("down", "j"),
			key.WithHelp(icons.Down, "move down")),
		Enter: key.NewBinding(key.WithKeys("enter"),
			key.WithHelp(icons.Enter, "Set current context")),
		KillPanel: key.NewBinding(key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("esc", "Close overlays or quit")),
		Left: key.NewBinding(key.WithKeys("left", "h"),
			key.WithHelp(icons.Left, "move left")),
		Login: key.NewBinding(key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "Login to cluster")),
		Move: key.NewBinding(key.WithKeys("m"),
			key.WithHelp("m", "Move context to session")),
		Pageup: key.NewBinding(key.WithKeys("pgup", "b"),
			key.WithHelp("b/pgup", "previous page")),
		Pagedown: key.NewBinding(key.WithKeys("pgdown"),
			key.WithHelp("b/pgdn", "next page")),
		Right: key.NewBinding(key.WithKeys("right", "l"),
			key.WithHelp(icons.Right, "move right")),
		Space: key.NewBinding(key.WithKeys(" "),
			key.WithHelp(icons.Space, "Change context namespace")),
		ShiftDel: key.NewBinding(key.WithKeys("X"),
			key.WithHelp("X", "force delete context")),
		Up: key.NewBinding(key.WithKeys("up", "k"),
			key.WithHelp(icons.Up, "move up")),
	}
}

func (m *Model) Help() dialog.HelpEntry {
	km := help.KeyMap(m.keymap)
	entry := dialog.HelpEntry{
		Keymap: &km,
		Title:  "Kubernetes Context Panel",
		Help: "The kubernetes context panel allows you to interact\nwith" +
			" contexts inside the kube-config for the current\nsession",
	}
	return entry
}

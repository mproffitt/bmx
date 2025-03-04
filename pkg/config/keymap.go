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

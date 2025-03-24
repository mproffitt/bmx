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

package table

import (
	"os/user"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mproffitt/bmx/pkg/components/createpanel"
	"github.com/mproffitt/bmx/pkg/components/dialog"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case createpanel.SuggestionsMsg:
		m.table, cmd = m.table.Update(msg.LastKey)
		cmds = append(cmds, cmd)

		// Get the suggestions from the table
		m.getSuggestions(&msg)
		msg.Focus = createpanel.Button

		// Send the message back to the panel
		m.panel, cmd = m.panel.Update(msg)
		cmds = append(cmds, cmd)

	case createpanel.ObserverMsg:
		// If the current focus was button
		// when we recieve this message then
		// we handle the creation of the new
		// session
		if msg.Focus == createpanel.Button {
			current := m.table.HighlightedRow().Data
			data := make(map[string]any)
			if map[string]any(current) != nil {
				for col := range current {
					data[col] = current[col]
				}
			}
			data["path"] = msg.Path
			data["command"] = msg.Command
			if msg.Name != "" {
				if data["name"] != msg.Name {
					user, _ := user.Current()
					data["owner"] = user.Username
				}
				data["name"] = msg.Name
			}

			// TODO: Convert callback to tea.Msg
			// This was written during the very first incarnation of the application
			// and as I've learned more about the way bubbletea operates it's become
			// a redundant methodology. Messages are a cleaner way of handling the
			// need to call back to parent handling methods
			return m, m.callback(data, m.config.CreateSessionKubeConfig)
		}

		var name, path string
		{
			if data, ok := m.table.HighlightedRow().Data[columnKeyName]; ok {
				name = data.(string)
			}
			if data, ok := m.table.HighlightedRow().Data[columnKeyPath]; ok {
				path = data.(string)
			}

		}
		if msg.Focus == createpanel.Name {
			if msg.Name != name {
				m.table = m.table.WithFilterInputValue(msg.Name)
			}
			msg.Path = path
		}

		// If last key on the panel was enter but
		// we didn't have the button in focus then
		// we allow for focusing the button so the
		// next enter keypress creates the session
		if msg.LastKey.String() == "enter" {
			msg.Focus = createpanel.Button
			msg.Name = name
			msg.Path = path
		}

		m.panel, cmd = m.panel.Update(msg)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:

		// Dialog needs to be handled first as it's an overlay
		// to the main window and takes precedence over all
		// other elements
		if m.dialog != nil {
			m.dialog, cmd = m.dialog.Update(msg)
			return m, cmd
		}

		switch {
		case key.Matches(msg, m.keymap.Quit):
			if m.dialog != nil {
				break
			}
			if m.isOverlay {
				return m, nil
			}
			cmds = append(cmds, tea.Quit)
		case key.Matches(msg, m.keymap.Help):
			m.displayHelp()
		case key.Matches(msg, m.keymap.Pagedown, m.keymap.Pageup):
			m.table, _ = m.table.Update(msg)
		default:
			m.panel, cmd = m.panel.Update(msg)
			cmds = append(cmds, cmd)
		}
	case dialog.DialogStatusMsg:
		if msg.Done {
			m.dialog = nil
		}
	case spinner.TickMsg:
		if m.spinner != nil {
			*m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
			m.drawTable()
		}
	case tea.WindowSizeMsg:
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
		m.width = msg.Width
		m.height = msg.Height
	default:
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

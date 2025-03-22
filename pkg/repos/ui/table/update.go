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
	"errors"
	"os"
	"os/user"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mproffitt/bmx/pkg/components/dialog"
	"github.com/mproffitt/bmx/pkg/exec"
	"k8s.io/utils/strings/slices"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
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
		case key.Matches(msg, m.keymap.Enter):
			if m.focus == Button {
				current := m.table.HighlightedRow().Data
				data := make(map[string]any)
				if map[string]any(current) != nil {
					for col := range current {
						data[col] = current[col]
					}
				}
				data["path"] = m.inputs.path.Value()
				data["command"] = m.inputs.command.Value()
				if filter := m.inputs.filter.Value(); filter != "" {
					if data["name"] != filter {
						user, _ := user.Current()
						data["owner"] = user.Username
					}
					data["name"] = filter
				}
				return m, m.callback(data, m.config.CreateSessionKubeConfig)
			}
		case key.Matches(msg, m.keymap.Help):
			m.displayHelp()
		case key.Matches(msg, m.keymap.ShiftTab):
			if m.current != nil {
				m.current.Blur()
			}
			switch m.focus {
			case Filter:
				m.focus = Button
				m.current = nil
			case Button:
				m.focus = Command
				m.current = &m.inputs.command
			case Path:
				m.focus = Filter
				m.current = &m.inputs.filter
			case Command:
				m.focus = Path
				m.current = &m.inputs.path
			}
			if m.current != nil {
				m.current.Focus()
			}
		case key.Matches(msg, m.keymap.Tab):
			if m.current != nil {
				m.current.Blur()
			}
			switch m.focus {
			case Filter:
				m.focus = Path
				m.current = &m.inputs.path
			case Path:
				m.focus = Command
				m.current = &m.inputs.command
			case Command:
				m.current = nil
				m.focus = Button
			case Button:
				m.focus = Filter
				m.current = &m.inputs.filter
			}
			if m.current != nil {
				m.current.Focus()
			}
		case key.Matches(msg, m.keymap.Pagedown, m.keymap.Pageup):
			fallthrough
		case key.Matches(msg, m.keymap.Up, m.keymap.Down):
			if m.focus == Filter {
				m.current.Blur()
				m.table, cmd = m.table.Update(msg)
				cmds = append(cmds, cmd)
				m.setValueFromTableRow()
				m.setSuggestions()
				m.current.Focus()
				break
			}
			// if not filter, fallthrough
			fallthrough
		default:
			*m.current, _ = m.current.Update(msg)
			currentInputValue := m.current.Value()
			var currentRowValue string
			{
				if data, ok := m.table.HighlightedRow().Data[columnKeyName]; ok {
					currentRowValue = data.(string)
				}
			}

			if m.focus == Filter && currentInputValue != currentRowValue {
				value := m.inputs.filter.Value()
				m.table = m.table.WithFilterInputValue(value)
			}

			var (
				options []exec.Completion
				err     error
			)
			switch m.focus {
			case Command:
				// For commands we allow triggering completion from
				// space, hyphen or slash. This should allow for
				// collection of sub-commands, options and paths
				// for command line completion
				allowed := []string{" ", "-", "/"}
				if slices.Contains(allowed, msg.String()) {
					options, err = exec.ZshCompletions(currentInputValue)
				}
			case Path:
				if msg.String() == "/" {
					options, err = exec.ZshCompletions(currentInputValue)
				}
			}

			if err != nil {
				if !errors.Is(err, exec.MissingZshError{}) {
					(*m.current).SetValue(err.Error())
				}
				break
			}
			suggestions := make([]string, len(options))
			for _, o := range options {
				// Paths are only allowed directories
				if m.focus == Path {
					finfo, err := os.Stat(o.Option)
					if err != nil || !finfo.IsDir() {
						continue
					}
				}
				suggestions = append(suggestions, o.Option)
			}

			// only set new suggestions so we're not overwriting
			// existing on every keypress
			if len(suggestions) > 0 {
				(*m.current).SetSuggestions(suggestions)
			}
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

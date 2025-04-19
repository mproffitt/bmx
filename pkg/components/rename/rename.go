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

package rename

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/components/overlay"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/theme"
	"github.com/mproffitt/bmx/pkg/tmux/ui/session"
	"github.com/mproffitt/bmx/pkg/tmux/ui/window"
)

type Model struct {
	config *config.Config
	input  textinput.Model
	model  session.Renamable
}

func New(what session.Renamable, config *config.Config) *Model {
	m := Model{
		input: textinput.New(),
		model: what,
	}
	m.input.Width = 30
	m.input.Focus()
	return &m
}

func (m *Model) GetSize() (int, int) {
	return 35, 5
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return nil, nil
		case "enter":
			value := m.input.Value()
			// don't continue if name is empty or matches original
			if value == "" || value == m.model.GetName() {
				return nil, nil
			}

			{
				err := m.model.Rename(value)
				if err != nil {
					cmd = helpers.NewErrorCmd(err)
				}

				if err == nil && m.model.GetName() == m.config.DefaultSession {
					err = m.config.SetDefaultSession(value)
					if err != nil {
						cmd = helpers.NewErrorCmd(err)
					}
				}
			}
			return nil, cmd
		default:
			m.input, cmd = m.input.Update(msg)
		}
	}
	return m, cmd
}

func (m *Model) View() string {
	content := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(theme.Colours.Green).
		Render(m.input.View())

	name := m.model.GetName()

	var t string
	switch m.model.(type) {
	case *session.Session:
		t = "session"
	case *window.Window:
		t = "window"
	case *window.Node:
		t = "pane"
	}
	title := fmt.Sprintf("Rename %s : %s", t, name)
	content = overlay.PlaceOverlay(2, 0, title, content, false)
	return content
}

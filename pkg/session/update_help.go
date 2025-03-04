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
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/dialog"
	"github.com/mproffitt/bmx/pkg/helpers"
)

func (m *model) displayHelp() tea.Cmd {
	message := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.config.Style.Title)).
		PaddingBottom(1).BorderBottom(true).
		Render("Key mappings")

	// create session help
	{
		helpmsg := help.New()
		helpmsg.Styles.FullKey = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Style.ContextListActiveTitle))
		helpmsg.Styles.FullDesc = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Style.ContextListActiveDescription))
		helpmsg.ShowAll = true
		session := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Style.FocusedColor)).Render("Session window")
		session = lipgloss.JoinVertical(lipgloss.Left, session, helpmsg.View(m.keymap))
		message = lipgloss.JoinVertical(lipgloss.Left, message, session)
	}
	// create context help
	{
		if m.config.ManageSessionKubeContext && m.context != nil {
			contextHelp := m.context.(helpers.UseHelp).Help()
			message = lipgloss.JoinVertical(lipgloss.Left, message, contextHelp)
		}
	}
	m.dialog = dialog.New(message, true, m.config, false, 2*config.DialogWidth)

	return nil
}

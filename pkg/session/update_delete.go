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
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/components/dialog"
	"github.com/mproffitt/bmx/pkg/components/overlay"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/muesli/reflow/wordwrap"
)

func (m *model) delete(msg tea.Msg) tea.Cmd {
	if m.focused == overlayPane {
		return nil
	}

	var cmd tea.Cmd
	if m.focused == contextPane {
		m.context, cmd = m.context.Update(msg)
		if m.overlay == nil {
			m.overlay = overlay.New(&m.context, m.focused)
			m.focused = overlayPane
		}
		return cmd
	}

	var t, name string

	switch m.active {
	case sessionManager:
		t = "session"
		name = m.session.Name
	case windowManager:
		t = "window"
		name = fmt.Sprintf("%s with index %d", m.window.Name, m.window.Index)
	}

	width := config.DialogWidth - 4
	message := "Cannnot delete the current active " + t
	// Dialog cannot delete active session, use kill instead
	m.dialog = dialog.NewOKDialog(wordwrap.String(message, width),
		m.config, config.DialogWidth)

	// TODO: This currently blocks deleting windows in the active session which isn't
	// really what I'm looking for. Deletion of any but the current active window
	// should be feasible so may take a little more configuration to make this work
	// in all scenarios
	if !m.session.Attached {
		m.deleting = true
		message = fmt.Sprintf("Are you sure you want to delete %s\n", t)
		builder := strings.Builder{}
		builder.WriteString(wordwrap.String(message, width))
		builder.WriteString(lipgloss.PlaceHorizontal(width, lipgloss.Center,
			lipgloss.NewStyle().
				Bold(true).
				Padding(1, 0).
				Foreground(m.config.Colours().BrightBlue).
				Render(name)))
		if m.config.CreateSessionKubeConfig {
			message = wordwrap.String("This will remove the associated kubeconfig"+
				" and log you out of all clusters\n", width)
			builder.WriteString(message)
		}

		m.dialog = dialog.NewConfirmDialog(builder.String(), m.config, config.DialogWidth)
	}
	m.dialog.Update(msg)
	return nil
}

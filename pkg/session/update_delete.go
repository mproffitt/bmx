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
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/dialog"
)

func (m *model) delete(msg tea.Msg) tea.Cmd {
	if m.focused == overlay {
		return nil
	}

	var cmd tea.Cmd
	if m.focused == contextPane {
		m.context, cmd = m.context.Update(msg)
		if m.overlay == nil {
			m.overlay = NewOverlayContainer(&m.context, m.focused)
			m.focused = overlay
		}
		return cmd
	}

	// Dialog cannot delete active session, use kill instead
	m.dialog = dialog.NewOKDialog(
		"Cannot delete the current active session", m.config, config.DialogWidth)
	if !m.session.Attached {
		m.deleting = true
		builder := strings.Builder{}
		builder.WriteString("Are you sure you want to delete session\n")
		builder.WriteString(lipgloss.PlaceHorizontal(config.DialogWidth, lipgloss.Center,
			lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(m.config.Style.FocusedColor)).
				Padding(1).
				Render(m.session.Name)))
		if m.config.CreateSessionKubeConfig {
			builder.WriteString("\nThis will remove the associated kubeconfig")
			builder.WriteString("and log you out of all clusters")
		}

		m.dialog = dialog.NewConfirmDialog(builder.String(), m.config, config.DialogWidth)
	}
	m.dialog.Update(msg)
	return nil
}

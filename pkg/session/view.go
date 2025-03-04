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

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/dialog"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/kubernetes/ui/panel"
	"github.com/mproffitt/bmx/pkg/tmux"
)

func (m *model) View() string {
	current := m.list.SelectedItem()
	if current == nil {
		current = m.list.Items()[0]
	}
	session := current.(list.DefaultItem).Title()
	panes, _ := tmux.SessionPanes(session)
	preview, _ := tmux.CapturePane(panes[0])

	m.preview.SetContent(preview)
	if m.config.ManageSessionKubeContext && m.context != nil {
		m.context = m.context.(*panel.Model).UpdateContextList(
			session, m.getSessionKubeconfig(session))
	}
	left := strings.Builder{}
	left.WriteString("\n")
	left.WriteString(m.styles.sessionlist.Render(m.list.View()) + "\n")

	right := strings.Builder{}
	right.WriteString("\n")
	switch m.focused {
	case sessionList, contextPane, overlay:
		right.WriteString(m.styles.viewportNormal.Render(m.preview.View()))
	case previewPane:
		right.WriteString(m.styles.viewportFocused.Render(m.preview.View()))
	}
	right.WriteString("\n")

	if m.config.ManageSessionKubeContext && m.context != nil {
		right.WriteString(m.context.View())
	}

	doc := lipgloss.JoinHorizontal(lipgloss.Top, left.String(), right.String())

	if m.dialog != nil {
		dw, _ := m.dialog.(*dialog.Dialog).GetSize()
		w := m.width/2 - max(dw, config.DialogWidth)/2
		return helpers.PlaceOverlay(w, 10, m.dialog.View(),
			doc, false)
	}

	if m.overlay != nil {
		w, h := m.overlay.model.GetSize()
		w = m.width/2 - max(w, config.DialogWidth)/2
		h = m.height/2 - max(h, 10)/2

		return helpers.PlaceOverlay(w, h, m.overlay.View(),
			doc, false)
	}

	return doc
}

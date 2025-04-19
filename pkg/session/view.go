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

	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/components/dialog"
	"github.com/mproffitt/bmx/pkg/components/overlay"
	"github.com/mproffitt/bmx/pkg/components/viewport"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/kubernetes/ui/panel"
	"github.com/mproffitt/bmx/pkg/tmux/ui/session"
	tmuxui "github.com/mproffitt/bmx/pkg/tmux/ui/window"
)

func (m *model) View() string {
	var index uint64 = 1
	switch m.active {
	case sessionManager:
		if selected, ok := m.list.SelectedItem().(*session.Session); ok {
			m.session = selected
			m.window = m.session.Windows[0]
		}
	case windowManager:
		if selected, ok := m.list.SelectedItem().(*tmuxui.Window); ok {
			m.window = selected
			index = uint64(m.window.Index)
		}
	}

	if !m.ready {
		if m.splash == nil {
			return ""
		}
		splash := m.splash.View()
		splash = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, splash)
		return splash
	}

	if m.config.ManageSessionKubeContext && m.context != nil {
		m.context = m.context.(*panel.Model).UpdateContextList(
			m.session.Name, m.getSessionKubeconfig(m.session.Name))
	}

	var left string
	{
		sessionlist := m.styles.sessionlist.MarginTop(1).Render(m.list.View())
		w, h := m.list.Width()+2, m.list.Height()
		left = viewport.New(w, h).
			SetContent(sessionlist).
			SetSize(w, h).
			View()
	}

	title := fmt.Sprintf("Preview : %s:%d", m.session.Name, index)
	if m.zoomed {
		lastpane := m.lastch
		maxPanes := m.manager.Session(m.session.Name).Window(index).Len()
		if int(m.lastch) >= maxPanes {
			lastpane = uint(maxPanes)
		}
		title = fmt.Sprintf("Preview : %s:%d.%d", m.session.Name, index, lastpane)
	}
	m.preview.SetTitle(title, viewport.Inline)
	m.makePreview(m.session.Name, index, m.lastch)

	right := strings.Builder{}
	{
		switch m.focused {
		case sessionList, contextPane, overlayPane:
			m.preview.Blur()
		case previewPane:
			m.preview.Focus()
		}

		right.WriteString(m.preview.View())
		if m.config.ManageSessionKubeContext && m.context != nil && !m.contextHidden {
			right.WriteString("\n" + m.context.View())
		}
	}

	doc := lipgloss.JoinHorizontal(lipgloss.Bottom, left, right.String())
	return m.viewOverlays(doc)
}

func (m *model) viewOverlays(doc string) string {
	// Always place the toast overlay even if we have other dialogs
	// so don't retyrn here
	if m.toast != nil {
		doc = overlay.PlaceOverlay(1, ((m.height - m.toast.Height) - 3), m.toast.View(), doc, false)
	}
	if m.dialog != nil {
		dw, _ := m.dialog.(*dialog.Dialog).GetSize()
		w := m.width/2 - max(dw, config.DialogWidth)/2
		return overlay.PlaceOverlay(w, 10, m.dialog.View(), doc, false)
	}

	if m.overlay != nil {
		w, h := m.overlay.Model.GetSize()
		w = m.width/2 - max(w, config.DialogWidth)/2
		h = m.height/2 - max(h, 10)/2

		return overlay.PlaceOverlay(w, h, m.overlay.View(),
			doc, false)
	}

	if m.renameOverlay != nil {
		w, h := m.renameOverlay.GetSize()
		w = m.width/2 - max(w, config.DialogWidth)/2
		h = m.height/2 - max(h, 10)/2

		return overlay.PlaceOverlay(w, h, m.renameOverlay.View(),
			doc, false)
	}

	return doc
}

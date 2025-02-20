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

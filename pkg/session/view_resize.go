package session

import (
	"math"

	"github.com/charmbracelet/bubbles/list"
	"github.com/mproffitt/bmx/pkg/kubernetes/ui/panel"
)

func (m *model) resize() {
	if m.session == nil {
		return
	}
	width := min(listWidth, int(float64(m.width)*.25))
	height := m.height // (m.height - padding)

	previewWidth := (m.width - width) - padding
	m.list.SetSize(width, height-padding)
	m.preview.SetSize(previewWidth, height)

	if m.config.ManageSessionKubeContext && !m.contextHidden {
		// look for 40% of the screen space
		sessionHeight := int(math.Ceil(float64(height) * kubernetesSessionHeight))

		// title + pager + rows + padding + border
		minheight := 2 + 2 + 4 + (padding * 2) + 0
		sessionHeight = max(minheight, sessionHeight)

		// subtract from that required elements
		sessionHeight -= (panel.PanelTitle + panel.PanelFooter)

		// calculate how many rows and cols we can use
		rows, cols, colWidth := 0, 0, 0
		{
			cols = int(math.Floor(float64(previewWidth-2) / float64(panel.KubernetesListWidth)))
			colWidth = int(math.Floor(float64(previewWidth-2) / float64(cols)))
			rows = int(math.Floor(float64(sessionHeight) / float64(panel.KubernetesRowHeight)))
		}

		if m.context == nil {
			session := m.list.SelectedItem().(list.DefaultItem).Title()
			m.context = panel.NewKubectxPane(m.config, session, rows, cols, colWidth)
		}
		m.preview.SetSize(previewWidth, (height-sessionHeight)-2)
		m.context = m.context.(*panel.Model).SetSize(previewWidth-6, sessionHeight, colWidth)
	}
}

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

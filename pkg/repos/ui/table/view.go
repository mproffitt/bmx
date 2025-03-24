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
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/components/dialog"
	"github.com/mproffitt/bmx/pkg/components/overlay"
	"github.com/mproffitt/bmx/pkg/config"
)

func (m *Model) View() string {
	if m.spinner != nil {
		spinner := fmt.Sprintf("%s%s%s\n\t- %s\n", m.spinner.View(),
			" ", m.styles.text.Render("Loading..."), strings.Join(m.paths, "\n\t- "))
		m.viewport = viewport.New(m.width, m.height)
		m.viewport.SetContent(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, spinner))
		return m.styles.viewport.Render(m.viewport.View())
	}

	body := strings.Builder{}
	if m.isOverlay {
		title := m.styles.title.Render("Create new session\n")
		body.WriteString(title)
	}

	subtract := 9
	if m.isOverlay {
		subtract = 11
	}
	var content string
	{
		viewport := viewport.New(m.width-4, m.height-subtract)
		viewport.SetContent(m.table.View())
		content = m.styles.viewport.Padding(0, 0, 1, 2).Render(viewport.View())
		body.WriteString(content + "\n")
	}
	m.panel.SetWidth(m.width)

	body.WriteString(m.panel.View())

	doc := m.styles.table.Render(body.String())
	if m.dialog != nil {
		dw, _ := m.dialog.(*dialog.Dialog).GetSize()
		w := m.width/2 - max(dw, config.DialogWidth)/2
		doc = overlay.PlaceOverlay(w, 10, m.dialog.View(),
			doc, false)
	}

	if m.isOverlay {
		m.viewport = viewport.New(m.width, m.height)
		m.viewport.SetContent(doc)
		return m.styles.viewport.Render(m.viewport.View())
	}

	return doc
}

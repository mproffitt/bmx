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

package panel

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/x/ansi"
	"github.com/mproffitt/bmx/pkg/kubernetes"
)

const (
	ellipsis = "…"

	kubernetesSymbol = "󱃾"
)

type ItemDelegate struct {
	Styles  list.DefaultItemStyles
	height  int
	spacing int
}

func NewItemDelegate() ItemDelegate {
	const defaultHeight = 2
	const defaultSpacing = 1
	return ItemDelegate{
		Styles:  list.NewDefaultItemStyles(),
		height:  defaultHeight,
		spacing: defaultSpacing,
	}
}

func (d ItemDelegate) Height() int {
	return d.height
}

func (d ItemDelegate) Spacing() int {
	return d.spacing
}

func (d ItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	var (
		title, desc string
		s           = &d.Styles
		i           list.DefaultItem
		ok          bool
	)

	if i, ok = item.(list.DefaultItem); !ok {
		return
	}
	title = i.Title()
	desc = i.Description()

	if m.Width() <= 0 {
		// short-circuit
		return
	}

	// Prevent text from exceeding list width
	textwidth := m.Width() - s.NormalTitle.GetPaddingLeft() - s.NormalTitle.GetPaddingRight() - 2
	title = ansi.Truncate(title, textwidth, ellipsis)

	// description
	{
		var lines []string
		for i, line := range strings.Split(desc, "\n") {
			if i >= d.height-1 {
				break
			}
			lines = append(lines, ansi.Truncate(line, textwidth, ellipsis))
		}
		desc = strings.Join(lines, "\n")
	}

	isSelected := index == m.Index()

	stitle := s.NormalTitle.Render(title)
	sdesc := s.NormalDesc.Render(desc)
	if isSelected {
		stitle = s.SelectedTitle.Render(title)
		sdesc = s.SelectedDesc.Render(desc)
	}

	symbol := " "
	if item.(kubernetes.KubeContext).IsCurrentContext {
		symbol = lipgloss.NewStyle().Foreground(lipgloss.Color("#326CE5")).Render(
			kubernetesSymbol)
	}

	if symbol != "" {
		if _, err := fmt.Fprintf(w, "%s%s\n %s", symbol, stitle, sdesc); err != nil {
			log.Error("failed to write to delegate", "error", err, "symbol", symbol, "title", stitle, "description", sdesc)
		}
		return
	}

	if _, err := fmt.Fprintf(w, "%s\n%s", stitle, sdesc); err != nil {
		log.Error("failed to write to delegate", "error", err, "title", stitle, "description", sdesc)
	}
}

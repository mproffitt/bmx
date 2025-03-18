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

package viewport

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/components/overlay"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/muesli/reflow/padding"
)

// This package is a wrapper to bubbles/viewport that allows for
// title placement on the border line

type TitlePos uint

const (
	None TitlePos = iota
	Border
	Inline
)

type Model struct {
	Content       string
	Title         string
	borderStyle   lipgloss.Border
	colours       config.ColourStyles
	height        int
	width         int
	viewport      viewport.Model
	titlePosition TitlePos
	focus         bool
}

func New(colours config.ColourStyles, width, height int) *Model {
	m := Model{
		colours:       colours,
		height:        height,
		titlePosition: None,
		width:         width,
		borderStyle:   lipgloss.RoundedBorder(),
	}
	return &m
}

func (m *Model) Blur() {
	m.focus = false
}

func (m *Model) BorderStyle(b lipgloss.Border) *Model {
	m.borderStyle = b
	return m
}

func (m *Model) Focus() {
	m.focus = true
}

func (m *Model) AtTop() bool {
	return m.viewport.AtTop()
}

func (m *Model) AtBottom() bool {
	return m.viewport.AtBottom()
}

func (m *Model) Init() tea.Cmd {
	return m.viewport.Init()
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	if !m.focus {
		return m, nil
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *Model) SetTitle(title string, pos TitlePos) *Model {
	m.Title = title
	m.titlePosition = pos

	return m
}

func (m *Model) SetContent(content string) *Model {
	m.Content = content
	return m
}

func (m *Model) SetSize(w, h int) *Model {
	m.width = w
	m.height = h
	return m
}

func (m *Model) GetSize() (int, int) {
	return m.width, m.height
}

func (m *Model) GetDrawableSize() (int, int) {
	switch m.titlePosition {
	case Inline:
		return m.width - 2, m.height - 4
	case Border:
		return m.width - 2, m.height - 2
	}
	return m.width - 2, m.height
}

func (m *Model) View() string {
	if m.width == 0 {
		return ""
	}

	borderColour := m.colours.Black
	if m.focus {
		borderColour = m.colours.Blue
	}

	title := lipgloss.NewStyle().Foreground(m.colours.Yellow).Render(m.Title)
	{
		if m.titlePosition == Inline {
			title = padding.String(title, uint(m.width-4))
			title = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(borderColour).
				PaddingLeft(2).
				Render(title)
		}
	}

	m.viewport.SetContent(m.Content)
	m.viewport.Width, m.viewport.Height = m.GetDrawableSize()

	var content string
	{
		style := lipgloss.NewStyle().
			Border(m.borderStyle, true).
			BorderForeground(borderColour)

		content = m.viewport.View()
		switch m.titlePosition {
		case Border:
			content = style.Render(content)
			content = overlay.PlaceOverlay(2, 0, title, content, false)

		case Inline:
			content = lipgloss.JoinVertical(lipgloss.Left, title, content)
			fallthrough
		case None:
			content = style.Render(content)
		}
	}
	return content
}

func (m *Model) PastBottom() bool {
	return m.viewport.PastBottom()
}

func (m *Model) ScrollPercent() float64 {
	return m.viewport.ScrollPercent()
}

func (m *Model) SetYOffset(n int) {
	m.viewport.SetYOffset(n)
}

func (m *Model) ViewDown() []string {
	return m.viewport.ViewDown()
}

func (m *Model) ViewUp() []string {
	return m.viewport.ViewUp()
}

func (m *Model) HalfViewDown() []string {
	return m.viewport.HalfViewDown()
}

func (m *Model) HalfViewUp() []string {
	return m.viewport.HalfViewUp()
}

func (m *Model) LineDown(n int) []string {
	return m.viewport.LineDown(n)
}

func (m *Model) LineUp(n int) []string {
	return m.viewport.LineUp(n)
}

func (m *Model) TotalLineCount() int {
	return m.viewport.TotalLineCount()
}

func (m *Model) VisibleLineCount() int {
	return m.viewport.VisibleLineCount()
}

func (m *Model) GotoTop() []string {
	return m.viewport.GotoTop()
}

func (m *Model) GotoBottom() []string {
	return m.viewport.GotoBottom()
}

// Future (not released upstream)

/*func (m *Model) HorizontalScrollPercent() float64 {
	return m.viewport.HorizontalScrollPercent()
}

func (m *Model) SetHorizontalStep(n int) {
	return m.viewport.SetHorizontalStep(n)
}

func (m *Model) MoveLeft(cols int) {
	m.viewport.MoveLeft(cols)
}

func (m *Model) MoveRight(cols int) {
	m.viewport.MoveRight(cols)
}

func (m *Model) ResetIndent() {
	m.viewport.ResetIndent()
}*/

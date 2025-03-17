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

package dialog

import (
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/helpers"

	tea "github.com/charmbracelet/bubbletea"
)

type Status int

const (
	NoChg Status = iota
	Confirm
	Cancel
)

const DialogHeight = 15

type Dialog struct {
	active     Status
	done       bool
	config     *config.Config
	hasConfirm bool
	height     int
	message    string
	standalone bool
	styles     styles
	viewport   viewport.Model
	width      int
}

type styles struct {
	activeButton lipgloss.Style
	button       lipgloss.Style
	dialog       lipgloss.Style
}

type DialogStatusMsg struct {
	Selected Status
	Done     bool
	Message  string
}

func DialogStatusCmd(message DialogStatusMsg) tea.Cmd {
	return func() tea.Msg {
		return message
	}
}

// Creates a new confirmation dialog
//
// This is a convenience method to create a confirmation dialog
// containing both yes and no buttons
func NewConfirmDialog(message string, config *config.Config, width int) tea.Model {
	return New(message, false, config, false, width)
}

// Creates a new OK dialog
//
// This is a convenience method to create a confirmation dialog
// that only has an OK button
func NewOKDialog(message string, config *config.Config, width int) tea.Model {
	return New(message, true, config, false, width)
}

// Create a new standalone confirmation dialog
//
// # For standalone applications only
//
// This is a convenience method for creating dialogs that run inside
// a TMUX popup window.
//
// This method should not be used from inside the main window.
// Use NewConfirmDialog for that
func NewStandaloneConfirmDialog(message string, config *config.Config, width int) tea.Model {
	return New(message, false, config, true, width)
}

// Create a new OK dialog
//
// # For standalone applications only
//
// This is a convenience method for creating an OK dialog
// that runs inside a TMUX popup window.
//
// This method should not be used from inside the main window.
// Use NewOKDialog for that
func NewStanaloneOKDialog(message string, config *config.Config, width int) tea.Model {
	return New(message, true, config, true, width)
}

func New(message string, cancelOnly bool, c *config.Config, standalone bool, width int) tea.Model {
	height := min(DialogHeight, lipgloss.Height(message)+1)
	d := Dialog{
		active:     Cancel,
		config:     c,
		hasConfirm: !cancelOnly,
		height:     height,
		message:    message,
		standalone: standalone,
		styles: styles{
			dialog: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(c.Colours().Black).
				Padding(1, 0),

			button: lipgloss.NewStyle().
				Foreground(c.Colours().Bg).
				Background(c.Colours().Fg).
				Padding(0, 3).
				MarginTop(1).
				MarginRight(2),

			activeButton: lipgloss.NewStyle().
				Foreground(c.Colours().BrightWhite).
				Background(c.Colours().BrightRed).
				MarginRight(2).
				MarginTop(1).
				Padding(0, 3).
				Underline(true),
		},
		viewport: viewport.New(width, height),
		width:    width,
	}
	if standalone {
		d.styles.dialog = lipgloss.NewStyle().Padding(1, 0)
	}
	return &d
}

func (m *Dialog) Init() tea.Cmd {
	return nil
}

func (m *Dialog) Overlay() helpers.UseOverlay {
	return m
}

func (m *Dialog) GetSize() (int, int) {
	return m.width, m.height
}

func (m *Dialog) SetSize(w, h int) {
	m.height = h
	m.width = w
}

func (m *Dialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "right", "tab":
			switch m.active {
			case Confirm:
				m.active = Cancel
			case Cancel:
				m.active = Confirm
			}
		case "y", "Y":
			if m.hasConfirm {
				m.active = Confirm
			}
			m.done = true
		case "n", "N", "o", "O":
			m.active = Cancel
			m.done = true
		case "enter":
			m.done = true
		default:
			m.viewport, _ = m.viewport.Update(msg)
		}
	}
	if m.standalone && m.done {
		return m, tea.Quit
	}
	return m, DialogStatusCmd(DialogStatusMsg{
		Selected: m.active,
		Done:     m.done,
	})
}

func (m *Dialog) Status() Status {
	return m.active
}

func (m *Dialog) View() string {
	var okButton, cancelButton string
	buttons := m.styles.activeButton.Render("Ok")
	if m.hasConfirm {
		switch m.active {
		case Confirm:
			okButton = m.styles.activeButton.Render("Yes")
			cancelButton = m.styles.button.Render("No")
		case Cancel:
			okButton = m.styles.button.Render("Yes")
			cancelButton = m.styles.activeButton.Render("No")
		}
		buttons = lipgloss.JoinHorizontal(
			lipgloss.Top,
			okButton, cancelButton,
		)
	}
	style := lipgloss.NewStyle().Width(m.width).Align(lipgloss.Left).PaddingLeft(1)
	if m.height != 0 {
		style = style.Height(m.height)
	}
	question := style.Render(m.message)
	m.viewport.SetContent(question)

	return m.styles.dialog.Render(lipgloss.JoinVertical(lipgloss.Center, m.viewport.View(), buttons))
}

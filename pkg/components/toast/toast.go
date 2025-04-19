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

package toast

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/mproffitt/bmx/pkg/components/icons"
	"github.com/mproffitt/bmx/pkg/theme"
	"github.com/muesli/reflow/truncate"
	"github.com/muesli/reflow/wordwrap"
)

// A Toast is simply a viewport with progress
// that lasts for a few seconds on screen and
// disappears without any interaction.

type ToastType int

const (
	Info ToastType = iota
	Success
	Warning
	Error
)

const (
	padding         = 2
	maxHeight       = 4
	maxWidth        = 35
	DefaultProgress = 0.01
	DefaultDuration = 55 * time.Millisecond
)

func newID() string {
	return uuid.NewString()[:8]
}

type (
	TickMsg     time.Time
	NewToastMsg struct {
		Type    ToastType
		Message string
	}
	FrameMsg struct {
		id string
	}
)

func NewToastCmd(t ToastType, m string) tea.Cmd {
	return func() tea.Msg {
		return NewToastMsg{
			Type:    t,
			Message: m,
		}
	}
}

type Model struct {
	id      string
	Message string
	Type    ToastType
	Icon    rune
	Height  int
	Width   int

	viewport          viewport.Model
	progress          progress.Model
	percent           float64
	styles            styles
	activeStyle       *lipgloss.Style
	completionCommand tea.Cmd

	progressPercent float64
	tickDuration    time.Duration
}

type styles struct {
	info    lipgloss.Style
	success lipgloss.Style
	warning lipgloss.Style
	error   lipgloss.Style
}

func New(t ToastType, message string) *Model {
	m := Model{
		id:              newID(),
		Message:         message,
		Type:            t,
		Width:           maxWidth,
		percent:         1.0,
		progressPercent: DefaultProgress,
		tickDuration:    DefaultDuration,

		styles: styles{
			info: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(theme.Colours.Blue),
			success: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(theme.Colours.Green),
			warning: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(theme.Colours.Yellow),
			error: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(theme.Colours.Red),
		},
	}

	var coloura, colourb string
	switch m.Type {
	case Info:
		m.Icon = icons.Info
		m.activeStyle = &m.styles.info
		coloura, colourb = "#9ece6a", "#7aa2f7"
	case Success:
		m.Icon = icons.Tick
		m.activeStyle = &m.styles.info
		coloura, colourb = "#b4f9f8", "#9ece6a"
	case Warning:
		m.Icon = icons.Warning
		m.activeStyle = &m.styles.warning
		coloura, colourb = "#f7768e", "#ff9e64"
	case Error:
		m.Icon = icons.Error
		m.activeStyle = &m.styles.error
		coloura, colourb = "#e0af68", "#f7768e"
	}

	content := fmt.Sprintf("%s %s", string(m.Icon), message)
	content = wordwrap.String(content, maxWidth-padding)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if i > 0 {
			lines[i] = "  " + line
		}
	}

	if len(lines) > 3 {
		lines = lines[:3]
		lines[2] = truncate.String(lines[2], uint(maxWidth-padding-1)) + string(icons.Ellipis)
	}

	m.Height = len(lines) + 1
	m.viewport = viewport.New(maxWidth, min(maxHeight, m.Height))

	m.Message = strings.Join(lines, "\n")

	m.progress = progress.New(
		progress.WithScaledGradient(coloura, colourb),
		progress.WithoutPercentage(),
		progress.WithFillCharacters('─', ' '),
	)
	m.progress.Width = 30
	return &m
}

func (m *Model) Init() tea.Cmd {
	return m.tickCmd()
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	switch msg := msg.(type) {
	case FrameMsg:
		if msg.id != m.id {
			break
		}
		m.percent -= m.progressPercent
		if m.percent <= .0 {
			return nil, m.completionCommand
		}
		return m, m.tickCmd()
	}
	return m, nil
}

func (m *Model) SetCompletionCommand(cmd tea.Cmd) *Model {
	m.completionCommand = cmd
	return m
}

func (m *Model) SetProgressSpeed(speed float64) *Model {
	m.progressPercent = speed
	return m
}

func (m *Model) SetTickDuration(duration time.Duration) *Model {
	m.tickDuration = duration
	return m
}

func (m *Model) View() string {
	if m.percent <= .0 {
		return ""
	}
	var colour lipgloss.AdaptiveColor
	switch m.Type {
	case Info:
		colour = theme.Colours.Blue
	case Success:
		colour = theme.Colours.Green
	case Warning:
		colour = theme.Colours.Yellow
	case Error:
		colour = theme.Colours.Red
	}
	content := lipgloss.NewStyle().
		Foreground(colour).
		Render(m.Message)
	content = lipgloss.JoinVertical(lipgloss.Left, content, m.progress.ViewAs(m.percent))

	m.viewport.SetContent(strings.TrimSpace(content))
	return m.activeStyle.Padding(0, 1, 0, 1).Render(m.viewport.View())
}

func (m *Model) tickCmd() tea.Cmd {
	return tea.Tick(m.tickDuration, func(t time.Time) tea.Msg {
		return FrameMsg{id: m.id}
	})
}

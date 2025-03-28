package toast

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/components/icons"
	"github.com/mproffitt/bmx/pkg/config"
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
	padding   = 2
	maxHeight = 4
	maxWidth  = 35
)

type (
	TickMsg     time.Time
	NewToastMsg struct {
		Type    ToastType
		Message string
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
	Message string
	Type    ToastType
	Icon    rune
	Height  int
	Width   int

	viewport    viewport.Model
	progress    progress.Model
	colours     config.ColourStyles
	percent     float64
	styles      styles
	activeStyle *lipgloss.Style
}

type styles struct {
	info    lipgloss.Style
	success lipgloss.Style
	warning lipgloss.Style
	error   lipgloss.Style
}

func New(t ToastType, message string, colours config.ColourStyles) *Model {
	m := Model{
		Message: message,
		Type:    t,
		Width:   maxWidth,
		colours: colours,
		percent: 1.0,
		styles: styles{
			info: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(colours.Blue),
			success: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(colours.Green),
			warning: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(colours.Yellow),
			error: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(colours.Red),
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
	switch msg.(type) {
	case TickMsg:
		m.percent -= 0.01
		if m.percent <= .0 {
			return nil, nil
		}
		return m, m.tickCmd()

	default:
		return m, nil
	}
}

func (m *Model) View() string {
	if m.percent <= .0 {
		return ""
	}
	var colour lipgloss.AdaptiveColor
	switch m.Type {
	case Info:
		colour = m.colours.Blue
	case Success:
		colour = m.colours.Green
	case Warning:
		colour = m.colours.Yellow
	case Error:
		colour = m.colours.Red
	}
	content := lipgloss.NewStyle().
		Foreground(colour).
		Render(m.Message)
	content = lipgloss.JoinVertical(lipgloss.Left, content, m.progress.ViewAs(m.percent))

	m.viewport.SetContent(strings.TrimSpace(content))
	return m.activeStyle.Padding(0, 1, 0, 1).Render(m.viewport.View())
}

func (m *Model) tickCmd() tea.Cmd {
	return tea.Tick(55*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

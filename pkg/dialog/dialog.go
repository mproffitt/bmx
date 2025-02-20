package dialog

import (
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

type Dialog struct {
	active     Status
	done       bool
	config     *config.Config
	hasConfirm bool
	height     int
	message    string
	standalone bool
	styles     styles
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

func New(message string, cancelOnly bool, c *config.Config, standalone bool, width int) tea.Model {
	d := Dialog{
		active:     Cancel,
		hasConfirm: !cancelOnly,
		config:     c,
		message:    message,
		standalone: standalone,
		styles: styles{
			dialog: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(lipgloss.Color(c.Style.DialogBorderColor)).
				Padding(1, 0),

			button: lipgloss.NewStyle().
				Foreground(lipgloss.Color(c.Style.ButtonInactiveForeground)).
				Background(lipgloss.Color(c.Style.ButtonInactiveBackground)).
				Padding(0, 3).
				MarginTop(1).
				MarginRight(2),

			activeButton: lipgloss.NewStyle().
				Foreground(lipgloss.Color(c.Style.ButtonActiveForeground)).
				Background(lipgloss.Color(c.Style.ButtonActiveBackground)).
				MarginRight(2).
				MarginTop(1).
				Padding(0, 3).
				Underline(true),
		},
		width: width,
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

	return m.styles.dialog.Render(lipgloss.JoinVertical(lipgloss.Center, question, buttons))
}

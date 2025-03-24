package rename

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/components/overlay"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/tmux/ui/session"
	"github.com/mproffitt/bmx/pkg/tmux/ui/window"
)

type Model struct {
	config *config.Config
	input  textinput.Model
	model  session.Renamable
}

func New(what session.Renamable, config *config.Config) *Model {
	m := Model{
		config: config,
		input:  textinput.New(),
		model:  what,
	}
	m.input.Width = 30
	m.input.Focus()
	return &m
}

func (m *Model) GetSize() (int, int) {
	return 35, 5
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return nil, nil
		case "enter":
			value := m.input.Value()
			// don't continue if name is empty or matches original
			if value == "" || value == m.model.GetName() {
				return nil, nil
			}

			{
				err := m.model.Rename(value)
				if err != nil {
					cmd = helpers.NewErrorCmd(err)
				}

				if err == nil && m.model.GetName() == m.config.DefaultSession {
					err = m.config.SetDefaultSession(value)
					if err != nil {
						cmd = helpers.NewErrorCmd(err)
					}
				}
			}
			return nil, cmd
		default:
			m.input, cmd = m.input.Update(msg)
		}
	}
	return m, cmd
}

func (m *Model) View() string {
	content := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(m.config.Colours().Green).
		Render(m.input.View())

	name := m.model.GetName()

	var t string
	switch m.model.(type) {
	case *session.Session:
		t = "session"
	case *window.Window:
		t = "window"
	case *window.Node:
		t = "pane"
	}
	title := fmt.Sprintf("Rename %s : %s", t, name)
	content = overlay.PlaceOverlay(2, 0, title, content, false)
	return content
}

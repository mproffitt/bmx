package splash

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/config"
)

const logo = `
__/\\\\\\\\\\\\\____/\\\\____________/\\\\__/\\\_______/\\\_
 _\/\\\/////////\\\_\/\\\\\\________/\\\\\\_\///\\\___/\\\/__
  _\/\\\_______\/\\\_\/\\\//\\\____/\\\//\\\___\///\\\\\\/____
   _\/\\\\\\\\\\\\\\__\/\\\\///\\\/\\\/_\/\\\_____\//\\\\______
    _\/\\\/////////\\\_\/\\\__\///\\\/___\/\\\______\/\\\\______
     _\/\\\_______\/\\\_\/\\\____\///_____\/\\\______/\\\\\\_____
      _\/\\\_______\/\\\_\/\\\_____________\/\\\____/\\\////\\\___
       _\/\\\\\\\\\\\\\/__\/\\\_____________\/\\\__/\\\/___\///\\\_
        _\/////////////____\///______________\///__\///_______\///__
`

type Model struct {
	spinner spinner.Model
	colours config.ColourStyles
}

func New(colours config.ColourStyles) *Model {
	spinner := spinner.New(spinner.WithSpinner(spinner.Meter))
	return &Model{
		spinner: spinner,
		colours: colours,
	}
}

func (m *Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	var c tea.Cmd
	m.spinner, c = m.spinner.Update(msg)
	return m, c
}

func (m *Model) View() string {
	content := lipgloss.JoinVertical(lipgloss.Center, logo, m.spinner.View())
	return lipgloss.NewStyle().Foreground(m.colours.Blue).Render(content)
}

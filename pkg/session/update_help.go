package session

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/dialog"
	"github.com/mproffitt/bmx/pkg/helpers"
)

func (m *model) displayHelp() tea.Cmd {
	message := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.config.Style.Title)).
		PaddingBottom(1).BorderBottom(true).
		Render("Key mappings")

	// create session help
	{
		helpmsg := help.New()
		helpmsg.Styles.FullKey = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Style.ContextListActiveTitle))
		helpmsg.Styles.FullDesc = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Style.ContextListActiveDescription))
		helpmsg.ShowAll = true
		session := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Style.FocusedColor)).Render("Session window")
		session = lipgloss.JoinVertical(lipgloss.Left, session, helpmsg.View(m.keymap))
		message = lipgloss.JoinVertical(lipgloss.Left, message, session)
	}
	// create context help
	{
		if m.config.ManageSessionKubeContext && m.context != nil {
			contextHelp := m.context.(helpers.UseHelp).Help()
			message = lipgloss.JoinVertical(lipgloss.Left, message, contextHelp)
		}
	}
	m.dialog = dialog.New(message, true, m.config, false, 2*config.DialogWidth)

	return nil
}

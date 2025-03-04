package dialog

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/config"
)

type UseHelp interface {
	Help() HelpEntry
}

// A HelpEntry represents an item to be added to the help
// dialog.
type HelpEntry struct {
	// The title to use for this help entry
	Title string

	// TRhe keymap associated with this help entry
	Keymap *help.KeyMap

	// An optional help message
	Help string
}

func HelpDialog(c *config.Config, entries ...HelpEntry) tea.Model {
	helpWidth := 2 * config.DialogWidth
	message := lipgloss.NewStyle().
		Foreground(lipgloss.Color(c.Style.Title)).
		BorderForeground(lipgloss.Color(c.Style.DialogBorderColor)).
		Border(lipgloss.RoundedBorder(), false, false, true, false).
		Render("    Help    ")
	message = lipgloss.PlaceHorizontal(helpWidth, lipgloss.Center, message)

	for _, entry := range entries {
		if entry.Help == "" && entry.Keymap == nil {
			continue
		}

		helpmsg := help.New()
		helpmsg.Styles.FullKey = lipgloss.NewStyle().Foreground(lipgloss.Color(c.Style.ContextListActiveTitle))
		helpmsg.Styles.FullDesc = lipgloss.NewStyle().Foreground(lipgloss.Color(c.Style.ContextListActiveDescription))
		helpmsg.ShowAll = true

		current := lipgloss.NewStyle().
			Foreground(lipgloss.Color(c.Style.FocusedColor)).
			MarginTop(1).
			Render(entry.Title)
		if entry.Help != "" {
			current = lipgloss.JoinVertical(lipgloss.Left, current, lipgloss.NewStyle().
				Foreground(lipgloss.Color(c.Style.ListNormalDescription)).
				MarginBottom(1).
				Render(entry.Help))
		}

		if entry.Keymap != nil {
			current = lipgloss.JoinVertical(lipgloss.Left, current, helpmsg.View(*entry.Keymap))
		}
		message = lipgloss.JoinVertical(lipgloss.Left, message, current)
	}

	return NewOKDialog(message, c, 2*config.DialogWidth)
}

package repos

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type keyMap struct {
	Down     key.Binding
	Enter    key.Binding
	Pageup   key.Binding
	Pagedown key.Binding
	Quit     key.Binding
	Up       key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	// Pageup and Pagedown currently do not work
	// with the panel pager. leaving them out for now
	return [][]key.Binding{
		{
			k.Enter, k.Quit,
		},
		{
			k.Up, k.Down, k.Pageup, k.Pagedown,
		},
	}
}

func mapKeys() keyMap {
	return keyMap{
		Down: key.NewBinding(key.WithKeys("down"),
			key.WithHelp("↓", "move down")),
		Enter: key.NewBinding(key.WithKeys("enter"),
			key.WithHelp("enter", "Set current context")),
		Pageup: key.NewBinding(key.WithKeys("pgup"),
			key.WithHelp("pgup", "previous page")),
		Pagedown: key.NewBinding(key.WithKeys("pgdown"),
			key.WithHelp("pgdn", "next page")),
		Up: key.NewBinding(key.WithKeys("up"),
			key.WithHelp("↑", "move up")),
	}
}

func (m *Model) Help() string {
	helpmsg := help.New()
	helpmsg.ShowAll = true
	help := lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Style.FocusedColor)).Render("Context Pane")
	help = lipgloss.JoinVertical(lipgloss.Left, help, helpmsg.View(m.keymap))
	return help
}

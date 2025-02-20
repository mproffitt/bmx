package panel

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type keyMap struct {
	Delete    key.Binding
	Down      key.Binding
	Enter     key.Binding
	Killpanel key.Binding
	Left      key.Binding
	Move      key.Binding
	Pageup    key.Binding
	Pagedown  key.Binding
	Right     key.Binding
	Space     key.Binding
	Up        key.Binding
	Login     key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Delete, k.Killpanel}
}

func (k keyMap) FullHelp() [][]key.Binding {
	// Pageup and Pagedown currently do not work
	// with the panel pager. leaving them out for now
	return [][]key.Binding{
		{
			k.Delete, k.Enter, k.Killpanel, k.Move, k.Space,
		},
		{
			k.Up, k.Down, k.Left, k.Right, k.Login,
		},
	}
}

func mapKeys() keyMap {
	return keyMap{
		Delete: key.NewBinding(key.WithKeys("delete", "x"),
			key.WithHelp("del/x", "Delete the current item")),
		Down: key.NewBinding(key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down")),
		Enter: key.NewBinding(key.WithKeys("enter"),
			key.WithHelp("↩", "Set current context")),
		Killpanel: key.NewBinding(key.WithKeys("ctrl+c", "esc"),
			key.WithHelp("esc", "Close overlays or quit")),
		Left: key.NewBinding(key.WithKeys("left", "h"),
			key.WithHelp("←/h", "move left")),
		Login: key.NewBinding(key.WithKeys("ctrl+l"),
			key.WithHelp("ctrl+l", "Login to cluster")),
		Move: key.NewBinding(key.WithKeys("m"),
			key.WithHelp("m", "Move context to session")),
		Pageup: key.NewBinding(key.WithKeys("pgup", "b"),
			key.WithHelp("b/pgup", "previous page")),
		Pagedown: key.NewBinding(key.WithKeys("pgdown"),
			key.WithHelp("b/pgdn", "next page")),
		Right: key.NewBinding(key.WithKeys("right", "l"),
			key.WithHelp("→/l", "move right")),
		Space: key.NewBinding(key.WithKeys(" "),
			key.WithHelp("␣", "Change context namespace")),
		Up: key.NewBinding(key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up")),
	}
}

func (m *Model) Help() string {
	helpmsg := help.New()
	helpmsg.Styles.FullKey = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Style.ContextListActiveTitle))
	helpmsg.Styles.FullDesc = lipgloss.NewStyle().Foreground(lipgloss.Color(m.config.Style.ContextListActiveDescription))
	helpmsg.ShowAll = true
	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.config.Style.FocusedColor)).
		PaddingTop(1).
		Render("Context Pane")
	help = lipgloss.JoinVertical(lipgloss.Left, help, helpmsg.View(m.keymap))
	return help
}

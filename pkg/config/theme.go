package config

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const (
	buttonPaddingHorizontal = 2
	buttonPaddingVertical   = 0
)

func getTheme(c *Config) *huh.Theme {
	button := lipgloss.NewStyle().
		Padding(buttonPaddingVertical, buttonPaddingHorizontal).
		MarginRight(1)
	theme := huh.ThemeCharm()
	theme.Group = theme.Group.
		PaddingBottom(2).
		BorderBottom(true).
		BorderBottomForeground(lipgloss.Color(c.Style.BorderFgColor))

	theme.Focused = huh.FieldStyles{
		Base: lipgloss.NewStyle().
			PaddingLeft(2).
			BorderStyle(lipgloss.ThickBorder()).
			BorderLeft(true).
			BorderForeground(lipgloss.Color(c.Style.BorderFgColor)),
		Card:                lipgloss.NewStyle().PaddingLeft(1),
		ErrorIndicator:      lipgloss.NewStyle().SetString(" *"),
		ErrorMessage:        lipgloss.NewStyle().SetString(" *"),
		SelectSelector:      lipgloss.NewStyle().SetString("> "),
		NextIndicator:       lipgloss.NewStyle().MarginLeft(1).SetString("→"),
		PrevIndicator:       lipgloss.NewStyle().MarginRight(1).SetString("←"),
		MultiSelectSelector: lipgloss.NewStyle().SetString("> ").Foreground(lipgloss.Color(c.Style.FocusedColor)),
		SelectedPrefix:      lipgloss.NewStyle().SetString("[•] "),
		UnselectedPrefix:    lipgloss.NewStyle().SetString("[ ] "),
		FocusedButton: button.Foreground(lipgloss.Color(c.Style.ButtonActiveForeground)).
			Background(lipgloss.Color(c.Style.ButtonActiveBackground)),
		BlurredButton: button.Foreground(lipgloss.Color(c.Style.ButtonInactiveForeground)).
			Background(lipgloss.Color(c.Style.ButtonInactiveBackground)),
		TextInput: huh.TextInputStyles{
			Cursor:      lipgloss.NewStyle().Foreground(lipgloss.Color("4")),
			Placeholder: lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
			Prompt:      lipgloss.NewStyle().Foreground(lipgloss.Color(c.Style.Foreground)),
			CursorText:  lipgloss.NewStyle().MarginLeft(1).Foreground(lipgloss.Color("9")),
		},
		Title:          lipgloss.NewStyle().Foreground(lipgloss.Color(c.Style.Title)),
		SelectedOption: lipgloss.NewStyle().Foreground(lipgloss.Color(c.Style.ListNormalSelectedTitle)),
	}
	theme.Blurred = theme.Focused
	theme.Blurred.Base = theme.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
	theme.Blurred.MultiSelectSelector = lipgloss.NewStyle().SetString("")
	theme.Blurred.SelectedOption = lipgloss.NewStyle().Foreground(lipgloss.Color(c.Style.FocusedColor))
	return theme
}

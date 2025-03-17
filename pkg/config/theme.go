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
		BorderBottomForeground(c.Colours().Black)

	theme.Focused = huh.FieldStyles{
		Base: lipgloss.NewStyle().
			PaddingLeft(2).
			BorderStyle(lipgloss.ThickBorder()).
			BorderLeft(true).
			BorderForeground(c.Colours().Black),
		Card:                lipgloss.NewStyle().PaddingLeft(1),
		ErrorIndicator:      lipgloss.NewStyle().SetString(" *"),
		ErrorMessage:        lipgloss.NewStyle().SetString(" *"),
		SelectSelector:      lipgloss.NewStyle().SetString("> "),
		NextIndicator:       lipgloss.NewStyle().MarginLeft(1).SetString("→"),
		PrevIndicator:       lipgloss.NewStyle().MarginRight(1).SetString("←"),
		MultiSelectSelector: lipgloss.NewStyle().SetString("> ").Foreground(c.Colours().BrightBlue),
		SelectedPrefix:      lipgloss.NewStyle().SetString("[•] "),
		UnselectedPrefix:    lipgloss.NewStyle().SetString("[ ] "),
		FocusedButton: button.Foreground(c.Colours().BrightWhite).
			Background(c.Colours().BrightRed),
		BlurredButton: button.Foreground(c.Colours().Bg).
			Background(c.Colours().Fg),
		TextInput: huh.TextInputStyles{
			Cursor:      lipgloss.NewStyle().Foreground(lipgloss.Color("4")),
			Placeholder: lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
			Prompt:      lipgloss.NewStyle().Foreground(c.Colours().Fg),
			CursorText:  lipgloss.NewStyle().MarginLeft(1).Foreground(c.Colours().Green),
		},
		Title:          lipgloss.NewStyle().Foreground(c.Colours().BrightYellow),
		SelectedOption: lipgloss.NewStyle().Foreground(c.Colours().BrightBlue),
	}
	theme.Blurred = theme.Focused
	theme.Blurred.Base = theme.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
	theme.Blurred.MultiSelectSelector = lipgloss.NewStyle().SetString("")
	theme.Blurred.SelectedOption = lipgloss.NewStyle().Foreground(c.Colours().BrightBlue)
	return theme
}

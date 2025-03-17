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
		Foreground(c.Colours().Yellow).
		BorderForeground(c.Colours().Black).
		Border(lipgloss.RoundedBorder(), false, false, true, false).
		Render("    Help    ")
	message = lipgloss.PlaceHorizontal(helpWidth, lipgloss.Center, message)

	for _, entry := range entries {
		if entry.Help == "" && entry.Keymap == nil {
			continue
		}

		helpmsg := help.New()
		helpmsg.Styles.FullKey = lipgloss.NewStyle().Foreground(c.Colours().BrightGreen)
		helpmsg.Styles.FullDesc = lipgloss.NewStyle().Foreground(c.Colours().BrightCyan)
		helpmsg.ShowAll = true

		current := lipgloss.NewStyle().
			Foreground(c.Colours().BrightBlue).
			MarginTop(1).
			Render(entry.Title)
		if entry.Help != "" {
			current = lipgloss.JoinVertical(lipgloss.Left, current, lipgloss.NewStyle().
				Foreground(c.Colours().White).
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

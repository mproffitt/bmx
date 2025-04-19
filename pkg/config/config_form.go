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
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/theme"
)

const (
	tabMainForm   = "Main"
	tabStylesForm = "Styles"
)

type configModel struct {
	config     *Config
	tabs       []string
	tabContent []string
	activeTab  int
	mainform   tea.Model
	styles     configModelStyles
}

type configModelStyles struct {
	docStyle         lipgloss.Style
	activeTabStyle   lipgloss.Style
	inactiveTabStyle lipgloss.Style
	windowStyle      lipgloss.Style

	tabBorder       lipgloss.Border
	tabGapBorder    lipgloss.Border
	activeTabBorder lipgloss.Border
	tabGap          lipgloss.Style
}

// Create a new configuration UI model
//
// The config UI is there to help guide users through the
// process of creating and managing the UI config.
//
// Although the primary config is fairly basic, this helps
// guide users in its creation.
func NewConfigModel(c *Config) *configModel {
	highlight := theme.Colours.Cyan
	m := configModel{
		config: c,
		tabs: []string{
			tabMainForm,
		},
		mainform: NewMainModel(c),
		styles: configModelStyles{
			docStyle: lipgloss.NewStyle().Padding(1, 2, 1, 2),
			windowStyle: lipgloss.NewStyle().
				BorderForeground(highlight).
				Padding(2, 0).
				Align(lipgloss.Left),

			tabBorder: lipgloss.Border{
				Top:      "─",
				Bottom:   "─",
				Left:     "│",
				Right:    "│",
				TopLeft:  "╭",
				TopRight: "╮",

				BottomLeft:  "┴",
				BottomRight: "┴",
			},
			activeTabBorder: lipgloss.Border{
				Top:         "─",
				Bottom:      " ",
				Left:        "│",
				Right:       "│",
				TopLeft:     "╭",
				TopRight:    "╮",
				BottomLeft:  "┘",
				BottomRight: "└",
			},
			tabGapBorder: lipgloss.Border{
				Top:         "",
				Left:        "",
				Right:       "",
				TopLeft:     "",
				TopRight:    "",
				Bottom:      "─",
				BottomLeft:  "─",
				BottomRight: "╮",
			},
		},
	}

	m.styles.windowStyle = m.styles.windowStyle.Border(lipgloss.RoundedBorder()).
		UnsetBorderTop()

	m.styles.inactiveTabStyle = lipgloss.NewStyle().Border(m.styles.tabBorder).
		BorderForeground(highlight).
		Padding(0, 1)
	m.styles.activeTabStyle = m.styles.inactiveTabStyle.
		Border(m.styles.activeTabBorder, true)
	m.styles.tabGap = m.styles.inactiveTabStyle.Border(m.styles.tabGapBorder, true)
	return &m
}

// Initialise the model and forms
func (c *configModel) Init() tea.Cmd {
	return tea.Batch(c.mainform.Init())
}

// Recieve and process updates from the application
func (c *configModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch c.tabs[c.activeTab] {
	case tabMainForm:
		c.mainform, cmd = c.mainform.Update(msg)
	}
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			_ = c.config.writeConfig(c.config.GetConfigFile())
			return c, tea.Quit
		case "tab", "shift+tab", "enter":
			_ = c.config.writeConfig(c.config.GetConfigFile())
		case ".":
			c.activeTab = min(c.activeTab+1, len(c.tabs)-1)
		case ",":
			c.activeTab = max(c.activeTab-1, 0)
		}
	}
	return c, cmd
}

// render the config screens
func (c *configModel) View() string {
	c.createTabContents()
	doc := strings.Builder{}

	var renderedTabs []string

	width := 80
	for i, t := range c.tabs {
		var style lipgloss.Style
		isFirst, isActive := i == 0, i == c.activeTab
		if isActive {
			style = c.styles.activeTabStyle
		} else {
			style = c.styles.inactiveTabStyle
		}
		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst && !isActive {
			border.BottomLeft = "├"
		}
		style = style.Border(border)
		renderedTabs = append(renderedTabs, style.Render(t))
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	spacer := strings.Repeat(" ", max(0, width-lipgloss.Width(row)-2))
	gap := c.styles.tabGap.Render(spacer)

	row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap) + "\n"
	doc.WriteString(row)
	doc.WriteString(c.styles.windowStyle.Width(width).Render(c.tabContent[c.activeTab]))
	return c.styles.docStyle.Render(doc.String())
}

func (c *configModel) createTabContents() {
	c.tabContent = make([]string, len(c.tabs))

	c.tabContent[0] = c.makeMainTab()
}

func (c *configModel) makeMainTab() string {
	return c.mainform.View()
}

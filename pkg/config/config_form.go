package config

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	stylesform tea.Model
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

func NewConfigModel(c *Config) *configModel {
	highlight := lipgloss.Color(c.Style.FocusedColor)
	m := configModel{
		config: c,
		tabs: []string{
			tabMainForm,
			tabStylesForm,
		},
		mainform:   NewMainModel(c),
		stylesform: NewStylesModel(c),
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

func (c *configModel) Init() tea.Cmd {
	return tea.Batch(c.mainform.Init(), c.stylesform.Init())
}

func (c *configModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch c.tabs[c.activeTab] {
	case tabMainForm:
		c.mainform, cmd = c.mainform.Update(msg)
	case tabStylesForm:
		c.stylesform, cmd = c.stylesform.Update(msg)
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
	c.tabContent[1] = c.makeStylesTab()
}

func (c *configModel) makeMainTab() string {
	return c.mainform.View()
}

func (c *configModel) makeStylesTab() string {
	return c.stylesform.View()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

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

package optionlist

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/helpers"
)

const (
	columnKeyName = "name"
	defaultWidth  = 20
)

var customBorder = table.Border{
	Top:            "",
	Left:           "",
	Right:          "",
	Bottom:         "",
	TopRight:       "",
	TopLeft:        "",
	BottomRight:    "",
	BottomLeft:     "",
	TopJunction:    "",
	LeftJunction:   "",
	RightJunction:  "",
	BottomJunction: "",
	InnerJunction:  "",
	InnerDivider:   "",
}

// Namespace/session list will be implemented as a single column table
// as it's more filterable than a list

type OptionModel struct {
	cols        []table.Column
	config      *config.Config
	filterInput textinput.Model
	height      int
	rows        []table.Row
	selected    string
	styles      optionStyles
	table       table.Model
	title       string
	width       int
}

type optionStyles struct {
	overlay lipgloss.Style
	filter  lipgloss.Style
	table   lipgloss.Style
}

type (
	Iterator func(yield func(int, Row) bool)
	Options  interface {
		Title() string
		Options() Iterator
	}
	Row interface {
		GetValue() string
	}
)

type Option struct {
	Value string
}

func (o Option) GetValue() string {
	return o.Value
}

func NewOptionModel[T Options](options T, config *config.Config) *OptionModel {
	n := OptionModel{
		config:      config,
		filterInput: textinput.New(),
		rows:        make([]table.Row, 0),
		styles: optionStyles{
			overlay: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(config.Colours().Black).
				Padding(0, 1),
			filter: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(config.Colours().Green),
			table: lipgloss.NewStyle().
				Align(lipgloss.Left).
				BorderForeground(config.Colours().Black).
				Foreground(config.Colours().BrightBlue).
				Margin(1).
				Padding(0, 2),
		},
		title: options.Title(),
	}

	n.filterInput.TextStyle = n.filterInput.TextStyle.UnsetMargins()

	maxLen := 0
	for _, v := range options.Options() {
		n.rows = append(n.rows, table.NewRow(table.RowData{columnKeyName: v.GetValue()}))
		if len(v.GetValue()) > maxLen {
			maxLen = len(v.GetValue())
		}
	}

	n.cols = []table.Column{
		table.NewColumn(columnKeyName, "", maxLen).
			WithFiltered(true),
	}

	n.styles.filter = n.styles.filter.Width(maxLen)

	n.table = table.New(n.cols).
		Border(customBorder).
		Filtered(true).
		Focused(true).
		WithBaseStyle(n.styles.table).
		WithFooterVisibility(false).
		WithHeaderVisibility(false).
		WithPageSize(20).
		WithRows(n.rows).
		SortByAsc(columnKeyName)

	return &n
}

func (n *OptionModel) Init() tea.Cmd {
	return nil
}

func (n *OptionModel) GetSize() (int, int) {
	return n.width, n.height
}

func (n *OptionModel) Overlay() helpers.UseOverlay {
	if len(n.rows) > 0 {
		return n
	}
	return nil
}

func (n *OptionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			current := n.table.HighlightedRow()
			filter := n.filterInput.Value()
			data := map[string]any(current.Data)
			// default to using the table
			if name, ok := data[columnKeyName].(string); ok {
				cmds = append(cmds, helpers.OverlayCmd(name))
			} else if filter != "" {
				cmds = append(cmds, helpers.OverlayCmd(filter))
			}
		case "up", "down":
		default:
			n.filterInput.Focus()
			n.filterInput, _ = n.filterInput.Update(msg)
			n.filterInput.Blur()
			n.table = n.table.WithFilterInput(n.filterInput)
		}
	case tea.WindowSizeMsg:
		n.width = msg.Width
		n.height = msg.Height
	}

	n.table, cmd = n.table.Update(msg)
	cmds = append(cmds, cmd)
	return n, tea.Batch(cmds...)
}

func (n *OptionModel) View() string {
	title := lipgloss.NewStyle().Padding(0, 2).
		Border(lipgloss.RoundedBorder(), false, false, true, false).
		Foreground(n.config.Colours().Yellow).Align(lipgloss.Center).
		Render(n.title)

	filter := n.styles.filter.Render(n.filterInput.View())
	body := lipgloss.JoinVertical(lipgloss.Center, title, n.table.View(), filter)
	return n.styles.overlay.Render(body)
}

func (n *OptionModel) Rows() int {
	return len(n.rows)
}

func (n *OptionModel) GetSelected() string {
	return n.selected
}

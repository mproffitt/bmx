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
	"github.com/mproffitt/bmx/pkg/kubernetes"
	"github.com/mproffitt/bmx/pkg/tmux"
)

const (
	columnKeyName = "name"
	defaultWidth  = 20
)

type OptionType int

const (
	Namespace OptionType = iota
	Session
	ClusterLogin
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
	context     string
	error       error
	filename    string
	filterInput textinput.Model
	height      int
	optionType  OptionType
	rows        []table.Row
	selected    string
	styles      optionStyles
	table       table.Model
	width       int
}

type optionStyles struct {
	overlay lipgloss.Style
	filter  lipgloss.Style
	table   lipgloss.Style
}

func getRowData(optionType OptionType, context, filename string) (int, []table.Row, error) {
	var (
		options []string
		err     error
		maxName int
		rows    []table.Row
	)
	switch optionType {
	case Namespace:
		ctx := kubernetes.GetFullName(context, filename)
		options, err = kubernetes.GetNamespaces(ctx, filename)

	case Session:
		sessions := tmux.ListSessions()
		options = make([]string, len(sessions))
		for i, v := range sessions {
			options[i] = v.Name
		}
	case ClusterLogin:
		options, err = kubernetes.TeleportClusterList()
	}

	for _, name := range options {
		rows = append(rows, table.NewRow(table.RowData{columnKeyName: name}))
		if len(name) > maxName {
			maxName = len(name)
		}
	}
	return maxName, rows, err
}

func NewOptionModel(optionType OptionType, config *config.Config, context, filename string) *OptionModel {
	maxName, rows, err := getRowData(optionType, context, filename)
	maxName = max(maxName, defaultWidth)
	n := OptionModel{
		cols: []table.Column{
			table.NewColumn(columnKeyName, "", maxName).
				WithFiltered(true),
		},

		config:      config,
		context:     context,
		error:       err,
		filename:    filename,
		filterInput: textinput.New(),
		optionType:  optionType,
		rows:        rows,
		styles: optionStyles{
			overlay: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(lipgloss.Color(config.Style.DialogBorderColor)).
				Padding(0, 1),
			filter: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(lipgloss.Color(config.Style.FilterBorder)).
				Width(maxName),
			table: lipgloss.NewStyle().
				Align(lipgloss.Left).
				BorderForeground(lipgloss.Color(config.Style.BorderFgColor)).
				Foreground(lipgloss.Color(config.Style.ContextListNormalTitle)).
				Margin(1).
				Padding(0, 2),
		},
	}

	n.filterInput.TextStyle = n.filterInput.TextStyle.UnsetMargins()

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

func (n *OptionModel) GetOptionType() OptionType {
	return n.optionType
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
	var titleContent string
	switch n.optionType {
	case Namespace:
		titleContent = "Namespaces"
	case Session:
		titleContent = "Sessions"
	case ClusterLogin:
		titleContent = "Clusters"
	}
	title := lipgloss.NewStyle().Padding(0, 2).
		Border(lipgloss.RoundedBorder(), false, false, true, false).
		Foreground(lipgloss.Color(n.config.Style.Title)).Align(lipgloss.Center).
		Render(titleContent)

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

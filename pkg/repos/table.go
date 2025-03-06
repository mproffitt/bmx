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

package repos

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/dialog"
	"github.com/mproffitt/bmx/pkg/helpers"
)

var customBorder = table.Border{
	Top:    "",
	Left:   "",
	Right:  "",
	Bottom: "",

	TopRight:    "",
	TopLeft:     "",
	BottomRight: "",
	BottomLeft:  "",

	TopJunction:    "",
	LeftJunction:   "",
	RightJunction:  "",
	BottomJunction: "",

	InnerJunction: "",
	InnerDivider:  "",
}

const (
	columnKeyName  = "name"
	columnKeyOwner = "owner"
	columnKeyUrl   = "url"
	columnKeyPath  = "path"

	maxWidth            = 20
	minHeight           = 10
	fixedVerticalMargin = 4
	pattern             = ".git/config"
)

type Model struct {
	sync.Mutex

	config      *config.Config
	callback    func(table.RowData, string, bool) tea.Cmd
	dialog      tea.Model
	filterInput textinput.Model
	height      int
	isOverlay   bool
	keymap      keyMap
	paths       []string
	rows        []table.Row
	spinner     *spinner.Model
	styles      styles
	table       table.Model
	viewport    viewport.Model
	width       int
}

type styles struct {
	table    lipgloss.Style
	spinner  lipgloss.Style
	title    lipgloss.Style
	text     lipgloss.Style
	viewport lipgloss.Style
	filter   lipgloss.Style
}

func New(config *config.Config, callback func(table.RowData, string, bool) tea.Cmd) *Model {
	spinner := spinner.New(spinner.WithSpinner(spinner.Meter))
	model := &Model{
		config:      config,
		callback:    callback,
		filterInput: textinput.New(),
		keymap:      mapKeys(),
		paths:       config.Paths,
		spinner:     &spinner,
		styles: styles{
			table: lipgloss.NewStyle().
				BorderForeground(lipgloss.Color(config.Style.BorderFgColor)),
			spinner: lipgloss.NewStyle().
				Foreground(lipgloss.Color(config.Style.Foreground)),
			text: lipgloss.NewStyle().
				Foreground(lipgloss.Color(config.Style.Foreground)),
			viewport: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(config.Style.BorderFgColor)),
			title: lipgloss.NewStyle().Padding(0, 0, 0, 1).
				Foreground(lipgloss.Color(config.Style.Title)).Align(lipgloss.Center),
			filter: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(lipgloss.Color(config.Style.BorderFgColor)),
		},
	}

	model.spinner.Style = model.styles.spinner

	// load repo paths in the background
	go model.loadData(model.paths)
	return model
}

func (m *Model) GetSize() (int, int) {
	return m.width, m.height
}

func (m *Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *Model) Overlay() helpers.UseOverlay {
	m.isOverlay = true
	return m
}

func (m *Model) HasActiveDialog() bool {
	return m.dialog != nil
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:

		// Dialog needs to be handled first as it's an overlay
		// to the main window and takes precedence over all
		// other elements
		if m.dialog != nil {
			m.dialog, cmd = m.dialog.Update(msg)
			// cmd, err = m.handleDialog(msg)
			return m, cmd
		}
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)

		switch {
		case key.Matches(msg, m.keymap.Quit):
			if m.dialog != nil {
				break
			}
			if m.isOverlay {
				return m, nil
			}
			cmds = append(cmds, tea.Quit)
		case key.Matches(msg, m.keymap.Enter):
			current := m.table.HighlightedRow()
			filter := m.table.GetCurrentFilter()
			return m, m.callback(current.Data, filter, m.config.CreateSessionKubeConfig)
		case key.Matches(msg, m.keymap.Help):
			m.displayHelp()
		case key.Matches(msg, m.keymap.Pagedown, m.keymap.Pageup):
			break
		case key.Matches(msg, m.keymap.Down, m.keymap.Up):
			break
		default:
			m.filterInput.Focus()
			m.filterInput, _ = m.filterInput.Update(msg)
			m.filterInput.Blur()
			m.table = m.table.WithFilterInput(m.filterInput)
		}
	case dialog.DialogStatusMsg:
		if msg.Done {
			m.dialog = nil
		}
	case spinner.TickMsg:
		if m.spinner != nil {
			*m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
			m.drawTable()
		}
	case tea.WindowSizeMsg:
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
		m.width = msg.Width
		m.height = msg.Height
	default:
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.spinner != nil {
		spinner := fmt.Sprintf("%s%s%s\n\t- %s\n", m.spinner.View(),
			" ", m.styles.text.Render("Loading..."), strings.Join(m.paths, "\n\t- "))
		m.viewport = viewport.New(m.width, m.height)
		m.viewport.SetContent(lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, spinner))
		return m.styles.viewport.Render(m.viewport.View())
	}

	body := strings.Builder{}
	if m.isOverlay {
		title := m.styles.title.Render("Create new session\n")
		body.WriteString(title)
	}

	subtract := 6
	if m.isOverlay {
		subtract = 8
	}
	tviewport := viewport.New(m.width-4, m.height-subtract)
	tviewport.SetContent(m.table.View())
	content := m.styles.viewport.Padding(0, 0, 1, 2).Render(tviewport.View())
	body.WriteString(content + "\n")

	filter := m.styles.filter.Width(m.width - 2).Render(m.filterInput.View())
	body.WriteString(filter)

	doc := m.styles.table.Render(body.String())
	if m.dialog != nil {
		dw, _ := m.dialog.(*dialog.Dialog).GetSize()
		w := m.width/2 - max(dw, config.DialogWidth)/2
		doc = helpers.PlaceOverlay(w, 10, m.dialog.View(),
			doc, false)
	}

	if m.isOverlay {
		m.viewport = viewport.New(m.width, m.height)
		m.viewport.SetContent(doc)
		return m.styles.viewport.Render(m.viewport.View())
	}

	return doc
}

func (m *Model) loadData(paths []string) {
	m.rows = make([]table.Row, 0)
	repositories, err := Find(paths, pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "err : %+v\n", err)
	}

	m.Lock()
	for _, repo := range repositories {
		m.rows = append(m.rows,
			table.NewRow(table.RowData{
				columnKeyName:  repo.Name,
				columnKeyOwner: repo.Owner,
				columnKeyUrl:   repo.Url,
				columnKeyPath:  repo.Path,
			}),
		)
	}
	m.Unlock()
}

func (m *Model) drawTable() {
	if len(m.rows) == 0 {
		return
	}

	subtract := 7
	if m.isOverlay {
		subtract = 9
	}
	maxName, maxOwner, maxUrl := 0, 0, 0
	for _, row := range m.rows {
		nameLen := len(row.Data[columnKeyName].(string))
		if nameLen > maxName {
			maxName = nameLen
		}

		ownerLen := len(row.Data[columnKeyOwner].(string))
		if ownerLen > maxOwner {
			maxOwner = ownerLen
			maxOwner = min(maxOwner, maxWidth)
		}
		// w := m.styles.table.GetHorizontalFrameSize()
		maxUrl = m.width - (maxName + maxOwner) - subtract
	}

	cols := []table.Column{
		table.NewColumn(columnKeyName, "Name", maxName).WithFiltered(true),
		table.NewColumn(columnKeyOwner, "Owner", maxOwner),
		table.NewColumn(columnKeyUrl, "Url", maxUrl).WithFiltered(true),
	}

	pageSize := max(20, m.height-subtract)
	m.table = table.New(cols).
		Border(customBorder).
		Filtered(true).
		Focused(true).
		WithBaseStyle(lipgloss.NewStyle().
			BorderForeground(lipgloss.Color(m.config.Style.BorderFgColor)).
			Foreground(lipgloss.Color(m.config.Style.Foreground)).
			Align(lipgloss.Left),
		).
		HighlightStyle(lipgloss.NewStyle().
			Background(lipgloss.Color(m.config.Style.BorderFgColor)).
			Foreground(lipgloss.Color(m.config.Style.ListNormalSelectedTitle)),
		).
		WithFooterVisibility(false).
		WithHeaderVisibility(false).
		WithPageSize(pageSize).
		WithRows(m.rows).
		SortByAsc(columnKeyOwner).ThenSortByAsc(columnKeyName)

	m.spinner = nil
}

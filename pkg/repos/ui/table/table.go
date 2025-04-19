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

package table

import (
	"fmt"
	"os"
	"sync"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evertras/bubble-table/table"
	"github.com/mproffitt/bmx/pkg/components/createpanel"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/repos"
	"github.com/mproffitt/bmx/pkg/theme"
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

type InputFocus int

const (
	Filter InputFocus = iota
	Path
	Command
	Button
)

type Model struct {
	sync.Mutex

	callback  func(map[string]any, bool) tea.Cmd
	columns   []table.Column
	config    *config.Config
	current   *textinput.Model
	dialog    tea.Model
	inputs    inputs
	focus     InputFocus
	height    int
	isOverlay bool
	keymap    keyMap
	panel     *createpanel.Model
	paths     []string
	rows      []table.Row
	spinner   *spinner.Model
	styles    styles
	table     table.Model
	viewport  viewport.Model
	width     int
}

type inputs struct {
	command textinput.Model
	filter  textinput.Model
	path    textinput.Model
}

type styles struct {
	activeButton lipgloss.Style
	button       lipgloss.Style
	filter       lipgloss.Style
	spinner      lipgloss.Style
	table        lipgloss.Style
	text         lipgloss.Style
	title        lipgloss.Style
	viewport     lipgloss.Style
}

func New(config *config.Config, callback func(map[string]any, bool) tea.Cmd) *Model {
	spinner := spinner.New(spinner.WithSpinner(spinner.Meter))
	model := &Model{
		config:   config,
		callback: callback,
		inputs: inputs{
			filter:  textinput.New(),
			path:    textinput.New(),
			command: textinput.New(),
		},
		focus:   Filter,
		keymap:  mapKeys(),
		paths:   config.Paths,
		spinner: &spinner,
		styles: styles{
			table: lipgloss.NewStyle().
				BorderForeground(theme.Colours.Black),
			spinner: lipgloss.NewStyle().
				Foreground(theme.Colours.Purple),
			text: lipgloss.NewStyle().
				Foreground(theme.Colours.BrightPurple),
			viewport: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.Colours.Black),
			title: lipgloss.NewStyle().Padding(0, 0, 0, 1).
				Foreground(theme.Colours.Yellow).Align(lipgloss.Center),
			filter: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(theme.Colours.Black),
			button: lipgloss.NewStyle().
				Foreground(theme.Colours.Bg).
				Background(theme.Colours.Fg).
				Border(lipgloss.RoundedBorder(), true).
				BorderBackground(theme.Colours.Bg).
				BorderForeground(theme.Colours.Black).
				Padding(0, 4).
				MarginLeft(2),

			activeButton: lipgloss.NewStyle().
				Foreground(theme.Colours.BrightWhite).
				Background(theme.Colours.BrightRed).
				Border(lipgloss.RoundedBorder(), true).
				BorderBackground(theme.Colours.Bg).
				BorderForeground(theme.Colours.Fg).
				Padding(0, 4).
				MarginLeft(2).
				Underline(true),
		},
	}
	model.panel = createpanel.New().
		WithObserver(model)
	model.inputs.filter.ShowSuggestions = true
	model.inputs.filter.KeyMap = model.getInputKeyMap()
	model.inputs.path.ShowSuggestions = true
	model.inputs.path.KeyMap = model.getInputKeyMap()

	model.inputs.command.ShowSuggestions = true
	model.inputs.command.KeyMap.AcceptSuggestion = key.NewBinding(key.WithKeys("right"))
	model.inputs.command.KeyMap.NextSuggestion = key.NewBinding(key.WithKeys("down"))
	model.inputs.command.KeyMap.PrevSuggestion = key.NewBinding(key.WithKeys("up"))

	model.current = &model.inputs.filter
	model.current.Focus()

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

func (m *Model) getSuggestions(msg *createpanel.SuggestionsMsg) {
	selected := m.table.HighlightedRow().Data
	values := map[string]any(selected)

	(*msg).Name = make([]string, 0)
	(*msg).Path = make([]string, 0)

	var (
		filter, path string
		ok           bool
	)

	if filter, ok = values[columnKeyName].(string); ok {
		(*msg).Name = []string{filter}
	}

	if path, ok = values[columnKeyPath].(string); ok {
		(*msg).Path = []string{path}
	}
}

func (m *Model) loadData(paths []string) {
	m.rows = make([]table.Row, 0)
	repositories, err := repos.Find(paths, pattern)
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

	subtract := 9
	if m.isOverlay {
		subtract = 12
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

	m.columns = []table.Column{
		table.NewColumn(columnKeyName, "Name", maxName).WithFiltered(true),
		table.NewColumn(columnKeyOwner, "Owner", maxOwner),
		table.NewColumn(columnKeyUrl, "Url", maxUrl),
	}

	pageSize := max(19, m.height-subtract)
	m.table = table.New(m.columns).
		Border(customBorder).
		Filtered(true).
		Focused(true).
		WithBaseStyle(lipgloss.NewStyle().
			BorderForeground(theme.Colours.Black).
			Foreground(theme.Colours.BrightPurple).
			Align(lipgloss.Left),
		).
		HighlightStyle(lipgloss.NewStyle().
			Background(theme.Colours.SelectionBg).
			Foreground(theme.Colours.BrightBlue),
		).
		WithFooterVisibility(false).
		WithHeaderVisibility(false).
		WithPageSize(pageSize).
		WithRows(m.rows).
		SortByAsc(columnKeyOwner).ThenSortByAsc(columnKeyName)

	m.spinner = nil
}

func (m *Model) getInputKeyMap() textinput.KeyMap {
	return textinput.KeyMap{
		CharacterForward:        key.NewBinding(key.WithKeys("right", "ctrl+f")),
		CharacterBackward:       key.NewBinding(key.WithKeys("left", "ctrl+b")),
		WordForward:             key.NewBinding(key.WithKeys("alt+right", "ctrl+right", "alt+f")),
		WordBackward:            key.NewBinding(key.WithKeys("alt+left", "ctrl+left", "alt+b")),
		DeleteWordBackward:      key.NewBinding(key.WithKeys("alt+backspace", "ctrl+w")),
		DeleteWordForward:       key.NewBinding(key.WithKeys("alt+delete", "alt+d")),
		DeleteAfterCursor:       key.NewBinding(key.WithKeys("ctrl+k")),
		DeleteBeforeCursor:      key.NewBinding(key.WithKeys("ctrl+u")),
		DeleteCharacterBackward: key.NewBinding(key.WithKeys("backspace", "ctrl+h")),
		DeleteCharacterForward:  key.NewBinding(key.WithKeys("delete", "ctrl+d")),
		LineStart:               key.NewBinding(key.WithKeys("home", "ctrl+a")),
		LineEnd:                 key.NewBinding(key.WithKeys("end", "ctrl+e")),
		Paste:                   key.NewBinding(key.WithKeys("ctrl+v")),
		AcceptSuggestion:        key.NewBinding(key.WithKeys("right")),
	}
}

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

package panel

import (
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/components/dialog"
	"github.com/mproffitt/bmx/pkg/components/overlay"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/kubernetes"
	"github.com/muesli/reflow/truncate"
)

const (
	KubernetesListWidth = 18
	KubernetesRowHeight = 3
	PanelTitle          = 2
	PanelFooter         = 4
)

type Model struct {
	activeItem int
	activeList int
	cols       int
	config     *config.Config
	context    string
	focused    bool
	force      bool
	height     int
	items      []kubernetes.KubeContext
	keymap     *keyMap
	kubeconfig string
	lists      []list.Model
	listWidth  int
	options    tea.Model
	optionType OptionType
	paginator  *paginator.Model
	rows       int
	session    string
	styles     contextStyles
	todelete   string
	tomove     string
	viewport   viewport.Model
	width      int
}

type contextStyles struct {
	delegates       delegates
	list            lipgloss.Style
	viewportFocused lipgloss.Style
	viewportNormal  lipgloss.Style
}

type delegates struct {
	active ItemDelegate
	base   ItemDelegate
	shaded ItemDelegate
}

func NewKubectxPane(c *config.Config, session string, rows, cols, columnWidth int) *Model {
	m := Model{
		activeItem: 0,
		activeList: 0,
		cols:       cols,
		config:     c,
		keymap:     mapKeys(),
		listWidth:  min(columnWidth, KubernetesListWidth),
		paginator: &paginator.Model{
			ActiveDot:    lipgloss.NewStyle().Foreground(c.Colours().BrightWhite).Render("•"),
			ArabicFormat: "%d/%d",
			InactiveDot:  lipgloss.NewStyle().Foreground(c.Colours().BrightBlack).Render("•"),
			KeyMap:       paginator.DefaultKeyMap,
			Page:         0,
			PerPage:      cols,
			Type:         paginator.Dots,
		},
		rows: rows,
		styles: contextStyles{
			delegates: delegates{},
			list:      lipgloss.NewStyle().Margin(1, 0, 0, 0).Width(columnWidth),
			viewportNormal: lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(c.Colours().Black).
				AlignHorizontal(lipgloss.Center).
				AlignVertical(lipgloss.Center).
				PaddingLeft(2).
				PaddingRight(2),
			viewportFocused: lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(c.Colours().Blue).
				AlignHorizontal(lipgloss.Center).
				AlignVertical(lipgloss.Center).
				PaddingLeft(2).
				PaddingRight(2),
		},
		viewport: viewport.New(0, 0),
	}

	m.styles.delegates.base = m.createBaseDelegate()
	m.styles.delegates.active = m.createActiveDelegate()
	m.styles.delegates.shaded = m.createShadedDelegate()
	return &m
}

func (m *Model) Blur() tea.Model {
	m.focused = false
	return m
}

func (m *Model) Focus() tea.Model {
	m.focused = true
	return m
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Overlay() helpers.UseOverlay {
	if m.options != nil {
		return m.options.(helpers.UseOverlay).Overlay()
	}

	if m.todelete != "" {
		if m.force {
			m.force = false
			return nil
		}
		builder := strings.Builder{}
		builder.WriteString("Are you sure you want to delete context\n")
		builder.WriteString(lipgloss.PlaceHorizontal(config.DialogWidth, lipgloss.Center,
			lipgloss.NewStyle().
				Bold(true).
				Foreground(m.config.Colours().BrightBlue).
				Padding(1).
				Render(m.todelete)))
		builder.WriteString("\ndeleting means you will no longer be logged in to this cluster")
		dialog := dialog.NewConfirmDialog(builder.String(), m.config, config.DialogWidth)
		return dialog.(helpers.UseOverlay)
	}

	return nil
}

func (m *Model) RequiresOverlay() bool {
	return m.options != nil || (m.todelete != "" && !m.force)
}

func (m *Model) GetSize() (int, int) {
	return m.width, m.height
}

func (m *Model) UpdateContextList(session, kubeconfig string) tea.Model {
	if session == m.session {
		return m
	}
	m.session = session
	m.kubeconfig = kubeconfig
	m.reloadContextList()

	m.activeItem = 0
	m.activeList = 0
	m.paginator.Page = 0
	m.lists = m.createKubeLists()
	pages := float64(len(m.items)) / float64(m.rows*m.cols)
	m.paginator.TotalPages = max(1, int(math.Ceil(pages)))
	if m.paginator.TotalPages > 1 {
		m.setActiveContextPage()
	}

	return m
}

func (m *Model) setActiveContextPage() {
	page := m.paginator.Page
	for i := range m.paginator.TotalPages {
		m.paginator.Page = i
		start, end := m.paginator.GetSliceBounds(len(m.lists))
		for j := start; j < end; j++ {
			items := m.lists[j].Items()
			for l := range len(items) {
				if items[l].(kubernetes.KubeContext).IsCurrentContext {
					m.activeItem = l
					m.activeList = j
					m.lists[m.activeList].Select((m.activeItem))
					return
				}
			}
		}
	}
	m.paginator.Page = page
}

func (m *Model) SetSize(width, height, columnWidth int) tea.Model {
	m.width = width
	m.height = height
	m.listWidth = columnWidth
	m.viewport.Height = height
	m.viewport.Width = width

	if m.rows*KubernetesRowHeight > m.height {
		// 4 = 2 lines for title, 2 lines for pagination
		m.rows = (m.height / KubernetesRowHeight) - 4
	}

	if (m.cols * m.listWidth) > m.width {
		m.cols = m.width / m.listWidth
		m.paginator.PerPage = m.cols
	}
	m.reloadContextList()
	return m
}

func (m *Model) View() string {
	cols := m.createPaginatedColumns()

	var title string
	{
		titlestring := "Kubernetes Contexts : " + m.kubeconfig
		titlestring = truncate.String(titlestring, uint(m.width))
		title = lipgloss.NewStyle().
			Foreground(m.config.Colours().Yellow).Align(lipgloss.Left).
			Render(titlestring)
	}

	paginated := lipgloss.NewStyle().Foreground(m.config.Colours().Blue).
		Padding(2).
		Render("No active contexts")
	paginated = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, paginated)

	if len(cols) > 0 {
		paginated = lipgloss.JoinHorizontal(lipgloss.Top, cols...)
		paginated = lipgloss.JoinVertical(lipgloss.Center, paginated, m.paginator.View())
	}

	m.viewport.SetContent(paginated)
	style := m.styles.viewportNormal
	if m.focused {
		style = m.styles.viewportFocused
	}
	doc := style.Render(m.viewport.View())
	return overlay.PlaceOverlay(2, 0, title, doc, false)
}

func (m *Model) createActiveDelegate() ItemDelegate {
	delegate := m.styles.delegates.base
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(m.config.Colours().BrightBlue)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(m.config.Colours().BrightWhite)
	return delegate
}

func (m *Model) createBaseDelegate() ItemDelegate {
	delegate := NewItemDelegate()
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(m.config.Colours().Blue)
	delegate.Styles.NormalDesc = delegate.Styles.NormalTitle.
		Foreground(m.config.Colours().BrightBlack)

	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		UnsetBorderLeft().
		PaddingLeft(2).
		Foreground(m.config.Colours().Blue)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		UnsetBorderLeft().
		PaddingLeft(2).
		Foreground(m.config.Colours().BrightBlack)
	return delegate
}

func (m *Model) createKubeLists() []list.Model {
	var (
		models  = make([]list.Model, 0)
		current int
	)

	for i := 0; i < len(m.items); i += m.rows {
		items := make([]list.Item, 0)
		for j := i; j < min(i+m.rows, len(m.items)); j++ {
			items = append(items, m.items[j])
		}

		l := list.New(items, m.styles.delegates.base, m.listWidth, (m.rows * KubernetesRowHeight))
		l.SetShowTitle(false)
		l.SetShowHelp(false)
		l.SetShowPagination(false)
		l.SetShowFilter(false)
		l.SetShowStatusBar(false)
		l.SetFilteringEnabled(false)

		models = append(models, l)
		current++
	}
	return models
}

func (m *Model) createPaginatedColumns() []string {
	var cols []string
	{
		start, end := m.paginator.GetSliceBounds(len(m.lists))
		if m.activeList < start {
			m.paginator.PrevPage()
		} else if m.activeList >= end {
			m.paginator.NextPage()
		}
		start, end = m.paginator.GetSliceBounds(len(m.lists))

		for i, item := range m.lists[start:end] {
			item.SetDelegate(m.styles.delegates.base)
			if m.focused && (i+start) == m.activeList {
				item.SetDelegate(m.styles.delegates.active)
			}
			cols = append(cols, m.styles.list.Render(item.View()))
		}
	}
	return cols
}

func (m *Model) createShadedDelegate() ItemDelegate {
	delegate := NewItemDelegate()
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(m.config.Colours().Black)

	delegate.Styles.NormalDesc = delegate.Styles.NormalTitle.
		Foreground(m.config.Colours().Black)
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(m.config.Colours().Black)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(m.config.Colours().Black)
	return delegate
}

func (m *Model) reloadContextList() {
	contexts, err := kubernetes.KubeContextList(
		m.config.ManageSessionKubeContext, m.kubeconfig)
	if err != nil {
		contexts = make([]kubernetes.KubeContext, 0)
	}

	m.items = contexts
	m.lists = m.createKubeLists()
	if len(m.lists) > 0 {
		if m.activeList >= len(m.lists) {
			m.activeList = len(m.lists) - 1
		}
		if m.activeItem >= len(m.lists[m.activeList].Items()) {
			m.activeItem = len(m.lists[m.activeList].Items()) - 1
		}
		m.lists[m.activeList].Select(m.activeItem)
	}
}

func (m *Model) switchContext() error {
	context := m.lists[m.activeList].SelectedItem().(kubernetes.KubeContext)
	contextName := context.Name
	err := kubernetes.SetCurrentContext(contextName, m.kubeconfig)
	m.reloadContextList()
	return err
}

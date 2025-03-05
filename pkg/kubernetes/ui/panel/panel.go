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
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/dialog"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/kubernetes"
	"github.com/mproffitt/bmx/pkg/optionlist"
	"github.com/mproffitt/bmx/pkg/tmux"
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
	error      error
	focused    bool
	height     int
	items      []kubernetes.KubeContext
	keymap     *keyMap
	kubeconfig string
	lists      []list.Model
	listWidth  int
	options    tea.Model
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
	k := Model{
		activeItem: 0,
		activeList: 0,
		cols:       cols,
		config:     c,
		keymap:     mapKeys(),
		listWidth:  min(columnWidth, KubernetesListWidth),
		paginator: &paginator.Model{
			ActiveDot:    lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•"),
			ArabicFormat: "%d/%d",
			InactiveDot:  lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•"),
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
				BorderForeground(lipgloss.Color(c.Style.BorderFgColor)).
				AlignHorizontal(lipgloss.Center).
				AlignVertical(lipgloss.Center).
				PaddingLeft(2).
				PaddingRight(2),
			viewportFocused: lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(c.Style.FocusedColor)).
				AlignHorizontal(lipgloss.Center).
				AlignVertical(lipgloss.Center).
				PaddingLeft(2).
				PaddingRight(2),
		},
		viewport: viewport.New(0, 0),
	}

	k.styles.delegates.base = k.createBaseDelegate()
	k.styles.delegates.active = k.createActiveDelegate()
	k.styles.delegates.shaded = k.createShadedDelegate()
	return &k
}

func (k *Model) Blur() tea.Model {
	k.focused = false
	return k
}

func (k *Model) Focus() tea.Model {
	k.focused = true
	return k
}

func (k *Model) GetError() error {
	return k.error
}

func (k *Model) Init() tea.Cmd {
	return nil
}

func (k *Model) Overlay() helpers.UseOverlay {
	if k.options != nil {
		return k.options.(helpers.UseOverlay).Overlay()
	}

	if k.todelete != "" {
		builder := strings.Builder{}
		builder.WriteString("Are you sure you want to delete context\n")
		builder.WriteString(lipgloss.PlaceHorizontal(config.DialogWidth, lipgloss.Center,
			lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(k.config.Style.FocusedColor)).
				Padding(1).
				Render(k.todelete)))
		builder.WriteString("\ndeleting means you will no longer be logged in to this cluster")
		dialog := dialog.NewConfirmDialog(builder.String(), k.config, config.DialogWidth)
		return dialog.(helpers.UseOverlay)
	}

	return nil
}

func (k *Model) RequiresOverlay() bool {
	return k.options != nil || k.todelete != ""
}

func (k *Model) GetSize() (int, int) {
	return k.width, k.height
}

func (k *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, k.keymap.Killpanel):
			if k.options != nil {
				k.options = nil
			}
			k.context = ""
			k.tomove = ""
		case key.Matches(msg, k.keymap.Left):
			k.activeItem = k.lists[k.activeList].Cursor()
			k.activeList = (k.activeList - 1)
			if k.activeList < 0 {
				k.activeList = len(k.lists) - 1
				k.paginator.Page = k.paginator.TotalPages - 1
			}
			if k.activeItem > len(k.lists[k.activeList].Items())-1 {
				k.activeItem = len(k.lists[k.activeList].Items()) - 1
			}
		case key.Matches(msg, k.keymap.Right):
			k.activeItem = k.lists[k.activeList].Cursor()
			k.activeList = (k.activeList + 1)
			if k.activeList > len(k.lists)-1 {
				k.activeList = 0
				k.paginator.Page = 0
			}

			if k.activeItem >= len(k.lists[k.activeList].Items())-1 {
				k.activeItem = len(k.lists[k.activeList].Items()) - 1
			}
		case key.Matches(msg, k.keymap.Up):
			if k.activeItem == 0 && k.activeList == 0 {
				k.activeList = len(k.lists) - 1
				k.paginator.Page = k.paginator.TotalPages - 1
				k.activeItem = len(k.lists[k.activeList].Items()) - 1
			} else if k.activeItem == 0 {
				k.activeList -= 1
				k.activeItem = len(k.lists[k.activeList].Items()) - 1
			} else {
				k.activeItem -= 1
			}
		case key.Matches(msg, k.keymap.Down):
			if k.activeList == len(k.lists)-1 && k.activeItem == len(k.lists[k.activeList].Items())-1 {
				k.activeList = 0
				k.activeItem = 0
				k.paginator.Page = 0
			} else if k.activeItem == len(k.lists[k.activeList].Items())-1 {
				k.activeList += 1
				k.activeItem = 0
			} else {
				k.activeItem += 1
			}

			// PAGE Up and Page down
		case key.Matches(msg, k.keymap.Pageup):
			k.paginator.PrevPage()
			k.activeList, _ = k.paginator.GetSliceBounds(len(k.lists))
			k.activeItem = 0
		case key.Matches(msg, k.keymap.Pagedown):
			k.paginator.NextPage()
			k.activeList, _ = k.paginator.GetSliceBounds(len(k.lists))
			k.activeItem = 0

			// END

		case key.Matches(msg, k.keymap.Delete):
			k.todelete = k.lists[k.activeList].SelectedItem().(list.DefaultItem).Title()
		case key.Matches(msg, k.keymap.Space):
			k.context = k.lists[k.activeList].SelectedItem().(list.DefaultItem).Title()
			k.options = optionlist.NewOptionModel(optionlist.Namespace, k.config, k.context, k.kubeconfig)
		case key.Matches(msg, k.keymap.Move):
			if k.options != nil {
				break
			}
			k.tomove = k.lists[k.activeList].SelectedItem().(list.DefaultItem).Title()
			k.options = optionlist.NewOptionModel(optionlist.Session, k.config, k.context, k.kubeconfig)
		case key.Matches(msg, k.keymap.Login):
			k.options = optionlist.NewOptionModel(optionlist.ClusterLogin, k.config, k.context, k.kubeconfig)
		}

		if len(k.lists) > 0 {
			k.lists[k.activeList].Select((k.activeItem))
		}

	case kubernetes.ContextChangeMsg:
		k.error = k.switchContext()
		k.lists[k.activeList].Select((k.activeItem))
	case kubernetes.ContextDeleteMsg:
		if k.todelete != "" {
			k.error = kubernetes.DeleteContext(k.todelete, k.kubeconfig)
			k.todelete = ""
			k.reloadContextList()
		}
	case helpers.OverlayMsg:
		switch value := msg.Message.(type) {
		case string:
			optionType := k.options.(*optionlist.OptionModel).GetOptionType()
			k.options = nil
			switch optionType {
			case optionlist.Session:
				// Get filename from session
				newconfig, err := kubernetes.CreateConfig(value)
				if err != nil {
					k.error = err
					break
				}
				if !tmux.HasSession(value) {
					home, _ := os.UserHomeDir()
					err := tmux.CreateSession(value, home, "", true, false)
					k.error = err
					cmds = append(cmds, helpers.ReloadSessionsCmd())
				}
				k.error = kubernetes.MoveContext(k.tomove, k.kubeconfig, newconfig)
				k.reloadContextList()

			case optionlist.Namespace:
				k.error = kubernetes.SetNamespace(k.context, value, k.kubeconfig)
				k.context = ""
				k.reloadContextList()

			case optionlist.ClusterLogin:
				k.error = kubernetes.TeleportClusterLogin(value)
				k.reloadContextList()
			}
		case dialog.Status:
			switch value {
			case dialog.Confirm:
				if k.todelete != "" {
					cmd = kubernetes.ContextDeleteCmd()
					cmds = append(cmds, cmd)
				}
			case dialog.Cancel:
				if k.todelete != "" {
					k.todelete = ""
				}
			}
		}
	}
	return k, tea.Batch(cmds...)
}

func (k *Model) UpdateContextList(session, kubeconfig string) tea.Model {
	if session == k.session {
		return k
	}
	k.session = session
	k.kubeconfig = kubeconfig
	k.reloadContextList()

	k.activeItem = 0
	k.activeList = 0
	k.paginator.Page = 0
	k.lists = k.createKubeLists()
	pages := float64(len(k.items)) / float64(k.rows*k.cols)
	k.paginator.TotalPages = max(1, int(math.Ceil(pages)))
	if k.paginator.TotalPages > 1 {
		k.setActiveContextPage()
	}

	return k
}

func (k *Model) setActiveContextPage() {
	page := k.paginator.Page
	for i := 0; i < k.paginator.TotalPages; i++ {
		k.paginator.Page = i
		start, end := k.paginator.GetSliceBounds(len(k.lists))
		for j := start; j < end; j++ {
			items := k.lists[j].Items()
			for l := 0; l < len(items); l++ {
				if items[l].(kubernetes.KubeContext).IsCurrentContext {
					k.activeItem = l
					k.activeList = j
					k.lists[k.activeList].Select((k.activeItem))
					return
				}
			}
		}
	}
	k.paginator.Page = page
}

func (k *Model) SetSize(width, height, columnWidth int) tea.Model {
	k.width = width
	k.height = height
	k.listWidth = columnWidth
	k.viewport.Height = height
	k.viewport.Width = width

	if k.rows*KubernetesRowHeight > k.height {
		// 4 = 2 lines for title, 2 lines for pagination
		k.rows = (k.height / KubernetesRowHeight) - 4
	}

	if (k.cols * k.listWidth) > k.width {
		k.cols = k.width / k.listWidth
		k.paginator.PerPage = k.cols
	}
	k.reloadContextList()
	return k
}

func (k *Model) View() string {
	cols := k.createPaginatedColumns()
	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(k.config.Style.Title)).Align(lipgloss.Left).
		Render("Kubernetes Contexts : " + k.kubeconfig)

	nocontexts := lipgloss.NewStyle().Foreground(lipgloss.Color(k.config.Style.FocusedColor)).
		Padding(2).
		Render("No active contexts")

	paginated := lipgloss.JoinVertical(lipgloss.Center, title, nocontexts)
	if len(cols) > 0 {
		pageContents := lipgloss.JoinHorizontal(lipgloss.Top, cols...)
		paginated = lipgloss.JoinVertical(
			lipgloss.Left, title,
			lipgloss.JoinVertical(lipgloss.Center, pageContents, k.paginator.View()))
	}
	k.viewport.SetContent(paginated)
	style := k.styles.viewportNormal
	if k.focused {
		style = k.styles.viewportFocused
	}
	return style.Render(k.viewport.View())
}

func (k *Model) createActiveDelegate() ItemDelegate {
	delegate := k.styles.delegates.base
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color(k.config.Style.ContextListActiveTitle))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color(k.config.Style.ContextListActiveDescription))
	return delegate
}

func (k *Model) createBaseDelegate() ItemDelegate {
	delegate := NewItemDelegate()
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color(k.config.Style.ContextListNormalTitle))

	delegate.Styles.NormalDesc = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color(k.config.Style.ContextListNormalDescription))
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		UnsetBorderLeft().
		PaddingLeft(2).
		Foreground(lipgloss.Color(k.config.Style.ContextListNormalTitle))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		UnsetBorderLeft().
		PaddingLeft(2).
		Foreground(lipgloss.Color(k.config.Style.ContextListNormalDescription))
	return delegate
}

func (k *Model) createKubeLists() []list.Model {
	var (
		models  = make([]list.Model, 0)
		current int
	)

	for i := 0; i < len(k.items); i += k.rows {
		items := make([]list.Item, 0)
		for j := i; j < min(i+k.rows, len(k.items)); j++ {
			items = append(items, k.items[j])
		}

		l := list.New(items, k.styles.delegates.base, k.listWidth, (k.rows * KubernetesRowHeight))
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

func (k *Model) createPaginatedColumns() []string {
	var cols []string
	{
		start, end := k.paginator.GetSliceBounds(len(k.lists))
		if k.activeList < start {
			k.paginator.PrevPage()
		} else if k.activeList >= end {
			k.paginator.NextPage()
		}
		start, end = k.paginator.GetSliceBounds(len(k.lists))

		for i, item := range k.lists[start:end] {
			item.SetDelegate(k.styles.delegates.base)
			if k.focused && (i+start) == k.activeList {
				item.SetDelegate(k.styles.delegates.active)
			}
			cols = append(cols, k.styles.list.Render(item.View()))
		}
	}
	return cols
}

func (k *Model) createShadedDelegate() ItemDelegate {
	delegate := NewItemDelegate()
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color(k.config.Style.ListShadedTitle))

	delegate.Styles.NormalDesc = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color(k.config.Style.ListShadedDescription))
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color(k.config.Style.ListShadedSelectedTitle))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color(k.config.Style.ListShadedSelectedDescription))
	return delegate
}

func (k *Model) reloadContextList() {
	contexts, err := kubernetes.KubeContextList(
		k.config.ManageSessionKubeContext, k.kubeconfig)
	if err != nil {
		contexts = make([]kubernetes.KubeContext, 0)
	}

	k.items = contexts
	k.lists = k.createKubeLists()
	if len(k.lists) > 0 {
		if k.activeList >= len(k.lists) {
			k.activeList = len(k.lists) - 1
		}
		if k.activeItem >= len(k.lists[k.activeList].Items()) {
			k.activeItem = len(k.lists[k.activeItem].Items()) - 1
		}
	}
}

func (k *Model) switchContext() error {
	context := k.lists[k.activeList].SelectedItem().(kubernetes.KubeContext)
	contextName := context.Name
	err := kubernetes.SetCurrentContext(contextName, k.kubeconfig)
	k.reloadContextList()
	return err
}

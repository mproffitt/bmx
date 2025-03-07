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

package session

import (
	"math"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/dialog"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/kubernetes"
	"github.com/mproffitt/bmx/pkg/kubernetes/ui/panel"
	"github.com/mproffitt/bmx/pkg/repos"
	"github.com/mproffitt/bmx/pkg/tmux"
)

const (
	listWidth               = 26
	padding                 = 2
	previewWidth            = 80
	previewHeight           = 30
	paddingMultiplier       = 5
	kubernetesSessionHeight = .4
)

type FocusType int

const (
	sessionList FocusType = iota
	previewPane
	contextPane
	overlay
	dialogp
	helpd
)

type model struct {
	config   *config.Config
	context  tea.Model
	deleting bool
	dialog   tea.Model
	filter   textinput.Model
	focused  FocusType
	height   int
	keymap   *keyMap
	lastch   uint
	list     list.Model
	overlay  *overlayContainer
	preview  viewport.Model
	session  tmux.Session

	styles styles
	width  int
	zoomed bool
}

type styles struct {
	sessionlist     lipgloss.Style
	viewportNormal  lipgloss.Style
	viewportFocused lipgloss.Style
	delegates       delegates
}

type delegates struct {
	normal list.DefaultDelegate
	shaded list.DefaultDelegate
}

func New(c *config.Config) *model {
	items := []list.Item{}
	m := model{
		config:  c,
		filter:  textinput.New(),
		focused: sessionList,
		keymap:  mapKeys(),
		list:    list.New(items, list.NewDefaultDelegate(), 0, 0),
		preview: viewport.New(0, 0),
		styles: styles{
			sessionlist: lipgloss.NewStyle().MarginRight(1),
			viewportNormal: lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(c.Style.BorderFgColor)).
				MarginRight(1),
			viewportFocused: lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(c.Style.FocusedColor)).
				MarginRight(1),
			delegates: delegates{},
		},
		lastch: 0,
		zoomed: false,
	}

	m.styles.delegates.normal = m.createListNormalDelegate()
	m.styles.delegates.shaded = m.createListShadedDelegate()

	m.setItems()
	m.list.SetDelegate(m.styles.delegates.normal)
	m.list.SetShowFilter(false)
	m.list.SetFilteringEnabled(false)
	m.list.SetShowHelp(false)
	m.list.SetShowPagination(true)
	m.list.SetShowStatusBar(false)
	m.list.SetShowTitle(false)

	return &m
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Overlay() helpers.UseOverlay {
	repos := repos.New(m.config, repos.RepoCallback)
	width := int(math.Ceil(float64(m.width) * .8))
	height := int(math.Ceil(float64(m.height) * .75))
	repos.SetSize(width, height)
	return repos.Overlay()
}

func (m *model) GetSize() (int, int) {
	return m.width, m.height
}

func (m *model) getSessionKubeconfig(session string) string {
	kubeconfig := tmux.GetTmuxEnvVar(session, "KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = kubernetes.DefaultConfigFile()
	}
	return kubeconfig
}

func (m *model) handleDialog(status dialog.Status) (tea.Cmd, error) {
	var (
		cmd tea.Cmd
		err error
	)
	m.dialog = nil

	if m.deleting {
		switch status {
		case dialog.Confirm:
			err = kubernetes.DeleteConfig(m.session.Name)
			if err == nil {
				err = tmux.KillSwitch(m.session.Name, m.config.DefaultSession)
			}
			m.list.Select(0)
			fallthrough
		case dialog.Cancel:
			m.deleting = false
		}
		m.setItems()
	}

	if m.overlay != nil {
		cmd = helpers.OverlayCmd(status)
	}
	return cmd, err
}

func (m *model) setItems() {
	sessions := tmux.ListSessions()
	items := make([]list.Item, len(sessions))
	sort.SliceStable(sessions, func(i, j int) bool {
		if sessions[i].Attached != sessions[j].Attached {
			return sessions[i].Attached
		}
		return sessions[i].Name < sessions[j].Name
	})
	for i, s := range sessions {
		items[i] = s
	}
	_ = m.list.SetItems(items)
}

// TODO: There is something in this that prevents
// the model from resizing correctly on small terminal
// windows. It would be good to understand what this is
// in order to better lay-out the frame and not have to
// disable if the term size is too small
func (m *model) resize() {
	width := min(listWidth, int(float64(m.width)*.25))
	height := (m.height - padding)
	m.list.SetSize(width, (m.height - padding))
	m.preview.Width = (m.width - width) - (2 * padding)
	m.preview.Height = height

	if m.config.ManageSessionKubeContext {
		// look for 40% of the screen space
		sessionHeight := int(math.Ceil(float64(height) * kubernetesSessionHeight))

		// title + pager + rows + padding + border
		minheight := 2 + 2 + 4 + (padding * 2) + 0
		sessionHeight = max(minheight, sessionHeight)

		// subtract from that required elements
		sessionHeight -= (panel.PanelTitle + panel.PanelFooter)

		// calculate how many rows and cols we can use
		rows, cols, colWidth := 0, 0, 0
		{
			cols = int(math.Floor(float64(m.preview.Width-2) / float64(panel.KubernetesListWidth)))
			colWidth = int(math.Floor(float64(m.preview.Width-2) / float64(cols)))
			rows = int(math.Floor(float64(sessionHeight)/float64(panel.KubernetesRowHeight))) - 1
		}

		if m.context == nil {
			session := m.list.SelectedItem().(list.DefaultItem).Title()
			m.context = panel.NewKubectxPane(m.config, session, rows, cols, colWidth)
		}
		m.preview.Height = m.preview.Height - sessionHeight - 2
		m.context = m.context.(*panel.Model).SetSize(m.preview.Width-4, sessionHeight, colWidth)
	}
}

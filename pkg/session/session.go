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

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/components/dialog"
	"github.com/mproffitt/bmx/pkg/components/overlay"
	"github.com/mproffitt/bmx/pkg/components/splash"
	"github.com/mproffitt/bmx/pkg/components/viewport"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/kubernetes"
	"github.com/mproffitt/bmx/pkg/repos"
	"github.com/mproffitt/bmx/pkg/repos/ui/table"
	"github.com/mproffitt/bmx/pkg/tmux"
	"github.com/mproffitt/bmx/pkg/tmux/ui/manager"
	"github.com/mproffitt/bmx/pkg/tmux/ui/session"
	tmuxui "github.com/mproffitt/bmx/pkg/tmux/ui/window"
)

const (
	listWidth               = 26
	padding                 = 2
	previewWidth            = 80
	previewHeight           = 30
	paddingMultiplier       = 5
	kubernetesSessionHeight = .4
)

const (
	sessionList overlay.FocusType = iota
	previewPane
	contextPane
	overlayPane
	dialogp
	helpd
)

type ActiveType int

const (
	sessionManager ActiveType = iota
	windowManager
)

type model struct {
	active          ActiveType
	config          *config.Config
	context         tea.Model
	contextHidden   bool
	deleting        bool
	dialog          tea.Model
	filter          textinput.Model
	focused         overlay.FocusType
	height          int
	keymap          *keyMap
	lastch          uint
	list            list.Model
	manager         *manager.Model
	managerIterator manager.Iterator
	overlay         *overlay.Container
	preview         *viewport.Model
	ready           bool
	session         *session.Session
	splash          *splash.Model
	window          *tmuxui.Window

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
	manager, Iterator := manager.New()
	items := []list.Item{}
	m := model{
		active:          sessionManager,
		config:          c,
		contextHidden:   !c.ManageSessionKubeContext,
		filter:          textinput.New(),
		focused:         sessionList,
		keymap:          mapKeys(),
		list:            list.New(items, list.NewDefaultDelegate(), 0, 0),
		manager:         manager,
		managerIterator: Iterator,
		preview:         viewport.New(c.Colours(), 0, 0),
		splash:          splash.New(c.Colours()),
		styles: styles{
			sessionlist: lipgloss.NewStyle().MarginRight(1),
			viewportNormal: lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(c.Colours().Black).
				MarginRight(1),
			viewportFocused: lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(c.Colours().Blue).
				MarginRight(1),
			delegates: delegates{},
		},
		lastch: 0,
		zoomed: false,
	}

	m.styles.delegates.normal = m.createListNormalDelegate()
	m.styles.delegates.shaded = m.createListShadedDelegate()

	m.list.SetDelegate(m.styles.delegates.normal)
	m.list.SetShowFilter(false)
	m.list.SetFilteringEnabled(false)
	m.list.SetShowHelp(false)
	m.list.SetShowPagination(true)
	m.list.SetShowStatusBar(false)
	m.list.SetShowTitle(false)
	go m.setItems()

	return &m
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(m.splash.Init(), m.manager.Init())
}

func (m *model) Overlay() helpers.UseOverlay {
	repos := table.New(m.config, repos.RepoCallback)
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

	switch status {
	case dialog.Confirm:
		switch m.active {
		case sessionManager:
			if m.deleting {
				err = kubernetes.DeleteConfig(m.session.Name)
				if err == nil {
					err = m.manager.KillSwitch(m.session.Name, m.config.DefaultSession)
				}
			}
			cmd = helpers.ReloadManagerCmd()
		case windowManager:
			break
		}
		m.list.Select(0)
		fallthrough
	case dialog.Cancel:
		m.deleting = false
	}

	if m.overlay != nil {
		cmd = helpers.OverlayCmd(status)
	}
	return cmd, err
}

func (m *model) Ready() bool {
	return m.manager.Ready && m.ready
}

func (m *model) setItems() {
	switch m.active {
	case sessionManager:
		m.setSessionItems()
	case windowManager:
		m.setWindowsItems(m.session.Name)
	}
}

func (m *model) setSessionItems() {
	sessions := m.manager.Items()
	items := make([]list.Item, len(sessions))
	for i, s := range sessions {
		s.Index = uint(i)
		items[i] = s
	}
	if len(items) > 0 {
		m.ready = true
	}
	_ = m.list.SetItems(items)
	index := 0
	if m.session != nil {
		index = int(m.session.Index)
	}
	m.list.Select(index)
}

func (m *model) setWindowsItems(session string) {
	windows := tmuxui.ListWindows(session)
	items := make([]list.Item, len(windows))
	for i, w := range windows {
		items[i] = w
	}
	_ = m.list.SetItems(items)
	m.list.Select(0)
}

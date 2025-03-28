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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/components/createpanel"
	"github.com/mproffitt/bmx/pkg/components/splash"
	"github.com/mproffitt/bmx/pkg/components/viewport"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/kubernetes"
	"github.com/mproffitt/bmx/pkg/repos"
	"github.com/mproffitt/bmx/pkg/repos/ui/table"
	"github.com/mproffitt/bmx/pkg/tmux"
	"github.com/mproffitt/bmx/pkg/tmux/ui/manager"
)

func New(c *config.Config) *model {
	manager, Iterator := manager.New()
	items := []list.Item{}
	m := model{
		active:          sessionManager,
		config:          c,
		contextHidden:   !c.ManageSessionKubeContext,
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
		lastch: manager.BaseIndex(),
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

// Overlay is used to present the overlay for creating new sessions
// or windows depending on which list is currently visible
func (m *model) Overlay() helpers.UseOverlay {
	width := int(math.Ceil(float64(m.width) * .8))
	height := int(math.Ceil(float64(m.height) * .75))
	switch m.active {
	case sessionManager:
		repos := table.New(m.config, repos.RepoCallback)
		repos.SetSize(width, height)
		return repos.Overlay()
	case windowManager:
		cp := createpanel.New(m.config.Colours()).
			WithTitle("Create new window", viewport.Inline)
		cp.SetWidth(width)
		return cp.Overlay()
	}
	return nil
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

func (m *model) Ready() bool {
	return m.manager.Ready && m.ready
}

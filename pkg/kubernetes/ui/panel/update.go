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
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mproffitt/bmx/pkg/components/dialog"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/kubernetes"
	"github.com/mproffitt/bmx/pkg/tmux"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.KillPanel):
			if m.options != nil {
				m.options = nil
			}
			m.context = ""
			m.tomove = ""
		case key.Matches(msg, m.keymap.Left):
			m.activeItem = m.lists[m.activeList].Cursor()
			m.activeList = (m.activeList - 1)
			if m.activeList < 0 {
				m.activeList = len(m.lists) - 1
				m.paginator.Page = m.paginator.TotalPages - 1
			}
			if m.activeItem > len(m.lists[m.activeList].Items())-1 {
				m.activeItem = len(m.lists[m.activeList].Items()) - 1
			}
		case key.Matches(msg, m.keymap.Right):
			m.activeItem = m.lists[m.activeList].Cursor()
			m.activeList = (m.activeList + 1)
			if m.activeList > len(m.lists)-1 {
				m.activeList = 0
				m.paginator.Page = 0
			}

			if m.activeItem >= len(m.lists[m.activeList].Items())-1 {
				m.activeItem = len(m.lists[m.activeList].Items()) - 1
			}
		case key.Matches(msg, m.keymap.Up):
			if m.activeItem == 0 && m.activeList == 0 {
				m.activeList = len(m.lists) - 1
				m.paginator.Page = m.paginator.TotalPages - 1
				m.activeItem = len(m.lists[m.activeList].Items()) - 1
			} else if m.activeItem == 0 {
				m.activeList -= 1
				m.activeItem = len(m.lists[m.activeList].Items()) - 1
			} else {
				m.activeItem -= 1
			}
		case key.Matches(msg, m.keymap.Down):
			if m.activeList == len(m.lists)-1 && m.activeItem == len(m.lists[m.activeList].Items())-1 {
				m.activeList = 0
				m.activeItem = 0
				m.paginator.Page = 0
			} else if m.activeItem == len(m.lists[m.activeList].Items())-1 {
				m.activeList += 1
				m.activeItem = 0
			} else {
				m.activeItem += 1
			}

			// PAGE Up and Page down
		case key.Matches(msg, m.keymap.Pageup):
			m.paginator.PrevPage()
			m.activeList, _ = m.paginator.GetSliceBounds(len(m.lists))
			m.activeItem = 0
		case key.Matches(msg, m.keymap.Pagedown):
			m.paginator.NextPage()
			m.activeList, _ = m.paginator.GetSliceBounds(len(m.lists))
			m.activeItem = 0

			// END

		case key.Matches(msg, m.keymap.Delete, m.keymap.ShiftDel):
			m.todelete = m.lists[m.activeList].SelectedItem().(list.DefaultItem).Title()
			if key.Matches(msg, m.keymap.ShiftDel) {
				m.force = true
				return m, kubernetes.ContextDeleteCmd()
			}
		case key.Matches(msg, m.keymap.Space):
			m.optionChooser(Namespace, &m.context)
		case key.Matches(msg, m.keymap.Move):
			m.optionChooser(Session, &m.tomove)
		case key.Matches(msg, m.keymap.Login):
			m.optionChooser(ClusterLogin, nil)
		}

		if len(m.lists) > 0 {
			m.lists[m.activeList].Select((m.activeItem))
		}

	case kubernetes.ContextChangeMsg:
		if err := m.switchContext(); err != nil {
			return m, helpers.NewErrorCmd(err)
		}
		m.lists[m.activeList].Select((m.activeItem))
	case kubernetes.ContextDeleteMsg:
		if m.todelete != "" {
			if err := kubernetes.DeleteContext(m.todelete, m.kubeconfig); err != nil {
				return m, helpers.NewErrorCmd(err)
			}
			m.todelete = ""
			m.reloadContextList()
		}
	case helpers.OverlayMsg:
		switch value := msg.Message.(type) {
		case string:
			m.options = nil
			switch m.optionType {
			case Session:
				// Get filename from session
				newconfig, err := kubernetes.CreateConfig(value)
				if err != nil {
					return m, helpers.NewErrorCmd(err)
				}
				if !tmux.HasSession(value) {
					home, _ := os.UserHomeDir()
					err := tmux.CreateSession(value, home, "", true, false)
					if err != nil {
						return m, helpers.NewErrorCmd(err)
					}
					cmds = append(cmds, helpers.ReloadManagerCmd())
				}
				err = kubernetes.MoveContext(m.tomove, m.kubeconfig, newconfig)
				if err != nil {
					return m, helpers.NewErrorCmd(err)
				}

				m.reloadContextList()

			case Namespace:
				if err := kubernetes.SetNamespace(m.context, value, m.kubeconfig); err != nil {
					return m, helpers.NewErrorCmd(err)
				}
				m.context = ""
				m.reloadContextList()

			case ClusterLogin:
				if err := kubernetes.TeleportClusterLogin(value); err != nil {
					return m, helpers.NewErrorCmd(err)
				}
				m.reloadContextList()
				m.setActiveContextPage()
			}
			m.optionType = None
		case dialog.Status:
			switch value {
			case dialog.Confirm:
				if m.todelete != "" {
					cmd = kubernetes.ContextDeleteCmd()
					cmds = append(cmds, cmd)
				}
			case dialog.Cancel:
				if m.todelete != "" {
					m.todelete = ""
				}
			}
		}
	}
	return m, tea.Batch(cmds...)
}

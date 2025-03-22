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
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mproffitt/bmx/pkg/components/dialog"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/kubernetes"
	"github.com/mproffitt/bmx/pkg/tmux/ui/manager"
	"github.com/mproffitt/bmx/pkg/tmux/ui/session"
	tmuxui "github.com/mproffitt/bmx/pkg/tmux/ui/window"
)

// Has the current overlay got an active dialog on it
type HasActiveDialog interface {
	HasActiveDialog() bool
}

// This is the main logic switch for all UI behaviour in the
// app. The purpose of this is to direct messages to the correct
// location, only updating enabled or focused views as and when
// required and blocking flow when not required.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd               tea.Cmd
		cmds              []tea.Cmd
		err               error
		sendOverlayUpdate = true
	)

	switch m.active {
	case sessionManager:
		if selected, ok := m.list.SelectedItem().(session.Session); ok {
			m.session = &selected
		}
		if m.session != nil {
			m.list.Select(int((*m.session).Index))
		}
	case windowManager:
		if selected, ok := m.list.SelectedItem().(*tmuxui.Window); ok {
			m.window = selected
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:

		// Dialog needs to be handled first as it's an overlay
		// to the main window and takes precedence over all
		// other elements
		if m.dialog != nil {
			m.dialog, cmd = m.dialog.Update(msg)
			return m, cmd
		}

		// Main window key handling
		var early bool
		cmd, early, err = m.switchKeyMessage(msg, &sendOverlayUpdate)
		if early {
			return m, cmd
		}
		cmds = append(cmds, cmd)

		// Switch focus
		switch m.focused {
		case sessionList:
			m.list, cmd = m.list.Update(msg)
			m.list.SetDelegate(m.styles.delegates.normal)
			cmds = append(cmds, cmd)
			if m.dialog == nil && key.Matches(msg, m.keymap.Enter) {
				err = m.session.Attach()
				cmds = append(cmds, tea.Quit)
			}
		case previewPane:
			m.preview, cmd = m.preview.Update(msg)
			m.list.SetDelegate(m.styles.delegates.shaded)
			cmds = append(cmds, cmd)
		case contextPane:
			m.context, cmd = m.context.Update(kubernetes.ContextChangeMsg{})
			cmds = append(cmds, cmd)
		case overlayPane:
			if !sendOverlayUpdate {
				break
			}
			var model tea.Model
			model, cmd = m.overlay.Model.Update(msg)
			m.overlay.Model = model.(helpers.UseOverlay)
			cmds = append(cmds, cmd)
		}

	case kubernetes.ContextDeleteMsg:
		m.context, cmd = m.context.Update(msg)
		cmds = append(cmds, cmd)
	case helpers.OverlayMsg:
		if m.overlay != nil {
			_, cmd = (*m.overlay.Parent).Update(msg)
			m.focused = m.overlay.Previous
			m.overlay = nil
			cmds = append(cmds, cmd)
		}
	case dialog.DialogStatusMsg:
		updateModel := false

		if m.overlay != nil {
			switch (*m.overlay.Parent).(type) {
			case *model:
				updateModel = true
			}
		}

		if m.focused == overlayPane && updateModel {
			var model tea.Model
			model, cmd = m.overlay.Model.Update(msg)
			m.overlay.Model = model.(helpers.UseOverlay)
			return m, cmd
		}
		if msg.Done {
			cmd, err = m.handleDialog(msg.Selected)
			cmds = append(cmds, cmd)
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		cmds = append(cmds, helpers.ReloadManagerCmd())
	case spinner.TickMsg:
		if m.splash != nil {
			m.splash, cmd = m.splash.Update(msg)
			cmds = append(cmds, cmd)
			if m.Ready() {
				m.splash = nil
				break
			}
		}
		if m.overlay != nil {
			var model tea.Model
			model, cmd = m.overlay.Model.Update(msg)
			m.overlay.Model = model.(helpers.UseOverlay)
			cmds = append(cmds, cmd)
		}
	case manager.ManagerReadyMsg:
		if msg.Ready {
			cmds = append(cmds, helpers.ReloadManagerCmd())
		}
	case helpers.ReloadManagerMsg:
		m.resize()
		m.setItems()
	case helpers.ErrorMsg:
		if m.focused == overlayPane {
			m.focused = m.overlay.Previous
			m.overlay = nil
		}
		err = msg.Error
	case helpers.SaveMsg:
		cmd = m.save()
		cmds = append(cmds, cmd)
	}

	// handle error in dialog
	if err != nil {
		m.dialog = dialog.NewOKDialog(err.Error(), m.config, config.DialogWidth)
		m.dialog, cmd = m.dialog.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

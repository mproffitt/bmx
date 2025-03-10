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
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/dialog"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/kubernetes"
	"github.com/mproffitt/bmx/pkg/kubernetes/ui/panel"
	"github.com/mproffitt/bmx/pkg/tmux"
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

	m.session = m.list.SelectedItem().(tmux.Session)
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

		// Main window key handling
		switch {
		case key.Matches(msg, m.keymap.Quit):
			switch m.focused {
			case overlay:
				// If overlay parent is of type `model` (session) then
				// skip the update to prevent self-referencial crash
				switch (*m.overlay.parent).(type) {
				case *model:
					break
				default:
					*m.overlay.parent, _ = (*m.overlay.parent).(helpers.UseOverlay).Update(msg)
				}

				// If the overlay model implements HasActiveDialog
				// and the overlay has an active dialog, then
				// don't continue parsing key messages and instead
				// just break out of the switch
				model, ok := m.overlay.model.(HasActiveDialog)
				if ok && model.HasActiveDialog() {
					break
				}
				m.focused = m.overlay.previous
				m.overlay = nil
				return m, nil
			default:
				cmds = append(cmds, tea.Quit)
			}
		case key.Matches(msg, m.keymap.Tab):
			switch m.focused {
			case sessionList:
				m.focused = previewPane
			case previewPane:
				m.focused = sessionList
				if m.config.ManageSessionKubeContext {
					m.focused = contextPane
					m.context = m.context.(*panel.Model).Focus()
				}
			case contextPane:
				m.focused = sessionList
				m.context = m.context.(*panel.Model).Blur()
			}
		case key.Matches(msg, m.keymap.ShiftTab):
			switch m.focused {
			case sessionList:
				m.focused = previewPane
				if m.config.ManageSessionKubeContext {
					m.focused = contextPane
					m.context = m.context.(*panel.Model).Focus()
				}
			case previewPane:
				m.focused = sessionList
			case contextPane:
				m.focused = previewPane
				m.context = m.context.(*panel.Model).Blur()
			}
		case key.Matches(msg, m.keymap.Enter):
			switch m.focused {
			case contextPane:
				m.context, cmd = m.context.Update(kubernetes.ContextChangeMsg{})
				cmds = append(cmds, cmd)
			case overlay:
				break
			default:
				if m.dialog == nil {
					err = tmux.AttachSession(m.session.Name)
					cmds = append(cmds, tea.Quit)
				}
			}
		case key.Matches(msg, m.keymap.CtrlN):
			model := tea.Model(m)
			m.overlay = NewOverlayContainer(&model, m.focused)
			cmd = m.overlay.model.(tea.Model).Init()
			cmds = append(cmds, cmd)
			m.focused = overlay
		case key.Matches(msg, m.keymap.Delete):
			m.delete(msg)
		case key.Matches(msg, m.keymap.Help):
			if m.focused == overlay {
				var model tea.Model
				model, cmd = m.overlay.model.Update(msg)
				m.overlay.model = model.(helpers.UseOverlay)
				return m, cmd
			}
			m.displayHelp()
		case key.Matches(msg, m.keymap.ToggleZoom):
			if m.focused == previewPane {
				m.zoomed = !m.zoomed
			}
		case key.Matches(msg, m.keymap.HideContext):
			if m.focused != overlay {
				m.contextHidden = !m.contextHidden
				m.resize()
			}
		default:
			switch m.focused {
			case contextPane:
				m.context, cmd = m.context.Update(msg)
				cmds = append(cmds, cmd)
				if m.overlay == nil && m.context.(*panel.Model).RequiresOverlay() {
					// don't send an update to the overlay on first creation
					sendOverlayUpdate = false
					m.overlay = NewOverlayContainer(&m.context, m.focused)
					m.focused = overlay
				}
			case previewPane:
				key := msg.String()
				if len(key) == 1 && key[0] >= '0' && key[0] <= '9' {
					m.lastch = (uint(key[0]-'0') + 9) % 10
				}
			}
		}

		// Switch focus
		switch m.focused {
		case sessionList:
			m.list, cmd = m.list.Update(msg)
			m.list.SetDelegate(m.styles.delegates.normal)
			cmds = append(cmds, cmd)
		case previewPane:
			m.preview, cmd = m.preview.Update(msg)
			m.list.SetDelegate(m.styles.delegates.shaded)
			cmds = append(cmds, cmd)
		case contextPane:
			// Don't resend the messages to context-pane as it leads to duplication
		case overlay:
			if !sendOverlayUpdate {
				break
			}
			var model tea.Model
			model, cmd = m.overlay.model.Update(msg)
			m.overlay.model = model.(helpers.UseOverlay)
			cmds = append(cmds, cmd)
		}

	case kubernetes.ContextDeleteMsg:
		m.context, cmd = m.context.Update(msg)
		cmds = append(cmds, cmd)
	case helpers.OverlayMsg:
		if m.overlay != nil {
			_, cmd = (*m.overlay.parent).Update(msg)
			m.focused = m.overlay.previous
			m.overlay = nil
			cmds = append(cmds, cmd)
		}
	case dialog.DialogStatusMsg:
		updateModel := false

		if m.overlay != nil {
			switch (*m.overlay.parent).(type) {
			case *model:
				updateModel = true
			}
		}

		if m.focused == overlay && updateModel {
			var model tea.Model
			model, cmd = m.overlay.model.Update(msg)
			m.overlay.model = model.(helpers.UseOverlay)
			return m, cmd
		}
		if msg.Done {
			cmd, err = m.handleDialog(msg.Selected)
			cmds = append(cmds, cmd)
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		m.resize()
	case spinner.TickMsg:
		if m.overlay != nil {
			var model tea.Model
			model, cmd = m.overlay.model.Update(msg)
			m.overlay.model = model.(helpers.UseOverlay)
			cmds = append(cmds, cmd)
		}
	case helpers.ReloadSessionsMsg:
		m.setItems()
	}

	if m.context != nil {
		ctxError := m.context.(*panel.Model).GetError()
		if ctxError != nil {
			err = ctxError
		}
	}

	// handle error in dialog
	if err != nil {
		m.dialog = dialog.NewOKDialog(err.Error(), m.config, config.DialogWidth)
		m.dialog, cmd = m.dialog.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

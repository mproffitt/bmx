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
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/mproffitt/bmx/pkg/components/overlay"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/kubernetes/ui/panel"
)

func (m *model) switchKeyMessage(msg tea.KeyMsg, sendOverlayUpdate *bool) (cmd tea.Cmd, returnEarly bool, err error) {
	var cmds []tea.Cmd
	switch {
	case key.Matches(msg, m.keymap.Quit):
		switch m.focused {
		case overlayPane:
			// If overlay parent is of type `model` (session) then
			// skip the update to prevent self-referencial crash
			switch (*m.overlay.Parent).(type) {
			case *model:
				break
			default:
				*m.overlay.Parent, cmd = (*m.overlay.Parent).(helpers.UseOverlay).Update(msg)
				cmds = append(cmds, cmd)
			}

			// If the overlay model implements HasActiveDialog
			// and the overlay has an active dialog, then
			// don't continue parsing key messages and instead
			// just break out of the switch
			model, ok := m.overlay.Model.(HasActiveDialog)
			if ok && model.HasActiveDialog() {
				break
			}
			m.focused = m.overlay.Previous
			m.overlay = nil
			returnEarly = true
		default:
			if msg.String() != "esc" {
				cmds = append(cmds, tea.Quit)
			}

			switch m.active {
			case windowManager:
				// switch back to session manager
				m.setSessionItems()
				m.active = sessionManager
				returnEarly = true
			default:
				cmds = append(cmds, tea.Quit)
			}
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
	case key.Matches(msg, m.keymap.CtrlN):
		// Create a New session by launching the create session
		// pane in an overlay window
		model := tea.Model(m)
		m.overlay = overlay.New(&model, m.focused)
		cmd = m.overlay.Model.(tea.Model).Init()
		cmds = append(cmds, cmd)
		m.focused = overlayPane
	case key.Matches(msg, m.keymap.CtrlS):
		cmd = helpers.SaveSessionsCmd()
		cmds = append(cmds, cmd)
	case key.Matches(msg, m.keymap.Delete):
		m.delete(msg)
	case key.Matches(msg, m.keymap.Help):
		if m.focused == overlayPane {
			var model tea.Model
			model, cmd = m.overlay.Model.Update(msg)
			m.overlay.Model = model.(helpers.UseOverlay)
			returnEarly = true
			cmds = append(cmds, cmd)
		}
		m.displayHelp()
	case key.Matches(msg, m.keymap.ToggleZoom):
		if m.focused == previewPane {
			m.zoomed = !m.zoomed
		}
	case key.Matches(msg, m.keymap.HideContext):
		if m.focused != overlayPane {
			m.contextHidden = !m.contextHidden
			m.resize()
		}
	case key.Matches(msg, m.keymap.SessionMode):
		if m.focused != overlayPane {
			switch m.active {
			case windowManager:
				m.setSessionItems()
				m.active = sessionManager
			}
			returnEarly = true
		}
	case key.Matches(msg, m.keymap.WindowMode):
		if m.focused != overlayPane {
			switch m.active {
			case sessionManager:
				m.setWindowsItems(m.session.Name)
				m.active = windowManager
			case windowManager:
				m.setSessionItems()
				m.active = sessionManager
			}
			returnEarly = true
		}
	case key.Matches(msg, m.keymap.Rename):
		if m.focused == sessionList {
			m.rename()
		}

	default:
		switch m.focused {
		case contextPane:
			m.context, cmd = m.context.Update(msg)
			cmds = append(cmds, cmd)
			if m.overlay == nil && m.context.(*panel.Model).RequiresOverlay() {
				// don't send an update to the overlay on first creation
				*sendOverlayUpdate = false
				m.overlay = overlay.New(&m.context, m.focused)
				m.focused = overlayPane
			}
		case previewPane:
			// TODO: This is currently only used for zooming the given
			// pane as part of the preview window.
			// This can be piggy-backed on to allow splitting and
			// deleting panes by providing options for s&x, S&X where
			// S & X provide options without the dialog.
			key := msg.String()
			log.Debug("got key", "key", key)
			if len(key) == 1 && key[0] >= '0' && key[0] <= '9' {
				m.lastch = ((uint(key[0]-'0') + 9) % 10) + m.manager.BaseIndex()
				log.Debug("using", "lastch", m.lastch, "BaseIndex", m.manager.BaseIndex())
			}
		}
	}
	return tea.Batch(cmds...), returnEarly, err
}

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

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd               tea.Cmd
		cmds              []tea.Cmd
		err               error
		sendOVerlayUpdate = true
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
			cmds = append(cmds, cmd)
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
				// skip if the overlay is active
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
			m.displayHelp()
		default:
			switch m.focused {
			case contextPane:
				m.context, cmd = m.context.Update(msg)
				cmds = append(cmds, cmd)
				if m.overlay == nil && m.context.(*panel.Model).RequiresOverlay() {
					// don't send an update to the overlay on first creation
					sendOVerlayUpdate = false
					m.overlay = NewOverlayContainer(&m.context, m.focused)
					m.focused = overlay
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
			if !sendOVerlayUpdate {
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
		m.dialog = dialog.New(err.Error(), true, m.config, false, config.DialogWidth)
		m.dialog, cmd = m.dialog.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

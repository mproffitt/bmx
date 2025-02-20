package session

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/dialog"
)

func (m *model) delete(msg tea.Msg) tea.Cmd {
	if m.focused == overlay {
		return nil
	}

	var cmd tea.Cmd
	if m.focused == contextPane {
		m.context, cmd = m.context.Update(msg)
		if m.overlay == nil {
			m.overlay = NewOverlayContainer(&m.context, m.focused)
			m.focused = overlay
		}
		return cmd
	}

	// Dialog cannot delete active session, use kill instead
	m.dialog = dialog.New(
		"Cannot delete the current active session",
		true, m.config, false, config.DialogWidth)
	if !m.session.Attached {
		m.deleting = true
		builder := strings.Builder{}
		builder.WriteString("Are you sure you want to delete session\n")
		builder.WriteString(lipgloss.PlaceHorizontal(config.DialogWidth, lipgloss.Center,
			lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(m.config.Style.FocusedColor)).
				Padding(1).
				Render(m.session.Name)))
		if m.config.CreateSessionKubeConfig {
			builder.WriteString("\nThis will remove the associated kubeconfig")
			builder.WriteString("and log you out of all clusters")
		}

		m.dialog = dialog.New(builder.String(),
			false, m.config, false, config.DialogWidth)
	}
	m.dialog.Update(msg)
	return nil
}

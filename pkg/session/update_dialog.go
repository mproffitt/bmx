package session

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mproffitt/bmx/pkg/components/dialog"
	"github.com/mproffitt/bmx/pkg/components/toast"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/kubernetes"
)

func (m *model) handleDialog(status dialog.Status) (tea.Cmd, error) {
	var (
		cmds []tea.Cmd
		err  error
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
				cmds = append(cmds, toast.NewToastCmd(toast.Info, "Deleted session "+m.session.Name))
			}
		case windowManager:
			err = m.session.KillWindow(m.window.Index)
			if err == nil {
				cmds = append(cmds, toast.NewToastCmd(
					toast.Info,
					fmt.Sprintf("Deleted window %q with index %d", m.window.Name, m.window.Index)))
			}
		}
		cmds = append(cmds, helpers.ReloadManagerCmd())
		fallthrough
	case dialog.Cancel:
		m.deleting = false
	}

	if m.overlay != nil {
		cmds = append(cmds, helpers.OverlayCmd(status))
	}
	return tea.Batch(cmds...), err
}

package helpers

import tea "github.com/charmbracelet/bubbletea"

type OverlayMsg struct {
	Message any
}

func NewOverlayMsg(message any) OverlayMsg {
	return OverlayMsg{
		Message: message,
	}
}

func OverlayCmd(message any) tea.Cmd {
	return func() tea.Msg {
		return NewOverlayMsg(message)
	}
}

type ReloadSessionsMsg struct{}

func ReloadSessionsCmd() tea.Cmd {
	return func() tea.Msg {
		return ReloadSessionsMsg{}
	}
}

type UseOverlay interface {
	Overlay() UseOverlay
	Update(tea.Msg) (tea.Model, tea.Cmd)
	GetSize() (int, int)
}

type UseHelp interface {
	Help() string
}

package preview

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mproffitt/bmx/pkg/components/viewport"
	"github.com/mproffitt/bmx/pkg/tmux/ui/window"
)

type Model struct {
	activePane rune
	height     int
	width      int
	window     *window.Window
	viewport   *viewport.Model
}

// Create a new preview window
func New(width, height int) *Model {
	m := Model{
		height: height,
		width:  width,
	}

	return &m
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) SetActivePane(pane rune) {
	m.activePane = pane
}

func (m *Model) SetHeight(height int) {
	m.height = height
}

func (m *Model) SetWindow(window *window.Window) {
	m.window = window
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	return m, nil
}

func (m *Model) View() string {
	return ""
}

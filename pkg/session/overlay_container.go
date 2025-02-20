package session

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mproffitt/bmx/pkg/helpers"
)

type overlayContainer struct {
	model    helpers.UseOverlay
	parent   *tea.Model
	previous FocusType
}

func (o *overlayContainer) View() string {
	if m, ok := o.model.(tea.Model); ok {
		return m.View()
	}
	return ""
}

func NewOverlayContainer(parent *tea.Model, previous FocusType) *overlayContainer {
	o := overlayContainer{
		parent:   parent,
		model:    (*parent).(helpers.UseOverlay).Overlay(),
		previous: previous,
	}
	return &o
}

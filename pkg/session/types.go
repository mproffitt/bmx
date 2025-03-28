package session

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/components/overlay"
	"github.com/mproffitt/bmx/pkg/components/rename"
	"github.com/mproffitt/bmx/pkg/components/splash"
	"github.com/mproffitt/bmx/pkg/components/toast"
	"github.com/mproffitt/bmx/pkg/components/viewport"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/tmux/ui/manager"
	"github.com/mproffitt/bmx/pkg/tmux/ui/session"
	tmuxui "github.com/mproffitt/bmx/pkg/tmux/ui/window"
)

// Has the current overlay got an active dialog on it
type HasActiveDialog interface {
	HasActiveDialog() bool
}

const (
	listWidth               = 26
	padding                 = 2
	previewWidth            = 80
	previewHeight           = 30
	paddingMultiplier       = 5
	kubernetesSessionHeight = .4
)

const (
	sessionList overlay.FocusType = iota
	previewPane
	contextPane
	overlayPane
	renamePane
	dialogp
	helpd
)

type ActiveType int

const (
	sessionManager ActiveType = iota
	windowManager
)

type model struct {
	config *config.Config // Config is central & shared

	// This is for the context pane
	// can I collapse these into the same unit?
	context       tea.Model
	contextHidden bool

	deleting bool

	dialog tea.Model

	// These two relate to the focus and whether
	// the session manager, or window manager is active
	active  ActiveType
	focused overlay.FocusType

	lastch uint
	list   list.Model

	// These kind of belong together as one component
	manager         *manager.Model
	managerIterator manager.Iterator
	session         *session.Session
	window          *tmuxui.Window
	ready           bool // This may be redundant

	preview *viewport.Model
	zoomed  bool

	overlay       *overlay.Container
	renameOverlay *rename.Model
	splash        *splash.Model
	toast         *toast.Model

	// related to main window
	height int
	keymap *keyMap
	styles styles
	width  int
}

type styles struct {
	sessionlist     lipgloss.Style
	viewportNormal  lipgloss.Style
	viewportFocused lipgloss.Style
	delegates       delegates
}

type delegates struct {
	normal list.DefaultDelegate
	shaded list.DefaultDelegate
}

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

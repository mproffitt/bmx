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

package helpers

import (
	tea "github.com/charmbracelet/bubbletea"
)

// ErrorMsg is a tea message used to instruct the program it is in an error state
type ErrorMsg struct {
	Error error
}

// NewErrorCmd is returned by components that are in error
func NewErrorCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return ErrorMsg{
			Error: err,
		}
	}
}

// OverlayMsg is a tea message that contains information from an overlay window
type OverlayMsg struct {
	Message any
}

// OverlayCmd is the tea command being returned from an overlay
func OverlayCmd(message any) tea.Cmd {
	return func() tea.Msg {
		return OverlayMsg{
			Message: message,
		}
	}
}

// ReloadManagerMsg instructs the session manager that it needs to reload components
type ReloadManagerMsg struct{}

// ReloadManagerCmd is the tea command to trigger a ReloadManagerMsg
func ReloadManagerCmd() tea.Cmd {
	return func() tea.Msg {
		return ReloadManagerMsg{}
	}
}

// ReloadWindowsMsg triggers the relaoding of session windows
type ReloadWindowsMsg struct{}

// ReloadWindowsCmd triggers the sending of a ReloadWindowsMsg
func ReloadWindowsCmd() tea.Cmd {
	return func() tea.Msg {
		return ReloadWindowsMsg{}
	}
}

// SaveMsg triggers the saving of session state
type SaveMsg struct{}

// SaveSessionsCmd sends the message that sessions should be saved to config
func SaveSessionsCmd() tea.Cmd {
	return func() tea.Msg {
		return SaveMsg{}
	}
}

// UseOverlay is the interface used to instruct the main window that the
// current component can use the overlay subsystem
type UseOverlay interface {
	Overlay() UseOverlay
	Update(tea.Msg) (tea.Model, tea.Cmd)
	GetSize() (int, int)
}

// Session is a light wrapper for a tmux session
type Session struct {
	Command string   `yaml:"command"`
	Name    string   `yaml:"name"`
	Path    string   `yaml:"path"`
	Windows []Window `yaml:"windows"`
}

// Window is a light wrapper for a tmux window
type Window struct {
	Layout string `yaml:"layout"`
	Name   string `yaml:"name"`
	Index  uint   `yaml:"index"`
	Panes  []Pane `yaml:"panes"`
}

// Pane is a light wrapper for a pane within a window
type Pane struct {
	CurrentCommand string `yaml:"pane_current_command"`
	CurrentPath    string `yaml:"pane_current_path"`
	StartCommand   string `yaml:"pane_start_command"`
	StartPath      string `yaml:"pane_start_path"`
	Title          string `yaml:"title"`
}

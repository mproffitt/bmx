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

type ErrorMsg struct {
	Error error
}

func NewErrorCmd(err error) tea.Cmd {
	return func() tea.Msg {
		return ErrorMsg{
			Error: err,
		}
	}
}

type OverlayMsg struct {
	Message any
}

func OverlayCmd(message any) tea.Cmd {
	return func() tea.Msg {
		return OverlayMsg{
			Message: message,
		}
	}
}

type ReloadManagerMsg struct{}

func ReloadManagerCmd() tea.Cmd {
	return func() tea.Msg {
		return ReloadManagerMsg{}
	}
}

type ReloadWindowsMsg struct{}

func ReloadWindowsCmd() tea.Cmd {
	return func() tea.Msg {
		return ReloadWindowsMsg{}
	}
}

type SaveMsg struct{}

func SaveSessionsCmd() tea.Cmd {
	return func() tea.Msg {
		return SaveMsg{}
	}
}

type UseOverlay interface {
	Overlay() UseOverlay
	Update(tea.Msg) (tea.Model, tea.Cmd)
	GetSize() (int, int)
}

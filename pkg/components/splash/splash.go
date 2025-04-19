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

package splash

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/theme"
)

const logo = `
__/\\\\\\\\\\\\\____/\\\\____________/\\\\__/\\\_______/\\\_
 _\/\\\/////////\\\_\/\\\\\\________/\\\\\\_\///\\\___/\\\/__
  _\/\\\_______\/\\\_\/\\\//\\\____/\\\//\\\___\///\\\\\\/____
   _\/\\\\\\\\\\\\\\__\/\\\\///\\\/\\\/_\/\\\_____\//\\\\______
    _\/\\\/////////\\\_\/\\\__\///\\\/___\/\\\______\/\\\\______
     _\/\\\_______\/\\\_\/\\\____\///_____\/\\\______/\\\\\\_____
      _\/\\\_______\/\\\_\/\\\_____________\/\\\____/\\\////\\\___
       _\/\\\\\\\\\\\\\/__\/\\\_____________\/\\\__/\\\/___\///\\\_
        _\/////////////____\///______________\///__\///_______\///__
`

type Model struct {
	spinner spinner.Model
}

func New() *Model {
	spinner := spinner.New(spinner.WithSpinner(spinner.Meter))
	return &Model{
		spinner: spinner,
	}
}

func (m *Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	var c tea.Cmd
	m.spinner, c = m.spinner.Update(msg)
	return m, c
}

func (m *Model) View() string {
	content := lipgloss.JoinVertical(lipgloss.Center, logo, m.spinner.View())
	return lipgloss.NewStyle().Foreground(theme.Colours.Blue).Render(content)
}

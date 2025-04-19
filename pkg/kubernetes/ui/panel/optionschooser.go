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

package panel

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mproffitt/bmx/pkg/components/optionlist"
	"github.com/mproffitt/bmx/pkg/helpers"
)

type OptionType int

const (
	None OptionType = iota
	ClusterLogin
	Namespace
	Session
)

func (m *Model) optionChooser(o OptionType, update *string) tea.Cmd {
	if update != nil {
		*update = m.lists[m.activeList].SelectedItem().(list.DefaultItem).Title()
	}
	m.optionType = o

	var err error
	var options optionlist.Options
	{
		switch m.optionType {
		case Namespace:
			options, err = m.getNamespaceList()

		case ClusterLogin:
			options, err = m.getClusterList()
		case Session:
			options, err = m.getSessionList()
		}
	}
	if err != nil {
		return helpers.NewErrorCmd(err)
	}
	m.options = optionlist.NewOptionModel(options)
	return nil
}

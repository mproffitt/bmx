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

package overlay

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mproffitt/bmx/pkg/helpers"
)

type FocusType int

type Container struct {
	Model    helpers.UseOverlay
	Parent   *tea.Model
	Previous FocusType
}

func (o *Container) View() string {
	if m, ok := o.Model.(tea.Model); ok {
		return m.View()
	}
	return ""
}

func New(parent *tea.Model, previous FocusType) *Container {
	o := Container{
		Model:    (*parent).(helpers.UseOverlay).Overlay(),
		Parent:   parent,
		Previous: previous,
	}
	return &o
}

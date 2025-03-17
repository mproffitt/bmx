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

package window

import (
	"fmt"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/tmux"
	"github.com/shirou/gopsutil/v4/process"
)

type (
	NodeType int
	CellType int
)

const (
	Row NodeType = iota
	Column
)

const (
	CellTypeRow CellType = iota
	CellTypeCol
)

type Node struct {
	Children     []*Node
	Height       int
	Index        uint
	PaneID       *int
	Type         NodeType
	Width        int
	X            int
	Y            int
	bordercolour lipgloss.TerminalColor
	celltype     CellType
	position     int
	viewport     viewport.Model
	session      string
	window       int
}

// Finds the pane with the given ID
func (n *Node) FindPane(id int) *Node {
	if n.HasChildren() {
		for _, v := range n.Children {
			if other := v.FindPane(id); other != nil {
				return other
			}
		}
	}
	if id == *n.PaneID {
		return n
	}
	return nil
}

// Get the contents of the pane via capture pane
func (n *Node) GetContents() string {
	var (
		content string
		err     error
	)
	if n.PaneID != nil {
		content, err = tmux.CapturePane(fmt.Sprintf("%%%d", *n.PaneID), n.Width)
		if err != nil {
			content = err.Error()
		}
	}

	return content
}

// Get the command running in the pane
func (n *Node) GetCommand(pid int32) string {
	p, err := process.NewProcess(pid)
	if err != nil {
		return ""
	}

	cmd, err := p.Cmdline()
	if err != nil {
		return ""
	}
	return cmd
}

func (n *Node) SetSessionName(session string) {
	n.session = session
}

func (n *Node) SetWindowIndex(window int) {
	n.window = window
}

// Get the list of all pane commands running in this window
func (n *Node) GetCommands() []string {
	commands := make([]string, 0)
	if n.HasChildren() {
		for _, child := range n.Children {
			commands = append(commands, child.GetCommands()...)
		}
		return commands
	}
	paneid := fmt.Sprintf("%%%d", *n.PaneID)
	pid := tmux.GetPanePid(paneid)

	commands = append(commands, n.GetCommand(pid))
	return commands
}

// True if this node has children
func (n *Node) HasChildren() bool {
	return len(n.Children) != 0
}

// Visual resize of the pane or all panes in window
func (n *Node) Resize(newWidth, newHeight, originalWidth, originalHeight int) *Node {
	scaleX := float64(newWidth) / float64(originalWidth)
	scaleY := float64(newHeight) / float64(originalHeight)

	n.X = int(float64(n.X) * scaleX)
	n.Y = int(float64(n.Y) * scaleY)
	n.Width = int(float64(n.Width) * scaleX)
	n.Height = int(float64(n.Height) * scaleY)

	for i := range n.Children {
		n.Children[i] = n.Children[i].Resize(newWidth, newHeight, originalWidth, originalHeight)
	}
	return n
}

// View the current pane or window
func (n *Node) View(position int, isCol bool) string {
	if n.HasChildren() {
		layout := make([]string, 0)
		for i, child := range n.Children {
			celltype := CellTypeRow
			if n.Type == Column {
				celltype = CellTypeCol
			}
			layout = append(layout, child.WithBorderColour(n.bordercolour).
				WithCellType(celltype).
				WithPosition(i).
				View(i, n.Type == Column))
		}
		switch n.Type {
		case Row:
			return lipgloss.JoinVertical(lipgloss.Left, layout...)
		case Column:
			return lipgloss.JoinHorizontal(lipgloss.Top, layout...)
		}
	}
	if n.PaneID != nil {
		n.viewport = viewport.New(n.Width, n.Height)
		contents := n.GetContents()
		n.viewport.SetContent(contents)

		/*sessionPane := fmt.Sprintf("%%%d", *n.PaneID)
		cmd := tmux.PaneCurrentCommand(sessionPane)
		if slices.Contains(tmux.CommonShells, cmd) {
			_ = n.viewport.GotoBottom()
		}*/
		if position == 0 {
			return n.viewport.View()
		}
		if isCol {
			return lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), false, false, false, true).
				BorderForeground(n.bordercolour).Render(n.viewport.View())
		}
		return lipgloss.NewStyle().BorderForeground(n.bordercolour).
			Border(lipgloss.RoundedBorder(), true, false, true, false).Render(n.viewport.View())
	}
	return ""
}

// Use the given colour for border separation
func (n *Node) WithBorderColour(c lipgloss.TerminalColor) *Node {
	n.bordercolour = c
	return n
}

// Use the cell type to support layout
func (n *Node) WithCellType(c CellType) *Node {
	n.celltype = c
	return n
}

// Set the pane position
func (n *Node) WithPosition(p int) *Node {
	n.position = p
	return n
}

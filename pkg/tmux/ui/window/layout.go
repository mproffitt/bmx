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
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/tmux"
)

const (
	RowStart    = '['
	ColumnStart = '{'
	Separator   = ','
	RowEnd      = ']'
	ColumnEnd   = '}'
)

var paneRegex = regexp.MustCompile(`(\d+)x(\d+),(\d+),(\d+)(?:,(\d+))?`)

type Layout struct {
	checksum  string
	Root      *Node
	bordercol lipgloss.TerminalColor
	Layout    string
	Commands  []string
}

// Get the panel layout of a given window
//
// This method calls out to `display-message` to get the panel
// layout for the window, then converts that into a node tree
func NewLayout(window string) (*Layout, error) {
	layoutStr, _, err := tmux.Exec([]string{
		"display-message", "-p", "-t", window, "#{window_layout}",
	})
	if err != nil {
		return nil, fmt.Errorf("%w %q", err, layoutStr)
	}

	l := Layout{
		Layout: layoutStr,
	}
	_, err = l.layout(layoutStr)
	if err != nil {
		return nil, err
	}
	return l.findCommandStrings()
}

// View all nested windows laid out as individual viewport windows
//
// This will eventually replace all the individual window views
// in `pkg/session`
func (l *Layout) View() string {
	return l.Root.
		WithBorderColour(l.bordercol).
		View(0, false)
}

// Resize all panes in the current window
//
// Note. This is for display. It does not resize
// the actual TMUX panes
func (l *Layout) Resize(w, h int) *Layout {
	l.Root.Resize(w, h, l.Root.Width, l.Root.Height)
	return l
}

// Set the border colour to use for display
func (l *Layout) WithBorderColour(c lipgloss.TerminalColor) *Layout {
	l.bordercol = c
	return l
}

func (l *Layout) layout(layout string) (*Layout, error) {
	parts := strings.SplitN(layout, ",", 2)
	if len(parts) < 2 {
		return nil, errors.New("invalid layout format")
	}
	l.checksum = parts[0]
	var (
		remaining string
		err       error
		node      Node
	)
	node, remaining, err = l.parseNode(parts[1])
	if err != nil {
		return nil, err
	}
	l.Root = &node
	if len(remaining) > 0 {
		return nil, errors.New("unexpected trailing characters in input")
	}
	return l, nil
}

func (l *Layout) parseChildren(input string) ([]Node, string, error) {
	var nodes []Node
	for len(input) > 0 && input[0] != RowEnd && input[0] != ColumnEnd {
		var node Node
		var err error
		node, input, err = l.parseNode(input)
		if err != nil {
			return nil, "", err
		}
		nodes = append(nodes, node)
		if len(input) > 0 {
			switch input[0] {
			case Separator:
				input = input[1:]
			case RowEnd, ColumnEnd:
				return nodes, input[1:], nil
			default:
				return nil, "", errors.New("unexpected character in input")
			}
		}
	}
	return nodes, input, nil
}

func (l *Layout) parseNode(input string) (Node, string, error) {
	matches := paneRegex.FindStringSubmatch(input)
	if matches == nil {
		return Node{}, "", errors.New("invalid format")
	}
	width, _ := strconv.Atoi(matches[1])
	height, _ := strconv.Atoi(matches[2])
	x, _ := strconv.Atoi(matches[3])
	y, _ := strconv.Atoi(matches[4])

	var pane *int
	if matches[5] != "" {
		id, _ := strconv.Atoi(matches[5])
		pane = &id
	}

	node := Node{
		Height: height,
		PaneID: pane,
		Width:  width,
		X:      x,
		Y:      y,
	}
	remaining := input[len(matches[0]):]

	if len(remaining) > 0 {
		switch remaining[0] {
		case RowStart:
			node.Type = Row
			var err error
			node.Children, remaining, err = l.parseChildren(remaining[1:])
			if err != nil {
				return Node{}, "", err
			}
		case ColumnStart:
			node.Type = Column
			var err error
			node.Children, remaining, err = l.parseChildren(remaining[1:])
			if err != nil {
				return Node{}, "", err
			}
		}
	}

	return node, remaining, nil
}

func (l *Layout) findCommandStrings() (*Layout, error) {
	for _, child := range l.Root.Children {
		l.Commands = append(l.Commands, child.GetCommands()...)
	}
	return l, nil
}

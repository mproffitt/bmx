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

package tmux

import (
	"fmt"
	"slices"
	"strings"
)

// Apply the given layout to a target window
func ApplyLayout(target, layout string) error {
	return ExecSilent([]string{
		"select-layout", "-t", target, layout,
	})
}

// Create a new window
func CreateWindow(target, path, command string, force bool) error {
	args := []string{
		"new-window", "-d", "-t", target,
	}
	if force {
		args = append(args, "-k")
	}
	if path != "" {
		args = append(args, "-c", path)
	}
	if command != "" {
		args = append(args, command)
	}
	return ExecSilent(args)
}

// Get the layout for a given window
func GetWindowLayout(target string) (string, error) {
	layout, _, err := Exec([]string{
		"display-message", "-p", "-t", target, "#{window_layout}",
	})
	if err != nil {
		return "", fmt.Errorf("%w %q", err, layout)
	}
	return layout, nil
}

// HasWindow checks if a window exists in the target session
func HasWindow(target string, window uint) bool {
	out, _, _ := Exec([]string{
		"list-windows", "-t", target, "-F", "#{window_index}",
	})
	windows := strings.Split(out, "\n")
	return slices.Contains(windows, fmt.Sprintf("%d", window))
}

// KillWindow kills the target window
func KillWindow(target string) {
	_ = ExecSilent([]string{
		"kill-window", "-t", target,
	})
}

// RenameWindow renames the target window
func RenameWindow(target, name string) {
	_ = ExecSilent([]string{
		"rename-window", "-t", target, name,
	})
}

// SplitWindow splits the target window, setting startPath and startCommand
// as appropriate
//
// If `vertical` is true, the window is split vertically otherwise a
// horizontal split is carried out
func SplitWindow(target, startPath, startCommand string, vertical bool) error {
	args := []string{
		"split-window", "-t", target,
	}
	if vertical {
		args = append(args, "-v")
	}
	if startPath != "" {
		args = append(args, "-c", startPath)
	}
	if startCommand != "" {
		args = append(args, startCommand)
	}

	return ExecSilent(args)
}

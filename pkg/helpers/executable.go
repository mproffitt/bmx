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
	"os"
	"os/exec"
	"path/filepath"
)

// ExecutableName gets the name of the current running program
func ExecutableName() string {
	return filepath.Base(os.Args[0])
}

// ExecString is primarily used for popups which
// by default start in the users home directory.
//
// This method is used to run popupe in the current terminal
// directory
func ExecString() string {
	executable := ExecutableName()
	cwd, _ := os.Getwd()
	path, err := exec.LookPath(executable)
	// Popups should be executed in the same directory as the
	// user location. We do this via `cd` to `tmux run`
	if err != nil {
		return "cd " + cwd + "; " + filepath.Join(cwd, executable)
	}
	return "cd " + cwd + "; " + path
}

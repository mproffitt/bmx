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

package exec

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/log"
)

type BmxExecError struct {
	Command        string
	Stdout, Stderr string
	error          error
}

func (t *BmxExecError) Error() string {
	var builder strings.Builder

	if t.Command != "" {
		builder.WriteString("command: " + t.Command)
		builder.WriteString("\n")
	}

	if t.Stdout != "" {
		builder.WriteString("stdout:")
		for _, line := range strings.Split(t.Stdout, "\n") {
			line = fmt.Sprintf("%s %s\n", strings.Repeat(" ", len("command:")), line)
			builder.WriteString(line)
		}
	}
	if t.Stderr != "" {
		builder.WriteString("stderr:")
		for _, line := range strings.Split(t.Stderr, "\n") {
			line = fmt.Sprintf("%s %s\n", strings.Repeat(" ", len("stdout:")), line)
			builder.WriteString(line)
		}
	}

	if t.error != nil && t.error.Error() != t.Stderr {
		builder.WriteString("error :")
		for _, line := range strings.Split(t.error.Error(), "\n") {
			line = fmt.Sprintf("%s %s\n", strings.Repeat(" ", len("stdout:")), line)
			builder.WriteString(line)
		}
	}

	return builder.String()
}

func execCmd(command string, args []string) (string, string, error) {
	log.Debug(command + " " + strings.Join(args, " "))
	cmd := exec.Command(command, args...)
	var stdout strings.Builder
	var stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	var err error
	{
		err = cmd.Run()
	}
	var o, e string
	{
		o = strings.TrimSpace(stdout.String())
		e = strings.TrimSpace(stderr.String())
	}

	if err != nil {
		return "", "", &BmxExecError{
			Command: command + " " + strings.Join(args, " "),
			Stdout:  o,
			Stderr:  e,
			error:   err,
		}
	}

	return o, e, nil
}

func execSilentCmd(command string, args []string) error {
	_, _, err := Exec(command, args)
	return err
}

var (
	Exec       = execCmd
	ExecSilent = execSilentCmd
)

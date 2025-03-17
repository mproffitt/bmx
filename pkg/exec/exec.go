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
	stdout, stderr string
	error          error
}

func (t *BmxExecError) Error() string {
	var builder strings.Builder

	if t.stdout != "" {
		builder.WriteString("stdout:")
		for _, line := range strings.Split(t.stdout, "\n") {
			line = fmt.Sprintf("%s %s\n", strings.Repeat(" ", len("stdout:")), line)
			builder.WriteString(line)
		}
	}
	if t.stderr != "" {
		builder.WriteString("stderr:")
		for _, line := range strings.Split(t.stderr, "\n") {
			line = fmt.Sprintf("%s %s\n", strings.Repeat(" ", len("stdout:")), line)
			builder.WriteString(line)
		}
	}

	if t.error != nil && t.error.Error() != t.stderr {
		builder.WriteString("error :")
		for _, line := range strings.Split(t.error.Error(), "\n") {
			line = fmt.Sprintf("%s %s\n", strings.Repeat(" ", len("stdout:")), line)
			builder.WriteString(line)
		}
	}

	return builder.String()
}

func Exec(command string, args []string) (string, string, error) {
	log.Debug(command + " " + strings.Join(args, " "))
	cmd := exec.Command(command, args...)
	var stdout strings.Builder
	var stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", "", &BmxExecError{
			stdout: stdout.String(),
			stderr: stderr.String(),
			error:  err,
		}
	}

	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), nil
}

func ExecSilent(command string, args []string) error {
	_, _, err := Exec(command, args)
	return err
}

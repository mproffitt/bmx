package exec

import (
	"fmt"
	"os/exec"
	"strings"
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

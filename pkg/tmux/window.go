package tmux

import "fmt"

func WindowLayout(window string) (string, error) {
	layoutStr, _, err := Exec([]string{
		"display-message", "-p", "-t", window, "#{window_layout}",
	})
	if err != nil {
		return "", fmt.Errorf("%w %q", err, layoutStr)
	}
	return layoutStr, nil
}

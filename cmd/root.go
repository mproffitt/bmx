// Copyright © 2025 Martin Proffitt <mproffitt@choclab.net>

package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/tmux"
	"github.com/spf13/cobra"
)

var (
	executable = helpers.ExecutableName()
	tmuxExec   = helpers.ExecString()
	tmsConfig  *config.Config
	noPopup    bool
)

var rootCmd = &cobra.Command{
	Use:   executable,
	Short: "Manage tmux sessions and kubernetes contexts",
	Long: fmt.Sprintf(`%s is a tmux session manager that creates
and manages kubernetes config files on a per session basis.

Sessions can be created from git repositories discovered on your system
placing the session inside that repository and setting the KUBECONFIG
environment variable (if required) to a new config specifically for that
session`, executable),
}

func Execute() {
	var err error
	tmsConfig, err = config.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config %q\n", err.Error())
		os.Exit(1)
	}

	err = rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error %q", err.Error())
		os.Exit(1)
	}
}

func createTitle(t string) string {
	return fmt.Sprintf("#[align=centre fg=%s] %s ", tmsConfig.Style.Title, t)
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&noPopup, "no-popup", "n", false,
		"don't run in tmux popup")
}

func run(m tea.Model) {
	can, w, h, err := canDrawOnTerminal()
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
		os.Exit(1)
	}
	if !can && noPopup {
		fmt.Fprintf(os.Stderr,
			"Unable to draw screen. Resize window to at least %d x %d\n"+
				"current screen size %d x %d\n",
			config.MinWidth, config.MinHeight, w, h,
		)
		// fatal error + sigwinch
		os.Exit(128 + 28)
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program:\n%s\n", err.Error())
		os.Exit(1)
	}
}

func canDrawOnTerminal() (bool, int, int, error) {
	// If we're in a popup, TMUX_PANE isn't set
	if os.Getenv("TMUX_PANE") == "" {
		return true, 0, 0, nil
	}
	out, _, err := tmux.Exec([]string{
		"display", "-p", "#{pane_width},#{pane_height}",
	})
	if err != nil {
		return false, 0, 0, err
	}
	size := strings.Split(out, ",")
	w, _ := strconv.Atoi(size[0])
	h, _ := strconv.Atoi(size[1])
	x := w > config.MinWidth
	y := h > config.MinHeight
	return (x && y), w, h, nil
}

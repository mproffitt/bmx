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

package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/spf13/cobra"
)

var (
	executable = helpers.ExecutableName()
	tmuxExec   = helpers.ExecString()
	bmxConfig  *config.Config
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
	if len(os.Getenv("DEBUG")) > 0 {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("fatal:", err)
			os.Exit(1)
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Error("failed to close file `debug.log`", "error", err)
			}
		}()
		log.SetLevel(log.DebugLevel)
		log.SetOutput(f)
	}

	var err error
	bmxConfig, err = config.New()
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
	return fmt.Sprintf("#[align=centre fg=%s] %s ", bmxConfig.Colours().Yellow.Dark, t)
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&noPopup, "no-popup", "n", false,
		"don't run in tmux popup")
}

func run(m tea.Model) {
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error running program:\n%s\n", err.Error())
		os.Exit(1)
	}
}

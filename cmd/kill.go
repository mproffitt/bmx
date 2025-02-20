// Copyright © 2025 Martin Proffitt <mproffitt@choclab.net>

package cmd

import (
	"fmt"
	"os"

	"github.com/mproffitt/bmx/pkg/dialog"
	"github.com/mproffitt/bmx/pkg/kubernetes"
	"github.com/mproffitt/bmx/pkg/tmux"
	"github.com/spf13/cobra"
)

var force bool

var killCmd = &cobra.Command{
	Use:   "kill",
	Short: "kill the current active session",
	Run: func(cmd *cobra.Command, args []string) {
		name := tmux.CurrentSession()
		if force {
			kill(name)
			return
		}

		if !noPopup {
			err := tmux.DisplayPopup("28", "8", "", tmsConfig.Style.BorderFgColor, []string{
				tmuxExec, "kill", "--no-popup",
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to run kill command. error was %q", err.Error())
				os.Exit(1)
			}
			return
		}

		m := dialog.New("Are you sure you want to kill the current session",
			false, tmsConfig, true, 27)
		run(m)
		fmt.Printf("%+v\n", m)
		if m.(*dialog.Dialog).Status() == dialog.Confirm {
			kill(name)
		}
	},
}

func init() {
	rootCmd.AddCommand(killCmd)

	killCmd.Flags().BoolVarP(&force, "force", "f", false, "force kill the current session (skips popup)")
}

func kill(name string) {
	err := kubernetes.DeleteConfig(name)
	if err == nil {
		err = tmux.KillSession(name)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to cleanup current session. error was %q", err.Error())
		os.Exit(1)
	}
}

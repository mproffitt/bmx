// Copyright © 2025 Martin Proffitt <mproffitt@choclab.net>

package cmd

import (
	"fmt"
	"os"

	"github.com/mproffitt/bmx/pkg/session"
	"github.com/mproffitt/bmx/pkg/tmux"
	"github.com/spf13/cobra"
)

// manageCmd represents the manage command
var manageCmd = &cobra.Command{
	Use:   "manage",
	Short: "run the session manager",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !noPopup {
			err := tmux.DisplayPopup("68%", "70%", createTitle("Session Manager"), tmsConfig.Style.BorderFgColor, []string{
				tmuxExec, "--no-popup", "manage",
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "sorry, an error occurred during execution. error was %s", err.Error())
				return err
			}
			return nil
		}

		m := session.New(tmsConfig)
		run(m)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(manageCmd)
}

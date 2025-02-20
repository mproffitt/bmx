// Copyright © 2025 Martin Proffitt <mproffitt@choclab.net>

package cmd

import (
	"fmt"
	"os"

	"github.com/mproffitt/bmx/pkg/tmux"
	"github.com/spf13/cobra"
)

var menu = [][]string{
	{
		"create new session", "c", "run '" + tmuxExec + " create'",
	},
	{
		"manage sessions", "m", "run '" + tmuxExec + " list'",
	},
	{
		"kill session", "k", "run '" + tmuxExec + " kill'",
	},
	{
		"force kill session", "K", "run '" + tmuxExec + " kill -f'",
	},
	{
		"refresh", "r", "run '" + tmuxExec + " refresh",
	},
}

// menuCmd represents the menu command
var menuCmd = &cobra.Command{
	Use:   "menu",
	Short: "show a tmux menu",
	Long:  `This is a conveniance function to access the manager via a tmux menu.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := tmux.DisplayMenu(
			createTitle("Session commands"),
			tmsConfig.Style.BorderFgColor,
			tmsConfig.Style.Foreground,
			"",
			menu,
		)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(menuCmd)
}

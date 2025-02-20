// Copyright © 2025 Martin Proffitt <mproffitt@choclab.net>

package cmd

import (
	"fmt"
	"os"

	"github.com/mproffitt/bmx/pkg/tmux"
	"github.com/spf13/cobra"
)

var (
	sendVars   bool
	refreshCmd = &cobra.Command{
		Use:   "refresh",
		Short: "refresh all active sessions",
		Long: `Refresh reloads all discovered tmux config files and optionally
iterates through all active sessions and ensures the KUBECONFIG environment
variable is present in the tmux session.

This will not set the shell environment variable for existing shells unless
the 'send-vars' flag is true.

If 'send-vars' is true, refresh will iterate through all panes across all
sessions and attempt to suspend the active process, then write the environment
variables using the 'tmux send-keys' command which will send the following
in sequence:

    - tmux send-keys -t <session_name>:<window_id>.<pane_id> C-z C-m
    - tmux send-keys -t <session_name>:<window_id>.<pane_id> export $(tmux show-env KUBECONFIG) C-m
    - tmux send-keys -t <session_name>:<window_id>.<pane_id> fg C-m

The behaviour of the 'sendVars' flag may be unpredictable with applications that
do not respond to being suspended. Use with caution.

Generally, if your shell is set up correctly, you should not need to use the
'send-vars' flag although it exists as a convenience function.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := tmux.Refresh(tmsConfig.CreateSessionKubeConfig, sendVars)
			if err != nil {

				fmt.Fprintf(os.Stderr, "failed to refresh sessions. error was %q", err.Error())
				os.Exit(1)
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(refreshCmd)
	refreshCmd.Flags().BoolVarP(&sendVars, "send-vars", "s", false, "send environment variables to all active panes")
}

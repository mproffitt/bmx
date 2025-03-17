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

	"github.com/mproffitt/bmx/pkg/tmux/ui/manager"
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
			manager, _ := manager.New()
			err := manager.Refresh(tmsConfig.CreateSessionKubeConfig, sendVars)
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

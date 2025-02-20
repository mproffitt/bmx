// Copyright © 2025 Martin Proffitt <mproffitt@choclab.net>
package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mproffitt/bmx/pkg/repos"
	"github.com/mproffitt/bmx/pkg/tmux"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "create a new tmux session",
	Long: `Create a new tmux session either from git repository or
using an arbitrary name.

If run with no arguments, a picker window will be shown, populated
with all git repositories discovered under the configured list of
paths.

Optionally you may specify one or more session names to create. if
this is the case, the last name in the list will be the active
session. This way, you can specify a name, path and command to
run for that session using colon separation in the form

    <session>:<path>:<command>

If the command contains spaces, you need to quote it to ensure that
is is maintained in full for the session and not split into different
sessions.

For example:

    example:/home/martin/src/example:'kubectl port-forward service/myservice 8443:https'

If an empty path is given, your user $HOME directory will be used
instead.

If a git repository is chosen, the entire session will use the path of
the repository as its base path. Otherwise, your home directory will
be the starting path for the session

If 'createSessionKubeConfig' is true in the configuration, a new
file will be created at '$HOME/.kube' with the name of the session and
this will be exported as the $KUBECONFIG environment variable`,

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			for _, value := range args {
				parts := make([]string, 3)
				bits := strings.Split(value, ":")
				copy(parts, bits)
				fmt.Printf("%+v\n", parts)

				_ = tmux.NewSessionOrAttach(map[string]any{
					"name":    parts[0],
					"path":    parts[1],
					"command": parts[2],
				}, "", tmsConfig.CreateSessionKubeConfig)
			}
			return
		}
		if noPopup {
			m := repos.New(tmsConfig, repos.RepoCallback)
			run(m)
			return
		}
		err := tmux.DisplayPopup("65%", "50%", createTitle("Create new session"), tmsConfig.Style.BorderFgColor, []string{
			tmuxExec, "--no-popup", "create",
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "sorry, an error occurred during execution. error was %q", err.Error())
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}

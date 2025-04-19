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
	"strings"

	"github.com/mproffitt/bmx/pkg/repos"
	"github.com/mproffitt/bmx/pkg/repos/ui/table"
	"github.com/mproffitt/bmx/pkg/theme"
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
				}, bmxConfig.CreateSessionKubeConfig)
			}
			return
		}
		if noPopup {
			m := table.New(bmxConfig, repos.RepoCallback)
			run(m)
			return
		}
		err := tmux.DisplayPopup("65%", "50%", createTitle("Create new session"), theme.Colours.Black.Dark, []string{
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

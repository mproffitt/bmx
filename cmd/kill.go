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

	"github.com/mproffitt/bmx/pkg/components/dialog"
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
			err := tmux.DisplayPopup("28", "8", "", tmsConfig.Colours().Black.Dark, []string{
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

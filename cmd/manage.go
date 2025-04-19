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

	"github.com/mproffitt/bmx/pkg/session"
	"github.com/mproffitt/bmx/pkg/theme"
	"github.com/mproffitt/bmx/pkg/tmux"
	"github.com/spf13/cobra"
)

// manageCmd represents the manage command
var manageCmd = &cobra.Command{
	Use:   "manage",
	Short: "run the session manager",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !noPopup {
			err := tmux.DisplayPopup("68%", "70%", createTitle("Session Manager"), theme.Colours.Black.Dark, []string{
				tmuxExec, "--no-popup", "manage",
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "sorry, an error occurred during execution. error was %s", err.Error())
				return err
			}
			return nil
		}

		m := session.New(bmxConfig)
		run(m)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(manageCmd)
}

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
			tmsConfig.Colours().Black.Dark,
			tmsConfig.Colours().Fg.Dark,
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

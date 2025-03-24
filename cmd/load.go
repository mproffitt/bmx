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
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/kubernetes"
	"github.com/mproffitt/bmx/pkg/tmux"
	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var baseIndex uint

// createCmd represents the create command
var loadCmd = &cobra.Command{
	Use:   "load",
	Short: "load existing sessions from config",
	Long:  `Load sessions saved previously from the manager into tmux`,

	Run: func(cmd *cobra.Command, args []string) {
		if tmux.IsRunning() {
			log.Warn("tmux server is currently running. will not continue")
			if bmxConfig.DefaultSession != "" {
				fmt.Fprintf(os.Stdout, bmxConfig.DefaultSession)
			}
			return
		}
		for !tmux.IsRunning() {
			if err := startTmux(); err != nil {
				log.Fatal("failed to start tmux server", "error", err)
			}
			<-time.After(10 * time.Millisecond)
		}
		baseIndex = tmux.GetBaseIndex()
		log.Info("Using base index", "baseIndex", baseIndex)

		var wg sync.WaitGroup
		{
			for _, session := range bmxConfig.Sessions {
				wg.Add(1)
				session := session
				createSession(session, &wg)
			}
			wg.Wait()
		}

		// Clean up any sessions not in the config file
		requiredSessions := make([]string, 0)
		for _, session := range bmxConfig.Sessions {
			requiredSessions = append(requiredSessions, session.Name)
		}

		for _, s := range tmux.ListSessions() {
			name := strings.Split(s, ",")[0]
			log.Info("checking", "session", name)
			if !slices.Contains(requiredSessions, name) {
				if err := tmux.KillSession(name); err != nil {
					log.Error("failed to kill session", "session", name)
				}
			}
		}

		if bmxConfig.DefaultSession != "" {
			fmt.Fprintf(os.Stdout, bmxConfig.DefaultSession)
		}
	},
}

func init() {
	rootCmd.AddCommand(loadCmd)
}

func createSession(session helpers.Session, wg *sync.WaitGroup) {
	defer wg.Done()
	for !tmux.HasSession(session.Name) {
		log.Info("creating", "session", session.Name)
		err := tmux.CreateSession(session.Name, session.Path,
			"", bmxConfig.CreateSessionKubeConfig, false)
		if err != nil {
			log.Error("failed to create", "session", session.Name, "error", err)
			return
		}

		<-time.After(10 * time.Millisecond)
	}

	if bmxConfig.CreateSessionKubeConfig {
		config, err := kubernetes.CreateConfig(session.Name)
		if err != nil {
			log.Error("failed to create or load kubeconfig", "error", err)
		}
		err = tmux.SetSessionEnvironment(session.Name, "KUBECONFIG", config)
		if err != nil {
			log.Error("failed to set KUBECONFIG", "session", session.Name, "error", err)
		}
	}

	sort.SliceStable(session.Windows, func(i, j int) bool {
		return session.Windows[i].Index < session.Windows[j].Index
	})

	for _, window := range session.Windows {
		window := window
		createWindow(session.Name, session.Path, window)
	}
}

func createWindow(session, sPath string, window helpers.Window) {
	targetWindow := fmt.Sprintf("%s:%d", session, window.Index)
	log.Info("creating", "window", targetWindow)
	windowName := window.Name
	layout := window.Layout

	var err error
	// create window if it does not already exist
	for !tmux.HasWindow(session, window.Index) {
		err = tmux.CreateWindow(targetWindow, sPath, "", true)
		if err != nil {
			log.Error("failed to create window", "error", err)
			return
		}
		<-time.After(10 * time.Millisecond)
	}

	if windowName != "" {
		tmux.RenameWindow(targetWindow, windowName)
	}

	// create panes
	for p, pane := range window.Panes {
		paneIndex := baseIndex + uint(p)
		createPane(paneIndex, targetWindow, pane)

		targetPane := fmt.Sprintf("%s.%d", targetWindow, paneIndex)
		log.Info("resizing", "pane", targetPane)
		tmux.MazimizeCurrentPane(targetPane)
	}

	// apply layout
	log.Info("applying layout", "window", targetWindow)
	err = tmux.ApplyLayout(targetWindow, layout)
	if err != nil {
		log.Error("failed to apply", "layout", layout, "error", err)
	}
}

func createPane(paneIndex uint, targetWindow string, pane helpers.Pane) {
	targetPane := fmt.Sprintf("%s.%d", targetWindow, paneIndex)
	sendCurrentPath := true

	startPath := pane.StartPath
	{
		if startPath == "" && pane.CurrentPath != "" {
			startPath = pane.CurrentPath
			sendCurrentPath = false
		}
	}
	exists := tmux.HasPane(targetWindow, paneIndex)
	{
		log.Info("creating", "pane", targetPane, "exists", exists)
		target := targetWindow
		if exists {
			target = targetPane
		}

		err := tmux.CreatePane(target, startPath, pane.StartCommand, exists)
		if err != nil {
			log.Error("failed to create pane", "error", err)
			return
		}
	}
	// Kill the original pane if it wasn't respawned
	if !exists && isFirst(paneIndex) {
		err := tmux.KillPane(targetPane)
		if err != nil {
			log.Error("failed to kill target", "pane", targetPane)
			return
		}
	}

	if pane.CurrentPath != pane.StartPath && sendCurrentPath {
		tmux.SendKeys(targetPane, "cd "+pane.CurrentPath)
	}
	if pane.CurrentCommand != pane.StartCommand {
		tmux.SendKeys(targetPane, pane.CurrentCommand)
	}
}

func isFirst(index uint) bool {
	return index-baseIndex == 0
}

func startTmux() error {
	tmux, _ := exec.LookPath("tmux")
	attr := os.ProcAttr{
		Dir: ".",
		Env: os.Environ(),
	}
	var err error
	process, err := os.StartProcess(tmux, []string{}, &attr)
	if err == nil {
		err = process.Release()
	}
	return err
}

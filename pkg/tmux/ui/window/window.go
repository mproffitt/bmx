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

package window

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/mproffitt/bmx/pkg/components/icons"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/tmux"
)

type Flag rune

const (
	Activity      Flag = '#'
	Bell          Flag = '!'
	Silence       Flag = '~'
	CurrentWindow Flag = '*'
	LastWindow    Flag = '-'
	Marked        Flag = 'M'
	Zoomed        Flag = 'Z'
)

func getPanesCountAsIcons(panes uint64) string {
	remainder := panes % 10
	tens := (panes - remainder) / 10
	message := ""
	if tens > 1 {
		message += string(icons.DigitalNumbers[tens])
	}
	message += string(icons.DigitalNumbers[remainder])
	return message
}

// Represents a TMUX window
type Window struct {
	Active    bool
	Command   string
	Index     uint64
	Flags     map[Flag]bool
	Name      string
	PaneCount uint64
	Session   string

	checksum string
	root     *Node
	layout   string

	bordercol lipgloss.AdaptiveColor
}

// Styles for window flag icons
type WindowStyles struct {
	ActiveTerm    lipgloss.Style
	Bell          lipgloss.Style
	CurrentWindow lipgloss.Style
	LastWindow    lipgloss.Style
	Marked        lipgloss.Style
	Silence       lipgloss.Style
	Zoomed        lipgloss.Style
}

func new(session, attrStr string) *Window {
	attributes := strings.Split(attrStr, ",")
	w := Window{
		Session: session,

		Active:  attributes[3] != "0",
		Command: "",
		Name:    attributes[2],

		Flags: map[Flag]bool{
			Activity:      false,
			Bell:          false,
			CurrentWindow: false,
			LastWindow:    false,
			Marked:        false,
			Silence:       false,
			Zoomed:        false,
		},
		bordercol: lipgloss.AdaptiveColor{Dark: "#ffffff", Light: "#000000"},
	}
	w.Index, _ = strconv.ParseUint(attributes[0], 10, 8)
	for _, f := range []Flag(attributes[1]) {
		w.Flags[f] = true
	}
	w.PaneCount, _ = strconv.ParseUint(attributes[4], 10, 8)
	w.layout, _ = tmux.GetWindowLayout(fmt.Sprintf("%s:%d", w.Session, w.Index))
	if _, err := w.parse(w.layout); err != nil {
		log.Debug("error parsing layout ", w.layout)
		return nil
	}
	return &w
}

func (w *Window) Attach() error {
	target := fmt.Sprintf("%s:%d", w.Session, w.Index)
	return tmux.SwitchClient(target)
}

func (w *Window) GetName() string {
	return w.Name
}

func (w *Window) HasFlag(flag Flag) bool {
	v, ok := w.Flags[flag]
	if !ok {
		return false
	}
	return v
}

func (w *Window) Len() int {
	return w.root.Len()
}

func (w *Window) ToHelperStruct() helpers.Window {
	window := helpers.Window{
		Name:   w.Name,
		Layout: w.layout,
		Index:  uint(w.Index),
		Panes:  w.getPaneDetails(),
	}
	return window
}

func (w *Window) Rename(newname string) error {
	err := tmux.ExecSilent([]string{
		"rename-window", "-t",
		fmt.Sprintf("%s:%d", w.Session, w.Index),
		newname,
	})
	return err
}

func (w *Window) Title() string {
	return w.Name
}

func (w *Window) Description() string {
	message := ""
	current := icons.TerminalIcon
	if w.Flags[CurrentWindow] {
		current = icons.ActiveTerminalIcon
	}
	message += string(current)

	if w.Flags[Activity] {
		message += " " + string(icons.ActivityIcon)
	}
	if w.Flags[Bell] {
		message += " " + string(icons.BellIcon)
	}

	if w.Flags[LastWindow] {
		message += " " + string(icons.LastIcon)
	}
	if w.Flags[Marked] {
		message += " " + string(icons.MarkedIcon)
	}
	if w.Flags[Silence] {
		message += " " + string(icons.SilenceIcon)
	}
	if w.Flags[Zoomed] {
		message += string(icons.ZoomIcon)
	}
	return message + " " + getPanesCountAsIcons(w.PaneCount)
}

func (w *Window) FilterValue() string {
	return w.Name
}

// List all windows in a given session
func ListWindows(session string) []*Window {
	windows := make([]*Window, 0)

	var l sync.Mutex
	w, err := tmux.ListWindows(session)
	if err != nil {
		return windows
	}

	var wg sync.WaitGroup
	for _, line := range w {
		wg.Add(1)
		go func() {
			defer wg.Done()
			window := new(session, line)

			l.Lock()
			windows = append(windows, window)
			l.Unlock()
		}()
	}
	wg.Wait()

	sort.SliceStable(windows, func(i, j int) bool {
		return windows[i].Index < windows[j].Index
	})
	return windows
}

func (w *Window) getPaneDetails() []helpers.Pane {
	return w.root.GetDetails()
}

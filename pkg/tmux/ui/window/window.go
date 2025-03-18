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
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
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

var (
	DigitalNumbers = [10]rune{
		'🯰', '🯱', '🯲', '🯳', '🯴',
		'🯵', '🯶', '🯷', '🯸', '🯹',
	}

	HsquareNumbers = [10]rune{
		'󰎣', '󰎦', '󰎩', '󰎬', '󰎮',
		'󰎰', '󰎵', '󰎸', '󰎻', '󰎾',
	}
)

const (
	ActivityIcon       = '󱅫'
	ActiveTerminalIcon = ''
	ApplicationIcon    = ''
	BellIcon           = '󰂞'
	CurrentIcon        = '󰖯'
	GitIcon            = '󰊢'
	HostIcon           = '󰒋'
	LastIcon           = '󰖰'
	MarkedIcon         = '󰃀'
	SilenceIcon        = '󰂛'
	TerminalIcon       = ''
	UserIcon           = ''
	ZoomIcon           = '󰁌'
)

func getPanesCountAsIcons(panes uint64) string {
	remainder := panes % 10
	tens := (panes - remainder) / 10
	message := ""
	if tens > 1 {
		message += string(DigitalNumbers[tens])
	}
	message += string(DigitalNumbers[remainder])
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
	// colours  *config.ColourStyles
	commands []string
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
	w.layout, _ = tmux.WindowLayout(fmt.Sprintf("%s:%d", w.Session, w.Index))
	if _, err := w.parse(w.layout); err != nil {
		log.Debug("error parsing layout ", w.layout)
		return nil
	}
	w.findCommandStrings()
	return &w
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

func (w *Window) MarshalYAML() (any, error) {
	raw := struct {
		Name     string   `yaml:"name"`
		Layout   string   `yaml:"layout"`
		Commands []string `yaml:"commands"`
	}{
		Name:     w.Name,
		Layout:   w.layout,
		Commands: w.commands,
	}
	return raw, nil
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
	current := TerminalIcon
	if w.Flags[CurrentWindow] {
		current = ActiveTerminalIcon
	}
	message += string(current)

	if w.Flags[Activity] {
		message += " " + string(ActivityIcon)
	}
	if w.Flags[Bell] {
		message += " " + string(BellIcon)
	}

	if w.Flags[LastWindow] {
		message += " " + string(LastIcon)
	}
	if w.Flags[Marked] {
		message += " " + string(MarkedIcon)
	}
	if w.Flags[Silence] {
		message += " " + string(SilenceIcon)
	}
	if w.Flags[Zoomed] {
		message += string(ZoomIcon)
	}
	return message + " " + getPanesCountAsIcons(w.PaneCount)
}

func (w *Window) FilterValue() string {
	return w.Name
}

// List all windows in a given session
func ListWindows(session string) []*Window {
	windows := make([]*Window, 0)

	args := []string{
		"list-windows", "-t", session, "-F",
		"#{window_index},#{window_flags},#{window_name},#{window_active},#{window_panes}",
	}

	out, _, err := tmux.Exec(args)
	if err != nil {
		return windows
	}
	for _, line := range strings.Split(out, "\n") {
		window := new(session, line)
		windows = append(windows, window)
	}

	return windows
}

// Get a list of session windows names
func SessionWindows(session string) ([]string, error) {
	args := []string{
		"list-windows", "-t", session, "-F", "#S:#I",
	}
	out, _, err := tmux.Exec(args)
	if err != nil {
		return []string{}, err
	}

	return strings.Split(out, "\n"), nil
}

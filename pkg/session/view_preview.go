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

package session

import (
	"strings"

	"github.com/charmbracelet/log"
	"github.com/mproffitt/bmx/pkg/theme"
	rftc "github.com/muesli/reflow/truncate"
)

func (m *model) makePreview(session string, window uint64, pane uint) {
	var preview string
	preview = m.makeZoomedOut(session, window)

	if m.zoomed {
		w, _ := m.preview.GetSize()
		log.Debug("loading content", "session", session, "window", window, "pane", pane)
		log.Debugf("session %+v", m.manager.Session(session))
		log.Debugf("window %+v", m.manager.Session(session).Window(window))

		win := m.manager.Session(session).Window(window)
		if pane > uint(win.Len()) {
			return
		}
		preview = win.FindPane(pane).GetContents()
		preview = truncate(preview, w-2)
	}

	m.preview = m.preview.SetContent(preview)
}

func (m *model) makeZoomedOut(session string, windowIndex uint64) string {
	window := m.manager.Session(session).Window(windowIndex)
	colour := theme.Colours.Black
	if m.focused == previewPane {
		colour = theme.Colours.Blue
	}

	w0, h0 := m.preview.GetSize()
	w, h := m.preview.GetDrawableSize()
	log.Debug("preview size(s)", w0, w, h0, h)
	window = window.Resize(w, h).
		WithBorderColour(colour)
	return window.View()
}

func truncate(what string, width int) string {
	if width > 0 {
		builder := strings.Builder{}
		for _, line := range strings.Split(what, "\n") {
			line = rftc.String(line, uint(width))
			builder.WriteString(line + "\n")
		}
		what = builder.String()
	}
	return what
}

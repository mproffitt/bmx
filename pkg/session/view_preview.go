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

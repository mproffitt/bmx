package repos

import "github.com/mproffitt/bmx/pkg/dialog"

func (m *Model) displayHelp() {
	entries := make([]dialog.HelpEntry, 0)
	entries = append(entries, m.Help())

	m.dialog = dialog.HelpDialog(m.config, entries...)
}

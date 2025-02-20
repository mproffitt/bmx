package session

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

func (m *model) createListNormalDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color(m.config.Style.ListNormalTitle))

	delegate.Styles.NormalDesc = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color(m.config.Style.ListNormalDescription))
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color(m.config.Style.ListNormalSelectedTitle))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color(m.config.Style.ListNormalSelectedDescription))
	return delegate
}

func (m *model) createListShadedDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color(m.config.Style.ListShadedTitle))
	delegate.Styles.NormalDesc = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color(m.config.Style.ListShadedDescription))
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color(m.config.Style.ListShadedSelectedTitle))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color(m.config.Style.ListShadedSelectedDescription))
	return delegate
}

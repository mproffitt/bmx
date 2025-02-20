package config

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

const (
	optionManageKubeConfig  = "create session kube config files"
	optionManageKubeContext = "manage kube contexts"
)

type mainmodel struct {
	config  *Config
	options []string
	paths   []string
	form    *huh.Form
	theme   *huh.Theme
	fields  []huh.Field
}

func NewMainModel(c *Config) *mainmodel {
	m := mainmodel{
		config: c,
		theme:  getTheme(c),
		paths:  make([]string, 2),
	}
	if c.CreateSessionKubeConfig {
		m.options = append(m.options, optionManageKubeConfig)
	}
	if c.ManageSessionKubeContext {
		m.options = append(m.options, optionManageKubeContext)
	}
	m.createOrUpdateForm()
	return &m
}

func (m *mainmodel) Init() tea.Cmd { return m.form.Init() }

func (m *mainmodel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
		form tea.Model
	)
	form, cmd = m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		if f.State == huh.StateCompleted {
			f.Init()
		}
		cmds = append(cmds, cmd)
	}

	m.config.CreateSessionKubeConfig = false
	m.config.ManageSessionKubeContext = false
	for _, option := range m.options {
		switch option {
		case optionManageKubeConfig:
			m.config.CreateSessionKubeConfig = true
		case optionManageKubeContext:
			m.config.ManageSessionKubeContext = true
		}
	}

	m.config.Paths = make([]string, 0)
	for _, path := range m.paths {
		if path != "" {
			m.config.Paths = append(m.config.Paths, path)
		}
	}
	return m, tea.Batch(cmds...)
}

func (m *mainmodel) View() string {
	form := m.form.View()
	if m.form.State == huh.StateCompleted || form == "" {
		m.createOrUpdateForm()
		m.form.Init()
		form = m.form.View()
	}
	return lipgloss.NewStyle().PaddingLeft(2).Render(form)
}

func (m *mainmodel) createOrUpdateForm() {
	if len(m.config.Paths) > 0 {
		m.paths = make([]string, len(m.config.Paths)+1)
		_ = copy(m.paths, m.config.Paths)
	}

	m.fields = make([]huh.Field, 0)
	m.fields = append(m.fields, huh.NewMultiSelect[string]().
		Options(huh.NewOptions(optionManageKubeConfig, optionManageKubeContext)...).
		Title("Manage kube config").
		Value(&m.options))

	for i := range cap(m.paths) {
		m.fields = append(m.fields, huh.NewFilePicker().
			CurrentDirectory("/").
			Title("Select directory containing git repos").
			Value(&m.paths[i]).
			FileAllowed(false).
			DirAllowed(true))
	}

	m.form = huh.NewForm(huh.NewGroup(m.fields...)).
		WithShowHelp(false).
		WithTheme(m.theme).
		WithHeight(3 * len(m.fields)).
		WithKeyMap(keymap())
}

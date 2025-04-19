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

// The primary configuration options tab
func NewMainModel(c *Config) *mainmodel {
	m := mainmodel{
		config: c,
		theme:  getTheme(),
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

	m.fields = append(m.fields, huh.NewInput().
		Title("Default fallback session").
		Value(&m.config.DefaultSession))

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

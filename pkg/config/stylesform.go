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
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

type stylesmodel struct {
	config  *Config
	options map[string]*string
	form    *huh.Form
	theme   *huh.Theme
	fields  map[string][]huh.Field
	groups  []*huh.Group
}

// The styles options tab
func NewStylesModel(c *Config) *stylesmodel {
	s := stylesmodel{
		config: c,
		theme:  huh.ThemeCharm(), // getTheme(c),
	}

	b, _ := yaml.Marshal(c.Style)
	_ = yaml.Unmarshal(b, &s.options)

	s.createOrUpdateForm()
	return &s
}

func (m *stylesmodel) Init() tea.Cmd {
	return m.form.Init()
}

func (m *stylesmodel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
		form tea.Model
	)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case ",", ".":
		default:
			form, cmd = m.form.Update(msg)
		}
	default:
		form, cmd = m.form.Update(msg)
	}
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		if f.State == huh.StateCompleted {
			f.Init()
		}
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m *stylesmodel) View() string {
	form := m.form.View()
	if m.form.State == huh.StateCompleted || form == "" {
		m.createOrUpdateForm()
		m.form.Init()
		form = m.form.View()
	}
	return lipgloss.NewStyle().PaddingLeft(2).Render(form)
}

const (
	listStylePrefix        = "list"
	contextListStylePrefix = "context"
	buttonStylePrefix      = "button"
	dialogStylePrefix      = "dialog"
	generalStyles          = "general"
	groupHeight            = 12
	groupWidth             = 40
)

func (m *stylesmodel) createOrUpdateForm() {
	m.fields = make(map[string][]huh.Field)
	for _, v := range []string{
		listStylePrefix, contextListStylePrefix,
		buttonStylePrefix, generalStyles,
	} {
		m.fields[v] = make([]huh.Field, 0)
	}

	for k := range m.options {
		input := huh.NewInput().
			Title(k).
			Value(m.options[k]).
			Validate(isHexColour).
			WithWidth(groupWidth)

		if strings.HasPrefix(k, listStylePrefix) {
			m.fields[listStylePrefix] = append(m.fields[listStylePrefix], input)
		} else if strings.HasPrefix(k, buttonStylePrefix) {
			m.fields[buttonStylePrefix] = append(m.fields[buttonStylePrefix], input)
		} else if strings.HasPrefix(k, contextListStylePrefix) {
			m.fields[contextListStylePrefix] = append(m.fields[contextListStylePrefix], input)
		} else {
			m.fields[generalStyles] = append(m.fields[generalStyles], input)
		}
	}

	longest := 0
	for _, v := range m.fields {
		if len(v) > longest {
			longest = len(v)
		}
	}

	m.groups = make([]*huh.Group, 0)
	for k, v := range m.fields {
		sort.SliceStable(v, func(i, j int) bool {
			return v[i].GetKey() < v[j].GetKey()
		})
		m.groups = append(m.groups, huh.NewGroup(v...).
			WithHeight(groupHeight).
			WithWidth(groupWidth).Title(k))
	}
	m.form = huh.NewForm(m.groups...).
		WithShowHelp(false).
		WithTheme(m.theme).
		WithKeyMap(keymap()).
		WithLayout(huh.LayoutDefault).
		WithHeight(groupHeight)
}

func isHexColour(s string) error {
	s = strings.ToLower(s)
	if len(s) < 4 || len(s) > 7 {
		return fmt.Errorf("invalid colour. must be between 4 and 7 characters")
	}
	for _, b := range []byte(s) {
		if b == '#' {
			continue
		}

		isHex := (b >= '0' && b <= '9' || b >= 'a' && b <= 'f')
		if !isHex {
			return fmt.Errorf("invalid hex colour")
		}
	}
	return nil
}

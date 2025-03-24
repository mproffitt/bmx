package createpanel

import (
	"errors"
	"os"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mproffitt/bmx/pkg/config"
	"github.com/mproffitt/bmx/pkg/exec"
	"github.com/mproffitt/bmx/pkg/helpers"
)

type Observing interface {
	// Update notifies the Observing object of any updates coming
	// from the current model
	Update(tea.Msg) (tea.Model, tea.Cmd)
}

type Focus int

const (
	Name Focus = iota
	Path
	Command
	Button
)

// CreateOutputMsg contains the values from
// the text input fields wrapped in a tea message
type ObserverMsg struct {
	Name    string
	Path    string
	Command string
	Focus   Focus
	LastKey tea.KeyMsg
}

// GetSuggestionsMsg is sent when the panel is
// requesting new suggestions
type SuggestionsMsg struct {
	Command []string
	Name    []string
	Path    []string
	Focus   Focus
	LastKey tea.KeyMsg
}

// GetSuggestionsCmd is triggered when the
// panel is requesting suggestions
func GetSuggestionsCmd() tea.Cmd {
	return func() tea.Msg {
		return SuggestionsMsg{}
	}
}

// CreateOutputCmd wraps the CreateOutputMsg values
func ObserverCmd(msg ObserverMsg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

type Model struct {
	colours  config.ColourStyles
	current  *textinput.Model
	focus    Focus
	inputs   inputs
	keymap   keyMap
	observer Observing
	width    int
	styles   styles
}

type inputs struct {
	command textinput.Model
	name    textinput.Model
	path    textinput.Model
}

type styles struct {
	input  lipgloss.Style
	button lipgloss.Style
	active lipgloss.Style
}

func New(colours config.ColourStyles) *Model {
	model := Model{
		colours: colours,
		inputs: inputs{
			command: textinput.New(),
			name:    textinput.New(),
			path:    textinput.New(),
		},
		keymap: mapKeys(),
		styles: styles{
			active: lipgloss.NewStyle().
				Foreground(colours.BrightRed).
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(colours.Red).
				Padding(0, 4).
				MarginLeft(2),
			button: lipgloss.NewStyle().
				Foreground(colours.Black).
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(colours.Black).
				Padding(0, 4).
				MarginLeft(2),
			input: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), true).
				BorderForeground(colours.Black),
		},
	}

	model.updateKeymap(&model.inputs.command)
	model.updateKeymap(&model.inputs.name)
	model.updateKeymap(&model.inputs.path)

	model.current = &model.inputs.name
	model.current.Focus()

	return &model
}

func (m *Model) updateKeymap(input *textinput.Model) {
	(*input).ShowSuggestions = true
	(*input).KeyMap.AcceptSuggestion = m.keymap.Right
	(*input).KeyMap.NextSuggestion = m.keymap.Down
	(*input).KeyMap.PrevSuggestion = m.keymap.Up
}

func (m *Model) Init() tea.Cmd { return nil }

func (m *Model) SetWidth(width int) {
	m.width = width
}

func (m *Model) Update(msg tea.Msg) (*Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case ObserverMsg:
		if msg.Command != "" {
			m.inputs.command.SetValue(msg.Command)
		}
		if msg.Name != "" {
			m.inputs.name.SetValue(msg.Name)
		}
		if msg.Path != "" {
			m.inputs.path.SetValue(msg.Path)
		}
		if msg.Focus != m.focus {
			m.SetFocus(msg.Focus)
		}
	case SuggestionsMsg:
		if len(msg.Command) > 0 {
			m.inputs.command.SetSuggestions(msg.Command)
		}

		if len(msg.Name) > 0 {
			name := m.inputs.name.Value()
			if len(name) == 0 || !strings.Contains(msg.Name[0], name) {
				m.inputs.name.SetValue(msg.Name[0])
			}

			m.inputs.name.SetSuggestions(msg.Name)
		}

		if len(msg.Path) > 0 {
			path := m.inputs.path.Value()
			if len(path) == 0 || !strings.HasPrefix(msg.Path[0], path) {
				m.inputs.path.SetValue(msg.Path[0])
			}
			m.inputs.path.SetSuggestions(msg.Path)
		}
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keymap.Enter):
			data := m.getState(msg)
			cmds = append(cmds, ObserverCmd(data))

		case key.Matches(msg, m.keymap.ShiftTab):
			if m.current != nil {
				m.current.Blur()
			}
			switch m.focus {
			case Name:
				m.focus = Button
				m.current = nil
			case Button:
				m.focus = Command
				m.current = &m.inputs.command
			case Path:
				m.focus = Name
				m.current = &m.inputs.name
			case Command:
				m.focus = Path
				m.current = &m.inputs.path
			}
			if m.current != nil {
				m.current.Focus()
			}
		case key.Matches(msg, m.keymap.Tab):
			if m.current != nil {
				m.current.Blur()
			}
			switch m.focus {
			case Name:
				m.focus = Path
				m.current = &m.inputs.path
			case Path:
				m.focus = Command
				m.current = &m.inputs.command
			case Command:
				m.current = nil
				m.focus = Button
			case Button:
				m.focus = Name
				m.current = &m.inputs.name
			}
			if m.current != nil {
				m.current.Focus()
			}
		case key.Matches(msg, m.keymap.Up, m.keymap.Down):
			if m.focus == Name {
				m.observer, cmd = m.observer.Update(SuggestionsMsg{
					Focus:   m.focus,
					LastKey: msg,
				})
				cmds = append(cmds, cmd)
				break
			}
			// if not filter, fallthrough
			fallthrough

		default:
			// only Enter is accepted as input to button
			if m.focus == Button {
				break
			}

			*m.current, cmd = m.current.Update(msg)
			cmds = append(cmds, cmd)

			name := m.inputs.name.Value()
			path := m.inputs.path.Value()
			if len(name) == 0 && len(path) > 0 {
				m.inputs.path.SetValue("")
			}

			// If we have an observer, allow it to interact with
			// our current state
			if m.observer != nil {
				data := m.getState(msg)
				m.observer, cmd = m.observer.Update(data)
				cmds = append(cmds, cmd)
			}

			var (
				options []exec.Completion
				err     error
				value   = m.current.Value()
			)

			// Both path and command inputs accept completions from the system
			switch m.focus {
			case Command:
				// For commands we allow triggering completion from
				// space, hyphen or slash. This should allow for
				// collection of sub-commands, options and paths
				// for command line completion
				allowed := []string{" ", "-", "/"}
				if slices.Contains(allowed, msg.String()) {
					options, err = exec.ZshCompletions(value)
				}
			case Path:
				if msg.String() == "/" {
					options, err = exec.ZshCompletions(value)
				}
			}

			if err != nil {
				// In general, if we're missing zsh we do nothing
				// but if any other error arises, we need to pass
				// that messaage back to the user
				if !errors.Is(err, exec.MissingZshError{}) {
					return m, helpers.NewErrorCmd(err)
				}
				break
			}

			// Convert any shell suggestions to tea input suggestions
			suggestions := make([]string, len(options))
			{
				for _, o := range options {
					// Paths are only allowed directories
					if m.focus == Path {
						finfo, err := os.Stat(o.Option)
						if err != nil || !finfo.IsDir() {
							continue
						}
					}
					suggestions = append(suggestions, o.Option)
				}
			}

			// only set new suggestions so we're not overwriting
			// existing on every keypress
			if len(suggestions) > 0 {
				(*m.current).SetSuggestions(suggestions)
			}

		}
	}
	return m, tea.Batch(cmds...)
}

func (m *Model) WithObserver(observer Observing) *Model {
	m.observer = observer
	return m
}

func (m *Model) View() string {
	nameWidth := int(float64(m.width) * .33)
	pathWidth := int((float64(m.width) * .66) - 2)
	commandWidth := m.width - 20

	m.inputs.name.Width = nameWidth
	m.inputs.path.Width = pathWidth
	m.inputs.command.Width = commandWidth

	name := m.styles.input.Width(nameWidth).Render(m.inputs.name.View())
	path := m.styles.input.Width(pathWidth).Render(m.inputs.path.View())
	command := m.styles.input.Width(commandWidth).Render(m.inputs.command.View())
	button := m.styles.button.Render("create")
	if m.focus == Button {
		button = m.styles.active.Render("create")
	}

	nameLine := lipgloss.JoinHorizontal(lipgloss.Top, name, path)
	commandLine := lipgloss.JoinHorizontal(lipgloss.Top, command, button)
	content := lipgloss.JoinVertical(lipgloss.Left, nameLine, commandLine)

	return content
}

func (m *Model) getState(msg tea.KeyMsg) ObserverMsg {
	return ObserverMsg{
		Command: m.inputs.command.Value(),
		Name:    m.inputs.name.Value(),
		Path:    m.inputs.path.Value(),
		Focus:   m.focus,
		LastKey: msg,
	}
}

func (m *Model) SetFocus(focus Focus) {
	m.focus = focus
	switch focus {
	case Name:
		m.current = &m.inputs.name
	case Path:
		m.current = &m.inputs.path
	case Command:
		m.current = &m.inputs.command
	case Button:
		m.current = nil
	}
}

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
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mproffitt/bmx/pkg/helpers"
	"gopkg.in/yaml.v3"
)

const (
	MinHeight   = 30
	MinWidth    = 110
	DialogWidth = 30
)

type Config struct {
	Paths                    []string `yaml:"paths"`
	CreateSessionKubeConfig  bool     `yaml:"createSessionKubeConfig"`
	DefaultSession           string   `yaml:"defaultSession"`
	ManageSessionKubeContext bool     `yaml:"manageSessionKubeContext"`
	Style                    Style    `yaml:"style"`
	filename                 string
}

type Style struct {
	BorderFgColor string `yaml:"borderFgColor"`
	FocusedColor  string `yaml:"focusedColor"`
	Foreground    string `yaml:"foreground"`
	Spinner       string `yaml:"spinner"`
	Title         string `yaml:"title"`
	FilterBorder  string `yaml:"filterBorder"`

	ListNormalTitle               string `yaml:"listNormalTitle"`
	ListNormalDescription         string `yaml:"listNormalDescription"`
	ListNormalSelectedTitle       string `yaml:"listNormalSelectedTitle"`
	ListNormalSelectedDescription string `yaml:"listNormalSelectedDescription"`
	ListShadedTitle               string `yaml:"listShadedTitle"`
	ListShadedDescription         string `yaml:"listShadedDescription"`
	ListShadedSelectedTitle       string `yaml:"listShadedSelectedTitle"`
	ListShadedSelectedDescription string `yaml:"listShadedSelectedDescription"`

	ContextListNormalTitle       string `yaml:"contextListNormalTitle"`
	ContextListNormalDescription string `yaml:"contextListNormalDescription"`
	ContextListActiveTitle       string `yaml:"contextListActiveTitle"`
	ContextListActiveDescription string `yaml:"contextListActiveDescription"`

	DialogBorderColor        string `yaml:"dialogBorderColor"`
	ButtonActiveForeground   string `yaml:"buttonActiveForeground"`
	ButtonActiveBackground   string `yaml:"buttonActiveBackground"`
	ButtonInactiveForeground string `yaml:"buttonInactiveForeground"`
	ButtonInactiveBackground string `yaml:"buttonInactiveBackground"`
}

const defaultContents = `
paths: []
createSessionKubeConfig: true
style:
    borderFgColor: "#414868"
    focusedColor: "#7aa2f7"
    foreground: "#bb9af7"
    title: "#ff9e64"
    spinner: "#f7768e"
    filterBorder: "#73daca"

    listNormalTitle: "#bb9af7"
    listNormalDescription: "#565f89"
    listNormalSelectedTitle: "#2ac3de"
    listNormalSelectedDescription: "#9aa5ce"
    listShadedTitle: "#414868"
    listShadedDescription: "#414868"
    listShadedSelectedTitle: "#414868"
    listShadedSelectedDescription: "#414868"

    contextListNormalTitle: "#7aa2f7"
    contextListNormalDescription: "#565f89"
    contextListActiveTitle: "#73daca"
    contextListActiveDescription: "#7dcfff"

    dialogBorderColor: "#565f89"
    buttonActiveBackground: "#f7768e"
    buttonActiveForeground: "#cfc9c2"
    buttonInactiveForeground: "#a9b1d6"
    buttonInactiveBackground: "#414868"
`

const configFilename = "config.yaml"

// Create a new config object
//
// This will create a new Config object and load the config from disk
// If the configuration can't be found, a UI will be shown to help
// create the config. This is then stored at `~/.config/{appname}/config.yaml`
func New() (*Config, error) {
	c := Config{}

	configDir, err := c.getConfigDir()
	if err != nil {
		return nil, err
	}
	c.filename = filepath.Join(configDir, configFilename)
	if _, err := os.Stat(c.filename); err != nil && os.IsNotExist(err) {
		if err = c.createDefaultConfigIfNotExist(configDir); err != nil {
			return nil, err
		}
	}
	err = c.loadConfig(c.filename)
	return &c, err
}

// Get the name of the config file being used
//
// This returns a full path to the config file on disk
func (c *Config) GetConfigFile() string {
	return c.filename
}

func (c *Config) createConfig() error {
	m := NewConfigModel(c)
	m.createTabContents()

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program %w", err)
	}
	return nil
}

func (c *Config) createDefaultConfigIfNotExist(configDir string) error {
	err := os.Mkdir(configDir, 0750)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create configDir %q %w", configDir, err)
	}

	c.filename = filepath.Join(configDir, configFilename)

	_, err = os.Stat(c.filename)
	if err != nil && os.IsNotExist(err) {
		content := []byte(defaultContents)
		if err = yaml.Unmarshal(content, c); err != nil {
			return fmt.Errorf("failed to load default config content %w", err)
		}
		if err := c.createConfig(); err != nil {
			return fmt.Errorf("failed to create default config %w", err)
		}
		if err = c.writeConfig(c.filename); err != nil {
			return fmt.Errorf("failed to write default config %q %w", c.filename, err)
		}
	}

	return nil
}

func (c *Config) getConfigDir() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to find user config directory: %w", err)
	}

	configDir := filepath.Join(userConfigDir, helpers.ExecutableName())
	_, err = os.Stat(configDir)
	if err != nil && os.IsNotExist(err) {
		if err = c.createDefaultConfigIfNotExist(configDir); err != nil {
			return "", err
		}
	}
	return configDir, nil
}

func (c *Config) loadConfig(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(content, c)
	return err
}

func (c *Config) writeConfig(filename string) error {
	contents, err := yaml.Marshal(*c)
	if err != nil {
		return err
	}
	err = os.WriteFile(filename, []byte(contents), 0640)
	if err != nil {
		return fmt.Errorf("failed to write config file %w", err)
	}

	return nil
}

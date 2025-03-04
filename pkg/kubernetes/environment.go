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

package kubernetes

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	contents = `apiVersion: v1
clusters: []
contexts: []
current-context: ""
kind: Config
preferences: {}
users: []
`
	defaultConfigDir  = ".kube"
	defaultConfigFile = "config"
)

// Creates a new empty kubernetes config file with a suffix
// of `session-name`
//
// If the ~/.kube directory does not exist, this will first be created
// with permissions of 0700 and the file created under that with the name
// `config-<sessionName>` and permissions of 0600
func CreateConfig(sessionName string) (string, error) {
	if err := createKubeDirIfNotExist(); err != nil {
		return "", err
	}
	home, _ := os.UserHomeDir()
	sessionFile := strings.Join([]string{defaultConfigFile, sessionName}, "-")
	configFile := filepath.Join(home, defaultConfigDir, sessionFile)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		err := os.WriteFile(configFile, []byte(contents), 0600)
		if err != nil {
			return "", err
		}
	}

	return configFile, nil
}

// Gets the kubernetes default configfile for the current user
//
// This function does not test if the default configfile exists
func DefaultConfigFile() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, defaultConfigDir, defaultConfigFile)
}

// Delete the config file with the current session name as suffix
func DeleteConfig(sessionName string) error {
	home, _ := os.UserHomeDir()
	sessionFile := strings.Join([]string{defaultConfigFile, sessionName}, "-")
	configFile := filepath.Join(home, defaultConfigDir, sessionFile)

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(configFile)
}

func createKubeDirIfNotExist() error {
	home, _ := os.UserHomeDir()
	kubeDir := filepath.Join(home, defaultConfigDir)
	if _, err := os.Stat(kubeDir); os.IsNotExist(err) {
		err := os.Mkdir(kubeDir, 0700)
		if err != nil {
			return fmt.Errorf("failed to create kubernetes directory %q %w", kubeDir, err)
		}
	}
	return nil
}

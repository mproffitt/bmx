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

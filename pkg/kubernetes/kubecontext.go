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
	"context"
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type (
	ContextChangeMsg struct{}
	ContextDeleteMsg struct{}
)

func ContextDeleteCmd() tea.Cmd {
	return func() tea.Msg {
		return ContextDeleteMsg{}
	}
}

// KubeContext is a generic struct that contains
// information about a context in a given kubeconfig
// file.
//
// The struct implements `bubbles::list.DefaultItem` interface
// and can be used directly in bubbletea lists
type KubeContext struct {
	Name             string
	User             string
	Host             string
	Namespace        string
	IsCurrentContext bool
	fullname         string
}

// Get the title of this list item
func (k KubeContext) Title() string {
	return k.Name
}

// Get the description value of this list item
func (k KubeContext) Description() string {
	return k.Namespace
}

// Get the value to fileter by
func (k KubeContext) FilterValue() string { return k.Name }

// Load contexts from a kubeconfig file
//
// # If `shouldManage` is false, this function will return an empty list
//
// This function loads contexts from the given kubeconfig file and returns
// a list of `KubeContext` items
func KubeContextList(shouldManage bool, filename string) ([]KubeContext, error) {
	list := make([]KubeContext, 0)
	var err error
	if shouldManage {
		err = listContexts(&list, filename)
	}
	return list, err
}

// Gets the fullname of a context
//
// Shortened names are used for display but for
// certain actions, the full name of the context is
// required.
func GetFullName(name, filename string) string {
	list, err := KubeContextList(true, filename)
	if err == nil {
		for _, c := range list {
			if c.Name == name {
				return c.fullname
			}
		}
	}
	return ""
}

// Use only for rest clients connecting to the cluster
func buildConfigFromFlags(context, filename string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{
			ExplicitPath: filename,
		},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}

// Use for modifying local files
func getApiConfig(filename string) (*clientcmd.PathOptions, *api.Config, error) {
	options := clientcmd.NewDefaultPathOptions()
	options.GlobalFile = filename
	options.EnvVar = ""
	config, err := options.GetStartingConfig()
	if err != nil {
		return nil, nil, err
	}
	return options, config, nil
}

// Changes the configs current context to the name provided
func SetCurrentContext(name, filename string) error {
	// Sets the current kubeconfig context to the value selected
	fullname := GetFullName(name, filename)
	if fullname == "" {
		return fmt.Errorf("context name %q does not exist in current config %q", name, filename)
	}

	options, config, err := getApiConfig(filename)
	config.CurrentContext = fullname
	if err == nil {
		err = clientcmd.ModifyConfig(options, *config, true)
	}
	return err
}

func GetCurrentContext(filename string) (string, error) {
	_, config, err := getApiConfig(filename)
	var current string
	{
		if err == nil {
			current = config.CurrentContext
		}
	}
	return current, err
}

func SetNamespace(ctx, namespace, filename string) error {
	// Sets the current kubeconfig context to the value selected
	fullname := GetFullName(ctx, filename)
	if fullname == "" {
		return fmt.Errorf("context name %q does not exist in current config %q", ctx, filename)
	}

	options, config, err := getApiConfig(filename)
	config.Contexts[fullname].Namespace = namespace
	if err == nil {
		err = clientcmd.ModifyConfig(options, *config, true)
	}
	return err
}

func DeleteContext(ctx, filename string) error {
	fullname := GetFullName(ctx, filename)
	if fullname == "" {
		return fmt.Errorf("context name %q does not exist in the current config %q", ctx, filename)
	}
	options, config, err := getApiConfig(filename)
	user := config.Contexts[fullname].AuthInfo
	cluster := config.Contexts[fullname].Cluster
	{
		delete(config.AuthInfos, user)
		// delete clusters
		// applications such as teleport share the cluster information
		// across multiple contexts. Only delete if it's the only instance using it
		found := false
		for name, context := range config.Contexts {
			if name == fullname {
				continue
			}
			if context.Cluster == cluster {
				found = true
			}
		}
		if !found {
			delete(config.Clusters, cluster)
		}
		// delete context
		delete(config.Contexts, fullname)
	}
	// unset current context if applicable
	if config.CurrentContext == fullname {
		config.CurrentContext = ""
	}
	if err == nil {
		err = clientcmd.ModifyConfig(options, *config, true)
	}

	return err
}

func listContexts(list *[]KubeContext, filename string) error {
	options := clientcmd.NewDefaultPathOptions()
	options.EnvVar = ""
	options.GlobalFile = filename
	config, err := options.GetStartingConfig()
	if err != nil {
		return errors.Wrap(err, "cannot get starting config")
	}

	contexts := make([]KubeContext, 0)

	for name, ctx := range config.Contexts {
		shortname := strings.Join(strings.Split(name, "-")[1:], "-")

		kctx := KubeContext{
			Name:             shortname,
			User:             ctx.AuthInfo,
			Host:             config.Clusters[ctx.Cluster].Server,
			Namespace:        ctx.Namespace,
			IsCurrentContext: name == config.CurrentContext,
			fullname:         name,
		}
		if kctx.Namespace == "" {
			kctx.Namespace = "default"
		}
		contexts = append(contexts, kctx)
	}

	sort.SliceStable(contexts, func(i, j int) bool {
		return contexts[i].Name < contexts[j].Name
	})
	*list = make([]KubeContext, len(contexts))
	copy(*list, contexts)
	return nil
}

// Get all namespaces listed in the given context
func GetNamespaces(ctx, filename string) ([]string, error) {
	var namespaces []string
	{
		config, err := buildConfigFromFlags(ctx, filename)
		if err != nil {
			return nil, err
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			return nil, err
		}

		list, err := clientset.CoreV1().
			Namespaces().
			List(context.Background(), metav1.ListOptions{})
		if err != nil {
			return nil, err
		}

		for _, v := range list.Items {
			namespaces = append(namespaces, v.GetName())
		}
	}
	return namespaces, nil
}

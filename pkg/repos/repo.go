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

package repos

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charlievieth/fastwalk"
	tea "github.com/charmbracelet/bubbletea"
	giturl "github.com/kubescape/go-git-url"
	"github.com/mproffitt/bmx/pkg/helpers"
	"github.com/mproffitt/bmx/pkg/tmux"
	git "gopkg.in/src-d/go-git.v4"
)

type Repository struct {
	Name  string
	Owner string
	Path  string
	Url   string
}

func Find(paths []string, pattern string) ([]Repository, error) {
	conf := fastwalk.Config{
		Follow: true,
	}

	repositories := make(chan Repository)
	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if _, err := os.Stat(filepath.Join(path, pattern)); os.IsNotExist(err) {
			return nil
		}

		if _, err := os.Stat(filepath.Join(path, "..", ".git")); err == nil {
			return fastwalk.SkipDir
		}

		repo, err := git.PlainOpen(path)
		if err != nil {
			return nil
		}

		repoUrl, err := repo.Remote("origin")
		if err != nil {
			return nil
		}

		url := repoUrl.Config().URLs[0]
		gitURL, err := giturl.NewGitURL(url)
		if err != nil {
			// skip unknown repo types
			return nil
		}

		repositories <- Repository{
			Name:  strings.ToLower(gitURL.GetRepoName()),
			Owner: strings.ToLower(gitURL.GetOwnerName()),
			Path:  path,
			Url:   url,
		}
		return err
	}

	var repoList []Repository
	go func() {
		for _, root := range paths {
			if err := fastwalk.Walk(&conf, root, walkFn); err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", root, err)
			}
		}
		close(repositories)
	}()

	for repo := range repositories {
		repoList = append(repoList, repo)
	}

	return unique(repoList), nil
}

func RepoCallback(data map[string]any, useKubeConfig bool) tea.Cmd {
	return func() tea.Msg {
		if p, ok := data["path"]; !ok || p == "" {
			path, _ := os.UserHomeDir()
			data["path"] = path
		}
		if err := tmux.NewSessionOrAttach(data, useKubeConfig); err != nil {
			return helpers.ErrorMsg{Error: err}
		}
		return tea.Quit()
	}
}

func unique(sample []Repository) []Repository {
	var unique []Repository
	sort.SliceStable(sample, func(i, j int) bool {
		return sample[i].Path < sample[j].Path
	})
	for _, v := range sample {
		for _, u := range unique {
			if v.Path == u.Path {
				continue
			}
		}
		unique = append(unique, v)
	}
	return unique
}

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
	"github.com/evertras/bubble-table/table"
	giturl "github.com/kubescape/go-git-url"
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

func RepoCallback(rowData table.RowData, filter string, useKubeConfig bool) tea.Cmd {
	return func() tea.Msg {
		data := make(map[string]any)
		if map[string]any(rowData) != nil {
			data = map[string]any(rowData)
		}
		if _, ok := data["path"]; !ok {
			path, _ := os.UserHomeDir()
			data["path"] = path
		}
		if err := tmux.NewSessionOrAttach(data, filter, useKubeConfig); err != nil {
			fmt.Println("error", err)
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

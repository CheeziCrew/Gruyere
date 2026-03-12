package screens

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/CheeziCrew/curd"
)

// RepoSelectModel wraps curd.RepoSelectModel for gruyere.
type RepoSelectModel = curd.RepoSelectModel

// NewRepoSelect creates a single-select repo picker for gruyere.
func NewRepoSelect(caller, rootPath string, parentOffset, termHeight int) curd.RepoSelectModel {
	return curd.NewRepoSelectModel(curd.RepoSelectConfig{
		Palette:      palette,
		RootPath:     rootPath,
		Caller:       caller,
		ParentOffset: parentOffset,
		TermHeight:   termHeight,
		Scanner:      scanRepos,
		SingleSelect: true,
	})
}

// scanRepos discovers all git repos under rootPath.
func scanRepos(rootPath string) ([]curd.RepoInfo, error) {
	entries, err := os.ReadDir(rootPath)
	if err != nil {
		return nil, err
	}

	var repos []curd.RepoInfo
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		repoPath := filepath.Join(rootPath, entry.Name())

		// Must be a git repo
		if _, err := os.Stat(filepath.Join(repoPath, ".git")); err != nil {
			continue
		}

		branch := getGitBranch(repoPath)

		repos = append(repos, curd.RepoInfo{
			Path:          repoPath,
			Name:          entry.Name(),
			Branch:        branch,
			DefaultBranch: "main",
		})
	}

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].Name < repos[j].Name
	})

	return repos, nil
}

func getGitBranch(repoPath string) string {
	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

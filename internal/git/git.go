package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CreateGitIgnore(path string) error {
	gitIgnorePath := filepath.Join(path, ".gitignore")
	return os.WriteFile(gitIgnorePath, []byte("*"), 0o600)
}

func GitRepoURL(wd string) (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = wd
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get git remote origin url: %w", err)
	}
	return normalizeGitRepoURL(string(output)), nil
}

func GitSubdirectory(wd string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-prefix")
	cmd.Dir = wd
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return filepath.Clean(strings.TrimSpace(string(output))), nil
}

// github
// git format git@github.com:jetpack-io/opensource.git
// https format https://github.com/jetpack-io/opensource.git

// bitbucket

// git@bitbucket.org:fargo3d/public.git
// https://bitbucket.org/fargo3d/public.git

// gh format is same as git
//
// normalized: github.com/jetpack-io/opensource
func normalizeGitRepoURL(repoURL string) string {
	result := strings.TrimSpace(repoURL)
	if strings.HasPrefix(result, "git@") {
		result = strings.TrimPrefix(strings.Replace(result, ":", "/", 1), "git@")
	} else {
		result = strings.TrimPrefix(result, "https://")
		result = strings.TrimPrefix(result, "http://")
	}

	// subdomain www is rarely used but the big 3 (github, gitlab, bitbucket)
	// allow it. Obviously this doesn't work for all subdomains.
	return strings.TrimSuffix(strings.TrimPrefix(result, "www."), ".git")
}

func IsInGitRepo(wd string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = wd
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) == "true"
}

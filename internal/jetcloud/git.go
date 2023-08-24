package jetcloud

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func createGitIgnore(wd string) error {
	gitIgnorePath := filepath.Join(wd, dirName, ".gitignore")
	return os.WriteFile(gitIgnorePath, []byte("*"), 0600)
}

func gitRepoURL(wd string) (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return normalizeGitRepoURL(string(output)), nil
}

func gitSubdirectory(wd string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-prefix")
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

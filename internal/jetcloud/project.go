package jetcloud

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"go.jetpack.io/envsec/internal/auth"
)

const dirName = ".jetpack"
const configName = "envsec.json"

type projectConfig struct {
	ID projectID `json:"id"`
}

func InitProject(ctx context.Context, user *auth.User, wd string) (projectID, error) {
	existing, err := ProjectID(wd)
	if err == nil {
		return nilProjectID,
			errors.Errorf("already initialized with project ID: %s", existing)
	} else if !os.IsNotExist(err) {
		return nilProjectID, err
	}

	dirPath := filepath.Join(wd, dirName)
	if err = os.MkdirAll(dirPath, 0700); err != nil {
		return nilProjectID, err
	}

	if err = createGitIgnore(wd); err != nil {
		return nilProjectID, err
	}

	repoURL, _ := gitRepoURL(wd)
	subdir, _ := gitSubdirectory(wd)

	projectID, err := newClient().newProjectID(ctx, user, repoURL, subdir)
	if err != nil {
		return nilProjectID, err
	}

	cfg := projectConfig{ID: projectID}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nilProjectID, err
	}
	return projectID, os.WriteFile(filepath.Join(dirPath, configName), data, 0600)
}

func ProjectID(wd string) (projectID, error) {
	data, err := os.ReadFile(filepath.Join(wd, dirName, configName))
	if err != nil {
		return nilProjectID, err
	}
	var cfg projectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nilProjectID, err
	}
	return cfg.ID, nil
}

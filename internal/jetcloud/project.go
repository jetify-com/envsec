package jetcloud

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"go.jetpack.io/envsec/internal/typeids"
	"go.jetpack.io/pkg/sandbox/auth/session"
)

var ErrProjectAlreadyInitialized = errors.New("project already initialized")

const dirName = ".jetpack.io"
const configName = "project.json"

type projectConfig struct {
	ProjectID typeids.ProjectID      `json:"project_id"`
	OrgID     typeids.OrganizationID `json:"org_id"`
}

func InitProject(ctx context.Context, tok *session.Token, dir string) (typeids.ProjectID, error) {
	if tok == nil {
		return typeids.NilProjectID, errors.Errorf("Please login first")
	}
	existing, err := ProjectID(dir)
	if err == nil {
		return existing, ErrProjectAlreadyInitialized
	} else if !os.IsNotExist(err) {
		return typeids.NilProjectID, err
	}

	dirPath := filepath.Join(dir, dirName)
	if err = os.MkdirAll(dirPath, 0700); err != nil {
		return typeids.NilProjectID, err
	}

	if err = createGitIgnore(dir); err != nil {
		return typeids.NilProjectID, err
	}

	repoURL, err := gitRepoURL(dir)
	if err != nil {
		return typeids.NilProjectID, err
	}
	subdir, _ := gitSubdirectory(dir)

	projectID, err := newClient().newProjectID(ctx, tok, repoURL, subdir)
	if err != nil {
		return typeids.NilProjectID, err
	}

	claims := tok.IDClaims()
	if claims == nil {
		return typeids.NilProjectID, errors.Errorf("token did not contain an org")
	}

	orgID, err := typeids.OrganizationIDFromString(tok.IDClaims().OrgID)
	if err != nil {
		return typeids.NilProjectID, err
	}

	cfg := projectConfig{ProjectID: projectID, OrgID: orgID}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return typeids.NilProjectID, err
	}
	return projectID, os.WriteFile(filepath.Join(dirPath, configName), data, 0600)
}

func ProjectConfig(wd string) (*projectConfig, error) {
	data, err := os.ReadFile(filepath.Join(wd, dirName, configName))
	if err != nil {
		return nil, err
	}
	var cfg projectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func ProjectID(wd string) (typeids.ProjectID, error) {
	cfg, err := ProjectConfig(wd)
	if err != nil {
		return typeids.NilProjectID, err
	}
	return cfg.ProjectID, nil
}

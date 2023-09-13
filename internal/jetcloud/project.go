package jetcloud

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"go.jetpack.io/auth/session"
	"go.jetpack.io/envsec/internal/typeids"
)

const dirName = ".jetpack"
const configName = "envsec.json"

type projectConfig struct {
	ProjectID typeids.ProjectID      `json:"project_id"`
	OrgID     typeids.OrganizationID `json:"org_id"`
}

func InitProject(ctx context.Context, tok *session.Token, wd string) (typeids.ProjectID, error) {
	existing, err := ProjectID(wd)
	if err == nil {
		return typeids.NilProjectID,
			errors.Errorf("already initialized with project ID: %s", existing)
	} else if !os.IsNotExist(err) {
		return typeids.NilProjectID, err
	}

	dirPath := filepath.Join(wd, dirName)
	if err = os.MkdirAll(dirPath, 0700); err != nil {
		return typeids.NilProjectID, err
	}

	if err = createGitIgnore(wd); err != nil {
		return typeids.NilProjectID, err
	}

	repoURL, _ := gitRepoURL(wd)
	subdir, _ := gitSubdirectory(wd)

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

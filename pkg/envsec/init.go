package envsec

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"go.jetpack.io/envsec/internal/flow"
	"go.jetpack.io/envsec/internal/git"
	"go.jetpack.io/pkg/api"
	"go.jetpack.io/pkg/id"
	"go.jetpack.io/typeid"
)

var (
	ErrProjectAlreadyInitialized = errors.New("project already initialized")
	errProjectNotInitialized     = errors.New("project not initialized")
)

const (
	dirName       = ".jetpack.io"
	configName    = "project.json"
	devConfigName = "dev.project.json"
)

type projectConfig struct {
	ProjectID id.ProjectID `json:"project_id"`
	OrgID     id.OrgID     `json:"org_id"`
}

func (e *Envsec) NewProject(ctx context.Context, force bool) error {
	var err error

	authClient, err := e.authClient()
	if err != nil {
		return err
	}

	tok, err := authClient.LoginFlowIfNeeded(ctx)
	if err != nil {
		return err
	}

	projectID, err := (&flow.Init{
		Client:                api.NewClient(ctx, e.APIHost, tok),
		PromptOverwriteConfig: !force && e.configExists(),
		Token:                 tok,
		WorkingDir:            e.WorkingDir,
	}).Run(ctx)
	if err != nil {
		return err
	}

	dirPath := filepath.Join(e.WorkingDir, dirName)
	if err = os.MkdirAll(dirPath, 0o700); err != nil {
		return err
	}

	if err = git.CreateGitIgnore(dirPath); err != nil {
		return err
	}

	orgID, err := typeid.Parse[id.OrgID](tok.IDClaims().OrgID)
	if err != nil {
		return err
	}
	return e.saveConfig(projectID, orgID)
}

func (e *Envsec) ProjectConfig() (*projectConfig, error) {
	data, err := os.ReadFile(e.configPath(e.WorkingDir))
	if errors.Is(err, os.ErrNotExist) {
		return nil, errProjectNotInitialized
	} else if err != nil {
		return nil, err
	}
	var cfg projectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (e *Envsec) configPath(wd string) string {
	return filepath.Join(wd, dirName, e.configName())
}

func (e *Envsec) configName() string {
	if e.IsDev {
		return devConfigName
	}
	return configName
}

func (e *Envsec) saveConfig(projectID id.ProjectID, orgID id.OrgID) error {
	cfg := projectConfig{ProjectID: projectID, OrgID: orgID}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	dirPath := filepath.Join(e.WorkingDir, dirName)
	return os.WriteFile(filepath.Join(dirPath, e.configName()), data, 0o600)
}

func (e *Envsec) configExists() bool {
	_, err := os.Stat(e.configPath(e.WorkingDir))
	return err == nil
}

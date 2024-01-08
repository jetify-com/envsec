package envsec

import (
	"context"
	"io"
)

type Envsec struct {
	APIHost    string
	Auth       AuthConfig
	IsDev      bool
	Stderr     io.Writer
	WorkingDir string

	store Store
}

type AuthConfig struct {
	Issuer   string
	ClientID string
	// TODO Audiences and Scopes
}

func (e *Envsec) SetStore(ctx context.Context, store Store) error {
	project, err := e.ProjectConfig()
	if project == nil {
		return err
	}

	authClient, err := e.authClient()
	if err != nil {
		return err
	}

	tok, err := authClient.LoginFlowIfNeededForOrg(ctx, project.OrgID.String())
	if err != nil {
		return err
	}

	store.Identify(ctx, e, tok)
	e.store = store
	return nil
}

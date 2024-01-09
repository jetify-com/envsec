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
	if err := store.Identify(ctx, e); err != nil {
		return err
	}
	e.store = store
	return nil
}

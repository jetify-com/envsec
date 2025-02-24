package envsec

import (
	"context"
	"io"

	"go.jetify.com/pkg/auth/session"
)

type Envsec struct {
	APIHost    string
	Auth       AuthConfig
	EnvID      EnvID
	IsDev      bool
	Stderr     io.Writer
	Store      Store
	WorkingDir string
}

type AuthConfig struct {
	Audience        []string
	Issuer          string
	ClientID        string
	SuccessRedirect string
	// TODO Audiences and Scopes
}

func (e *Envsec) InitForUser(ctx context.Context) (*session.Token, error) {
	return e.Store.InitForUser(ctx, e)
}

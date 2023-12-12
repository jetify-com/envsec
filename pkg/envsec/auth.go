package envsec

import (
	"go.jetpack.io/pkg/auth"
)

func (e *Envsec) authClient() (*auth.Client, error) {
	return auth.NewClient(e.Auth.Issuer, e.Auth.ClientID)
}

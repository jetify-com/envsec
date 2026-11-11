// Package claims provides safe access to a session token's ID claims.
package claims

import (
	"errors"

	"go.jetify.com/pkg/auth/session"
)

// ErrMissingIDClaims is returned when a session token has no usable ID token.
var ErrMissingIDClaims = errors.New(
	"session has no valid ID token. Please log in again with `envsec auth login`",
)

// FromToken returns the ID claims of tok, or ErrMissingIDClaims if the token
// is nil or its ID token is missing or can't be parsed.
func FromToken(tok *session.Token) (*session.IDClaims, error) {
	if tok == nil {
		return nil, ErrMissingIDClaims
	}
	c := tok.IDClaims()
	if c == nil {
		return nil, ErrMissingIDClaims
	}
	return c, nil
}

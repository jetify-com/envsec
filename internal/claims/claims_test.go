package claims

import (
	"errors"
	"testing"

	"go.jetify.com/pkg/auth/session"
)

func TestFromTokenMissing(t *testing.T) {
	for name, tok := range map[string]*session.Token{
		"nil token":     nil,
		"empty IDToken": {},
		"bad IDToken":   {IDToken: "not-a-jwt"},
	} {
		if _, err := FromToken(tok); !errors.Is(err, ErrMissingIDClaims) {
			t.Errorf("%s: FromToken() error = %v, want ErrMissingIDClaims", name, err)
		}
	}
}

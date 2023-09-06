// Copyright 2023 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package auth

import (
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"go.jetpack.io/envsec/internal/envvar"
	"go.jetpack.io/envsec/internal/typeids"
)

var orgIDClaim = envvar.Get(
	"ENVSEC_ORG_ID_CLAIM",
	"org_id",
)

var ErrNotLoggedIn = errors.New("not logged in")
var ErrNoOrgID = errors.New("no org ID")

type User struct {
	AccessToken *jwt.Token
	IDToken     *jwt.Token
}

func (a *Authenticator) GetUser() (*User, error) {
	unverifiedTokens := &tokenSet{}
	if err := parseFile(a.getAuthFilePath(), unverifiedTokens); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.Wrapf(
				ErrNotLoggedIn,
				"you must be logged in to use this command. Run `%s`",
				a.AuthCommandHint,
			)
		}
		return nil, err
	}
	// Attempt to parse and verify the tokens.
	user, err := a.verifyAndBuildUser(unverifiedTokens)
	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		return nil, err
	}

	// If the token is expired, refresh the tokens and try again.
	if errors.Is(err, jwt.ErrTokenExpired) {
		unverifiedTokens, err = a.RefreshTokens()
		if err != nil {
			return nil, err
		}

		user, err = a.verifyAndBuildUser(unverifiedTokens)
		if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
			return nil, err
		}
	}

	return user, nil
}

func (u *User) String() string {
	return u.Email()
}

func (u *User) Email() string {
	if u == nil || u.IDToken == nil {
		return ""
	}
	return u.IDToken.Claims.(jwt.MapClaims)["email"].(string)
}

func (u *User) ID() string {
	if u == nil || u.IDToken == nil {
		return ""
	}
	return u.IDToken.Claims.(jwt.MapClaims)["sub"].(string)
}

func (u *User) OrgID() (typeids.OrganizationID, error) {
	if u == nil || u.AccessToken == nil {
		return typeids.NilOrganizationID, ErrNotLoggedIn
	}
	id, ok := u.AccessToken.Claims.(jwt.MapClaims)[orgIDClaim].(string)
	if !ok {
		return typeids.NilOrganizationID, ErrNoOrgID
	}
	return typeids.OrganizationIDFromString(id)
}

func (a *Authenticator) verifyAndBuildUser(tokens *tokenSet) (*User, error) {

	jwks, err := a.fetchJWKSWithCache()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	accessToken, err := jwt.Parse(tokens.AccessToken, jwks.Keyfunc)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	idToken, err := jwt.Parse(tokens.IDToken, jwks.Keyfunc)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &User{
		AccessToken: accessToken,
		IDToken:     idToken,
	}, nil
}

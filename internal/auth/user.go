// Copyright 2023 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package auth

import (
	"fmt"
	"os"

	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

type User struct {
	filesystemTokens *tokenSet
	IDToken          *jwt.Token
}

func (a *Authenticator) GetUser() (*User, error) {
	filesystemTokens := &tokenSet{}
	if err := parseFile(a.getAuthFilePath(), filesystemTokens); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf(
				"you must be logged in to use this command. Run `%s`", a.AuthCommandHint,
			)
		}
		return nil, err
	}
	// Attempt to parse and verify the ID token.
	IDToken, err := a.parseToken(filesystemTokens.IDToken)
	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		return nil, err
	}

	// If the token is expired, refresh the tokens and try again.
	if errors.Is(err, jwt.ErrTokenExpired) {
		filesystemTokens, err = a.RefreshTokens()
		if err != nil {
			return nil, err
		}
		IDToken, err = a.parseToken(filesystemTokens.IDToken)
		if err != nil {
			return nil, err
		}
	}

	return &User{filesystemTokens: filesystemTokens, IDToken: IDToken}, nil
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

func (u *User) OrgID() string {
	if u == nil || u.IDToken == nil {
		return ""
	}
	return u.IDToken.Claims.(jwt.MapClaims)["org_id"].(string)
}

func (a *Authenticator) parseToken(stringToken string) (*jwt.Token, error) {
	jwksURL := fmt.Sprintf(
		"https://%s/.well-known/jwks.json",
		a.Domain,
	)
	// TODO: Cache this
	jwks, err := keyfunc.Get(jwksURL, keyfunc.Options{})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	token, err := jwt.Parse(stringToken, jwks.Keyfunc)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return token, nil
}

// Copyright 2023 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package auth

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// Authenticator performs various auth0 login flows to authenticate users.
type Authenticator struct {
	AppName         string
	AuthCommandHint string
	ClientID        string
	Domain          string
	Scope           string
	Audience        string
}

// DeviceAuthFlow starts decide auth flow
func (a *Authenticator) DeviceAuthFlow(ctx context.Context, w io.Writer) error {
	resp, err := a.requestDeviceCode()
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "\nYour auth code is: %s\n\n", resp.UserCode)
	a.showVerificationURL(resp.VerificationURIComplete, w)

	tokenSuccess, err := a.requestTokens(ctx, resp)
	if err != nil {
		return err
	}

	if err = ensureDirExists(filepath.Dir(a.getAuthFilePath()), 0700, true); err != nil {
		return err
	}

	if err = writeFile(a.getAuthFilePath(), tokenSuccess); err != nil {
		return err
	}

	fmt.Fprintln(w, "You are now authenticated.")
	return nil
}

// Use existing refresh tokens to cycle all tokens. This will fail if refresh
// tokens are missing or expired. Handle accordingly
func (a *Authenticator) RefreshTokens() (*tokenSet, error) {
	tokens := &tokenSet{}
	if err := parseFile(a.getAuthFilePath(), tokens); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil,
				fmt.Errorf("you must have previously logged in to use this command")
		}
		return nil, err
	}

	tokens, err := a.doRefreshToken(tokens.RefreshToken)
	if err != nil {
		if errors.Is(err, errExpiredOrInvalidRefreshToken) {
			return nil, fmt.Errorf("our refresh token is expired or invalid. "+
				"Please log in again using `%s`", a.AuthCommandHint)
		}
		return nil, err
	}

	return tokens, writeFile(a.getAuthFilePath(), tokens)
}

func (a *Authenticator) Logout() error {
	return os.Remove(a.getAuthFilePath())
}

func ensureDirExists(path string, perm fs.FileMode, chmod bool) error {
	if err := os.MkdirAll(path, perm); err != nil && !errors.Is(err, fs.ErrExist) {
		return errors.WithStack(err)
	}
	if chmod {
		if err := os.Chmod(path, perm); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

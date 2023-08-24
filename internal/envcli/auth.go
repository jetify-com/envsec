// Copyright 2023 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.jetpack.io/envsec/internal/auth"
	"go.jetpack.io/envsec/internal/envvar"
)

func authCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "envsec auth commands",
	}

	cmd.AddCommand(loginCmd())
	cmd.AddCommand(logoutCmd())
	cmd.AddCommand(refreshCmd())
	cmd.AddCommand(whoAmICmd())

	return cmd
}

func loginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to envsec",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return newAuthenticator().DeviceAuthFlow(
				cmd.Context(),
				cmd.OutOrStdout(),
			)
		},
	}

	return cmd
}

func logoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "logout from envsec",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := newAuthenticator().Logout()
			if err == nil {
				fmt.Fprintln(cmd.OutOrStdout(), "Logged out successfully")
			}
			return err
		},
	}

	return cmd
}

// This is for debugging purposes only. Hidden.
func refreshCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "refresh",
		Short:  "Refresh credentials",
		Args:   cobra.ExactArgs(0),
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := newAuthenticator().RefreshTokens()
			return err
		},
	}

	return cmd
}

func whoAmICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Show the current user",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			user, err := newAuthenticator().GetUser()
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), user)
			return nil
		},
	}

	return cmd
}

func newAuthenticator() *auth.Authenticator {
	return &auth.Authenticator{
		AppName:         "envsec",
		AuthCommandHint: "devbox auth login",
		ClientID: envvar.Get(
			"ENVSEC_AUTH_CLIENT_ID",
			"5PusB4fMm6BQ8WbTFObkTI0JUDi9ahPC",
		),
		Domain: envvar.Get(
			"ENVSEC_AUTH_DOMAIN",
			"auth.jetpack.io",
		),
		Scope: envvar.Get(
			"ENVSEC_AUTH_SCOPE",
			"openid offline_access email profile",
		),
		Audience: envvar.Get(
			"ENVSEC_AUTH_AUDIENCE",
			"https://api.jetpack.io",
		),
	}
}

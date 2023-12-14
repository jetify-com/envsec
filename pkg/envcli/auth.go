// Copyright 2023 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.jetpack.io/envsec/internal/build"
	"go.jetpack.io/pkg/auth"
	"go.jetpack.io/pkg/envvar"
)

func authCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "envsec auth commands",
	}

	cmd.AddCommand(loginCmd())
	cmd.AddCommand(logoutCmd())
	cmd.AddCommand(whoAmICmd())

	return cmd
}

func loginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to envsec",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := newAuthClient()
			if err != nil {
				return err
			}

			_, err = client.LoginFlow()
			if err == nil {
				fmt.Fprintln(cmd.OutOrStdout(), "Logged in successfully")
			}
			return err
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
			client, err := newAuthClient()
			if err != nil {
				return err
			}

			err = client.LogoutFlow()
			if err == nil {
				fmt.Fprintln(cmd.OutOrStdout(), "Logged out successfully")
			}
			return err
		},
	}

	return cmd
}

type whoAmICmdFlags struct {
	showTokens bool
}

func whoAmICmd() *cobra.Command {
	flags := &whoAmICmdFlags{}
	cmd := &cobra.Command{
		Use:   "whoami",
		Short: "Show the current user",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return defaultEnvsec(cmd).
				WhoAmI(cmd.Context(), cmd.OutOrStdout(), flags.showTokens)
		},
	}

	cmd.Flags().BoolVar(
		&flags.showTokens,
		"show-tokens",
		false,
		"Show the access, id, and refresh tokens",
	)

	return cmd
}

func newAuthClient() (*auth.Client, error) {
	issuer := envvar.Get("ENVSEC_ISSUER", build.Issuer())
	clientID := envvar.Get("ENVSEC_CLIENT_ID", build.ClientID())
	// TODO: Consider making scopes and audience configurable:
	// "ENVSEC_AUTH_SCOPE" = "openid offline_access email profile"
	// "ENVSEC_AUTH_AUDIENCE" = "https://api.jetpack.io",
	return auth.NewClient(
		issuer,
		clientID,
		[]string{"openid", "offline_access", "email", "profile"},
	)
}

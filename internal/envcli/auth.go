// Copyright 2023 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.jetpack.io/envsec/internal/envvar"
	"go.jetpack.io/pkg/sandbox/auth"
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
			client, err := newAuthClient()
			if err != nil {
				return err
			}

			tok, err := client.GetSession(cmd.Context())
			if err != nil {
				return fmt.Errorf("error: %w. Run `envsec auth login` to log in", err)
			}
			idClaims := tok.IDClaims()

			fmt.Fprintf(cmd.OutOrStdout(), "Logged in\n")
			fmt.Fprintf(cmd.OutOrStdout(), "User ID: %s\n", idClaims.Subject)

			if idClaims.OrgID != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Org ID: %s\n", idClaims.OrgID)
			}

			if idClaims.Email != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Email: %s\n", idClaims.Email)
			}

			if idClaims.Name != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", idClaims.Name)
			}

			if flags.showTokens {
				fmt.Fprintf(cmd.OutOrStdout(), "Access Token: %s\n", tok.AccessToken)
				fmt.Fprintf(cmd.OutOrStdout(), "ID Token: %s\n", tok.IDToken)
				fmt.Fprintf(cmd.OutOrStdout(), "Refresh Token: %s\n", tok.RefreshToken)
			}

			return nil
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
	issuer := envvar.Get("ENVSEC_ISSUER", "https://accounts.jetpack.io")
	clientID := envvar.Get("ENVSEC_CLIENT_ID", "ff3d4c9c-1ac8-42d9-bef1-f5218bb1a9f6")
	// TODO: Consider making scopes and audience configurable:
	// "ENVSEC_AUTH_SCOPE" = "openid offline_access email profile"
	// "ENVSEC_AUTH_AUDIENCE" = "https://api.jetpack.io",
	return auth.NewClient(issuer, clientID)
}

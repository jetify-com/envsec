package envsec

import (
	"context"
	"fmt"
	"io"

	"go.jetpack.io/pkg/auth"
	"go.jetpack.io/pkg/jetcloud"
)

func (e *Envsec) authClient() (*auth.Client, error) {
	return auth.NewClient(
		e.Auth.Issuer,
		e.Auth.ClientID,
		[]string{"openid", "offline_access", "email", "profile"},
	)
}

func (e *Envsec) WhoAmI(
	ctx context.Context,
	w io.Writer,
	showTokens bool,
) error {
	client, err := e.authClient()
	if err != nil {
		return err
	}

	tok, err := client.GetSession(ctx)
	if err != nil {
		return fmt.Errorf("error: %w, run `envsec auth login`", err)
	}

	idClaims := tok.IDClaims()

	fmt.Fprintf(w, "Logged in\n")
	fmt.Fprintf(w, "User ID: %s\n", idClaims.Subject)

	if idClaims.OrgID != "" {
		fmt.Fprintf(w, "Org ID: %s\n", idClaims.OrgID)
	}

	if idClaims.Email != "" {
		fmt.Fprintf(w, "Email: %s\n", idClaims.Email)
	}

	if idClaims.Name != "" {
		fmt.Fprintf(w, "Name: %s\n", idClaims.Name)
	}

	apiClient := jetcloud.Client{APIHost: e.APIHost, IsDev: e.IsDev}

	member, err := apiClient.GetMember(ctx, tok, tok.IDClaims().Subject)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Org name: %s\n", member.Organization.Name)

	if showTokens {
		fmt.Fprintf(w, "Access Token: %s\n", tok.AccessToken)
		fmt.Fprintf(w, "ID Token: %s\n", tok.IDToken)
		fmt.Fprintf(w, "Refresh Token: %s\n", tok.RefreshToken)
	}
	return nil
}

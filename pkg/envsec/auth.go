package envsec

import (
	"context"
	"fmt"
	"io"

	"go.jetify.com/pkg/api"
	"go.jetify.com/pkg/auth"
)

func (e *Envsec) AuthClient() (*auth.Client, error) {
	return auth.NewClient(
		e.Auth.Issuer,
		e.Auth.ClientID,
		[]string{"openid", "offline_access", "email", "profile"},
		e.Auth.SuccessRedirect,
		e.Auth.Audience,
	)
}

func (e *Envsec) WhoAmI(
	ctx context.Context,
	w io.Writer,
	showTokens bool,
) error {
	authClient, err := e.AuthClient()
	if err != nil {
		return err
	}

	tok, err := authClient.GetSession(ctx)
	if err != nil {
		return err
	}

	idClaims := tok.IDClaims()

	_, _ = fmt.Fprintf(w, "Logged in\n")
	_, _ = fmt.Fprintf(w, "User ID: %s\n", idClaims.Subject)

	if idClaims.OrgID != "" {
		_, _ = fmt.Fprintf(w, "Org ID: %s\n", idClaims.OrgID)
	}

	if idClaims.Email != "" {
		_, _ = fmt.Fprintf(w, "Email: %s\n", idClaims.Email)
	}

	if idClaims.Name != "" {
		_, _ = fmt.Fprintf(w, "Name: %s\n", idClaims.Name)
	}

	apiClient := api.NewClient(ctx, e.APIHost, tok)

	member, err := apiClient.GetMember(ctx, tok.IDClaims().Subject)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, "Org name: %s\n", member.Organization.Name)

	if showTokens {
		_, _ = fmt.Fprintf(w, "Access Token: %s\n", tok.AccessToken)
		_, _ = fmt.Fprintf(w, "ID Token: %s\n", tok.IDToken)
		_, _ = fmt.Fprintf(w, "Refresh Token: %s\n", tok.RefreshToken)
	}
	return nil
}

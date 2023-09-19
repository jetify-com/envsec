package awsfed

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity/types"
	"go.jetpack.io/envsec/internal/envvar"
	"go.jetpack.io/envsec/internal/filecache"
	"go.jetpack.io/pkg/sandbox/auth/session"
)

const cacheKey = "awsfed"

type AWSFed struct {
	AccountID      string
	IdentityPoolID string
	LegacyProvider string
	Provider       string
	Region         string
}

func New() *AWSFed {
	return &AWSFed{
		AccountID:      "984256416385",
		IdentityPoolID: "us-west-2:8111c156-085b-4ac5-b94d-f823205f6261",
		LegacyProvider: "auth.jetpack.io",
		Provider: envvar.Get(
			"ENVSEC_AUTH_DOMAIN",
			"accounts.jetpack.io",
		),
		Region: "us-west-2",
	}
}

func (a *AWSFed) AWSCreds(
	ctx context.Context,
	tok *session.Token,
) (*types.Credentials, error) {
	cache := filecache.New("envsec")
	if cachedCreds, err := cache.Get(cacheKey); err == nil {
		var creds types.Credentials
		if err := json.Unmarshal(cachedCreds, &creds); err == nil {
			return &creds, nil
		}
	}

	svc := cognitoidentity.New(cognitoidentity.Options{
		Region: a.Region,
	})

	logins := map[string]string{}
	if tok.IDClaims() == nil {
		// skip
	} else if tok.IDClaims().Issuer == fmt.Sprintf("https://%s/", a.LegacyProvider) {
		logins[a.LegacyProvider] = tok.IDToken
	} else {
		logins[a.Provider] = tok.IDToken
	}

	getIdoutput, err := svc.GetId(
		ctx,
		&cognitoidentity.GetIdInput{
			AccountId:      &a.AccountID,
			IdentityPoolId: &a.IdentityPoolID,
			Logins:         logins,
		},
	)
	if err != nil {
		return nil, err
	}

	output, err := svc.GetCredentialsForIdentity(
		ctx,
		&cognitoidentity.GetCredentialsForIdentityInput{
			IdentityId: getIdoutput.IdentityId,
			Logins:     logins,
		},
	)
	if err != nil {
		return nil, err
	}

	if creds, err := json.Marshal(output.Credentials); err != nil {
		return nil, err
	} else if err := cache.SetT(
		cacheKey,
		creds,
		*output.Credentials.Expiration,
	); err != nil {
		return nil, err
	}

	return output.Credentials, nil
}

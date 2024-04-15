package awsfed

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity/types"
	"go.jetpack.io/pkg/auth/session"
	"go.jetpack.io/pkg/envvar"
	"go.jetpack.io/pkg/filecache"
)

const cacheKeyPrefix = "awsfed"

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
			"accounts.jetify.com",
		),
		Region: "us-west-2",
	}
}

func (a *AWSFed) AWSCredsWithLocalCache(
	ctx context.Context,
	tok *session.Token,
) (*types.Credentials, error) {
	cache := filecache.New[*types.Credentials]("envsec/aws-creds")
	return cache.GetOrSetWithTime(
		cacheKey(tok),
		func() (*types.Credentials, time.Time, error) {
			outputCreds, err := a.AWSCreds(ctx, tok.IDToken)
			if err != nil {
				return nil, time.Time{}, err
			}
			return outputCreds, *outputCreds.Expiration, nil
		},
	)
}

// AWSCreds behaves similar to AWSCredsWithLocalCache but it takes a JWT from input
// rather than reading from a file or cache. This is to allow web services use
// this package without having to write every user's JWT in a cache or a file.
func (a *AWSFed) AWSCreds(
	ctx context.Context,
	idToken string,
) (*types.Credentials, error) {
	svc := cognitoidentity.New(
		cognitoidentity.Options{
			Region: a.Region,
		},
	)

	logins := map[string]string{
		a.Provider: idToken,
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

	return output.Credentials, nil
}

func cacheKey(t *session.Token) string {
	id := ""
	if claims := t.IDClaims(); claims != nil && claims.OrgID != "" {
		id = claims.OrgID
	} else {
		id = fmt.Sprintf("%x", sha256.Sum256([]byte(t.IDToken)))
	}

	return fmt.Sprintf("%s-%s", cacheKeyPrefix, id)
}

package awsfed

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentity/types"
	"github.com/golang-jwt/jwt/v5"
)

type AWSFed struct {
	AccountId      string
	IdentityPoolId string
	Provider       string
	Region         string
}

func New() *AWSFed {
	return &AWSFed{
		AccountId:      "984256416385",
		IdentityPoolId: "us-west-2:8111c156-085b-4ac5-b94d-f823205f6261",
		Provider:       "auth.jetpack.io",
		Region:         "us-west-2",
	}
}

func (a *AWSFed) AWSCreds(
	ctx context.Context,
	token *jwt.Token,
) (*types.Credentials, error) {
	svc := cognitoidentity.New(cognitoidentity.Options{
		Region: a.Region,
	})

	logins := map[string]string{a.Provider: token.Raw}
	getIdoutput, err := svc.GetId(
		ctx,
		&cognitoidentity.GetIdInput{
			AccountId:      &a.AccountId,
			IdentityPoolId: &a.IdentityPoolId,
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

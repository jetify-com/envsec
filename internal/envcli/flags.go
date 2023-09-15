// Copyright 2023 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"context"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetpack.io/auth/session"
	"go.jetpack.io/envsec"
	"go.jetpack.io/envsec/internal/awsfed"
	"go.jetpack.io/envsec/internal/jetcloud"
	"go.jetpack.io/envsec/internal/typeids"
)

// to be composed into xyzCmdFlags structs
type configFlags struct {
	projectId string
	orgId     string
	envName   string
}

func (f *configFlags) register(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(
		&f.projectId,
		"project-id",
		"",
		"Project id to namespace secrets by",
	)

	cmd.PersistentFlags().StringVar(
		&f.orgId,
		"org-id",
		"",
		"Organization id to namespace secrets by",
	)

	cmd.PersistentFlags().StringVar(
		&f.envName,
		"environment",
		"dev",
		"Environment name, such as dev or prod",
	)
}

func (f *configFlags) validateProjectID(orgID typeids.OrganizationID) (string, error) {
	if f.projectId != "" {
		return f.projectId, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", errors.WithStack(err)
	}
	config, err := jetcloud.ProjectConfig(wd)
	if errors.Is(err, os.ErrNotExist) {
		return "", errors.Errorf(
			"Project ID not specified. You must run `envsec init` or specify --project-id in this directory",
		)
	} else if err != nil {
		return "", errors.WithStack(err)
	}

	if config.OrgID != orgID {
		// Validate that the project ID belongs to the org ID
		return "", errors.Errorf(
			"Project ID %s does not belong to organization %s",
			config.ProjectID,
			orgID,
		)
	}
	return config.ProjectID.String(), nil
}

type cmdConfig struct {
	Store envsec.Store
	EnvId envsec.EnvId
}

func (f *configFlags) genConfig(ctx context.Context) (*cmdConfig, error) {
	var tok *session.Token
	var err error

	if f.orgId == "" {
		client, err := newAuthClient()
		if err != nil {
			return nil, err
		}

		tok = client.GetSession()
		if tok == nil {
			return nil, errors.Errorf(
				"To use envsec you must log in (`envsec auth login`) or specify --project-id and --org-id",
			)
		}
	}

	ssmConfig, err := genSSMConfigForUser(ctx, tok)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	s, err := envsec.NewStore(ctx, ssmConfig)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if tok != nil && f.orgId == "" {
		f.orgId = tok.IDClaims().OrgID
	}

	orgID, err := typeids.OrganizationIDFromString(f.orgId)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	projectID, err := f.validateProjectID(orgID)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	envid, err := envsec.NewEnvId(projectID, f.orgId, f.envName)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &cmdConfig{
		Store: s,
		EnvId: envid,
	}, nil
}

func genSSMConfigForUser(
	ctx context.Context,
	tok *session.Token,
) (*envsec.SSMConfig, error) {
	if tok == nil {
		return &envsec.SSMConfig{}, nil
	}
	fed := awsfed.New()
	creds, err := fed.AWSCreds(ctx, tok)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &envsec.SSMConfig{
		AccessKeyId:     *creds.AccessKeyId,
		SecretAccessKey: *creds.SecretKey,
		SessionToken:    *creds.SessionToken,
		Region:          fed.Region,
	}, nil
}

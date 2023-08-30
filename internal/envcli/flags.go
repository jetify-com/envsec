// Copyright 2023 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"context"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec"
	"go.jetpack.io/envsec/internal/auth"
	"go.jetpack.io/envsec/internal/awsfed"
	"go.jetpack.io/envsec/internal/jetcloud"
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

func (f *configFlags) validateProjectID() (string, error) {
	if f.projectId != "" {
		return f.projectId, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", errors.WithStack(err)
	}
	id, err := jetcloud.ProjectID(wd)
	if errors.Is(err, os.ErrNotExist) {
		return "", errors.Errorf(
			"Project ID not specified. You must run `envsec init` or specify --project-id in this directory",
		)
	} else if err != nil {
		return "", errors.WithStack(err)
	}
	return id.String(), nil
}

type cmdConfig struct {
	Store envsec.Store
	EnvId envsec.EnvId
}

func (f *configFlags) genConfig(ctx context.Context) (*cmdConfig, error) {
	var user *auth.User
	var err error
	if f.orgId == "" {
		user, err = newAuthenticator().GetUser()
		if errors.Is(err, auth.ErrNotLoggedIn) {
			return nil, errors.Errorf(
				"To use envsec you must log in (`envsec auth login`) or specify --project-id and --org-id",
			)
		} else if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	ssmConfig, err := genSSMConfigForUser(ctx, user)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	s, err := envsec.NewStore(ctx, ssmConfig)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if user != nil && f.orgId == "" {
		f.orgId = user.OrgID()
	}

	projectID, err := f.validateProjectID()
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
	user *auth.User,
) (*envsec.SSMConfig, error) {
	if user == nil {
		return &envsec.SSMConfig{}, nil
	}
	fed := awsfed.New()
	creds, err := fed.AWSCreds(ctx, user.AccessToken)
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

// Copyright 2023 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"context"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec"
	"go.jetpack.io/envsec/internal/jetcloud"
	"go.jetpack.io/envsec/internal/typeids"
	"go.jetpack.io/envsec/pkg/awsfed"
	"go.jetpack.io/pkg/sandbox/auth/session"
)

// to be composed into xyzCmdFlags structs
type configFlags struct {
	projectID string
	orgID     string
	envName   string
}

func (f *configFlags) register(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(
		&f.projectID,
		"project-id",
		"",
		"Project id to namespace secrets by",
	)

	cmd.PersistentFlags().StringVar(
		&f.orgID,
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
	if f.projectID != "" {
		return f.projectID, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", errors.WithStack(err)
	}
	config, err := jetcloud.ProjectConfig(wd)
	if errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf(
			"project ID not specified. You must run `envsec init` or specify --project-id in this directory",
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
	EnvID envsec.EnvID
}

func (f *configFlags) genConfig(ctx context.Context) (*cmdConfig, error) {
	var tok *session.Token
	var err error

	if f.orgID == "" {
		client, err := newAuthClient()
		if err != nil {
			return nil, err
		}

		tok, err = client.GetSession(ctx)
		if err != nil {
			return nil, fmt.Errorf(
				"error: %w. To use envsec you must log in (`envsec auth login`) or specify --project-id and --org-id",
				err,
			)
		}
	}

	ssmConfig, err := awsfed.GenSSMConfigFromToken(ctx, tok, true /*useCache*/)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	store, err := envsec.NewStore(ctx, ssmConfig)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if tok != nil && f.orgID == "" {
		f.orgID = tok.IDClaims().OrgID
	}

	orgID, err := typeids.OrganizationIDFromString(f.orgID)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	projectID, err := f.validateProjectID(orgID)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	envid, err := envsec.NewEnvID(projectID, f.orgID, f.envName)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &cmdConfig{
		Store: store,
		EnvID: envid,
	}, nil
}

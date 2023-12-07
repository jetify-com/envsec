// Copyright 2023 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.jetpack.io/envsec"
	"go.jetpack.io/envsec/internal/build"
	"go.jetpack.io/envsec/pkg/awsfed"
	"go.jetpack.io/pkg/auth/session"
	"go.jetpack.io/pkg/envvar"
	"go.jetpack.io/pkg/id"
	"go.jetpack.io/pkg/jetcloud"
	"go.jetpack.io/typeid"
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

func (f *configFlags) validateProjectID(orgID id.OrgID) (string, error) {
	if f.projectID != "" {
		return f.projectID, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", errors.WithStack(err)
	}
	c := jetcloud.Client{APIHost: build.JetpackAPIHost(), IsDev: build.IsDev}
	config, err := c.ProjectConfig(wd)
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

type CmdConfig struct {
	Store    envsec.Store
	EnvID    envsec.EnvID
	EnvNames []string
}

func (f *configFlags) genConfig(cmd *cobra.Command) (*CmdConfig, error) {
	if bootstrappedConfig != nil {
		return bootstrappedConfig, nil
	}

	ctx := cmd.Context()
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

	var store envsec.Store
	if envvar.Bool("ENVSEC_USE_AWS_STORE") {
		// Temporary hack to enable the AWS store
		ssmConfig, err := awsfed.GenSSMConfigFromToken(ctx, tok, true /*useCache*/)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		store, err = envsec.NewStore(ctx, ssmConfig)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	} else {
		store, err = envsec.NewStore(ctx, envsec.NewJetpackAPIConfig(tok.AccessToken))
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if tok != nil && f.orgID == "" {
		f.orgID = tok.IDClaims().OrgID
	}

	orgID, err := typeid.Parse[id.OrgID](f.orgID)
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

	envNames := []string{"dev", "prod", "preview"}
	if cmd.Flags().Changed(environmentFlagName) {
		envNames = []string{envid.EnvName}
	}

	return &CmdConfig{
		Store:    store,
		EnvID:    envid,
		EnvNames: envNames,
	}, nil
}

var bootstrappedConfig *CmdConfig

// BootstrapConfig is used to set the config for all commands that use genConfig
// Useful for using envsec programmatically.
func BootstrapConfig(cmdConfig *CmdConfig) {
	bootstrappedConfig = cmdConfig
}

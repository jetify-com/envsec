// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envsec

import (
	"context"
	"path"

	"github.com/pkg/errors"
)

const pathPrefix = "/jetpack-data/env"

// Uniquely identifies an environment in which we store environment variables.
type EnvID struct {
	// A string that uniquely identifies the project to which the environment belongs.
	ProjectID string
	// A string that uniquely identifies the organization to which the environment belongs.
	OrgID string
	// A name that uniquely identifies the environment within the project.
	// Usually one of: 'dev', 'prod'.
	EnvName string
}

func NewEnvID(projectID string, orgID string, envName string) (EnvID, error) {
	if projectID == "" {
		return EnvID{}, errors.New("ProjectId can not be empty")
	}
	return EnvID{
		ProjectID: projectID,
		OrgID:     orgID,
		EnvName:   envName,
	}, nil
}

type Store interface {
	// List all environmnent variables and their values associated with the given envId.
	List(ctx context.Context, envID EnvID) ([]EnvVar, error)
	// Set the value of an environment variable.
	Set(ctx context.Context, envID EnvID, name string, value string) error
	// Set the values of multiple environment variables.
	SetAll(ctx context.Context, envID EnvID, values map[string]string) error
	// Get the value of an environment variable.
	Get(ctx context.Context, envID EnvID, name string) (string, error)
	// Get the values of multiple environment variables.
	GetAll(ctx context.Context, envID EnvID, names []string) ([]EnvVar, error)
	// Delete an environment variable.
	Delete(ctx context.Context, envID EnvID, name string) error
	// Delete multiple environment variables.
	DeleteAll(ctx context.Context, envID EnvID, names []string) error
}

type EnvVar struct {
	Name  string
	Value string
}

func NewStore(ctx context.Context, config Config) (Store, error) {
	switch config := config.(type) {
	case *SSMConfig:
		return newSSMStore(ctx, config)
	case *JetpackAPIConfig:
		return newJetpackAPIStore(config), nil
	default:
		return nil, errors.Errorf("unsupported store type: %T", config)
	}
}

type Config interface {
	IsEnvStoreConfig() bool
}

type SSMConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	KmsKeyID        string

	VarPathFn       func(envId EnvID, varName string) string
	PathNamespaceFn func(envId EnvID) string
}

// SSMStore implements interface Config (compile-time check)
var _ Config = (*SSMConfig)(nil)

func (c *SSMConfig) IsEnvStoreConfig() bool {
	return true
}

func (c *SSMConfig) varPath(envID EnvID, varName string) string {
	if c.VarPathFn != nil {
		return c.VarPathFn(envID, varName)
	}
	return path.Join(
		c.pathNamespace(envID),
		envID.ProjectID,
		envID.EnvName,
		varName,
	)
}

func (c *SSMConfig) pathNamespace(envID EnvID) string {
	if c.PathNamespaceFn != nil {
		return c.PathNamespaceFn(envID)
	}
	return path.Join(pathPrefix, envID.OrgID)
}

func (c *SSMConfig) hasDefaultPaths() bool {
	return c.VarPathFn == nil && c.PathNamespaceFn == nil
}

type JetpackAPIConfig struct {
	endpoint string
	token    string
}

// prodJetpackAPIEndpoint is used for production.
const prodJetpackAPIEndpoint = "https://api.jetpack.io"

// localhostJetpackAPIEndpoint is used for local development.
// const localhostJetpackAPIEndpoint = "http://localhost:8080"

// JetpackAPIStore implements interface Config (compile-time check)
var _ Config = (*JetpackAPIConfig)(nil)

func (c *JetpackAPIConfig) IsEnvStoreConfig() bool {
	return true
}

func NewJetpackAPIConfig(token string) *JetpackAPIConfig {
	return &JetpackAPIConfig{
		endpoint: prodJetpackAPIEndpoint,
		// endpoint: localhostJetpackAPIEndpoint,
		token: token,
	}
}

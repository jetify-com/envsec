// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envsec

import (
	"context"

	"github.com/pkg/errors"
)

// Uniquely identifies an environment in which we store environment variables.
type EnvId struct {
	// A string that uniquely identifies the project to which the environment belongs.
	ProjectId string
	// A string that uniquely identifies the organization to which the environment belongs.
	OrgId string
	// A name that uniquely identifies the environment within the project.
	// Usually one of: 'dev', 'prod'.
	EnvName string
}

func NewEnvId(projectId string, orgId string, envName string) (EnvId, error) {
	if projectId == "" {
		return EnvId{}, errors.New("ProjectId can not be empty")
	}
	return EnvId{
		ProjectId: projectId,
		OrgId:     orgId,
		EnvName:   envName,
	}, nil
}

type Store interface {
	// List all environmnent variables and their values associated with the given envId.
	List(ctx context.Context, envId EnvId) ([]EnvVar, error)
	// Set the value of an environment variable.
	Set(ctx context.Context, envId EnvId, name string, value string) error
	// Set the values of multiple environment variables.
	SetAll(ctx context.Context, envId EnvId, values map[string]string) error
	// Get the value of an environment variable.
	Get(ctx context.Context, envId EnvId, name string) (string, error)
	// Set the values of multiple environment variables.
	GetAll(ctx context.Context, envId EnvId, names []string) ([]EnvVar, error)
	// Delete an environment variable.
	Delete(ctx context.Context, envId EnvId, name string) error
	// Delete multiple environment variables.
	DeleteAll(ctx context.Context, envId EnvId, names []string) error
}

type EnvVar struct {
	Name  string
	Value string
}

func NewStore(ctx context.Context, config Config) (Store, error) {
	switch config := config.(type) {
	case *SSMConfig:
		return newSSMStore(ctx, config)
	default:
		return nil, errors.Errorf("unsupported store type: %T", config)
	}
}

type Config interface {
	IsEnvStoreConfig() bool
}

type SSMConfig struct {
	Region          string
	AccessKeyId     string
	SecretAccessKey string
	SessionToken    string
	KmsKeyId        string
}

// SSMStore implements interface Config (compile-time check)
var _ Config = (*SSMConfig)(nil)

func (c *SSMConfig) IsEnvStoreConfig() bool {
	return true
}

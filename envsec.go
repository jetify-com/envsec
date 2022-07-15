package envsec

import (
	"context"

	"github.com/pkg/errors"
)

// Uniquely identifies an environment in which we store environment variables.
type EnvId struct {
	// A string that uniquely identifies the project to which the environment belongs.
	ProjectId string
	// A name that uniquely identifies the environment within the project.
	// Usually one of: 'dev', 'prod'.
	EnvName string
}

type Store interface {
	// List all environmnent variables and their values associated with the given envId.
	List(ctx context.Context, envId EnvId) (map[string]string, error)
	// Set the value of an environment variable.
	Set(ctx context.Context, envId EnvId, name string, value string) error
	// Set the values of multiple environment variables.
	SetAll(ctx context.Context, envId EnvId, values map[string]string) error
	// Delete an environment variable.
	Delete(ctx context.Context, envId EnvId, name string) error
	// Delete multiple environment variables.
	DeleteAll(ctx context.Context, envId EnvId, names []string) error
}

func NewStore(config Config) (Store, error) {
	switch config.(type) {
	case *SSMConfig:
		// Remove org argument when fully transitioned to projects
		return newSSMStore("jetpack-io", config.(*SSMConfig))
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
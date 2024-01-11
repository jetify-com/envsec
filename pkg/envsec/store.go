package envsec

import (
	"context"
	"errors"

	"go.jetpack.io/pkg/auth/session"
)

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
	// InitForUser initializes the store for current user.
	InitForUser(ctx context.Context, e *Envsec) (*session.Token, error)
}

type EnvVar struct {
	Name  string
	Value string
}

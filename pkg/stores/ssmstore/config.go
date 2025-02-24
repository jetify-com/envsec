package ssmstore

import (
	"path"

	"go.jetify.com/envsec/pkg/envsec"
)

const pathPrefix = "/jetpack-data/env"

type SSMConfig struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	KmsKeyID        string

	VarPathFn       func(envId envsec.EnvID, varName string) string
	PathNamespaceFn func(envId envsec.EnvID) string
}

func (c *SSMConfig) varPath(envID envsec.EnvID, varName string) string {
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

func (c *SSMConfig) pathNamespace(envID envsec.EnvID) string {
	if c.PathNamespaceFn != nil {
		return c.PathNamespaceFn(envID)
	}
	return path.Join(pathPrefix, envID.OrgID)
}

func (c *SSMConfig) hasDefaultPaths() bool {
	return c.VarPathFn == nil && c.PathNamespaceFn == nil
}

package envsec

import (
	"path"
)

const PATH_PREFIX = "/jetpack-data/env"

func GetVarPath(envId EnvId, varName string) string {
	return path.Join(projectPath(envId), envId.EnvName, varName)
}

func projectPath(envId EnvId) string {
	return path.Join(PATH_PREFIX, envId.ProjectId)
}

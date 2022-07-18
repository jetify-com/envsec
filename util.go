package envsec

import (
	"path"
)

func (e *EnvId) Path() string {
	return path.Join(e.ProjectId, e.EnvName)
}

func GetVarPath(envId EnvId, varName string) string {
	return path.Join(PATH_PREFIX, envId.Path(), varName)
}

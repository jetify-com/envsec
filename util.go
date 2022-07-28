package envsec

import (
	"path"
	"strings"
)

const PATH_PREFIX = "/jetpack-data/env"

func varPath(envId EnvId, varName string) string {
	return path.Join(projectPath(envId), envId.EnvName, varName)
}

func projectPath(envId EnvId) string {
	return path.Join(PATH_PREFIX, envId.ProjectId)
}

func envPath(envId EnvId) string {
	return path.Join(projectPath(envId), envId.EnvName)
}

func nameFromPath(path string) string {
	subpaths := strings.Split(path, "/")
	if len(subpaths) == 0 {
		return ""
	}
	return subpaths[len(subpaths)-1]
}

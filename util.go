// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envsec

import (
	"path"
	"strings"
)

const PATH_PREFIX = "/jetpack-data/env"

// varPath is the path for a given variable.
// TODO: Allow customization of this function so that launchpad can use it.
// Launchpad does projectID/envName/varName.
func varPath(envId EnvId, varName string) string {
	return path.Join(pathNamespace(envId), envId.ProjectId, envId.EnvName, varName)
}

// pathNamespace is the path prefix for all variables for a given organization.
// TODO: Allow customization of this function so that launchpad can use it.
// Launchpad uses projectID as prefix.
func pathNamespace(envId EnvId) string {
	return path.Join(PATH_PREFIX, envId.OrgId)
}

func nameFromPath(path string) string {
	subpaths := strings.Split(path, "/")
	if len(subpaths) == 0 {
		return ""
	}
	return subpaths[len(subpaths)-1]
}

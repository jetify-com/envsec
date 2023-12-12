// Copyright 2023 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package build

import (
	"os"
	"strconv"
)

var forceProd, _ = strconv.ParseBool(os.Getenv("ENVSEC_PROD"))

// These variables are set by the build script.
var (
	IsDev      = Version == "0.0.0-dev" && !forceProd
	Version    = "0.0.0-dev"
	Commit     = "none"
	CommitDate = "unknown"
)

func Issuer() string {
	if IsDev {
		return "https://laughing-agnesi-vzh2rap9f6.projects.oryapis.com"
	}
	return "https://accounts.jetpack.io"
}

func ClientID() string {
	if IsDev {
		return "3945b320-bd31-4313-af27-846b67921acb"
	}
	return "ff3d4c9c-1ac8-42d9-bef1-f5218bb1a9f6"
}

func JetpackAPIHost() string {
	if IsDev {
		return "https://api.jetpack.dev"
	}
	return "https://api.jetpack.io"
}

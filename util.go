// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envsec

import (
	"strings"
)

func nameFromPath(path string) string {
	subpaths := strings.Split(path, "/")
	if len(subpaths) == 0 {
		return ""
	}
	return subpaths[len(subpaths)-1]
}

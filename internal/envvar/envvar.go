// Copyright 2023 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envvar

import (
	"os"
)

// Get gets the value of an environment variable.
// If it's empty, it will return the given default value instead.
func Get(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		val = def
	}

	return val
}

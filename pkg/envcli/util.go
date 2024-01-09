// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"os"

	"github.com/pkg/errors"
)

func fileExists(path string) (bool, error) {
	fileinfo, err := os.Stat(path)
	if err == nil {
		if !fileinfo.IsDir() {
			// It is a file!
			return true, nil
		}
		// It is a directory
		return false, nil
	}

	// No such file was found:
	if err != nil && os.IsNotExist(err) {
		return false, nil
	}

	// Some other error:
	return false, errors.WithStack(err)
}

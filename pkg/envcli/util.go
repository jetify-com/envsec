// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envcli

import (
	"context"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go.jetpack.io/envsec/pkg/envsec"
)

const nameRegexStr = "^[a-zA-Z_][a-zA-Z0-9_]*"

var nameRegex = regexp.MustCompile(nameRegexStr)

func SetEnvMap(ctx context.Context, store envsec.Store, envID envsec.EnvID, envMap map[string]string) error {
	err := ensureValidNames(lo.Keys(envMap))
	if err != nil {
		return errors.WithStack(err)
	}

	err = store.SetAll(ctx, envID, envMap)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}

func ensureValidNames(names []string) error {
	for _, name := range names {

		// Any variation of jetpack_ or JETPACK_ prefix is not allowed
		lowerName := strings.ToLower(name)
		if strings.HasPrefix(lowerName, "jetpack_") {
			return errors.Errorf(
				"name %s cannot start with JETPACK_ (or lowercase)",
				name,
			)
		}

		if !nameRegex.MatchString(name) {
			return errors.Errorf(
				"name %s must match the regular expression: %s ",
				name,
				nameRegexStr,
			)
		}
	}
	return nil
}

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

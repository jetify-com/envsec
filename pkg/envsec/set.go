package envsec

import (
	"context"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go.jetpack.io/envsec/internal/tux"
)

func (e *Envsec) Set(ctx context.Context, name string, value string) error {
	return e.SetMap(ctx, map[string]string{name: value})
}

func (e *Envsec) SetMap(ctx context.Context, envMap map[string]string) error {
	err := ensureValidNames(lo.Keys(envMap))
	if err != nil {
		return errors.WithStack(err)
	}

	err = e.Store.SetAll(ctx, e.EnvID, envMap)
	if err != nil {
		return errors.WithStack(err)
	}
	insertedNames := lo.Keys(envMap)
	return tux.WriteHeader(e.Stderr,
		"[DONE] Set environment %s %v in environment: %s\n",
		tux.Plural(insertedNames, "variable", "variables"),
		strings.Join(tux.QuotedTerms(insertedNames), ", "),
		strings.ToLower(e.EnvID.EnvName),
	)
}

func (e *Envsec) SetFromArgs(ctx context.Context, args []string) error {
	envMap, err := parseSetArgs(args)
	if err != nil {
		return errors.WithStack(err)
	}
	return e.SetMap(ctx, envMap)
}

func ValidateSetArgs(args []string) error {
	for _, arg := range args {
		k, _, ok := strings.Cut(arg, "=")
		if !ok || k == "" {
			return errors.Errorf(
				"argument %s must have an '=' to be of the form NAME=VALUE",
				arg,
			)
		}
	}

	return nil
}

func parseSetArgs(args []string) (map[string]string, error) {
	envMap := map[string]string{}
	for _, arg := range args {
		key, val, _ := strings.Cut(arg, "=")
		if strings.HasPrefix(val, "\\@") {
			val = strings.TrimPrefix(val, "\\")
		} else if strings.HasPrefix(val, "@") {
			file := strings.TrimPrefix(val, "@")
			if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
				return nil, errors.Errorf(
					"@ syntax is used for setting a secret from a file. file %s "+
						"does not exist. If your value starts with @, escape it with "+
						"a backslash, e.g. %s='\\%s'",
					file,
					key,
					val,
				)
			}
			c, err := os.ReadFile(file)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to read file %s", file)
			}
			val = string(c)
		}
		envMap[key] = val
	}
	return envMap, nil
}

const nameRegexStr = "^[a-zA-Z_][a-zA-Z0-9_]*"

var nameRegex = regexp.MustCompile(nameRegexStr)

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

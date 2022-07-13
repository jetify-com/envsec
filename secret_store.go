package envsec

import (
	"encoding/base64"
	"os"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"go.jetpack.io/axiom/opensource/cmd/jetpack/viewer"
	"go.jetpack.io/axiom/opensource/kubevalidate"
	"go.jetpack.io/axiom/opensource/proto/api"
)

// TODO savil. Delete.
type SecretFile struct {
	// The name of the stored parameter:
	//    * Parameter names are case sensitive.
	//    * Parameter names may only include the following symbols and letters a-zA-Z0-9_.-
	//    * A parameter name can't include spaces.
	Path string

	// User supplied description of parameter.
	Description string

	// Last modification date and time.
	LastModified *time.Time
}

type envSecType string

const (
	envSecType_EnvVar envSecType = "ENVIRONMENT_VARIABLE"
	envSecType_File   envSecType = "FILE"
)

type EnvStore struct {
	store *parameterStore
}

func NewEnvStore(vc viewer.Context, config *api.ParameterStoreConfig) (*EnvStore, error) {
	var p string
	if vc.OrgDomain() != "" {
		domain, err := kubevalidate.ToValidName(vc.OrgDomain())
		if err != nil {
			return nil, errors.WithStack(err)
		}

		p = path.Join("/jetpack.io/secrets", domain)
	} else {
		email, err := kubevalidate.ToValidName(vc.Email())
		if err != nil {
			return nil, errors.WithStack(err)
		}

		p = path.Join("/jetpack.io/secrets", email)
	}

	store := &EnvStore{
		store: newParameterStore(config, p),
	}
	return store, nil
}

// Returns names of stored env-vars
func (s *EnvStore) List(vc viewer.Context, environment string) ([]string, error) {
	filters := buildParameterFilters(envSecType_EnvVar, vc, environment)
	parameters, err := s.store.listParameters(vc, filters)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var names []string
	for _, p := range parameters {
		if name, defined := p.resolveParameterTag("name"); defined {
			names = append(names, name)
		}
	}
	return names, nil
}

// Returns values associated with requested env-vars
func (s *EnvStore) Get(vc viewer.Context, environment string, names []string) (map[string]string, error) {
	filters := buildParameterFilters(envSecType_EnvVar, vc, environment)
	filters = append(filters, types.ParameterStringFilter{
		Key:    aws.String("tag:name"),
		Values: names,
	})

	parameters, err := s.store.listParameters(vc, filters)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(parameters) == 0 {
		return map[string]string{}, nil
	}

	values, err := s.store.loadParametersValues(vc, parameters)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	result := map[string]string{}
	for id, value := range values {
		for _, p := range parameters {
			if p.id == id {
				if name, defined := p.resolveParameterTag("name"); defined {
					result[name] = value
				}
			}
		}
	}
	return result, nil
}

// Stores or updates an env-var
func (s *EnvStore) Set(
	vc viewer.Context,
	environment string,
	projectID string,
	name string,
	value string,
) error {
	secretTags, err := buildSecretTags(vc, environment)
	if err != nil {
		return errors.WithStack(err)
	}
	// appending project ID tag to secret tags
	secretTags["ProjectID"] = projectID

	filters := buildParameterFilters(envSecType_EnvVar, vc, environment)
	filters = append(filters, types.ParameterStringFilter{
		Key:    aws.String("tag:name"),
		Values: []string{name},
	})

	parameters, err := s.store.listParameters(vc, filters)
	if err != nil {
		return errors.WithStack(err)
	}

	if len(parameters) == 0 {
		tags := buildParameterTags(envSecType_EnvVar, secretTags)
		tags = append(tags, types.Tag{
			Key: aws.String("name"), Value: aws.String(name),
		})

		// New parameter definition
		parameter := &parameter{
			tags: tags,
		}
		return s.store.newParameter(vc, parameter, value)
	}

	if len(parameters) == 1 {
		// Parameter with the same name is already defined
		parameter := parameters[0]
		return s.store.storeParameterValue(vc, parameter, value)
	}

	// Internal error: duplicate ambiguous definitions
	return errors.WithStack(errors.Errorf("duplicate definitions for environment secret %s", name))
}

// TODO savil. Rename to Delete.
// Deletes stored environment secrets
func (s *EnvStore) DeleteEnvironmentSecrets(vc viewer.Context, environment string, names []string) error {
	filters := buildParameterFilters(envSecType_EnvVar, vc, environment)
	filters = append(filters, types.ParameterStringFilter{
		Key:    aws.String("tag:name"),
		Values: names,
	})

	parameters, err := s.store.listParameters(vc, filters)
	if err != nil {
		return errors.WithStack(err)
	}

	if 0 < len(parameters) {
		return s.store.deleteParameters(vc, parameters)
	}
	return nil
}

// TODO savil. Delete.
// Returns information about stored secret files
func (s *EnvStore) ListSecretFiles(vc viewer.Context, environment string) ([]*SecretFile, error) {
	filters := buildParameterFilters(envSecType_File, vc, environment)
	parameters, err := s.store.listParameters(vc, filters)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var variables []*SecretFile
	for _, p := range parameters {
		if path, defined := p.resolveParameterTag("filename"); defined {
			variables = append(variables, &SecretFile{
				Path:         path,
				Description:  p.description,
				LastModified: p.lastModified,
			})
		}
	}
	return variables, nil
}

// Returns content associated with a set of secret files
func (s *EnvStore) LoadSecretFilesContent(vc viewer.Context, environment string, filenames []string) (map[string]string, error) {
	filters := buildParameterFilters(envSecType_File, vc, environment)
	filters = append(filters, types.ParameterStringFilter{
		Key:    aws.String("tag:filename"),
		Values: filenames,
	})

	parameters, err := s.store.listParameters(vc, filters)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if 0 < len(parameters) {
		values, err := s.store.loadParametersValues(vc, parameters)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		result := map[string]string{}
		for id, content := range values {
			for _, p := range parameters {
				if p.id == id {
					if filename, defined := p.resolveParameterTag("filename"); defined {
						result[filename] = content
					}
				}
			}
		}
		return result, nil
	}
	return map[string]string{}, nil
}

// TODO savil. Delete.
// Stores or updates an environment secret
func (s *EnvStore) StoreSecretFile(vc viewer.Context, environment string, v *SecretFile) error {
	secretTags, err := buildSecretTags(vc, environment)
	if err != nil {
		return errors.WithStack(err)
	}

	filename := path.Base(v.Path)
	filters := buildParameterFilters(envSecType_File, vc, environment)
	filters = append(filters, types.ParameterStringFilter{
		Key:    aws.String("tag:filename"),
		Values: []string{filename},
	})

	parameters, err := s.store.listParameters(vc, filters)
	if err != nil {
		return errors.WithStack(err)
	}

	content, err := os.ReadFile(v.Path)
	if err != nil {
		return errors.WithStack(err)
	}

	if len(parameters) == 0 {
		tags := buildParameterTags(envSecType_File, secretTags)
		tags = append(tags, types.Tag{
			Key: aws.String("filename"), Value: aws.String(filename),
		})

		// New parameter definition
		parameter := &parameter{
			description: v.Description,
			tags:        tags,
		}
		return s.store.newParameter(vc, parameter, base64.StdEncoding.EncodeToString(content))
	}

	if len(parameters) == 1 {
		// Parameter with the same name is already defined
		parameter := parameters[0]
		if v.Description != "" {
			// TODO: find a way to remove the description if the user so desires
			parameter.description = v.Description
		}
		return s.store.storeParameterValue(vc, parameter, base64.StdEncoding.EncodeToString(content))
	}

	// Internal error: duplicate ambiguous definitions
	return errors.WithStack(errors.Errorf("duplicate definitions for secret file %s", filename))
}

// TODO savil. Delete.
// Deletes a stored environment secret
func (s *EnvStore) DeleteSecretFile(vc viewer.Context, environment string, filename string) error {
	filters := buildParameterFilters(envSecType_File, vc, environment)
	filters = append(filters, types.ParameterStringFilter{
		Key:    aws.String("tag:filename"),
		Values: []string{filename},
	})

	parameters, err := s.store.listParameters(vc, filters)
	if err != nil {
		return errors.WithStack(err)
	}

	vc.Logger().BoldPrintf("Parameters %v\n", parameters)

	if 0 < len(parameters) {
		return s.store.deleteParameters(vc, parameters)
	}
	return nil
}

func buildParameterFilters(secretKind envSecType, vc viewer.Context, environment string) []types.
	ParameterStringFilter {
	filters := []types.ParameterStringFilter{{
		Key:    aws.String("tag:kind"),
		Values: []string{string(secretKind)},
	}}
	if vc.OrgDomain() != "" {
		filters = append(filters, types.ParameterStringFilter{
			Key:    aws.String("tag:org"),
			Values: []string{vc.OrgDomain()},
		})
	} else {
		filters = append(filters, types.ParameterStringFilter{
			Key:    aws.String("tag:email"),
			Values: []string{vc.Email()},
		})
	}
	if environment != "" {
		filters = append(filters, types.ParameterStringFilter{
			Key:    aws.String("tag:environment"),
			Values: []string{environment},
		})
	}
	return filters
}

func buildParameterTags(secretKind envSecType, secretTags map[string]string) []types.Tag {
	var parameterTags []types.Tag
	for tag, value := range secretTags {
		parameterTags = append(parameterTags, types.Tag{
			Key: aws.String(tag), Value: aws.String(value),
		})
	}
	parameterTags = append(parameterTags, types.Tag{
		Key: aws.String("kind"), Value: aws.String(string(secretKind)),
	})
	return parameterTags
}

func buildSecretTags(vc viewer.Context, environment string) (map[string]string, error) {
	var tags map[string]string
	if vc.OrgDomain() != "" {
		tags = map[string]string{
			"org":   vc.OrgDomain(),
			"email": vc.Email(),
		}
	} else {
		tags = map[string]string{
			"email": vc.Email(),
		}
	}

	if environment != "" {
		tags["environment"] = environment
	}
	return tags, nil
}

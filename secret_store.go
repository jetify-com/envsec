package envsec

import (
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"go.jetpack.io/axiom/opensource/cmd/jetpack/viewer"
	"go.jetpack.io/axiom/opensource/kubevalidate"
)

type envSecType string

const (
	envSecType_EnvVar envSecType = "ENVIRONMENT_VARIABLE"
)

type EnvStore struct {
	store *parameterStore
}

func NewEnvStore(vc viewer.Context, config *ParameterStoreConfig) (*EnvStore, error) {
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

func (s *EnvStore) Get(vc viewer.Context, environment string) (map[string]string, error) {
	filters := buildParameterFilters(envSecType_EnvVar, vc, environment)

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
	return errors.WithStack(errors.Errorf("duplicate definitions for environment variable %s", name))
}

// Deletes stored environment
func (s *EnvStore) Delete(vc viewer.Context, environment string, names []string) error {
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

func buildParameterFilters(kind envSecType, vc viewer.Context, environment string) []types.ParameterStringFilter {

	filters := []types.ParameterStringFilter{{
		Key:    aws.String("tag:kind"),
		Values: []string{string(kind)},
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

func buildParameterTags(kind envSecType, secretTags map[string]string) []types.Tag {
	var parameterTags []types.Tag
	for tag, value := range secretTags {
		parameterTags = append(parameterTags, types.Tag{
			Key: aws.String(tag), Value: aws.String(value),
		})
	}
	parameterTags = append(parameterTags, types.Tag{
		Key: aws.String("kind"), Value: aws.String(string(kind)),
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

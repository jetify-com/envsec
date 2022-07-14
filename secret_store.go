package envsec

import (
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"go.jetpack.io/axiom/opensource/cmd/jetpack/viewer"
	"go.jetpack.io/axiom/opensource/kubevalidate"
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

	paramStore, err := newParameterStore(config, p)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	store := &EnvStore{
		store: paramStore,
	}
	return store, nil
}

func (s *EnvStore) List(vc viewer.Context, environment string) (map[string]string, error) {
	filters := buildParameterFilters(vc, environment)

	parameters, err := s.store.listParameters(vc, filters)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if len(parameters) == 0 {
		return map[string]string{}, nil
	}

	// We need to loadParameterValues in chunks of 10, due to AWS API limits
	paramChunks := lo.Chunk(parameters, 10)
	valueChunks := []map[string]string{}
	for _, paramChunk := range paramChunks {
		values, err := s.store.loadParametersValues(vc, paramChunk)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		valueChunks = append(valueChunks, values)
	}
	values := lo.Assign(valueChunks...)

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

	filters := buildParameterFilters(vc, environment)
	filters = append(filters, types.ParameterStringFilter{
		Key:    aws.String("tag:name"),
		Values: []string{name},
	})

	parameters, err := s.store.listParameters(vc, filters)
	if err != nil {
		return errors.WithStack(err)
	}

	if len(parameters) == 0 {
		tags := buildParameterTags(secretTags)
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
	filters := buildParameterFilters(vc, environment)
	filters = append(filters, types.ParameterStringFilter{
		Key:    aws.String("tag:name"),
		Values: names,
	})

	parameters, err := s.store.listParameters(vc, filters)
	if err != nil {
		return errors.WithStack(err)
	}

	if 0 == len(parameters) {
		// early return, we are done
		return nil
	}

	paramChunks := lo.Chunk(parameters, 10)
	for _, paramChunk := range paramChunks {
		err = s.store.deleteParameters(vc, paramChunk)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func buildParameterFilters(vc viewer.Context, environment string) []types.ParameterStringFilter {

	filters := []types.ParameterStringFilter{}
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

func buildParameterTags(secretTags map[string]string) []types.Tag {
	var parameterTags []types.Tag
	for tag, value := range secretTags {
		parameterTags = append(parameterTags, types.Tag{
			Key: aws.String(tag), Value: aws.String(value),
		})
	}
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

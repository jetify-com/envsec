package envsec

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

// This dummy project id is temporary until the project ID in jetconfig.yaml
// comes out from behind feature gate.
const DUMMY_PROJECT_ID = "proj_00000000"

type SSMStore struct {
	store *parameterStore
}

// SSMStore implements interface Store (compile-time check)
var _ Store = (*SSMStore)(nil)

func newSSMStore(ctx context.Context, config *SSMConfig) (*SSMStore, error) {
	paramStore, err := newParameterStore(ctx, config)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	store := &SSMStore{
		store: paramStore,
	}
	return store, nil
}

func (s *SSMStore) List(ctx context.Context, envId EnvId) (map[string]string, error) {

	// TODO Reconcile the filters in buildParameterFilters and in listParameters.
	// Lets unify them in one function.
	filters := buildParameterFilters(envId)

	parameters, err := s.store.listParameters(ctx, projectPath(envId), filters)
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
		values, err := s.store.loadParametersValues(ctx, projectPath(envId), paramChunk)
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
func (s *SSMStore) Set(
	ctx context.Context,
	envId EnvId,
	name string,
	value string,
) error {
	secretTags := buildSecretTags(envId)
	parameterKey := GetVarPath(envId, name)

	// New parameter definition
	tags := buildParameterTags(secretTags)
	tags = append(tags, types.Tag{
		Key: aws.String("name"), Value: aws.String(name),
	})
	parameter := &parameter{
		tags: tags,
		id:   parameterKey,
	}
	return s.store.newParameter(ctx, parameter, value)
}

func (s *SSMStore) SetAll(ctx context.Context, envId EnvId, values map[string]string) error {
	// For now we implement by issuing multiple calls to Set()
	// Make more efficient either by implementing a batch call to the underlying API, or
	// by concurrently calling Set()

	var multiErr error
	for name, value := range values {
		err := s.Set(ctx, envId, name, value)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return multiErr
}

func (s *SSMStore) Delete(ctx context.Context, envId EnvId, name string) error {
	return s.DeleteAll(ctx, envId, []string{name})
}

func (s *SSMStore) DeleteAll(ctx context.Context, envId EnvId, names []string) error {
	filters := buildParameterFilters(envId)
	filters = append(filters, types.ParameterStringFilter{
		Key:    aws.String("tag:name"),
		Values: names,
	})

	parameters, err := s.store.listParameters(ctx, projectPath(envId), filters)
	if err != nil {
		return errors.WithStack(err)
	}

	if len(parameters) == 0 {
		// early return, we are done
		return nil
	}

	paramChunks := lo.Chunk(parameters, 10)
	for _, paramChunk := range paramChunks {
		err = s.store.deleteParameters(ctx, paramChunk)
		if err != nil {
			return errors.WithStack(err)
		}
	}

	return nil
}

func buildParameterFilters(envId EnvId) []types.ParameterStringFilter {
	filters := []types.ParameterStringFilter{}
	if envId.OrgId != "" {
		filters = append(filters, types.ParameterStringFilter{
			Key:    aws.String("tag:org-id"),
			Values: []string{envId.OrgId},
		})
	}
	if envId.EnvName != "" {
		filters = append(filters, types.ParameterStringFilter{
			Key:    aws.String("tag:env-name"),
			Values: []string{envId.EnvName},
		})
	}

	if envId.ProjectId != DUMMY_PROJECT_ID {
		filters = append(
			filters, types.ParameterStringFilter{
				Key:    aws.String("tag:project-id"),
				Values: []string{envId.ProjectId},
			},
		)
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

func buildSecretTags(envId EnvId) map[string]string {
	tags := map[string]string{}
	if envId.OrgId != "" {
		tags["org-id"] = envId.OrgId
	}
	if envId.EnvName != "" {
		tags["env-name"] = envId.EnvName
	}
	// appending project ID tag to secret tags
	tags["project-id"] = envId.ProjectId
	return tags
}

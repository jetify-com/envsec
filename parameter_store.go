package envsec

import (
	"context"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/pkg/errors"
)

type parameter struct {
	id           string
	description  string
	lastModified *time.Time
	tags         []types.Tag
}

func (p *parameter) resolveParameterTag(tag string) (string, bool) {
	parameterTag, defined := lo.Find(p.tags, func(t types.Tag) bool {
		return aws.ToString(t.Key) == tag
	})

	if !defined {
		return "", false
	}
	return aws.ToString(parameterTag.Value), true
}

type ParameterStoreConfig struct {
	Region          string
	AccessKeyId     string
	SecretAccessKey string
	SessionToken    string
	KmsKeyId        string
}

type parameterStore struct {
	config *ParameterStoreConfig
	path   string
	client *ssm.Client
}

// Parameter values are limited in size to 4KB
const parameterValueMaxLength = 4 * 1024

// New parameter store for current user/organization.
func newParameterStore(config *ParameterStoreConfig, path string) *parameterStore {
	return &parameterStore{
		config: config,
		path:   path,
		client: newSsmClient(config),
	}
}

// Returns information about stored parameters.
func (s *parameterStore) listParameters(ctx context.Context, filters []types.ParameterStringFilter) ([]*parameter, error) {
	return s.describeParameters(ctx, filters...)
}

// Returns values associated with a set of parameters
func (s *parameterStore) loadParametersValues(ctx context.Context, parameters []*parameter) (map[string]string, error) {
	if len(parameters) == 0 {
		return map[string]string{}, nil
	}

	var paths []string
	for _, p := range parameters {
		paths = append(paths, p.id)
	}

	// Retrieve stored parameters values
	output, err := s.client.GetParameters(ctx, &ssm.GetParametersInput{
		Names:          paths,
		WithDecryption: true,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "error executing AWS SSM query [path='%v', parameters='%v']", s.path, parameters)
	}

	if 0 < len(output.InvalidParameters) {
		return nil, errors.WithStack(errors.Errorf("invalid parameters supplied: '%v'", parameters))
	}

	values := map[string]string{}
	for _, p := range output.Parameters {
		values[aws.ToString(p.Name)] = aws.ToString(p.Value)
	}
	return values, nil
}

// Defines a new stored parameter.
func (s *parameterStore) newParameter(ctx context.Context, v *parameter, value string) error {
	if parameterValueMaxLength < len(value) {
		return errors.New("parameter values are limited in size to 4KB")
	}

	id, err := generateParameterId(s.path)
	if err != nil {
		return errors.WithStack(err)
	}

	input := &ssm.PutParameterInput{
		Name:        id,
		Description: aws.String(v.description),
		Type:        types.ParameterTypeSecureString,
		Value:       aws.String(value),
		KeyId:       aws.String(s.config.KmsKeyId),
		Tags:        v.tags,
	}

	_, err = s.client.PutParameter(ctx, input)
	return errors.WithStack(err)
}

// Defines or updates a stored parameter.
// parameter values are limited in size to 4 KB.
func (s *parameterStore) storeParameterValue(ctx context.Context, v *parameter, value string) error {
	if parameterValueMaxLength < len(value) {
		return errors.New("parameter values are limited in size to 4KB")
	}

	input := &ssm.PutParameterInput{
		Name:        aws.String(v.id),
		Description: aws.String(v.description),
		Overwrite:   true,
		Value:       aws.String(value),
	}

	_, err := s.client.PutParameter(ctx, input)
	return errors.WithStack(err)
}

// Delete a stored parameter from the system.
func (s *parameterStore) deleteParameters(ctx context.Context, parameters []*parameter) error {
	var paths []string
	for _, p := range parameters {
		paths = append(paths, p.id)
	}

	input := &ssm.DeleteParametersInput{
		Names: paths,
	}
	_, err := s.client.DeleteParameters(ctx, input)
	return errors.WithStack(err)
}

func newSsmClient(config *ParameterStoreConfig) *ssm.Client {
	return ssm.New(
		ssm.Options{
			Region: config.Region,
			Credentials: credentials.NewStaticCredentialsProvider(
				config.AccessKeyId,
				config.SecretAccessKey,
				config.SessionToken,
			),
		},
	)
}

func (s *parameterStore) describeParameters(ctx context.Context, additionalFilters ...types.ParameterStringFilter) ([]*parameter, error) {
	filters := []types.ParameterStringFilter{
		{
			Key:    aws.String("Type"),
			Values: []string{"SecureString"},
		},
	}
	filters = append(filters, additionalFilters...)

	var parameters []*parameter
	result, nextToken, err := s.executeDescribeParametersRequest(ctx, filters, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	parameters = append(parameters, result...)

	for aws.ToString(nextToken) != "" {
		result, nextToken, err = s.executeDescribeParametersRequest(ctx, filters, nextToken)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		parameters = append(parameters, result...)
	}

	return parameters, nil
}

func (s *parameterStore) executeDescribeParametersRequest(ctx context.Context, filters []types.ParameterStringFilter, nextToken *string) ([]*parameter, *string, error) {
	output, err := s.client.DescribeParameters(ctx, &ssm.DescribeParametersInput{
		ParameterFilters: filters,
		NextToken:        nextToken,
	})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error executing AWS SSM query [DescribeParameters '%v']", s.path)
	}

	var parameters []*parameter
	for _, p := range output.Parameters {
		listTagsInput := &ssm.ListTagsForResourceInput{
			ResourceId:   p.Name,
			ResourceType: types.ResourceTypeForTaggingParameter,
		}
		tags, err := s.client.ListTagsForResource(ctx, listTagsInput)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "error executing AWS SSM query [ListTagsForResource '%v']", p.Name)
		}

		parameters = append(parameters, &parameter{
			id:           aws.ToString(p.Name),
			description:  aws.ToString(p.Description),
			lastModified: p.LastModifiedDate,
			tags:         tags.TagList,
		})
	}
	return parameters, output.NextToken, nil
}

func generateParameterId(p string) (*string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return aws.String(path.Join(p, id.String())), nil
}

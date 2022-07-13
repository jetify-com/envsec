package envsec

import (
	"context"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/aws-sdk-go/aws"
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
		return aws.StringValue(t.Key) == tag
	})

	if !defined {
		return "", false
	}
	return aws.StringValue(parameterTag.Value), true
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
}

// Parameter values are limited in size to 4KB
const parameterValueMaxLength = 4 * 1024

// New parameter store for current user/organization.
func newParameterStore(config *ParameterStoreConfig, path string) *parameterStore {
	return &parameterStore{
		config: config,
		path:   path,
	}
}

// Returns information about stored parameters.
func (s *parameterStore) listParameters(ctx context.Context, filters []types.ParameterStringFilter) ([]*parameter, error) {
	client := s.newSsmClient()
	return s.describeParameters(ctx, client, filters...)
}

// Returns values associated with a set of parameters
func (s *parameterStore) loadParametersValues(ctx context.Context, parameters []*parameter) (map[string]string, error) {
	client := s.newSsmClient()

	if len(parameters) == 0 {
		return map[string]string{}, nil
	}

	var paths []string
	for _, p := range parameters {
		paths = append(paths, p.id)
	}

	// Retrieve stored parameters values
	output, err := client.GetParameters(ctx, &ssm.GetParametersInput{
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
		values[aws.StringValue(p.Name)] = aws.StringValue(p.Value)
	}
	return values, nil
}

// Defines a new stored parameter.
func (s *parameterStore) newParameter(ctx context.Context, v *parameter, value string) error {
	if parameterValueMaxLength < len(value) {
		return errors.New("parameter values are limited in size to 4KB")
	}

	client := s.newSsmClient()

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

	_, err = client.PutParameter(ctx, input)
	return errors.WithStack(err)
}

// Defines or updates a stored parameter.
// parameter values are limited in size to 4 KB.
func (s *parameterStore) storeParameterValue(ctx context.Context, v *parameter, value string) error {
	if parameterValueMaxLength < len(value) {
		return errors.New("parameter values are limited in size to 4KB")
	}

	client := s.newSsmClient()

	input := &ssm.PutParameterInput{
		Name:        aws.String(v.id),
		Description: aws.String(v.description),
		Overwrite:   true,
		Value:       aws.String(value),
	}

	_, err := client.PutParameter(ctx, input)
	return errors.WithStack(err)
}

// Delete a stored parameter from the system.
func (s *parameterStore) deleteParameters(ctx context.Context, parameters []*parameter) error {
	client := s.newSsmClient()

	var paths []string
	for _, p := range parameters {
		paths = append(paths, p.id)
	}

	input := &ssm.DeleteParametersInput{
		Names: paths,
	}
	_, err := client.DeleteParameters(ctx, input)
	return errors.WithStack(err)
}

func (s *parameterStore) newSsmClient() *ssm.Client {
	return ssm.New(
		ssm.Options{
			Region: s.config.Region,
			Credentials: credentials.NewStaticCredentialsProvider(
				s.config.AccessKeyId,
				s.config.SecretAccessKey,
				s.config.SessionToken,
			),
		},
	)
}

func (s *parameterStore) describeParameters(ctx context.Context, client *ssm.Client, additionalFilters ...types.ParameterStringFilter) ([]*parameter, error) {
	filters := []types.ParameterStringFilter{
		{
			Key:    aws.String("Type"),
			Values: []string{"SecureString"},
		},
	}
	filters = append(filters, additionalFilters...)

	var parameters []*parameter
	result, nextToken, err := s.executeDescribeParametersRequest(ctx, client, filters, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	parameters = append(parameters, result...)

	for aws.StringValue(nextToken) != "" {
		result, nextToken, err = s.executeDescribeParametersRequest(ctx, client, filters, nextToken)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		parameters = append(parameters, result...)
	}

	return parameters, nil
}

func (s *parameterStore) executeDescribeParametersRequest(ctx context.Context, client *ssm.Client, filters []types.ParameterStringFilter, nextToken *string) ([]*parameter, *string, error) {
	output, err := client.DescribeParameters(ctx, &ssm.DescribeParametersInput{
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
		tags, err := client.ListTagsForResource(ctx, listTagsInput)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "error executing AWS SSM query [ListTagsForResource '%v']", p.Name)
		}

		parameters = append(parameters, &parameter{
			id:           aws.StringValue(p.Name),
			description:  aws.StringValue(p.Description),
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

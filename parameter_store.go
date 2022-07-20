package envsec

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/pkg/errors"
	"github.com/samber/lo"
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

type parameterStore struct {
	config *SSMConfig
	client *ssm.Client
}

// Parameter values are limited in size to 4KB
const parameterValueMaxLength = 4 * 1024

// New parameter store for current user/organization.
func newParameterStore(ctx context.Context, config *SSMConfig) (*parameterStore, error) {
	awsConfig, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	client := ssm.NewFromConfig(awsConfig, func(o *ssm.Options) {
		if config.Region != "" {
			o.Region = config.Region
		}

		if (config.AccessKeyId != "" && config.SecretAccessKey != "") || config.SessionToken != "" {
			o.Credentials = credentials.NewStaticCredentialsProvider(
				config.AccessKeyId,
				config.SecretAccessKey,
				config.SessionToken,
			)
		}
	})

	return &parameterStore{
		config: config,
		client: client,
	}, nil /* no error */
}

// Returns values associated with a set of parameters
func (s *parameterStore) loadParametersValues(
	ctx context.Context,
	pathPrefix string,
	parameters []*parameter,
) (map[string]string, error) {
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
		return nil, errors.Wrapf(err, "error executing AWS SSM query [path='%v' parameters='%v']", pathPrefix,
			parameters)
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
// parameter values are limited in size to 4 KB.
func (s *parameterStore) newParameter(ctx context.Context, v *parameter, value string) error {
	if parameterValueMaxLength < len(value) {
		return errors.New("parameter values are limited in size to 4KB")
	}

	input := &ssm.PutParameterInput{
		Name:        aws.String(v.id),
		Description: aws.String(v.description),
		Type:        types.ParameterTypeSecureString,
		Value:       aws.String(value),
		KeyId:       aws.String(s.config.KmsKeyId),
		Tags:        v.tags,
	}

	_, err := s.client.PutParameter(ctx, input)
	if err != nil {
		var paeError *types.ParameterAlreadyExists
		if errors.As(err, &paeError) {
			// parameter already exists calling put parameter with overwrite flag
			return s.overwriteParameterValue(ctx, v, value)
		}
		return errors.WithStack(err)
	}
	return errors.WithStack(err)
}

// Updates a stored parameter.
func (s *parameterStore) overwriteParameterValue(ctx context.Context, v *parameter, value string) error {

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

// Returns the stored parameters.
func (s *parameterStore) listParameters(
	ctx context.Context,
	path string,
	additionalFilters []types.ParameterStringFilter,
) ([]*parameter, error) {
	filters := []types.ParameterStringFilter{
		{
			Key:    aws.String("Type"),
			Values: []string{"SecureString"},
		},
		{
			Key:    aws.String("Path"),
			Option: aws.String("Recursive"),
			Values: []string{path},
		},
	}
	filters = append(filters, additionalFilters...)

	var parameters []*parameter
	result, nextToken, err := s.executeDescribeParametersRequest(ctx, path, filters, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	parameters = append(parameters, result...)

	for aws.ToString(nextToken) != "" {
		result, nextToken, err = s.executeDescribeParametersRequest(ctx, path, filters, nextToken)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		parameters = append(parameters, result...)
	}

	return parameters, nil
}

func (s *parameterStore) executeDescribeParametersRequest(
	ctx context.Context,
	path string,
	filters []types.ParameterStringFilter,
	nextToken *string,
) ([]*parameter, *string, error) {

	output, err := s.client.DescribeParameters(ctx, &ssm.DescribeParametersInput{
		ParameterFilters: filters,
		NextToken:        nextToken,
	})
	if err != nil {
		return nil, nil, errors.Wrapf(err, "error executing AWS SSM query [DescribeParameters '%v']", path)
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

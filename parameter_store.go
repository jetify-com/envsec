package envsec

import (
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/pkg/errors"
	"go.jetpack.io/axiom/opensource/cmd/jetpack/viewer"
	"go.jetpack.io/axiom/opensource/goutil/errorutil"
	"go.jetpack.io/axiom/opensource/proto/api"
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

type parameterStore struct {
	config *api.ParameterStoreConfig
	path   string
}

// Parameter values are limited in size to 4KB
const parameterValueMaxLength = 4 * 1024

// New parameter store for current user/organization.
func newParameterStore(config *api.ParameterStoreConfig, path string) *parameterStore {
	return &parameterStore{
		config: config,
		path:   path,
	}
}

// Returns information about stored parameters.
func (s *parameterStore) listParameters(vc viewer.Context, filters []types.ParameterStringFilter) ([]*parameter, error) {
	client := s.newSsmClient()
	return s.describeParameters(vc, client, filters...)
}

// Resolve a parameter by name.
func (s *parameterStore) resolveParameter(vc viewer.Context, filters []types.ParameterStringFilter) (*parameter, error) {
	client := s.newSsmClient()

	parameters, err := s.describeParameters(vc, client, filters...)
	if err != nil {
		return nil, errors.Wrapf(err, "error executing AWS SSM query: '%v'", filters)
	}

	if len(parameters) == 0 {
		return nil, errors.WithStack(errorutil.NewUserErrorf("stored parameter not defined %v", filters))
	} else if 1 < len(parameters) {
		return nil, errors.WithStack(errorutil.NewUserErrorf("duplicate definitions for qualified parameter %v", filters))
	}

	return parameters[0], nil
}

// Returns values associated with a set of parameters
func (s *parameterStore) loadParametersValues(vc viewer.Context, parameters []*parameter) (map[string]string, error) {
	client := s.newSsmClient()

	if len(parameters) == 0 {
		return map[string]string{}, nil
	}

	var paths []string
	for _, p := range parameters {
		paths = append(paths, p.id)
	}

	// Retrieve stored parameters values
	output, err := client.GetParameters(vc, &ssm.GetParametersInput{
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
func (s *parameterStore) newParameter(vc viewer.Context, v *parameter, value string) error {
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

	_, err = client.PutParameter(vc, input)
	return errors.WithStack(err)
}

// Defines or updates a stored parameter.
// parameter values are limited in size to 4 KB.
func (s *parameterStore) storeParameterValue(vc viewer.Context, v *parameter, value string) error {
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

	_, err := client.PutParameter(vc, input)
	return errors.WithStack(err)
}

// Delete a stored parameter from the system.
func (s *parameterStore) deleteParameters(vc viewer.Context, parameters []*parameter) error {
	client := s.newSsmClient()

	var paths []string
	for _, p := range parameters {
		paths = append(paths, p.id)
	}

	input := &ssm.DeleteParametersInput{
		Names: paths,
	}
	_, err := client.DeleteParameters(vc, input)
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

func (s *parameterStore) describeParameters(vc viewer.Context, client *ssm.Client, additionalFilters ...types.ParameterStringFilter) ([]*parameter, error) {
	filters := []types.ParameterStringFilter{
		{
			Key:    aws.String("Type"),
			Values: []string{"SecureString"},
		},
	}
	filters = append(filters, additionalFilters...)

	var parameters []*parameter
	result, nextToken, err := s.executeDescribeParametersRequest(vc, client, filters, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	parameters = append(parameters, result...)

	for aws.StringValue(nextToken) != "" {
		result, nextToken, err = s.executeDescribeParametersRequest(vc, client, filters, nextToken)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		parameters = append(parameters, result...)
	}

	return parameters, nil
}

func (s *parameterStore) executeDescribeParametersRequest(vc viewer.Context, client *ssm.Client, filters []types.ParameterStringFilter, nextToken *string) ([]*parameter, *string, error) {
	output, err := client.DescribeParameters(vc, &ssm.DescribeParametersInput{
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
		tags, err := client.ListTagsForResource(vc, listTagsInput)
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

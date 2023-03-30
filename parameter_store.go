// Copyright 2022 Jetpack Technologies Inc and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package envsec

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/smithy-go"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

const emptyStringValuePlaceholder = "__###EMPTY_STRING###__"

type parameter struct {
	id          string
	description string
	tags        []types.Tag
}

type parameterStore struct {
	config *SSMConfig
	client *ssm.Client
}

// Parameter values are limited in size to 4KB
const parameterValueMaxLength = 4 * 1024

var FaultyParamError = errors.New("Faulty Parameter")

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
		Value:       awsSSMParamStoreValue(value),
		Tags:        v.tags,
	}

	// Set the KmsKeyId only when it is present. Otherwise, aws sdk uses the default KMS key
	// since we specify "SecureString" type.
	if s.config.KmsKeyId != "" {
		input.KeyId = aws.String(s.config.KmsKeyId)
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
		Overwrite:   lo.ToPtr(true),
		Value:       awsSSMParamStoreValue(value),
	}
	_, err := s.client.PutParameter(ctx, input)
	return errors.WithStack(err)
}

func (s *parameterStore) ListByPath(ctx context.Context, path string) ([]EnvVar, error) {
	// Create the request object:
	req := &ssm.GetParametersByPathInput{
		Path:           aws.String(path),
		WithDecryption: lo.ToPtr(true),
		Recursive:      lo.ToPtr(true),
	}

	// Start with empty results
	results := []EnvVar{}

	// Paginate through the results:
	paginator := ssm.NewGetParametersByPathPaginator(s.client, req)
	for paginator.HasMorePages() {
		// Issue the request for the next page:
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return results, errors.WithStack(err)
		}

		// Append results:
		params := resp.Parameters
		for _, p := range params {
			results = append(results, EnvVar{
				Name:  aws.ToString(p.Name), // TODO: Full path?
				Value: awsSSMParamStoreValueToString(p.Value),
			})
		}
	}
	sort(results)
	return results, nil
}

func (s *parameterStore) listByTags(ctx context.Context, envId EnvId) ([]EnvVar, error) {
	// Create the request object:
	req := &ssm.DescribeParametersInput{
		ParameterFilters: buildFilters(envId),
	}

	varNames := []string{}
	// Paginate through the results:
	paginator := ssm.NewDescribeParametersPaginator(s.client, req)
	for paginator.HasMorePages() {
		// Issue the request for the next page:
		resp, err := paginator.NextPage(ctx)
		if err != nil {
			return []EnvVar{}, errors.WithStack(err)
		}
		// Append results:
		for _, p := range resp.Parameters {
			// AWS returns the parameter path as its "name":
			varName := nameFromPath(aws.ToString(p.Name))
			varNames = append(varNames, varName)
		}
	}

	return s.getAll(ctx, envId, varNames)
}

func (s *parameterStore) getAll(ctx context.Context, envId EnvId, varNames []string) ([]EnvVar, error) {
	// Start with empty results
	results := []EnvVar{}
	paths := lo.Map(varNames, func(name string, _ int) string {
		return varPath(envId, name)
	})

	// Due to AWS API limits, chunk into groups of 10
	chunks := lo.Chunk(paths, 10)
	for _, chunk := range chunks {

		// Create the request object:
		req := &ssm.GetParametersInput{
			Names:          chunk,
			WithDecryption: lo.ToPtr(true),
		}
		// Issue the request:
		resp, err := s.client.GetParameters(ctx, req)
		if err != nil {
			// For now an error short circuits the entire thing, but we could be more careful
			// and return values that were successfully retrieved, even if others failed.
			return results, errors.WithStack(err)
		}

		// Append results:
		for _, p := range resp.Parameters {
			results = append(results, EnvVar{
				Name:  nameFromPath(aws.ToString(p.Name)),
				Value: awsSSMParamStoreValueToString(p.Value),
			})
		}
	}
	sort(results)
	return results, nil
}

func (s *parameterStore) deleteAll(ctx context.Context, envId EnvId, varNames []string) error {
	paths := lo.Map(varNames, func(name string, _ int) string {
		return varPath(envId, name)
	})
	// Due to AWS API limits, chunk into groups of 10
	chunks := lo.Chunk(paths, 10)
	var multiErr error
	for _, chunk := range chunks {
		// Create the request object:
		req := &ssm.DeleteParametersInput{
			Names: chunk,
		}

		// Issue the request:
		_, err := s.client.DeleteParameters(ctx, req)
		if err != nil {
			var awsErr smithy.APIError
			if errors.As(err, &awsErr) {
				if awsErr.ErrorCode() == "AccessDeniedException" {
					faultyParam := getFaultyParameter(awsErr.ErrorMessage())
					return errors.Wrap(FaultyParamError, faultyParam)
				}
			}
			multiErr = multierror.Append(multiErr, err)
			continue
		}
	}
	// We could also return the list of deleted parameters
	return multiErr
}

// Implement interface Lister from text/collate
type envVars []EnvVar

func (e envVars) Len() int {
	return len(e)
}

func (e envVars) Bytes(i int) []byte {
	return []byte(e[i].Name)
}

func (e envVars) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

func sort(vars envVars) {
	c := collate.New(language.English, collate.Loose, collate.Numeric)
	c.Sort(vars)
}

func getFaultyParameter(message string) string {
	resourceParts := strings.Split(message, "/")
	nameParts := strings.Split(resourceParts[len(resourceParts)-1], " ")
	return nameParts[0]
}

// AWS SSM Param store doesn't allow empty strings so we use a placeholder
// instead
func awsSSMParamStoreValue(s string) *string {
	if s == "" {
		return aws.String(emptyStringValuePlaceholder)
	}
	return aws.String(s)
}

func awsSSMParamStoreValueToString(s *string) string {
	if *s == emptyStringValuePlaceholder {
		return ""
	}
	return *s
}

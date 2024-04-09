// Copyright 2024 Jetify Inc. and contributors. All rights reserved.
// Use of this source code is governed by the license in the LICENSE file.

package ssmstore

import (
	"context"

	cognitoTypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentity/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"go.jetpack.io/envsec/pkg/awsfed"
	"go.jetpack.io/envsec/pkg/envsec"
	"go.jetpack.io/pkg/auth/session"
)

type SSMStore struct {
	store *parameterStore
}

// SSMStore implements interface Store (compile-time check)
var _ envsec.Store = (*SSMStore)(nil)

func (s *SSMStore) InitForUser(ctx context.Context, e *envsec.Envsec) (*session.Token, error) {
	client, err := e.AuthClient()
	if err != nil {
		return nil, errors.WithStack(err)
	}
	tok, err := client.LoginFlowIfNeeded(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	ssmConfig, err := genSSMConfigFromToken(ctx, tok, true /*useCache*/)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	paramStore, err := newParameterStore(ctx, ssmConfig)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	s.store = paramStore
	return tok, nil
}

func (s *SSMStore) List(ctx context.Context, envID envsec.EnvID) ([]envsec.EnvVar, error) {
	if s.store.config.hasDefaultPaths() {
		return s.store.listByPath(ctx, envID)
	}
	return s.store.listByTags(ctx, envID)
}

func (s *SSMStore) Get(ctx context.Context, envID envsec.EnvID, name string) (string, error) {
	vars, err := s.GetAll(ctx, envID, []string{name})
	if err != nil {
		return "", errors.WithStack(err)
	}
	if len(vars) == 0 {
		return "", nil
	}
	return vars[0].Value, nil
}

func (s *SSMStore) GetAll(ctx context.Context, envID envsec.EnvID, names []string) ([]envsec.EnvVar, error) {
	return s.store.getAll(ctx, envID, names)
}

func (s *SSMStore) Set(
	ctx context.Context,
	envID envsec.EnvID,
	name string,
	value string,
) error {
	path := s.store.config.varPath(envID, name)

	// New parameter definition
	tags := buildTags(envID, name)
	parameter := &parameter{
		tags: tags,
		id:   path,
	}
	return s.store.newParameter(ctx, parameter, value)
}

func (s *SSMStore) SetAll(ctx context.Context, envID envsec.EnvID, values map[string]string) error {
	// For now we implement by issuing multiple calls to Set()
	// Make more efficient either by implementing a batch call to the underlying API, or
	// by concurrently calling Set()

	var multiErr error
	for name, value := range values {
		err := s.Set(ctx, envID, name, value)
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return multiErr
}

func (s *SSMStore) Delete(ctx context.Context, envID envsec.EnvID, name string) error {
	return s.DeleteAll(ctx, envID, []string{name})
}

func (s *SSMStore) DeleteAll(ctx context.Context, envID envsec.EnvID, names []string) error {
	return s.store.deleteAll(ctx, envID, names)
}

func buildTags(envID envsec.EnvID, varName string) []types.Tag {
	tags := []types.Tag{}
	if envID.ProjectID != "" {
		tags = append(tags, types.Tag{
			Key:   lo.ToPtr("project-id"),
			Value: lo.ToPtr(envID.ProjectID),
		})
	}
	if envID.OrgID != "" {
		tags = append(tags, types.Tag{
			Key:   lo.ToPtr("org-id"),
			Value: lo.ToPtr(envID.OrgID),
		})
	}
	if envID.EnvName != "" {
		tags = append(tags, types.Tag{
			Key:   lo.ToPtr("env-name"),
			Value: lo.ToPtr(envID.EnvName),
		})
	}

	if varName != "" {
		tags = append(tags, types.Tag{
			Key:   lo.ToPtr("name"),
			Value: lo.ToPtr(varName),
		})
	}

	return tags
}

func genSSMConfigFromToken(
	ctx context.Context,
	tok *session.Token,
	useCache bool,
) (*SSMConfig, error) {
	if tok == nil {
		return &SSMConfig{}, nil
	}
	fed := awsfed.New()
	var creds *cognitoTypes.Credentials
	var err error
	if useCache {
		creds, err = fed.AWSCredsWithLocalCache(ctx, tok)
	} else {
		creds, err = fed.AWSCreds(ctx, tok.IDToken)
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &SSMConfig{
		AccessKeyID:     *creds.AccessKeyId,
		SecretAccessKey: *creds.SecretKey,
		SessionToken:    *creds.SessionToken,
		Region:          fed.Region,
	}, nil
}

package jetstore

import (
	"context"

	"connectrpc.com/connect"
	"go.jetpack.io/envsec/pkg/envsec"
	"go.jetpack.io/pkg/api"
	secretsv1alpha1 "go.jetpack.io/pkg/api/gen/priv/secrets/v1alpha1"
	"go.jetpack.io/pkg/api/gen/priv/secrets/v1alpha1/secretsv1alpha1connect"
	"go.jetpack.io/pkg/auth/session"
)

type JetpackAPIStore struct {
	client secretsv1alpha1connect.SecretsServiceClient
}

// JetpackAPIStore implements interface Store (compile-time check)
var _ envsec.Store = (*JetpackAPIStore)(nil)

func (j *JetpackAPIStore) InitForUser(
	ctx context.Context,
	envsec *envsec.Envsec,
) (*session.Token, error) {
	project, err := envsec.ProjectConfig()
	if project == nil {
		return nil, err
	}

	authClient, err := envsec.AuthClient()
	if err != nil {
		return nil, err
	}

	tok, err := authClient.LoginFlowIfNeededForOrg(ctx, project.OrgID.String())
	if err != nil {
		return nil, err
	}
	j.client = api.NewClient(ctx, envsec.APIHost, tok).SecretsService()
	return tok, nil
}

func (j JetpackAPIStore) List(ctx context.Context, envID envsec.EnvID) ([]envsec.EnvVar, error) {
	resp, err := j.client.ListSecrets(
		ctx,
		connect.NewRequest(&secretsv1alpha1.ListSecretsRequest{ProjectId: envID.ProjectID}),
	)
	if err != nil {
		return nil, err
	}
	result := []envsec.EnvVar{}
	for _, secret := range resp.Msg.Secrets {
		if v := secret.EnvironmentValues[envID.EnvName]; len(v) > 0 {
			result = append(
				result, envsec.EnvVar{
					Name:  secret.Name,
					Value: string(v),
				},
			)
		}
	}
	return result, nil
}

func (j JetpackAPIStore) Set(ctx context.Context, envID envsec.EnvID, name string, value string) error {
	_, err := j.client.PatchSecret(
		ctx, connect.NewRequest(
			&secretsv1alpha1.PatchSecretRequest{
				ProjectId: envID.ProjectID,
				Secret: &secretsv1alpha1.Secret{
					Name: name,
					EnvironmentValues: map[string][]byte{
						envID.EnvName: []byte(value),
					},
				},
			},
		),
	)
	return err
}

func (j JetpackAPIStore) SetAll(ctx context.Context, envID envsec.EnvID, values map[string]string) error {
	patchActions := []*secretsv1alpha1.Action{}
	for name, value := range values {
		patchActions = append(
			patchActions, &secretsv1alpha1.Action{
				Action: &secretsv1alpha1.Action_PatchSecret{
					PatchSecret: &secretsv1alpha1.PatchSecretRequest{
						ProjectId: envID.ProjectID,
						Secret: &secretsv1alpha1.Secret{
							Name: name,
							EnvironmentValues: map[string][]byte{
								envID.EnvName: []byte(value),
							},
						},
					},
				},
			},
		)
	}

	_, err := j.client.Batch(
		ctx, connect.NewRequest(&secretsv1alpha1.BatchRequest{Actions: patchActions}),
	)
	return err
}

func (j JetpackAPIStore) Get(ctx context.Context, envID envsec.EnvID, name string) (string, error) {
	vars, err := j.List(ctx, envID)
	if err != nil {
		return "", err
	}
	for _, v := range vars {
		if v.Name == name {
			return v.Value, nil
		}
	}
	return "", nil
}

func (j JetpackAPIStore) GetAll(ctx context.Context, envID envsec.EnvID, names []string) ([]envsec.EnvVar, error) {
	vars, err := j.List(ctx, envID)
	if err != nil {
		return nil, err
	}
	result := []envsec.EnvVar{}
	for _, v := range vars {
		for _, name := range names {
			if v.Name == name {
				result = append(result, v)
			}
		}
	}
	return result, nil
}

func (j JetpackAPIStore) Delete(ctx context.Context, envID envsec.EnvID, name string) error {
	_, err := j.client.DeleteSecret(
		ctx, connect.NewRequest(
			&secretsv1alpha1.DeleteSecretRequest{
				ProjectId:    envID.ProjectID,
				SecretName:   name,
				Environments: []string{envID.EnvName},
			},
		),
	)
	return err
}

func (j JetpackAPIStore) DeleteAll(ctx context.Context, envID envsec.EnvID, names []string) error {
	deleteActions := []*secretsv1alpha1.Action{}
	for _, name := range names {
		deleteActions = append(
			deleteActions, &secretsv1alpha1.Action{
				Action: &secretsv1alpha1.Action_DeleteSecret{
					DeleteSecret: &secretsv1alpha1.DeleteSecretRequest{
						ProjectId:    envID.ProjectID,
						SecretName:   name,
						Environments: []string{envID.EnvName},
					},
				},
			},
		)
	}

	_, err := j.client.Batch(
		ctx, connect.NewRequest(&secretsv1alpha1.BatchRequest{Actions: deleteActions}),
	)
	return err
}

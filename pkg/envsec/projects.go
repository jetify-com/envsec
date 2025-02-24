package envsec

import (
	"context"
	"io"

	"connectrpc.com/connect"
	"go.jetify.com/envsec/internal/tux"
	"go.jetify.com/pkg/api"
	v1alpha1 "go.jetify.com/pkg/api/gen/priv/projects/v1alpha1"
)

func (e *Envsec) DescribeCurrentProject(
	ctx context.Context,
	w io.Writer,
) error {
	project, err := e.ProjectConfig()
	if err != nil {
		return err
	}

	authClient, err := e.AuthClient()
	if err != nil {
		return err
	}

	tok, err := authClient.LoginFlowIfNeededForOrg(ctx, project.OrgID.String())
	if err != nil {
		return err
	}

	apiClient := api.NewClient(ctx, e.APIHost, tok)
	response, err := apiClient.ProjectsClient().GetProject(ctx,
		connect.NewRequest(&v1alpha1.GetProjectRequest{
			Id: project.ProjectID.String(),
		}))
	if err != nil {
		return err
	}

	tux.FTable(w, [][]string{
		{"Project", response.Msg.GetProject().GetName()},
		{"project ID", project.ProjectID.String()},
		{"Org ID", project.OrgID.String()},
	})

	return nil
}

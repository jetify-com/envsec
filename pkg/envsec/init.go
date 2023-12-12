package envsec

import (
	"context"
	"errors"
	"fmt"

	"go.jetpack.io/pkg/jetcloud"
)

type InitProjectArgs struct {
	Force bool
}

func (e *Envsec) NewProject(ctx context.Context, force bool) error {
	var err error

	client, err := e.authClient()
	if err != nil {
		return err
	}

	tok, err := client.GetSession(ctx)
	if err != nil {
		return fmt.Errorf("error: %w, run `envsec auth login`", err)
	}

	c := jetcloud.Client{APIHost: e.APIHost, IsDev: e.IsDev}
	projectID, err := c.InitProject(ctx, jetcloud.InitProjectArgs{
		Dir:   e.WorkingDir,
		Force: force,
		Token: tok,
	})
	if errors.Is(err, jetcloud.ErrProjectAlreadyInitialized) {
		fmt.Fprintf(
			e.Stderr,
			"Warning: project already initialized ID=%s\n",
			projectID,
		)
	} else if err != nil {
		return err
	} else {
		fmt.Fprintf(e.Stderr, "Initialized project ID=%s\n", projectID)
	}
	return nil
}

package jetcloud

import (
	"fmt"
	"os"

	"go.jetpack.io/envsec/internal/auth"
	typeid "go.jetpack.io/typeid/typed"
)

type projectPrefix struct{}

func (projectPrefix) Type() string { return "proj" }

type projectID struct{ typeid.TypeID[projectPrefix] }

var nilProjectID = projectID{typeid.Nil[projectPrefix]()}

type client struct{}

func newClient() *client {
	return &client{}
}

func (c *client) newProjectID(user *auth.User, repo, subdir string) (projectID, error) {
	fmt.Fprintf(os.Stderr, "Creating new project for repo=%s subdir=%s\n", repo, subdir)
	// TODO this will fetch project ID from an API
	tid, _ := typeid.New[projectPrefix]()
	return projectID{tid}, nil
}

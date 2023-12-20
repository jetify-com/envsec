package flow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/samber/lo"
	"go.jetpack.io/envsec/internal/git"
	"go.jetpack.io/pkg/api"
	membersv1alpha1 "go.jetpack.io/pkg/api/gen/priv/members/v1alpha1"
	projectsv1alpha1 "go.jetpack.io/pkg/api/gen/priv/projects/v1alpha1"
	"go.jetpack.io/pkg/auth/session"
	"go.jetpack.io/pkg/id"
	"go.jetpack.io/typeid"
)

// flow:
// 0. Ask if you want to overwrite existing config [y/N]
// 1. Link to an existing project? [Y/n]
// 2a. What project would you like to link to? (sorted by repo/dir match)
// 2b. What’s the name of your new project?

type Init struct {
	Client                *api.Client
	PromptOverwriteConfig bool
	Token                 *session.Token
	WorkingDir            string
}

func (i *Init) Run(ctx context.Context) (id.ProjectID, error) {
	createProject, err := i.confirmSetupProjectPrompt()
	if err != nil {
		return id.ProjectID{}, err
	}
	if !createProject {
		return id.ProjectID{}, errors.New("aborted")
	}

	member, err := i.Client.GetMember(ctx, i.Token.IDClaims().Subject)
	if err != nil {
		return id.ProjectID{}, err
	}

	// TODO: printOrgNotice will be a team picker once that is implemented.
	i.printOrgNotice(member)
	linkToExisting, err := i.linkToExistingPrompt()
	if err != nil {
		return id.ProjectID{}, err
	}
	if linkToExisting {
		return i.showExistingListPrompt(ctx)
	}
	return i.createNewPrompt(ctx, member)
}

func (i *Init) confirmSetupProjectPrompt() (bool, error) {
	if i.PromptOverwriteConfig {
		return boolPrompt(
			fmt.Sprintf("Project already exists. Reset project in %s", i.WorkingDir),
			"n",
		)
	}
	return boolPrompt(
		fmt.Sprintf("Setup project in %s", i.WorkingDir),
		"y",
	)
}

func (i *Init) printOrgNotice(member *membersv1alpha1.Member) {
	fmt.Fprintf(
		os.Stderr,
		"Initializing project in org %s\n",
		member.Organization.Name,
	)
}

func (i *Init) linkToExistingPrompt() (bool, error) {
	return boolPrompt("Link to an existing project", "y")
}

func (i *Init) showExistingListPrompt(
	ctx context.Context,
) (id.ProjectID, error) {
	orgID, err := typeid.Parse[id.OrgID](i.Token.IDClaims().OrgID)
	if err != nil {
		return id.ProjectID{}, err
	}

	projects, err := i.Client.ListProjects(ctx, orgID)
	if err != nil {
		return id.ProjectID{}, err
	}

	repo, err := git.GitRepoURL(i.WorkingDir)
	if err != nil {
		return id.ProjectID{}, err
	}

	directory, err := git.GitSubdirectory(i.WorkingDir)
	if err != nil {
		return id.ProjectID{}, err
	}

	sort.SliceStable(projects, func(i, j int) bool {
		if projects[i].GetRepo() == repo &&
			projects[i].GetDirectory() == directory {
			return true
		}
		return projects[i].GetRepo() == repo && projects[j].GetRepo() != repo
	})

	prompt := promptui.Select{
		Label: "What project would you like to link to",
		Items: lo.Map(projects, func(proj *projectsv1alpha1.Project, _ int) string {
			item := strings.TrimSpace(proj.GetName())
			if item == "" {
				item = "untitled"
			}
			if proj.GetRepo() != "" {
				item += " repo: " + proj.GetRepo()
			}
			if proj.GetDirectory() != "" && proj.GetDirectory() != "." {
				item += " dir: " + proj.GetDirectory()
			}
			return item + " id: " + proj.GetId()
		}),
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return id.ProjectID{}, err
	}

	projectID, err := typeid.Parse[id.ProjectID](projects[idx].GetId())
	if err != nil {
		return id.ProjectID{}, err
	}

	fmt.Fprintf(os.Stderr, "Linked to project %s\n", projects[idx].GetName())
	return projectID, nil
}

func (i *Init) createNewPrompt(
	ctx context.Context,
	member *membersv1alpha1.Member,
) (id.ProjectID, error) {
	prompt := promptui.Prompt{
		Label:   "What’s the name of your new project",
		Default: filepath.Base(i.WorkingDir),
		Validate: func(name string) error {
			if name == "" {
				return errors.New("project name cannot be empty")
			}
			return nil
		},
	}

	name, err := prompt.Run()
	if err != nil {
		return id.ProjectID{}, err
	}

	orgID, err := typeid.Parse[id.OrgID](i.Token.IDClaims().OrgID)
	if err != nil {
		return id.ProjectID{}, err
	}

	repo, err := git.GitRepoURL(i.WorkingDir)
	if err != nil {
		return id.ProjectID{}, err
	}

	directory, err := git.GitSubdirectory(i.WorkingDir)
	if err != nil {
		return id.ProjectID{}, err
	}

	project, err := i.Client.CreateProject(
		ctx,
		orgID,
		repo,
		directory,
		name,
	)
	if err != nil {
		return id.ProjectID{}, err
	}

	projectID, err := typeid.Parse[id.ProjectID](project.GetId())
	if err != nil {
		return id.ProjectID{}, err
	}

	fmt.Fprintf(
		os.Stderr,
		"Created project %s in org %s\n",
		project.GetName(),
		member.GetOrganization().GetName(),
	)
	return projectID, nil
}

func boolPrompt(label, defaultResult string) (bool, error) {
	prompt := promptui.Prompt{
		Label:     label,
		IsConfirm: true,
		Default:   defaultResult,
	}

	result, err := prompt.Run()
	// promptui.ErrAbort is returned when user enters "n" which is valid.
	if err != nil && !errors.Is(err, promptui.ErrAbort) {
		return false, err
	}
	if result == "" {
		result = defaultResult
	}

	return strings.ToLower(result) == "y", nil
}

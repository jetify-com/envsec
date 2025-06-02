package flow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"go.jetify.com/envsec/internal/git"
	"go.jetify.com/pkg/api"
	membersv1alpha1 "go.jetify.com/pkg/api/gen/priv/members/v1alpha1"
	projectsv1alpha1 "go.jetify.com/pkg/api/gen/priv/projects/v1alpha1"
	"go.jetify.com/pkg/auth/session"
	"go.jetify.com/pkg/ids"
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

func (i *Init) Run(ctx context.Context) (ids.ProjectID, error) {
	createProject, err := i.confirmSetupProjectPrompt()
	if err != nil {
		return ids.ProjectID{}, err
	}
	if !createProject {
		return ids.ProjectID{}, errors.New("aborted")
	}

	member, err := i.Client.GetMember(ctx, i.Token.IDClaims().Subject)
	if err != nil {
		return ids.ProjectID{}, err
	}

	// TODO: printOrgNotice will be a team picker once that is implemented.
	i.printOrgNotice(member)
	orgID, err := ids.ParseOrgID(i.Token.IDClaims().OrgID)
	if err != nil {
		return ids.ProjectID{}, err
	}

	projects, err := i.Client.ListProjects(ctx, orgID)
	if err != nil {
		return ids.ProjectID{}, err
	}
	if len(projects) > 0 {
		linkToExisting, err := i.linkToExistingPrompt()
		if err != nil {
			return ids.ProjectID{}, err
		}
		if linkToExisting {
			return i.showExistingListPrompt(projects)
		}
	}
	return i.createNewPrompt(ctx, member)
}

func (i *Init) confirmSetupProjectPrompt() (bool, error) {
	if i.PromptOverwriteConfig {
		return boolPrompt(
			fmt.Sprintf("Project already exists. Reset project in %s", i.WorkingDir),
			false,
		)
	}
	return boolPrompt(
		fmt.Sprintf("Setup project in %s", i.WorkingDir),
		true,
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
	return boolPrompt("Link to an existing project", true)
}

func (i *Init) showExistingListPrompt(
	projects []*projectsv1alpha1.Project,
) (ids.ProjectID, error) {
	// Ignore errors, it's fine if not in repo or git not installed.
	repo, _ := git.GitRepoURL(i.WorkingDir)
	directory, _ := git.GitSubdirectory(i.WorkingDir)

	sort.SliceStable(projects, func(i, j int) bool {
		if projects[i].GetRepo() == repo &&
			projects[i].GetDirectory() == directory {
			return true
		}
		return projects[i].GetRepo() == repo && projects[j].GetRepo() != repo
	})

	prompt := &survey.Select{
		Message: "What project would you like to link to?",
		Options: formatProjectItems(projects),
	}

	idx := 0
	if err := survey.AskOne(prompt, &idx); err != nil {
		return ids.ProjectID{}, err
	}

	projectID, err := ids.ParseProjectID(projects[idx].GetId())
	if err != nil {
		return ids.ProjectID{}, err
	}
	name := projects[idx].GetName()
	if name == "" {
		name = "untitled"
	}
	fmt.Fprintf(os.Stderr, "Linked to project %s\n", name)
	return projectID, nil
}

func (i *Init) createNewPrompt(
	ctx context.Context,
	member *membersv1alpha1.Member,
) (ids.ProjectID, error) {
	prompt := &survey.Input{
		Message: "What’s the name of your new project?",
		Default: filepath.Base(i.WorkingDir),
	}

	name := ""
	if err := survey.AskOne(prompt, &name); err != nil {
		return ids.ProjectID{}, err
	}

	orgID, err := ids.ParseOrgID(i.Token.IDClaims().OrgID)
	if err != nil {
		return ids.ProjectID{}, err
	}

	// Ignore errors, it's fine if not in repo or git not installed.
	repo, _ := git.GitRepoURL(i.WorkingDir)
	directory, _ := git.GitSubdirectory(i.WorkingDir)

	project, err := i.Client.CreateProject(
		ctx,
		orgID,
		repo,
		directory,
		strings.TrimSpace(name),
	)
	if err != nil {
		return ids.ProjectID{}, err
	}

	projectID, err := ids.ParseProjectID(project.GetId())
	if err != nil {
		return ids.ProjectID{}, err
	}

	fmt.Fprintf(
		os.Stderr,
		"Created project %s in org %s\n",
		project.GetName(),
		member.GetOrganization().GetName(),
	)
	return projectID, nil
}

func boolPrompt(label string, defaultResult bool) (bool, error) {
	result := false
	prompt := &survey.Confirm{
		Message: label,
		Default: defaultResult,
	}
	return result, survey.AskOne(prompt, &result)
}

func formatProjectItems(projects []*projectsv1alpha1.Project) []string {
	longestNameLength := 0
	for _, proj := range projects {
		name := proj.GetName()
		if name == "" {
			name = "untitled"
		}
		if l := len(name); l > longestNameLength {
			longestNameLength = l
		}
	}
	// Add padding
	table := make([][]string, len(projects))
	for idx, proj := range projects {
		name := strings.TrimSpace(proj.GetName())
		if name == "" {
			name = "untitled"
		}

		table[idx] = []string{
			color.HiGreenString(
				fmt.Sprintf("%-"+fmt.Sprintf("%d", longestNameLength)+"s", name),
			),
			color.HiBlueString("id:"),
			proj.GetId(),
		}
	}

	rows := []string{}
	for _, cols := range table {
		rows = append(rows, strings.Join(cols, " "))
	}
	return rows
}

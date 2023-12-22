package flow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
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
	orgID, err := typeid.Parse[id.OrgID](i.Token.IDClaims().OrgID)
	if err != nil {
		return id.ProjectID{}, err
	}

	projects, err := i.Client.ListProjects(ctx, orgID)
	if err != nil {
		return id.ProjectID{}, err
	}
	if len(projects) > 0 {
		linkToExisting, err := i.linkToExistingPrompt()
		if err != nil {
			return id.ProjectID{}, err
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
	projects []*projectsv1alpha1.Project,
) (id.ProjectID, error) {
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

	prompt := promptui.Select{
		Label: "What project would you like to link to",
		Items: formatProjectItems(projects),
		Size:  10,
		Templates: &promptui.SelectTemplates{
			Active: "\U000025B8 {{ . }}",
		},
	}

	idx, _, err := prompt.Run()
	if err != nil {
		return id.ProjectID{}, err
	}

	projectID, err := typeid.Parse[id.ProjectID](projects[idx].GetId())
	if err != nil {
		return id.ProjectID{}, err
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
) (id.ProjectID, error) {
	prompt := promptui.Prompt{
		Label:   "What’s the name of your new project",
		Default: filepath.Base(i.WorkingDir),
		Validate: func(name string) error {
			if strings.TrimSpace(name) == "" {
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

func formatProjectItems(projects []*projectsv1alpha1.Project) []string {
	longestNameLength := 0
	longestRepoLength := 0
	longestDirLength := 0
	for _, proj := range projects {
		name := proj.GetName()
		if name == "" {
			name = "untitled"
		}
		if l := len(name); l > longestNameLength {
			longestNameLength = l
		}
		if l := len(proj.GetRepo()); l > longestRepoLength {
			longestRepoLength = l
		}
		if l := len(proj.GetDirectory()); l > longestDirLength {
			longestDirLength = l
		}
	}
	// Add padding
	table := make([][]string, len(projects))
	for idx, proj := range projects {
		name := proj.GetName()
		if name == "" {
			name = "untitled"
		}
		table[idx] = []string{
			color.HiGreenString(
				fmt.Sprintf("%-"+fmt.Sprintf("%d", longestNameLength)+"s", name),
			),
			color.HiBlueString("repo:"),
			fmt.Sprintf("%-"+fmt.Sprintf("%d", longestRepoLength)+"s", proj.GetRepo()),
			color.HiBlueString("dir:"),
			fmt.Sprintf("%-"+fmt.Sprintf("%d", longestDirLength)+"s", proj.GetDirectory()),
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

package typeids

import typeid "go.jetpack.io/typeid/typed"

type projectPrefix struct{}

func (projectPrefix) Type() string { return "proj" }

type ProjectID struct{ typeid.TypeID[projectPrefix] }

var NilProjectID = ProjectID{typeid.Nil[projectPrefix]()}

type organizationPrefix struct{}

func (organizationPrefix) Type() string { return "org" }

type OrganizationID struct {
	typeid.TypeID[organizationPrefix]
}

var NilOrganizationID = OrganizationID{typeid.Nil[organizationPrefix]()}

func OrganizationIDFromString(s string) (OrganizationID, error) {
	id, err := typeid.FromString[organizationPrefix](s)
	return OrganizationID{id}, err
}

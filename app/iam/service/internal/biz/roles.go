package biz

import "fmt"

var (
	validOrganizationRoles = map[string]bool{"owner": true, "admin": true, "member": true}
	validProjectRoles      = map[string]bool{"admin": true, "developer": true, "viewer": true}
)

func ValidateOrganizationRole(role string) error {
	if !validOrganizationRoles[role] {
		return fmt.Errorf("invalid organization role %q; allowed: owner, admin, member", role)
	}
	return nil
}

func ValidateProjectRole(role string) error {
	if !validProjectRoles[role] {
		return fmt.Errorf("invalid project role %q; allowed: admin, developer, viewer", role)
	}
	return nil
}

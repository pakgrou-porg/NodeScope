// Package auth centralizes NodeScope authorization vocabulary.
package auth

import "fmt"

// Role is fleet-wide. Host-scoped grants are intentionally not part of Release 1.
type Role string

const (
	RoleViewer        Role = "viewer"
	RoleOperator      Role = "operator"
	RoleAdministrator Role = "administrator"
)

func (r Role) Valid() bool {
	switch r {
	case RoleViewer, RoleOperator, RoleAdministrator:
		return true
	default:
		return false
	}
}

func (r Role) Rank() int {
	switch r {
	case RoleViewer:
		return 1
	case RoleOperator:
		return 2
	case RoleAdministrator:
		return 3
	default:
		return 0
	}
}

// Allows reports whether a role satisfies the minimum role required for an
// operation. Invalid roles are never authorized.
func (r Role) Allows(minimum Role) bool {
	return r.Valid() && minimum.Valid() && r.Rank() >= minimum.Rank()
}

// Requirement declares an action's minimum role and whether its payload must
// be represented by the audit-intent transaction before dispatch.
type Requirement struct {
	Action        string
	MinimumRole   Role
	RequiresAudit bool
}

var requirements = map[string]Requirement{
	"fleet.read":               {Action: "fleet.read", MinimumRole: RoleViewer, RequiresAudit: false},
	"host.read":                {Action: "host.read", MinimumRole: RoleViewer, RequiresAudit: false},
	"alert.acknowledge":        {Action: "alert.acknowledge", MinimumRole: RoleOperator, RequiresAudit: true},
	"collection_interval.set":  {Action: "collection_interval.set", MinimumRole: RoleOperator, RequiresAudit: true},
	"storage_baseline.refresh": {Action: "storage_baseline.refresh", MinimumRole: RoleOperator, RequiresAudit: true},
	"runtime.approve":          {Action: "runtime.approve", MinimumRole: RoleAdministrator, RequiresAudit: true},
	"route.manage":             {Action: "route.manage", MinimumRole: RoleAdministrator, RequiresAudit: true},
	"credential.manage":        {Action: "credential.manage", MinimumRole: RoleAdministrator, RequiresAudit: true},
	"user_invite.manage":       {Action: "user_invite.manage", MinimumRole: RoleAdministrator, RequiresAudit: true},
	"backup.initiate":          {Action: "backup.initiate", MinimumRole: RoleAdministrator, RequiresAudit: true},
	"audit.read":               {Action: "audit.read", MinimumRole: RoleViewer, RequiresAudit: false},
}

func RequirementFor(action string) (Requirement, error) {
	requirement, ok := requirements[action]
	if !ok {
		return Requirement{}, fmt.Errorf("unknown NodeScope action %q", action)
	}
	return requirement, nil
}

func Authorize(role Role, action string) (Requirement, error) {
	requirement, err := RequirementFor(action)
	if err != nil {
		return Requirement{}, err
	}
	if !role.Allows(requirement.MinimumRole) {
		return Requirement{}, fmt.Errorf("role %q cannot perform %q", role, action)
	}
	return requirement, nil
}

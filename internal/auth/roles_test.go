package auth

import "testing"

func TestRoleHierarchy(t *testing.T) {
	cases := []struct {
		role    Role
		minimum Role
		allows  bool
	}{
		{RoleViewer, RoleViewer, true},
		{RoleViewer, RoleOperator, false},
		{RoleOperator, RoleViewer, true},
		{RoleOperator, RoleOperator, true},
		{RoleOperator, RoleAdministrator, false},
		{RoleAdministrator, RoleViewer, true},
		{RoleAdministrator, RoleOperator, true},
		{RoleAdministrator, RoleAdministrator, true},
		{Role("unknown"), RoleViewer, false},
	}

	for _, tc := range cases {
		if got := tc.role.Allows(tc.minimum); got != tc.allows {
			t.Fatalf("role %q minimum %q: expected %t, got %t", tc.role, tc.minimum, tc.allows, got)
		}
	}
}

func TestOperatorActionsRequireAuditIntent(t *testing.T) {
	requirement, err := Authorize(RoleOperator, "collection_interval.set")
	if err != nil {
		t.Fatalf("expected Operator to set collection interval, received %v", err)
	}
	if !requirement.RequiresAudit {
		t.Fatal("expected interval change to require an audit intent")
	}
}

func TestViewerCannotChangeCollectionInterval(t *testing.T) {
	if _, err := Authorize(RoleViewer, "collection_interval.set"); err == nil {
		t.Fatal("expected Viewer interval change to be rejected")
	}
}

func TestAdministratorOnlyRouteManagement(t *testing.T) {
	if _, err := Authorize(RoleOperator, "route.manage"); err == nil {
		t.Fatal("expected Operator route management to be rejected")
	}
	if _, err := Authorize(RoleAdministrator, "route.manage"); err != nil {
		t.Fatalf("expected Administrator route management to be allowed, received %v", err)
	}
}

func TestUnknownActionFailsClosed(t *testing.T) {
	if _, err := Authorize(RoleAdministrator, "arbitrary.command.execute"); err == nil {
		t.Fatal("expected unknown action to fail closed")
	}
}

package mcpserver

import (
	"context"
	"testing"
)

func TestFleetStatusAllowsViewer(t *testing.T) {
	service := &MemoryService{Hosts: []FleetHost{{ID: "framework", Name: "Framework", Freshness: "fresh"}}}
	server := Server{Service: service}
	_, output, err := server.fleetStatus(WithPrincipal(context.Background(), Principal{ID: "viewer-1", Role: RoleViewer}), nil, emptyInput{})
	if err != nil {
		t.Fatalf("viewer fleet status: %v", err)
	}
	if len(output.Hosts) != 1 || output.Hosts[0].ID != "framework" {
		t.Fatalf("unexpected host output %#v", output)
	}
}

func TestAcknowledgeRequiresOperator(t *testing.T) {
	service := &MemoryService{}
	server := Server{Service: service}
	_, _, err := server.acknowledgeAlert(WithPrincipal(context.Background(), Principal{ID: "viewer-1", Role: RoleViewer}), nil, acknowledgeInput{AlertID: "alert-1"})
	if err == nil {
		t.Fatal("viewer acknowledgement must be denied")
	}
	if len(service.Actions) != 0 {
		t.Fatalf("denied action must not reach service: %#v", service.Actions)
	}
}

func TestOperatorActionUsesPrincipal(t *testing.T) {
	service := &MemoryService{}
	server := Server{Service: service}
	_, output, err := server.setCollectionInterval(WithPrincipal(context.Background(), Principal{ID: "operator-1", Role: RoleOperator}), nil, intervalInput{HostID: "asus", IntervalSeconds: 10})
	if err != nil || !output.Audited {
		t.Fatalf("operator action failed: %#v %v", output, err)
	}
	if len(service.Actions) != 1 || service.Actions[0] != "operator-1:interval:asus:10" {
		t.Fatalf("unexpected action %#v", service.Actions)
	}
}

package mcpserver

import (
	"context"
	"strconv"
)

type MemoryService struct {
	Hosts   []FleetHost
	Actions []string
}

func (service *MemoryService) FleetStatus(_ context.Context, _ Principal) ([]FleetHost, error) {
	return append([]FleetHost(nil), service.Hosts...), nil
}
func (service *MemoryService) AcknowledgeAlert(_ context.Context, principal Principal, alertID, _ string) error {
	service.Actions = append(service.Actions, principal.ID+":acknowledge:"+alertID)
	return nil
}
func (service *MemoryService) SetCollectionInterval(_ context.Context, principal Principal, hostID string, seconds int) error {
	service.Actions = append(service.Actions, principal.ID+":interval:"+hostID+":"+strconv.Itoa(seconds))
	return nil
}
func (service *MemoryService) RefreshStorageBaseline(_ context.Context, principal Principal, hostID string, _ bool) error {
	service.Actions = append(service.Actions, principal.ID+":baseline:"+hostID)
	return nil
}

package mcpserver

import (
	"context"
	"fmt"
	"strconv"
)

type MemoryService struct {
	Hosts   []FleetHost
	Actions []string
}

func (service *MemoryService) FleetStatus(_ context.Context, _ Principal) ([]FleetHost, error) {
	return append([]FleetHost(nil), service.Hosts...), nil
}
func (service *MemoryService) HostStatus(_ context.Context, _ Principal, hostID string) (FleetHost, error) {
	for _, host := range service.Hosts {
		if host.ID == hostID {
			return host, nil
		}
	}
	return FleetHost{}, fmt.Errorf("host %q was not found", hostID)
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

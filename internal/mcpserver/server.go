// Package mcpserver exposes NodeScope's authorized, metadata-only fleet tools
// through the Model Context Protocol. Tool outputs intentionally omit prompts,
// completions, bearer tokens, and raw request/response payloads.
package mcpserver

import (
	"context"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pakgrou-porg/nodescope/internal/auth"
)

// Role aliases the central authorization vocabulary to preserve public MCP
// compatibility while preventing an independent role hierarchy.
type Role = auth.Role

const (
	RoleViewer        = auth.RoleViewer
	RoleOperator      = auth.RoleOperator
	RoleAdministrator = auth.RoleAdministrator
)

type Principal struct {
	ID   string
	Role Role
}

type principalKey struct{}

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, principal)
}
func principalFrom(ctx context.Context) (Principal, error) {
	principal, ok := ctx.Value(principalKey{}).(Principal)
	if !ok || principal.ID == "" {
		return Principal{}, fmt.Errorf("MCP principal is required")
	}
	return principal, nil
}
func requireRole(ctx context.Context, minimum Role) (Principal, error) {
	principal, err := principalFrom(ctx)
	if err != nil {
		return Principal{}, err
	}
	if !principal.Role.Allows(minimum) {
		return Principal{}, fmt.Errorf("%s role is required", minimum)
	}
	return principal, nil
}

type FleetHost struct {
	ID                     string     `json:"id"`
	Name                   string     `json:"name"`
	Platform               string     `json:"platform"`
	Freshness              string     `json:"freshness"`
	LatestServerReceipt    *time.Time `json:"latest_server_receipt,omitempty"`
	MetricCount            int64      `json:"metric_count"`
	UnavailableMetricCount int64      `json:"unavailable_metric_count"`
	StaleMetricCount       int64      `json:"stale_metric_count"`
	ClockOffsetSeconds     *float64   `json:"clock_offset_seconds,omitempty"`
	ClockOffsetQuality     *string    `json:"clock_offset_quality,omitempty"`
}

type Service interface {
	FleetStatus(context.Context, Principal) ([]FleetHost, error)
	HostStatus(context.Context, Principal, string) (FleetHost, error)
	AcknowledgeAlert(context.Context, Principal, string, string) error
	SetCollectionInterval(context.Context, Principal, string, int) error
	RefreshStorageBaseline(context.Context, Principal, string, bool) error
}

type Server struct{ Service Service }

func (server Server) New() *mcp.Server {
	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "nodescope", Version: "0.1.0"}, nil)
	mcp.AddTool(mcpServer, &mcp.Tool{Name: "nodescope_fleet_status", Description: "Read explicit NodeScope fleet freshness and metric-quality status. No inference prompt or response content is returned."}, server.fleetStatus)
	mcp.AddTool(mcpServer, &mcp.Tool{Name: "nodescope_acknowledge_alert", Description: "Acknowledge a NodeScope alert as an authorized operator. The action is audited."}, server.acknowledgeAlert)
	mcp.AddTool(mcpServer, &mcp.Tool{Name: "nodescope_set_collection_interval", Description: "Set the global or host-specific collection interval from 1 through 60 seconds as an authorized operator."}, server.setCollectionInterval)
	mcp.AddTool(mcpServer, &mcp.Tool{Name: "nodescope_refresh_storage_baseline", Description: "Refresh a reviewed learned storage baseline as an authorized operator."}, server.refreshStorageBaseline)
	return mcpServer
}

type emptyInput struct{}
type fleetStatusOutput struct {
	Hosts []FleetHost `json:"hosts"`
}

func (server Server) fleetStatus(ctx context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, fleetStatusOutput, error) {
	principal, err := requireRole(ctx, RoleViewer)
	if err != nil {
		return nil, fleetStatusOutput{}, err
	}
	hosts, err := server.Service.FleetStatus(ctx, principal)
	return nil, fleetStatusOutput{Hosts: hosts}, err
}

type acknowledgeInput struct {
	AlertID string `json:"alert_id" jsonschema:"the alert identifier to acknowledge"`
	Note    string `json:"note,omitempty" jsonschema:"optional operator note without secrets or inference content"`
}
type actionOutput struct {
	Status  string `json:"status"`
	Audited bool   `json:"audited"`
}

func (server Server) acknowledgeAlert(ctx context.Context, _ *mcp.CallToolRequest, input acknowledgeInput) (*mcp.CallToolResult, actionOutput, error) {
	principal, err := requireRole(ctx, RoleOperator)
	if err != nil {
		return nil, actionOutput{}, err
	}
	if input.AlertID == "" {
		return nil, actionOutput{}, fmt.Errorf("alert_id is required")
	}
	err = server.Service.AcknowledgeAlert(ctx, principal, input.AlertID, input.Note)
	return nil, actionOutput{Status: "acknowledged", Audited: err == nil}, err
}

type intervalInput struct {
	HostID          string `json:"host_id,omitempty" jsonschema:"optional host identifier; omit for the global default"`
	IntervalSeconds int    `json:"interval_seconds" jsonschema:"collection interval from 1 to 60 seconds"`
}

func (server Server) setCollectionInterval(ctx context.Context, _ *mcp.CallToolRequest, input intervalInput) (*mcp.CallToolResult, actionOutput, error) {
	principal, err := requireRole(ctx, RoleOperator)
	if err != nil {
		return nil, actionOutput{}, err
	}
	if input.IntervalSeconds < 1 || input.IntervalSeconds > 60 {
		return nil, actionOutput{}, fmt.Errorf("interval_seconds must be from 1 to 60")
	}
	err = server.Service.SetCollectionInterval(ctx, principal, input.HostID, input.IntervalSeconds)
	return nil, actionOutput{Status: "collection_interval_updated", Audited: err == nil}, err
}

type baselineInput struct {
	HostID           string `json:"host_id" jsonschema:"host identifier"`
	AcknowledgedDiff bool   `json:"acknowledged_diff" jsonschema:"must be true after an operator reviewed the baseline diff"`
}

func (server Server) refreshStorageBaseline(ctx context.Context, _ *mcp.CallToolRequest, input baselineInput) (*mcp.CallToolResult, actionOutput, error) {
	principal, err := requireRole(ctx, RoleOperator)
	if err != nil {
		return nil, actionOutput{}, err
	}
	if input.HostID == "" || !input.AcknowledgedDiff {
		return nil, actionOutput{}, fmt.Errorf("host_id and acknowledged_diff=true are required")
	}
	err = server.Service.RefreshStorageBaseline(ctx, principal, input.HostID, input.AcknowledgedDiff)
	return nil, actionOutput{Status: "storage_baseline_refreshed", Audited: err == nil}, err
}

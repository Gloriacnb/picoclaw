package mcp

import (
	"context"
	"time"

	"github.com/sipeed/picoclaw/pkg/config"
	runtimeevents "github.com/sipeed/picoclaw/pkg/events"
)

const mcpEventPublishTimeout = 100 * time.Millisecond

func (m *Manager) publishServerEvent(
	kind runtimeevents.Kind,
	serverName string,
	cfg config.MCPServerConfig,
	toolCount int,
	err error,
) {
	if m == nil || m.runtimeEvents == nil {
		return
	}

	severity := runtimeevents.SeverityInfo
	if err != nil {
		severity = runtimeevents.SeverityError
	}
	payload := ServerEventPayload{
		Server:    serverName,
		Type:      mcpTransportType(cfg),
		URL:       cfg.URL,
		Command:   cfg.Command,
		ToolCount: toolCount,
	}
	if err != nil {
		payload.Error = err.Error()
	}

	ctx, cancel := context.WithTimeout(context.Background(), mcpEventPublishTimeout)
	defer cancel()
	m.runtimeEvents.Publish(ctx, runtimeevents.Event{
		Kind:     kind,
		Source:   runtimeevents.Source{Component: "mcp", Name: serverName},
		Severity: severity,
		Payload:  payload,
	})
}

func (m *Manager) publishToolDiscovered(serverName string, cfg config.MCPServerConfig, toolName string) {
	if m == nil || m.runtimeEvents == nil {
		return
	}
	payload := ServerEventPayload{
		Server:  serverName,
		Type:    mcpTransportType(cfg),
		URL:     cfg.URL,
		Command: cfg.Command,
		Tool:    toolName,
	}
	ctx, cancel := context.WithTimeout(context.Background(), mcpEventPublishTimeout)
	defer cancel()
	m.runtimeEvents.Publish(ctx, runtimeevents.Event{
		Kind:     runtimeevents.KindMCPToolDiscovered,
		Source:   runtimeevents.Source{Component: "mcp", Name: serverName},
		Severity: runtimeevents.SeverityInfo,
		Payload:  payload,
	})
}

func mcpTransportType(cfg config.MCPServerConfig) string {
	if cfg.Type != "" {
		return cfg.Type
	}
	if cfg.URL != "" {
		return "sse"
	}
	if cfg.Command != "" {
		return "stdio"
	}
	return ""
}

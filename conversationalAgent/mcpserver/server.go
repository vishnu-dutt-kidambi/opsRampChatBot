// Package mcpserver exposes the OpsRamp ChatBot tools as a Model Context Protocol
// (MCP) server. This allows any MCP-compatible client (Claude Desktop, VS Code
// Copilot, Cursor, custom agents, etc.) to discover and call our 11 operational
// tools over stdio or HTTP.
//
// The tools themselves are not reimplemented — this package wraps the existing
// tools.Execute() dispatcher, converting between MCP request/response types and
// the project's internal ToolCall format.
package mcpserver

import (
	"context"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"opsramp-agent/juniper"
	"opsramp-agent/knowledge"
	"opsramp-agent/opsramp"
	"opsramp-agent/tools"
)

// StartStdio starts the MCP server on stdin/stdout (stdio transport).
// This is the standard transport for tools like Claude Desktop and VS Code.
func StartStdio(client *opsramp.Client, kb *knowledge.KnowledgeBase, jc ...*juniper.Client) error {
	s := newMCPServer(client, kb, jc...)
	log.Println("[MCP] Starting HPE Autopilot MCP Server on stdio...")
	return server.ServeStdio(s)
}

// StartHTTP starts the MCP server on an HTTP endpoint using Streamable HTTP transport.
func StartHTTP(addr string, client *opsramp.Client, kb *knowledge.KnowledgeBase, jc ...*juniper.Client) error {
	s := newMCPServer(client, kb, jc...)
	log.Printf("[MCP] Starting HPE Autopilot MCP Server on %s ...\n", addr)
	httpServer := server.NewStreamableHTTPServer(s)
	return httpServer.Start(addr)
}

// newMCPServer creates the MCP server with all OpsRamp tools registered.
func newMCPServer(client *opsramp.Client, kb *knowledge.KnowledgeBase, jc ...*juniper.Client) *server.MCPServer {
	var juniperClient *juniper.Client
	if len(jc) > 0 {
		juniperClient = jc[0]
	}
	s := server.NewMCPServer(
		"hpe-autopilot-mcp-server",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithRecovery(),
	)

	// Register all tools from the existing tool definitions.
	// We convert each Ollama-format tool schema into an mcp.Tool.
	for _, def := range tools.GetToolDefinitions() {
		td := def // capture loop variable
		mcpTool := convertTool(td)
		s.AddTool(mcpTool, makeHandler(client, kb, juniperClient, td.Function.Name))
	}

	return s
}

// convertTool translates an Ollama-format tool definition into an mcp.Tool.
func convertTool(def tools.Tool) mcp.Tool {
	opts := []mcp.ToolOption{
		mcp.WithDescription(def.Function.Description),
	}

	for name, prop := range def.Function.Parameters.Properties {
		isRequired := false
		for _, r := range def.Function.Parameters.Required {
			if r == name {
				isRequired = true
				break
			}
		}

		propOpts := []mcp.PropertyOption{
			mcp.Description(prop.Description),
		}
		if isRequired {
			propOpts = append(propOpts, mcp.Required())
		}
		if len(prop.Enum) > 0 {
			propOpts = append(propOpts, mcp.Enum(prop.Enum...))
		}

		// All our tool parameters are strings
		opts = append(opts, mcp.WithString(name, propOpts...))
	}

	return mcp.NewTool(def.Function.Name, opts...)
}

// makeHandler returns an MCP tool handler that delegates to tools.ExecuteWithOptions().
func makeHandler(client *opsramp.Client, kb *knowledge.KnowledgeBase, jc *juniper.Client, toolName string) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract string arguments from the MCP request
		args := make(map[string]string)
		for k, v := range request.GetArguments() {
			if s, ok := v.(string); ok {
				args[k] = s
			} else {
				args[k] = fmt.Sprintf("%v", v)
			}
		}

		// Build the internal ToolCall
		call := tools.ToolCall{
			Name:      toolName,
			Arguments: args,
		}

		// Delegate to the existing tool execution logic
		opts := tools.ExecuteOptions{
			KB:      kb,
			Juniper: jc,
		}
		result, err := tools.ExecuteWithOptions(client, call, opts)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(result), nil
	}
}

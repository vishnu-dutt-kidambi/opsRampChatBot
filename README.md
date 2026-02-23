# OpsRamp ChatBot

An AI-powered CLI that lets you query your IT infrastructure using natural language. Built with Go + Ollama, using a tool-calling agent pattern over mock OpsRamp APIs — now with **RAG-powered knowledge base** for operations runbooks and **MCP server mode** for integration with Claude Desktop, VS Code, and other MCP-compatible clients.

## Architecture

```
User Question
     │
     ▼
┌─────────────┐
│ LLM (Ollama) │ ── receives tool descriptions in system prompt
└──────┬──────┘
       │ tool_call (e.g., search_alerts state=Critical)
       ▼
┌─────────────┐
│ Tool Router  │ ── dispatches to mock OpsRamp client or RAG pipeline
└──────┬──────┘
       │ JSON result
       ▼
┌─────────────┐              ┌──────────────────┐
│ LLM (Ollama) │              │ Knowledge Base   │
│ summarizes   │◄────────────│ PDF → chunks →   │
│ tool results │              │ embed → search   │
└──────┬──────┘              └──────────────────┘
       │
       ▼
  Final Answer

Dual-Mode Architecture:
┌───────────────────────────────────────────────────────┐
│                  opsramp-agent binary                  │
├───────────┬──────────┬──────────────┬─────────────────┤
│ CLI REPL  │ Web UI   │ MCP (stdio)  │ MCP (HTTP)      │
│ (default) │ (--web)  │ (--mcp)      │ (--mcp-http)    │
├───────────┴──────────┴──────────────┴─────────────────┤
│              Shared Tool Layer (8 tools)              │
│              OpsRamp Client + Knowledge Base          │
└───────────────────────────────────────────────────────┘
```

## Quick Start

### Native (recommended for Mac — uses Apple GPU)

```bash
cd conversationalAgent

# 1. Install Ollama + pull models (~5GB)
make setup

# 2. Run the agent (choose a mode)
make run        # Terminal REPL
make web        # Browser chat UI on http://localhost:8080
make mcp        # MCP server (stdio) — for Claude Desktop, VS Code, etc.
make mcp-http   # MCP server (HTTP) on http://localhost:8081
```

### Docker

```bash
# Mac (recommended) — App in Docker, Ollama native on host (fast, GPU)
make docker-web-mac

# Full Docker — Everything in containers (portable, CPU-only on Mac)
make docker-setup    # First time: pull Ollama image + download models
make docker-web      # Start Web UI at http://localhost:8080
```

## Example Questions

| Category | Example |
|----------|---------|
| **Alerts** | "Show me all critical alerts" |
| **Alerts** | "Any unacknowledged warnings?" |
| **Resources** | "List all AWS resources" |
| **Resources** | "What Kubernetes nodes do we have?" |
| **Incidents** | "What incidents are open?" |
| **Incidents** | "Are there any SLA-breached tickets?" |
| **Investigation** | "Investigate web-server-prod-01" |
| **Investigation** | "What's wrong with db-primary-01?" |
| **Overview** | "Give me an environment summary" |
| **Capacity** | "Predict capacity for db-primary-01" |
| **Capacity** | "Which resources are at risk?" |
| **Knowledge Base** | "What is the runbook for high CPU usage?" |
| **Knowledge Base** | "How do I troubleshoot disk space full?" |

## Simulated Environment

The mock data simulates a mid-size enterprise with:

- **22 resources** across AWS, Azure, GCP, and on-prem (VMware)
- **8 active alerts** (3 Critical, 3 Warning, 2 Info)
- **7 incidents** (5 Open, 2 Resolved)
- Resource types: Linux, Windows, Azure SQL, Azure Functions, VMware ESXi
- Roles: web servers, app servers, databases, cache, message queue, Kubernetes, CI/CD

## Project Structure

```
opsRampChatBot/                         # Umbrella repo (Go workspace)
├── go.work                             # Go workspace: ties both modules together
├── Dockerfile                          # Multi-stage Docker build (root context)
├── docker-compose.yml                  # Docker orchestration (Ollama + Web/MCP/CLI)
├── Makefile                            # Docker commands (docker-web-mac, docker-setup, etc.)
├── .dockerignore                       # Keeps Docker build context lean
├── README.md                           # This file
├── .vscode/
│   ├── settings.json                   # VS Code Go toolchain settings
│   ├── launch.json                     # Debug configs for both projects
│   └── mcp.json                        # MCP server config for VS Code Copilot
│
├── conversationalAgent/                # OpsRamp ChatBot (main project)
│   ├── main.go                         # CLI entry point + mode routing (CLI/Web/MCP)
│   ├── server.go                       # Web server (chat API + embedded UI)
│   ├── Makefile                        # Native build & run commands
│   ├── go.mod                          # Module: opsramp-agent (depends on pdf-qa-agent)
│   ├── agent/
│   │   └── agent.go                    # LLM orchestrator (tool-calling loop)
│   ├── tools/
│   │   └── tools.go                    # Tool definitions + execution dispatcher (8 tools)
│   ├── mcpserver/
│   │   └── server.go                   # MCP server — wraps tools for stdio/HTTP transport
│   ├── knowledge/
│   │   └── knowledge.go                # RAG pipeline (embedder, vector store, chunker, PDF reader)
│   ├── opsramp/
│   │   ├── models.go                   # OpsRamp API data models
│   │   ├── client.go                   # Mock API client with filtering
│   │   └── forecast.go                 # Capacity forecasting (linear regression)
│   ├── mockdata/
│   │   ├── alerts.go                   # Realistic alert data
│   │   ├── resources.go                # Multi-cloud resource inventory
│   │   ├── incidents.go                # Incident/ticket data
│   │   └── metric_history.go           # 30-day metric series
│   ├── runbooks/
│   │   └── opsramp_operations_runbook.pdf  # Operations runbook (9 sections)
│   ├── web/
│   │   └── index.html                  # Browser-based chat UI (go:embed)
│   └── generate_runbook.py             # Script to regenerate runbook PDF
│
└── pdfReaderAIAgent/                   # PDF RAG library (shared dependency)
    ├── main.go                         # Standalone PDF Q&A CLI
    ├── rag/
    │   ├── pdf.go                      # PDF text extraction
    │   ├── chunker.go                  # Text chunking
    │   ├── embeddings.go               # Ollama embedding client
    │   └── vectorstore.go              # Cosine similarity vector store
    ├── go.mod                          # Module: pdf-qa-agent
    └── README.md
```

## How It Works

1. **Tool-Calling Pattern**: The LLM receives descriptions of available tools (search_alerts, search_resources, etc.) in its system prompt. When a user asks a question, the LLM decides which tool(s) to call and with what parameters.

2. **Mock OpsRamp Client**: Instead of calling real OpsRamp APIs, the client operates on in-memory mock data that matches real OpsRamp API response schemas. Swapping to real APIs requires zero model changes.

3. **Multi-Turn Conversation**: The agent maintains conversation history, supporting follow-up questions and contextual responses.

4. **Investigation Correlation**: The `investigate_resource` tool combines resource details + alerts + incidents + metrics into a single comprehensive report.

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `OLLAMA_HOST` | `http://localhost:11434` | Ollama server URL |
| `LLM_MODEL` | `llama3.1` | LLM model for tool-calling |
| `EMBEDDING_MODEL` | `nomic-embed-text` | Ollama model for RAG embeddings |
| `RUNBOOK_DIR` | `runbooks` | Directory containing PDF runbook files |

## Available Tools

| Tool | OpsRamp API Equivalent | Description |
|------|----------------------|-------------|
| `search_alerts` | `GET /api/v2/.../alerts/search` | Filter alerts by state, priority, resource |
| `search_resources` | `GET /api/v2/.../resources/search` | Find resources by cloud, region, type, tags |
| `get_resource_details` | `GET /api/v2/.../resources/{id}` | Deep resource info with metrics |
| `search_incidents` | `GET /api/v2/.../incidents/search` | Filter incidents by status, priority |
| `investigate_resource` | Composite query | Full investigation report |
| `get_environment_summary` | Dashboard API | High-level environment overview |
| `predict_capacity` | Metrics + linear regression | Forecast resource usage & days until threshold |
| `search_knowledge_base` | RAG over PDF runbooks | Search operations runbooks via embeddings |

## MCP Server Mode

The agent can run as a [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server, exposing all 8 tools to any MCP-compatible client without requiring a local LLM. The MCP client's own LLM decides which tools to call.

### Transports

| Transport | Flag | Use Case |
|-----------|------|----------|
| **stdio** | `--mcp` | Claude Desktop, VS Code Copilot, local MCP clients |
| **HTTP** | `--mcp-http` | Remote agents, multi-hop MCP, browser-based clients |

### Claude Desktop Configuration

Add to `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "opsramp": {
      "command": "/path/to/opsramp-agent",
      "args": ["--mcp"],
      "env": {
        "OLLAMA_HOST": "http://localhost:11434"
      }
    }
  }
}
```

### VS Code / Copilot Configuration

Add to `.vscode/mcp.json` in the `opsRampChatBot` workspace root:

```json
{
  "servers": {
    "opsramp": {
      "type": "stdio",
      "command": "${workspaceFolder}/conversationalAgent/opsramp-agent",
      "args": ["--mcp"],
      "env": {
        "OLLAMA_HOST": "http://localhost:11434",
        "EMBEDDING_MODEL": "nomic-embed-text",
        "RUNBOOK_DIR": "${workspaceFolder}/conversationalAgent/runbooks"
      }
    }
  }
}
```

### How MCP Mode Differs from CLI/Web Mode

| Aspect | CLI / Web Mode | MCP Mode |
|--------|---------------|----------|
| **LLM** | Built-in (Ollama llama3.1) | Client provides its own (Claude, GPT-4, etc.) |
| **Tool calling** | Agent orchestrates tool loop | MCP client orchestrates |
| **Transport** | Terminal / HTTP chat API | stdio or Streamable HTTP (MCP protocol) |
| **Conversation** | Multi-turn with history | Stateless per tool call |
| **Use case** | Standalone chatbot | Composable tool server |

## RoadMap Ideas

- [x] **Phase 1**: Mock data + CLI + basic tool-calling
- [x] **Phase 2**: Web UI with browser-based chat (go:embed)
- [x] **Phase 3**: Capacity forecasting with linear regression
- [x] **Phase 4**: Knowledge base — RAG over PDF runbooks (current)
- [x] **Phase 5**: MCP server mode — dual-mode binary serves as both standalone chatbot AND MCP tool server
  - Stdio transport for Claude Desktop, VS Code Copilot (--mcp flag)
  - Streamable HTTP transport for remote MCP clients (--mcp-http flag)
  - All 8 tools auto-converted from Ollama format to MCP format via mcp-go SDK
- [ ] **Phase 6**: Multi-MCP agent architecture — generic agent orchestrator discovers and composes MCP servers dynamically
  - MCP Gateway (auth, rate-limiting, observability across servers)
  - Off-the-shelf MCP servers for Jira, PagerDuty, Slack (replace custom integrations)
  - Agent mesh — specialist agents (Ops, Dev, Security) coordinating via shared MCP infrastructure
- [ ] **Phase 7**: Real OpsRamp API integration (OAuth2 + tenant config)
- [ ] **Phase 8**: Proactive insights + recommendations
- [ ] **Phase 9**: Actionable operations (acknowledge alerts, create incidents)
- [ ] **Phase 10**: Slack/Teams integration
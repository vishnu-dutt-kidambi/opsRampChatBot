# HPE Autopilot

An AI-powered IT operations assistant that lets you query your infrastructure using natural language. Ask about alerts, resources, incidents, capacity forecasts, network health, and operations runbooks — the agent figures out what to look up and responds with clear, actionable answers.

Built with **Go + Ollama**, using a tool-calling agent pattern over OpsRamp APIs (currently mock) and Juniper switch telemetry. Also runs as an **MCP server** for integration with VS Code Copilot, Claude Desktop, and other MCP-compatible clients.

## Web UI

![HPE Autopilot Web UI](conversationalAgent/screenshots/web-ui.png)

## MCP Server (VS Code Copilot)

![HPE Autopilot MCP Server](conversationalAgent/screenshots/mcp-server.png)

## What You Can Ask

- **Alerts** — "Show me all critical alerts" · "Any P0 alerts?"
- **Resources** — "List all AWS resources" · "Show HPE GreenLake servers"
- **Incidents** — "Show open incidents" · "Any urgent tickets?"
- **Investigation** — "Investigate web-server-prod-01" · "Why is the DB slow?"
- **Environment** — "Give me an environment summary"
- **Capacity Forecast** — "Predict capacity for db-primary-01" · "Which resources are at risk?"
- **Knowledge Base** — "What is the runbook for high CPU?" · "How to fix disk full?" · "Escalation contacts?"
- **Network Correlation** — "Correlate network for k8s-node-04" · "Is the network causing latency?"
- **Blast Radius** — "What's the blast radius for k8s-node-04?" · "How many users are affected?"
- **Guided Remediation** — "Give me a remediation plan for k8s-node-04" · "How do I fix the network issue?"
- **End-to-End** — "Why is the GreenLake portal slow?" (agent chains 6 tools autonomously)

## Capabilities

| Tool | Description |
|------|-------------|
| **Search Alerts** | Filter by state (Critical/Warning), priority, resource |
| **Search Resources** | Find servers across AWS, Azure, HPE GreenLake, on-prem |
| **Resource Details** | Deep-dive into configuration, metrics, tags |
| **Search Incidents** | Filter tickets by status, priority, SLA |
| **Investigate Resource** | Correlated view of alerts + incidents + metrics for a resource |
| **Environment Summary** | High-level infrastructure health dashboard |
| **Capacity Forecasting** | Linear regression on 30-day metric history to predict CPU/memory/disk exhaustion |
| **Knowledge Base (RAG)** | Retrieval-augmented generation over operations runbooks (PDF) using vector embeddings |
| **Network Correlation** | Correlate server issues with Juniper switch telemetry (packet loss, CRC errors, link flaps, latency, jitter, duplex) |
| **Blast Radius** | Map impact of infrastructure issues across applications, services, and user groups via dependency graph traversal |
| **Guided Remediation** | Generate step-by-step remediation plans with exact Junos CLI commands, risk levels, and approval gates |

## How MCP Mode Works

In MCP mode, **Copilot (or Claude) is the LLM** — not Ollama. HPE Autopilot acts as a tool server only.

```
You (in VS Code / Claude Desktop)
 │
 ▼
Copilot / Claude  ← decides which tools to call
 │
 │  MCP protocol (stdio or HTTP)
 ▼
hpe-autopilot --mcp  ← executes tools, returns JSON
 │
 ├─ search_alerts, search_resources, etc. → OpsRamp data
 └─ search_knowledge_base → Ollama embeddings (only Ollama use in MCP mode)
 │
 ▼
Copilot / Claude  ← summarizes results back to you
```

- **Ollama LLM is NOT used** in MCP mode — Copilot's own model handles reasoning and tool selection
- **Ollama is only needed** for the embedding model (`nomic-embed-text`) that powers runbook search
- The 11 tools + their descriptions are advertised via MCP's `initialize` handshake

---

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

# Other commands
make build      # Build the Go binary
make clean      # Remove build artifacts
make test       # Run tests
make fmt        # Format Go code
make vet        # Run Go vet
make lint       # Run all linters (fmt + vet)
make help       # Show all available commands
```

### Docker (from repo root)

```bash
# Mac (recommended) — App in Docker, Ollama native on host (fast, GPU)
make docker-web-mac

# Full Docker — Everything in containers (portable, CPU-only on Mac)
make docker-setup    # First time: pull Ollama image + download models (~8GB)
make docker-build    # Build the Docker image
make docker-web      # Start Web UI at http://localhost:8080
make docker-mcp      # Start MCP HTTP server at http://localhost:8081
make docker-cli      # Run interactive CLI in Docker
make docker-logs     # Show logs for all running services
make docker-down     # Stop all containers
make docker-clean    # Stop containers and remove images + volumes
```

## Simulated Environment

The mock data simulates a mid-size enterprise with:

- **22 resources** across AWS, Azure, HPE GreenLake, and on-prem (VMware)
- **8 active alerts** (3 Critical, 3 Warning, 2 Info)
- **7 incidents** (5 Open, 2 Resolved)
- Resource types: Linux, Windows, Azure SQL, Azure Functions, VMware ESXi
- Roles: web servers, app servers, databases, cache, message queue, Kubernetes, CI/CD

---

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
│                HPE Autopilot binary                    │
├───────────┬──────────┬──────────────┬─────────────────┤
│ CLI REPL  │ Web UI   │ MCP (stdio)  │ MCP (HTTP)      │
│ (default) │ (--web)  │ (--mcp)      │ (--mcp-http)    │
├───────────┴──────────┴──────────────┴─────────────────┤
│             Shared Tool Layer (11 tools)              │
│          OpsRamp Client + Juniper + Knowledge Base    │
└───────────────────────────────────────────────────────┘
```

## Project Structure

```
HPEAutopilot/                           # Umbrella repo (Go workspace)
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
├── conversationalAgent/                # HPE Autopilot (main project)
│   ├── main.go                         # CLI entry point + mode routing (CLI/Web/MCP)
│   ├── server.go                       # Web server (chat API + embedded UI)
│   ├── Makefile                        # Native build & run commands
│   ├── go.mod                          # Module: opsramp-agent (depends on pdf-qa-agent)
│   ├── agent/
│   │   └── agent.go                    # LLM orchestrator (tool-calling loop)
│   ├── tools/
│   │   └── tools.go                    # Tool definitions + execution dispatcher (11 tools)
│   ├── mcpserver/
│   │   └── server.go                   # MCP server — wraps tools for stdio/HTTP transport
│   ├── knowledge/
│   │   └── knowledge.go                # RAG pipeline (embedder, vector store, chunker, PDF reader)
│   ├── opsramp/
│   │   ├── models.go                   # OpsRamp API data models
│   │   ├── client.go                   # Mock API client with filtering
│   │   └── forecast.go                 # Capacity forecasting (linear regression)
│   ├── juniper/
│   │   ├── models.go                   # Juniper network + blast radius + remediation models
│   │   └── client.go                   # Juniper client (correlation, blast radius, remediation)
│   ├── mockdata/
│   │   ├── alerts.go                   # Realistic alert data
│   │   ├── resources.go                # Multi-cloud resource inventory
│   │   ├── incidents.go                # Incident/ticket data
│   │   ├── metric_history.go           # 30-day metric series
│   │   ├── network.go                  # Juniper switch telemetry + port mappings
│   │   └── dependencies.go             # Infrastructure dependency graph (blast radius topology)
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
| `correlate_network` | Juniper Mist API | Correlate server issues with switch port telemetry |
| `blast_radius` | Dependency graph traversal | Map impact across apps, services, and user groups |
| `get_remediation_plan` | Junos CLI generation | Step-by-step remediation with commands and approvals |

## MCP Server Mode

HPE Autopilot can run as a [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server, exposing all 11 tools to any MCP-compatible client without requiring a local LLM. The MCP client's own LLM decides which tools to call.

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

Add to `.vscode/mcp.json` in the `HPEAutopilot` workspace root:

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
| **Use case** | Standalone assistant | Composable tool server |

## RoadMap Ideas

- [x] **Phase 1**: Mock data + CLI + basic tool-calling
- [x] **Phase 2**: Web UI with browser-based chat (go:embed)
- [x] **Phase 3**: Capacity forecasting with linear regression
- [x] **Phase 4**: Knowledge base — RAG over PDF runbooks (current)
- [x] **Phase 5**: MCP server mode — dual-mode binary serves as both standalone assistant AND MCP tool server
  - Stdio transport for Claude Desktop, VS Code Copilot (--mcp flag)
  - Streamable HTTP transport for remote MCP clients (--mcp-http flag)
  - All 11 tools auto-converted from Ollama format to MCP format via mcp-go SDK
- [x] **Phase 6**: Cross-domain intelligence — Juniper network correlation, blast radius analysis, and guided remediation
  - Network correlation: match servers to Juniper switch ports, analyze port-level telemetry
  - Blast radius: traverse infrastructure dependency graph to map affected apps/services/users
  - Guided remediation: generate Junos CLI commands with risk levels and approval gates
  - End-to-end scenario: "Why is the GreenLake portal slow?" → 6 autonomous tool calls → actionable fix
- [ ] **Phase 7**: Multi-MCP agent architecture — generic agent orchestrator discovers and composes MCP servers dynamically
  - MCP Gateway (auth, rate-limiting, observability across servers)
  - Off-the-shelf MCP servers for Jira, PagerDuty, Slack (replace custom integrations)
  - Agent mesh — specialist agents (Ops, Dev, Security) coordinating via shared MCP infrastructure
- [ ] **Phase 8**: Real OpsRamp API integration (OAuth2 + tenant config)
- [ ] **Phase 9**: Actionable operations (acknowledge alerts, create incidents)
- [ ] **Phase 10**: Slack/Teams integration
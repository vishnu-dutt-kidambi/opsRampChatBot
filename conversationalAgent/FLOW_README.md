# HPE ChatBot — Flow Documentation

> **Generated:** March 10, 2026
> **Model:** Configurable — default `llama3.1`; native Mac setup installs `mistral` (set `LLM_MODEL` env var to override)
> **Ollama:** localhost:11434 (set `OLLAMA_HOST` to override)
> **Purpose:** Documents every agent flow end-to-end — from user question, through Ollama tool-calling, to final answer. Includes real captured responses for regression testing.

---

## Table of Contents

1. [Architecture Overview](#1-architecture-overview)
2. [How a Request Flows Through the System](#2-how-a-request-flows-through-the-system)
3. [Flow 1: Search Alerts](#3-flow-1-search-alerts)
4. [Flow 2: Search Resources](#4-flow-2-search-resources)
5. [Flow 3: Get Resource Details](#5-flow-3-get-resource-details)
6. [Flow 4: Search Incidents](#6-flow-4-search-incidents)
7. [Flow 5: Investigate Resource](#7-flow-5-investigate-resource)
8. [Flow 6: Environment Summary](#8-flow-6-environment-summary)
9. [Flow 7: Predict Capacity (Single Resource)](#9-flow-7-predict-capacity-single-resource)
10. [Flow 8: Predict Capacity (All Resources — At-Risk)](#10-flow-8-predict-capacity-all-resources--at-risk)
11. [Flow 9: Search Knowledge Base (RAG)](#11-flow-9-search-knowledge-base-rag)
12. [Flow 10: Meta Questions (No Tool Call)](#12-flow-10-meta-questions-no-tool-call)
13. [Flow 11: MCP Server Mode](#13-flow-11-mcp-server-mode)
14. [Flow 12: Correlate Network (Juniper)](#14-flow-12-correlate-network-juniper)
15. [Flow 13: Blast Radius Analysis](#15-flow-13-blast-radius-analysis)
16. [Flow 14: Guided Remediation](#16-flow-14-guided-remediation)
17. [Flow 15: End-to-End — "Why is the GreenLake portal slow?"](#17-flow-15-end-to-end--why-is-the-greenlake-portal-slow)
18. [Ollama API Request/Response Format](#18-ollama-api-requestresponse-format)
19. [Testing Reference — All Captured Responses](#19-testing-reference--all-captured-responses)

---

## 1. Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        User Interface                           │
│          (Terminal REPL  or  Web UI  or  MCP Client)            │
└──────────────────────────┬──────────────────────────────────────┘
                           │ user question (string) or MCP tool call
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                      agent.Agent.Ask()                          │
│              OR  mcpserver (direct tool dispatch)               │
│                                                                 │
│  CLI/Web path:                                                  │
│  1. Append user message to conversation history                 │
│  2. Loop (max 8 rounds):                                        │
│     a. Send history + tool schemas → Ollama /api/chat           │
│     b. If response has tool_calls → execute tool → append       │
│        tool result to history → continue loop                   │
│     c. If response is plain text with tool name → parse &       │
│        execute (fallback for Mistral-style models)              │
│     d. If response is plain text (no tool) → return as answer   │
│                                                                 │
│  MCP path:                                                      │
│  1. Receive MCP CallToolRequest (tool name + arguments)         │
│  2. Convert to internal ToolCall → tools.Execute()              │
│  3. Return JSON result as MCP CallToolResult                    │
└──────────────────────────┬──────────────────────────────────────┘
                           │
              ┌────────────┼────────────┐
              ▼            ▼            ▼
     ┌──────────────┐ ┌──────────┐ ┌──────────────┐
     │  Ollama LLM  │ │  Tool    │ │  OpsRamp     │
     │  (configur-  │ │  Router  │ │  Mock Client │
     │   able)      │ │          │ │              │
     │              │ │          │ │              │
     │ /api/chat    │ │ tools.go │ │ client.go    │
     │ with tools[] │ │ Execute()│ │ forecast.go  │
     └──────────────┘ └──────────┘ └──────────────┘
```

### Components

| File | Role |
|------|------|
| `main.go` | Entry point — loads mock data, creates client & agent, loads knowledge base PDFs, routes to CLI/web/MCP mode |
| `server.go` | HTTP server (`--web` mode) — serves embedded HTML + `/api/chat` and `/api/clear` endpoints |
| `mcpserver/server.go` | MCP server (`--mcp` / `--mcp-http` mode) — wraps tools as MCP protocol handlers using mcp-go SDK |
| `agent/agent.go` | Orchestrator — manages conversation history, calls Ollama, dispatches tools (including KB), handles the tool-calling loop |
| `tools/tools.go` | Tool definitions (JSON schemas for LLM) + `Execute()` dispatcher + per-tool execution handlers (11 tools) |
| `opsramp/client.go` | Mock OpsRamp API client — search/filter/investigate methods operating on in-memory data |
| `opsramp/models.go` | Data structures mirroring OpsRamp API v2 (Alert, Resource, Incident, MetricSeries, etc.) |
| `opsramp/forecast.go` | Capacity forecasting engine — linear regression, R² confidence, threshold prediction |
| `juniper/models.go` | Juniper network models, blast radius types, remediation plan types |
| `juniper/client.go` | Juniper Mist client — network correlation, blast radius analysis, remediation plan generation |
| `knowledge/knowledge.go` | RAG pipeline — PDF extraction, text chunking, Ollama embeddings, vector store, cosine similarity search |
| `mockdata/*.go` | Mock data generators — alerts, resources, incidents, 30-day metric history, network switches, dependency graph |
| `runbooks/*.pdf` | Operations runbook PDFs loaded into the knowledge base at startup |
| `web/index.html` | Browser chat UI (embedded in binary via `go:embed`) — 11 quick-action buttons |

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OLLAMA_HOST` | `http://localhost:11434` | Ollama API endpoint |
| `LLM_MODEL` | `llama3.1` | LLM model for chat (native `make setup` installs `mistral`) |
| `EMBEDDING_MODEL` | `nomic-embed-text` | Embedding model for RAG knowledge base |
| `RUNBOOK_DIR` | `runbooks` | Directory containing PDF runbook files |

### Build & Run

The project provides **two** Makefiles — a root-level one for convenience and a `conversationalAgent/` one with the full target set. All root targets delegate via `$(MAKE) -C conversationalAgent`.

**From the repository root** (`make help`):

| Category | Command | Description |
|----------|---------|-------------|
| **Native** | `make setup` | Install Ollama + pull models (~5 GB) |
| | `make run` | Terminal REPL |
| | `make web` | Browser chat UI on `http://localhost:8080` |
| | `make mcp` | MCP server (stdio — Claude Desktop, VS Code) |
| | `make mcp-http` | MCP server (HTTP on `http://localhost:8081`) |
| | `make build` | Build the Go binary |
| | `make clean` | Remove build artifacts |
| | `make test` | Run tests |
| | `make fmt` / `make vet` / `make lint` | Code quality |
| **Docker** | `make docker-web-mac` | Recommended for Mac (native Ollama + Apple GPU) |
| | `make docker-web` | Full Docker mode (CPU-only, ~60 s/response) |
| | `make docker-setup` | Start Ollama container + pull models |
| | `make docker-mcp` | MCP HTTP server in Docker |
| | `make docker-cli` | Interactive CLI in Docker |
| | `make docker-down` | Stop all containers |
| | `make docker-clean` | Stop + remove images/volumes |

### Web UI Quick Actions

The browser chat UI (`--web` mode) shows a welcome screen with 11 quick-click buttons for common queries:

| Button | Query sent |
|--------|-----------|
| 🔴 Critical alerts | "Show me all critical alerts" |
| 📊 Environment summary | "Give me an environment summary" |
| 🔍 Investigate server | "Investigate web-server-prod-01" |
| 🎫 Open incidents | "Show me all open incidents" |
| 📈 Capacity forecast | "Predict capacity for db-primary-01" |
| ⚠️ At-risk resources | "Which servers are at risk of running out of capacity?" |
| 📖 High CPU runbook | "What is the runbook for high CPU usage?" |
| 🔌 Network correlation | "Correlate network for k8s-node-04" |
| 💥 Blast radius | "What is the blast radius for k8s-node-04?" |
| 🛠️ Remediation plan | "Give me a remediation plan for k8s-node-04" |
| 🚀 End-to-end investigation | "Why is the GreenLake portal slow?" |

---

## 2. How a Request Flows Through the System

Every user question follows this exact path:

### Step 1: User sends question
- **Terminal mode:** stdin → `main.go` REPL → `agent.Ask(question)`
- **Web mode:** HTTP POST `/api/chat` → `server.go` handler → `agent.Ask(question)`
- **MCP mode:** MCP client sends `tools/call` → `mcpserver/server.go` handler → `tools.Execute()` directly (no LLM involved server-side)

### Step 2: Agent prepares Ollama request
```go
// agent.go — callLLM()
reqBody := chatRequest{
    Model:    a.model,       // configurable via LLM_MODEL env (default: "llama3.1")
    Messages: a.history,     // system prompt + conversation so far
    Tools:    tools.GetToolDefinitions(),  // 11 tool schemas
    Stream:   false,
}
// POST to ${OLLAMA_HOST}/api/chat (default: http://localhost:11434/api/chat)
```

The **system prompt** (first message, role=system) tells the LLM:
- Its identity (HPE Operations Assistant)
- How to handle meta questions without tools
- Rules for tool usage (never fabricate data, always use tools for real queries)
- How to format answers after receiving tool results

The **tools array** contains JSON Schema definitions for all 11 tools. Ollama sends these to the LLM so it knows what functions are available and what parameters they accept.

### Step 3: Ollama responds with either a tool call OR a text answer

**Option A — Tool call response:**
```json
{
  "message": {
    "role": "assistant",
    "content": "",
    "tool_calls": [{
      "function": {
        "name": "search_alerts",
        "arguments": {"state": "Critical"}
      }
    }]
  }
}
```

**Option B — Direct text answer (no tools needed):**
```json
{
  "message": {
    "role": "assistant",
    "content": "I'm an HPE Operations Assistant. I have 10 capabilities spanning monitoring, network analysis, and guided remediation..."
  }
}
```

### Step 4: If tool call → execute and loop back

```
Agent receives tool_call → tools.Execute(client, toolCall)
  → executeSearchAlerts(client, args)
    → client.SearchAlerts("Critical", "", "", "")
      → filters in-memory alert slice
      → returns []Alert
    → marshal to JSON summary
  → return JSON string

Agent appends tool result to history as role="tool" message
Agent calls Ollama AGAIN with updated history (now includes tool result)
Ollama reads the JSON tool result and generates a human-friendly summary
```

### Step 5: Final answer returned to user

The loop continues until either:
- The LLM responds with plain text (no tool_calls) → that's the final answer
- Max 8 rounds reached → returns a "try a more specific question" message

### Context Management

| Mechanism | Value | Purpose |
|-----------|-------|---------|
| `maxToolResultLen` | 4000 chars | Truncates large tool JSON results to prevent context bloat |
| `maxHistoryMessages` | 20 messages | Trims oldest messages when history grows too long |
| `maxRounds` | 8 | Prevents infinite tool-calling loops |

---

## 3. Flow 1: Search Alerts

### User Query Examples
- "Show me all critical alerts"
- "Are there any alerts about CPU usage?"
- "What warnings do we have?"
- "Any P0 alerts?"

### Flow Diagram

```
User: "Show me all critical alerts"
  │
  ▼
agent.Ask() → callLLM()
  │
  ▼
Ollama receives: system prompt + user message + 11 tool schemas
Ollama decides: call search_alerts with {"state": "Critical"}
  │
  ▼
agent.executeTool() → tools.Execute()
  │
  ▼
executeSearchAlerts(client, {"state": "Critical"})
  → client.SearchAlerts("Critical", "", "", "")
  → loops through client.Alerts, filters where CurrentState == "Critical"
  → returns 4 matching alerts
  → marshals to JSON: {results: [...], count: 4}
  │
  ▼
Tool result appended to history as role="tool"
callLLM() called again with history including tool result
  │
  ▼
Ollama reads JSON, generates human summary with bullet points
Returns role="assistant" with text answer
```

### Ollama Debug Trace (Captured)

**Round 1 — LLM decides to call tool:**
```
[debug] Raw Ollama response (413 bytes):
{"model":"llama3.1","message":{"role":"assistant","content":"",
  "tool_calls":[{"function":{"name":"search_alerts","arguments":{"state":"Critical"}}}]},
  "prompt_eval_count":1801,"eval_count":18}
[debug] Parsed — Role: "assistant", Content length: 0, ToolCalls: 1
-> Calling tool: search_alerts(state="Critical")
```

**Round 2 — LLM summarizes tool result:**
```
[debug] Raw Ollama response (1206 bytes):
{"model":"llama3.1","message":{"role":"assistant","content":"Based on the tool output..."}}
[debug] Parsed — Role: "assistant", Content length: 904, ToolCalls: 0
```

### Tool JSON Schema (sent to Ollama)
```json
{
  "type": "function",
  "function": {
    "name": "search_alerts",
    "description": "Search and filter OpsRamp alerts...",
    "parameters": {
      "type": "object",
      "properties": {
        "state": {"type": "string", "enum": ["Critical", "Warning", "Ok", "Info"]},
        "priority": {"type": "string", "enum": ["P0", "P1", "P2", "P3", "P4", "P5"]},
        "resource_name": {"type": "string"},
        "query": {"type": "string"}
      }
    }
  }
}
```

### Tool Execution Code Path
```
tools.Execute("search_alerts", args)
  → executeSearchAlerts(client, args)
    → client.SearchAlerts(state, priority, resourceName, query)
      → opsramp/client.go: loops through client.Alerts
      → applies filters: state match, priority match, resource name partial match, query text search
      → matchesAlertQuery(): searches subject, description, serviceName, problemArea, component, metric, resource name
    → builds []alertSummary (id, subject, state, priority, resource, cloud, elapsed, acknowledged, ticketed)
    → returns JSON: {"results": [...], "count": N}
```

### Captured Response
```
Query: "Show me all critical alerts"

Agent Answer:
Based on the tool output, here are all critical alerts:

• CPU utilization exceeded 95% on web-server-prod-01: This alert has been active
  for 3 hours and 45 minutes. The resource is in AWS.
• Disk usage at 92% on db-primary-01: This alert has been active for 5 hours and
  30 minutes, with a priority of P0 (High). The resource is in AWS and is not
  ticketed yet.
• PING failed for monitoring-agent-staging-01: This alert has been active for
  50 minutes. The resource is on-premises and is not acknowledged or ticketed yet.
• Container restart loop detected on greenlake-portal pod: This alert has been
  active for 2 hours and 5 minutes, with a priority of P0 (High). The resource
  is on HPE GreenLake.

I recommend checking the resources immediately to prevent any further issues. You
may also want to create tickets for db-primary-01 and monitoring-agent-staging-01
as they are not yet ticketed.
```

---

## 4. Flow 2: Search Resources

### User Query Examples
- "List all resources running in AWS"
- "Show HPE GreenLake servers"
- "What do we have in production?"
- "Show servers in critical state"

### Flow Diagram

```
User: "List all resources running in AWS"
  │
  ▼
Ollama decides: call search_resources with {"cloud": "AWS"}
  │
  ▼
executeSearchResources(client, {"cloud": "AWS"})
  → client.SearchResources("AWS", "", "", "", "", "")
  → filters where r.Cloud == "AWS"
  → returns 10 resources
  → marshals to JSON with: id, name, type, cloud, region, instanceSize, state, cpu%, mem%, disk%
  │
  ▼
Ollama summarizes → final answer
```

### Tool Execution Code Path
```
tools.Execute("search_resources", args)
  → executeSearchResources(client, args)
    → client.SearchResources(cloud, region, type, state, tag, query)
      → opsramp/client.go: loops through client.Resources
      → applies filters: cloud match, region match, type partial match,
        state match, tag match (matchesTag), query search (matchesResourceQuery)
      → matchesResourceQuery(): searches name, hostname, IP, OS, type, cloud, region, tags
    → builds []resourceSummary (id, name, type, cloud, region, instanceSize, state, cpu%, mem%, disk%)
    → returns JSON: {"results": [...], "count": N}
```

### Captured Response
```
Query: "List all resources running in AWS"

Agent Answer:
Here are the resources running in AWS:

* Active Resources:
  + web-server-prod-01 (97.3% CPU, 62.1% Memory, 45% Disk)
  + app-server-prod-01 (42.5% CPU, 58.3% Memory, 38.2% Disk)
  + app-server-prod-02 (55% CPU, 88% Memory, 41.5% Disk)
  + web-server-prod-02 (35.2% CPU, 50.4% Memory, 42% Disk)
  + db-primary-01 (68.2% CPU, 75% Memory, 92% Disk)
  + db-replica-01 (30.5% CPU, 55% Memory, 78% Disk)
  + redis-cache-01 (15% CPU, 72% Memory, 10% Disk)
  + rabbitmq-prod-01 (22% CPU, 45% Memory, 30% Disk)
  + elasticsearch-prod-01 (55% CPU, 80% Memory, 70% Disk)
  + api-gateway-prod (28% CPU, 35% Memory, 20% Disk)
```

---

## 5. Flow 3: Get Resource Details

### User Query Examples
- "Show me details of db-primary-01"
- "What's the configuration of the redis server?"
- "Tell me about web-server-prod-01"

### Flow Diagram

```
User: "Show me details of db-primary-01"
  │
  ▼
Ollama decides: call get_resource_details with {"resource_name": "db-primary-01"}
  │
  ▼
executeGetResourceDetails(client, {"resource_name": "db-primary-01"})
  → client.GetResourceByName("db-primary-01")
  → partial match on Name or HostName (containsIgnoreCase)
  → returns full Resource struct with all fields
  → marshals entire Resource object to JSON (all fields: tags, metrics, config, etc.)
  │
  ▼
Ollama summarizes → final answer
```

### Key Difference from search_resources
- `search_resources` returns a **summary** (id, name, type, cloud, region, state, CPU/mem/disk)
- `get_resource_details` returns the **complete** Resource object (all tags, all metrics, IP, OS, hostname, instance size, everything)

### Captured Response
```
Query: "Show me details of db-primary-01"

Agent Answer:
Based on the tool output, here are the details of db-primary-01:

* Name: db-primary-01
* Host Name: db-primary-01
* IP Address: 10.0.2.10
* Type: Linux (Ubuntu 22.04)
* State: active
* Status: managed
* Resource Type: Linux
* Cloud: AWS
* Region: us-east-1
* Instance Size: r6g.4xlarge

Metrics:
* CPU Utilization: 68.2%
* Memory Utilization: 75%
* Disk Utilization: 92%
* Network In (Mbps): 500
* Network Out (Mbps): 450

It appears that db-primary-01 is a PostgreSQL database server running in
production, managed by OpsRamp. The metrics suggest that it's experiencing
high CPU utilization and disk usage.

Next Steps:
1. Investigate why db-primary-01 is experiencing high CPU utilization and disk usage.
2. Check for any available updates or patches for the PostgreSQL engine.
3. Consider scaling up the instance size to improve performance, if necessary.
```

---

## 6. Flow 4: Search Incidents

### User Query Examples
- "Show me all open incidents"
- "Any P1 incidents right now?"
- "Which incidents have SLA breaches?"
- "Show recent incidents"

### Flow Diagram

```
User: "Show me all open incidents"
  │
  ▼
Ollama decides: call search_incidents with {"status": "Open"}
  │
  ▼
executeSearchIncidents(client, {"status": "Open"})
  → client.SearchIncidents("Open", "", "")
  → filters where inc.Status == "Open"
  → returns 5 incidents
  → builds []incidentSummary with: id, subject, status, subStatus, priority,
    assignedTo, category, SLA breach flags, createdDate
  │
  ▼
Ollama summarizes → final answer
```

### Special: Natural Language Query Handling
The `matchesIncidentQuery()` function handles natural language terms:
```go
// These terms all map to Status="Open" filter:
recentTerms := []string{"recent", "latest", "new", "current", "active", "today"}
```
So "show me recent incidents" → matches all Open incidents.

### Tool Execution Code Path
```
tools.Execute("search_incidents", args)
  → executeSearchIncidents(client, args)
    → client.SearchIncidents(status, priority, query)
      → filters: status match, priority match, query search
      → matchesIncidentQuery(): handles natural language terms,
        searches subject, description, category.Name, subCategory.Name, assignedTo.Name
      → safely dereferences *Category pointers (nil check)
    → builds []incidentSummary with SLA breach info from *SLADetails
    → returns JSON: {"results": [...], "count": N}
```

### Captured Response
```
Query: "Show me all open incidents"

Agent Answer:
Here are the open incidents:

* INC-20260219-001: Production web tier degraded - high CPU on web-server-prod-01
  + Status: Open, In Progress
  + Priority: High
* INC-20260219-002: Database disk space critical - db-primary-01 at 92%
  + Status: Open, Waiting for Approval
  + Priority: Urgent (SLA Resolution Breach)
* INC-20260219-003: GreenLake portal pod crash loop in HPE GreenLake Kubernetes cluster
  + Status: Open, Investigation
  + Priority: Urgent
* INC-20260219-004: Staging server unreachable - monitoring-agent-staging-01
  + Status: Open, Escalated
  + Priority: Normal
* INC-20260218-005: SSL certificate renewal for api.acme.com
  + Status: Open, In Progress
  + Priority: High

Next Steps: Review the status and priority of each incident. If an incident
requires attention, assign a resource to address it promptly.
```

---

## 7. Flow 5: Investigate Resource

### User Query Examples
- "Investigate web-server-prod-01"
- "What's going on with the database server?"
- "Why is the production site slow?"
- "Diagnose app-server-prod-02"

### Flow Diagram

```
User: "Investigate web-server-prod-01"
  │
  ▼
Ollama decides: call investigate_resource with {"resource_name": "web-server-prod-01"}
  │
  ▼
executeInvestigateResource(client, {"resource_name": "web-server-prod-01"})
  → client.InvestigateResource("web-server-prod-01")
    → GetResourceByID() || GetResourceByName()   ← resolves name to resource
    → GetAlertsForResource(r.ID)                  ← finds all alerts for this resource
    → GetIncidentsForResource(r.ID)               ← finds all incidents for this resource
    → returns InvestigationReport{Resource, Alerts, Incidents, Metrics}
  → builds investigationResult with:
    - Resource details (name, IP, OS, cloud, region, instance, tags)
    - Metrics (CPU, memory, disk, network in/out)
    - Active alerts (id, subject, state, priority, elapsed)
    - Related incidents (id, subject, status, priority, assignedTo, SLA breaches)
  │
  ▼
Ollama summarizes with next steps → final answer
```

### This is the "deep dive" tool
It combines data from 3 sources in a single call:
1. **Resource metadata** — config, tags, metrics
2. **Active alerts** — what's currently alarming on this resource
3. **Related incidents** — tickets linked to this resource

### Captured Response
```
Query: "Investigate web-server-prod-01"

Agent Answer:
Here are the investigation results for web-server-prod-01:

* Resource Details:
  + IP Address: 10.0.1.101
  + Operating System: Ubuntu 22.04
  + Cloud Provider: AWS
  + Region: us-east-1
  + Instance Size: c5.2xlarge
  + Tags: Environment: production, Role: web, Team: platform, Tier: frontend
* Performance Metrics:
  + CPU Utilization: 97.3%
  + Memory Usage: 62.1%
  + Disk Space: 45% used
  + Network In/Out: 850.5 Mbps / 1200.3 Mbps
* Active Alerts:
  + ALR-20260219-001: Critical alert for high CPU utilization (95% exceeded)
* Related Incidents:
  + INC-20260219-001: Open incident for production web tier degradation
    due to high CPU on web-server-prod-01

Next Steps:
1. Investigate the cause of high CPU utilization and take corrective action.
2. Review the related incident (INC-20260219-001) and update the status.
3. Consider scaling up the instance size or adding more resources.
```

---

## 8. Flow 6: Environment Summary

### User Query Examples
- "Give me an environment summary"
- "What's the overall health of our infrastructure?"
- "Show me a dashboard"

### Flow Diagram

```
User: "Give me an environment summary"
  │
  ▼
Ollama decides: call get_environment_summary (no arguments)
  │
  ▼
executeGetSummary(client)
  → client.GetSummary()
    → counts alerts by state (critical, warning, info)
    → counts incidents by status (open, resolved)
    → counts resources by cloud provider
    → returns EnvironmentSummary struct
  │
  ▼
Ollama summarizes → final answer
```

### Tool Execution Code Path
```
tools.Execute("get_environment_summary", args)
  → executeGetSummary(client)
    → client.GetSummary()
      → loops through all alerts → counts Critical, Warning, Info
      → loops through all incidents → counts Open vs Resolved
      → loops through all resources → counts per Cloud provider
    → returns EnvironmentSummary{
        TotalResources, TotalAlerts, CriticalAlerts, WarningAlerts,
        InfoAlerts, OpenIncidents, ResolvedIncidents, CloudBreakdown
      }
    → marshals to JSON
```

### Captured Response
```
Query: "Give me an environment summary"

Agent Answer:
Based on the environment summary, here's a breakdown of your infrastructure:

* Total Resources: 22
* Total Alerts: 8
  + Critical Alerts: 4
  + Warning Alerts: 4
  + Info Alerts: 0
* Open Incidents: 5
* Resolved Incidents: 2

Cloud Breakdown:
* AWS: 10 resources
* Azure: 3 resources
* HPE GreenLake: 5 resources
* OnPrem: 4 resources
```

---

## 9. Flow 7: Predict Capacity (Single Resource)

### User Query Examples
- "Predict capacity for web-server-prod-01"
- "When will db-primary-01 run out of disk?"
- "Show CPU forecast for the web server"

### Flow Diagram

```
User: "Predict capacity for web-server-prod-01"
  │
  ▼
Ollama decides: call predict_capacity with
  {"resource_name": "web-server-prod-01", "threshold": "90"}
  │
  ▼
executePredictCapacity(client, args)
  → client.PredictCapacity("web-server-prod-01", "", 90)
    → GetResourceByName("web-server-prod-01")     ← resolves to resource
    → GetMetricHistoryForResource("res-001")       ← finds 3 series (cpu, memory, disk)
    → for each series:
        CapacityForecast(series, "web-server-prod-01", 90.0)
          → linearRegression(dataPoints)           ← least-squares on 30 daily values
          → returns slope, intercept, R²
          → classifies trend: Rising (slope > 0.5) / Stable / Declining
          → if currentValue >= threshold → "Already exceeded"
          → else: daysAhead = (threshold - intercept) / slope - (n-1)
          → predictDate(daysAhead)                 ← adds days to Feb 20, 2026
          → buildRecommendation()                  ← urgency-based advice
    → returns []ForecastResult (3 results: cpu, memory, disk)
  │
  ▼
Ollama summarizes per-metric forecasts → final answer
```

### Why the Math Runs in Go, Not the LLM

The LLM **does not** perform the prediction. It would be unreliable — LLMs hallucinate numbers, give inconsistent results across runs, and can't do precise arithmetic on 30 data points. Instead:

| Layer | Responsibility |
|-------|---------------|
| **LLM** | Understands user intent → decides to call `predict_capacity` |
| **Go code** (`CapacityForecast()`) | Runs linear regression → produces exact, deterministic numbers |
| **LLM again** | Reads the structured JSON result → writes a human-friendly summary |

The LLM is the **decision-maker** and **communicator**. The Go code is the **calculator**. This is the standard pattern across production AI agents (OpenAI function-calling, LangChain tools, etc.) — tools do the accurate computation, the LLM reasons about when to use them and how to present results.

### What Is Linear Regression?

Linear regression finds the **straight line that best fits** a set of data points. Given scattered points on a graph, it draws the one line that minimizes the total distance between itself and all the points.

```
                                              •  ← actual data point
100% |                                    •  /
     |                                 • / •
 90% |· · · · · · · · · · · · · · · ·/· threshold · ·
     |                            • /     
 80% |                         • /
     |                      • / •
 70% |                   • /
     |                • /
 60% |  •  •  •    • /         ← best-fit line: y = slope·x + intercept
     +----+----+----+----+----+----+
     0    5   10   15   20   25   29
                   day index
```

The line is defined by two numbers:
- **slope** — how steeply the line tilts (how much the metric changes per day)
- **intercept** — where the line crosses day 0 (the starting value)

Once you have the line equation `y = slope × x + intercept`, you can plug in any future `x` (day) to predict what `y` (metric value) will be.

### Step-by-Step: How `linearRegression()` Works

**Input:** 30 daily data points, e.g., web-server-prod-01 CPU: `[(0, 60.1), (1, 61.4), (2, 63.2), ..., (29, 97.3)]`

**Step 1 — Accumulate four sums** by looping through all points:

```go
for i, p := range points {
    x := float64(i)        // day index: 0, 1, 2, ..., 29
    sumX  += x              // Σx    = 0+1+2+...+29 = 435
    sumY  += p.Value        // Σy    = 60.1+61.4+...+97.3 ≈ 2370
    sumXY += x * p.Value    // Σ(xy) = 0×60.1 + 1×61.4 + ... ≈ 37500
    sumX2 += x * x          // Σ(x²) = 0+1+4+...+841 = 8555
}
```

These four sums capture everything needed to fit the line. `sumXY` is especially important — it measures whether large x-values (later days) tend to pair with large y-values (higher usage). If they do, the slope is positive.

**Step 2 — Compute slope and intercept** using the least-squares formulas:

```
slope     = (n·Σ(xy) − Σx·Σy) / (n·Σ(x²) − (Σx)²)
intercept = (Σy − slope·Σx) / n
```

For web-server-prod-01 CPU:
- `denom = 30 × 8555 − 435² = 256650 − 189225 = 67425`
- `slope = (30 × 37500 − 435 × 2370) / 67425 ≈ 1.28` → CPU grows ~1.28% per day
- `intercept = (2370 − 1.28 × 435) / 30 ≈ 60.4` → line starts at ~60.4% on day 0

So the best-fit line is: **y = 1.28x + 60.4**

**Step 3 — Compute R² (goodness of fit)** — how well the line explains the data:

```
R² = 1 − (Σ(y − ŷ)² / Σ(y − ȳ)²)

where:
  ŷ = predicted value on the line (slope × x + intercept)
  ȳ = average of all y values
```

- **Numerator** (ssRes): sum of squared distances from each point to the line — small means good fit
- **Denominator** (ssTot): sum of squared distances from each point to the flat average — total spread

| R² Value | Meaning |
|----------|---------|
| 1.0 | Perfect — all points lie exactly on the line |
| 0.9+ | Strong linear trend (our CPU example: 0.986) |
| 0.5 | Weak — line explains half the variation |
| 0.0 | No linear relationship at all |

### Step-by-Step: How `CapacityForecast()` Uses the Regression

After `linearRegression()` returns slope, intercept, and R², `CapacityForecast()` makes predictions:

**1. Classify the trend:**
```
slope > 0.5  → "Rising"      (metric growing more than 0.5%/day)
slope < -0.5 → "Declining"   (metric shrinking more than 0.5%/day)
otherwise    → "Stable"      (flat — noise within ±0.5%/day)
```

**2. Check if already exceeded:**
```
if currentValue >= threshold (90%) → "Already exceeded", no prediction needed
```
For web-server-prod-01 CPU: 97.3% ≥ 90% → already past the threshold.

**3. Check if declining:**
```
if slope ≤ 0 → "Not projected to breach" — metric is going down, will never hit 90%
```

**4. Predict the date (rising metrics):**

The line equation is `y = slope × x + intercept`. We want to find which `x` gives `y = threshold`:

```
x_threshold = (threshold − intercept) / slope
days_ahead  = x_threshold − 29                  ← 29 is the last day's index
```

**Concrete example — app-server-prod-02 memory** (slope ≈ 1.01, intercept ≈ 58, current ≈ 88%):

```
x_threshold = (90 − 58) / 1.01 ≈ 31.7
days_ahead  = 31.7 − 29 = 2.7 → ceil(2.7) = 3 days
predicted_date = Feb 20, 2026 + 3 days = Feb 23, 2026
```

**Visual for this example:**
```
 90% |· · · · · · · · · · · · · · · · ·✕ threshold
     |                              •  /
 88% |                          • (day 29 = today)
     |                      •  /       ↑
 85% |                   • /      3 days ahead = Feb 23
     |                • /
     |             • /
 58% |  •  •     /                     ← intercept (~58%)
     +----+----+----+----+
     0   10   20   29  31.7
```

### Recommendation Logic

Based on `daysToThreshold`, the tool generates urgency-tagged advice:

```
≤ 3 days   → "CRITICAL — immediate action required"
≤ 7 days   → "URGENT — action needed this week"
≤ 14 days  → "WARNING — plan capacity within 2 weeks"
> 14 days  → "Plan capacity expansion"
```

Plus metric-specific actions:
- **CPU** → "scale up, add replicas, investigate high-CPU processes"
- **Disk** → "expand volume, enable log rotation, archive old data"
- **Memory** → "increase instance memory, investigate leaks, tune settings"

### What the LLM Receives (Tool Response)

After all this math, the tool returns a JSON string like:

```json
{
  "currentValue": 97.3,
  "daysToThreshold": 0,
  "predictedDate": "Already exceeded",
  "trend": "Rising",
  "rSquared": 0.986,
  "dailyGrowthRate": 1.28,
  "recommendation": "web-server-prod-01 is already at 97.3% (threshold: 90%). Immediate action required..."
}
```

The LLM reads this pre-computed result and translates it into natural language. It never sees the raw 30 data points or runs any math — it simply communicates the answer.

### Captured Response
```
Query: "Predict capacity for web-server-prod-01"

Agent Answer:
Based on the tool results, here are the capacity predictions for web-server-prod-01:

* CPU Utilization: Already at 97.3% (threshold: 90%). Immediate action required
  to scale up, add capacity, or clean up resources.
* Memory Utilization: Will reach 90% in approximately 102 days with a daily
  growth rate of 0.26%. Consider increasing instance memory, investigating
  memory leaks, or tuning application memory settings.
* Disk Utilization: Will reach 90% in approximately 169 days with a daily
  growth rate of 0.26%. Plan capacity expansion by expanding disk volume,
  enabling log rotation, archiving old data, or cleaning up temp files.

Please take immediate action to address the high CPU utilization and plan for
future capacity expansion based on these predictions.
```

---

## 10. Flow 8: Predict Capacity (All Resources — At-Risk)

### User Query Examples
- "Which resources are at risk of running out of capacity?"
- "Can you predict which resource could be at risk looking at the historic data?"
- "Show capacity forecast for all monitored resources"

### Flow Diagram

```
User: "Which resources are at risk of running out of capacity?"
  │
  ▼
Ollama decides: call predict_capacity with
  {"resource_name": "", "threshold": "90"}    ← empty name = ALL resources
  │
  ▼
executePredictCapacity(client, args)
  → name is empty → client.PredictAllCapacity("", 90)
    → collects unique resource IDs from MetricHistory
    → for each resource (5 resources × 3 metrics = 15 forecasts):
        CapacityForecast(series, resourceName, 90.0)
    → sorts results: Rising trends first, then by DaysToThresh ascending
    → returns sorted []ForecastResult
  │
  ▼
Returns JSON: {"forecasts": [...], "count": 15, "scope": "all monitored resources"}
  │
  ▼
Ollama reads 15 forecasts, identifies the critical ones → final answer
```

### Key Design Decision
The `resource_name` parameter is **optional**. When omitted, the tool runs `PredictAllCapacity()` which:
1. Scans ALL resources with metric history (5 resources, 15 metric series)
2. Runs linear regression on each
3. Sorts by risk: `Rising` trend first, then `DaysToThresh` ascending (soonest breach first)
4. Returns everything in one JSON response

This was critical — before this, the LLM would try to call `predict_capacity` once per resource and run out of the 8-round limit.

### Mock Data Trends (for reference)
| Resource | Metric | 30-Day Trend | Current | Prediction |
|----------|--------|-------------|---------|------------|
| web-server-prod-01 | CPU | 60% → 97% (sharp rise) | 97.3% | Already exceeded |
| db-primary-01 | Disk | 32% → 92% (~2%/day) | 92% | Already exceeded |
| app-server-prod-02 | Memory | 58% → 88% (leak) | 88% | ~0 days (imminent) |
| db-primary-01 | CPU | ~50% → ~70% (moderate) | 68.2% | ~46 days |
| app-server-prod-02 | CPU | ~40% → ~55% (gentle) | 55% | ~62 days |
| k8s-node-04 | All | Flat/stable | ~40% | Not projected |
| azure-sql-prod-01 | CPU | 50% → 78% (growing) | 78% | ~10 days |

### Captured Response
```
Query: "Which resources are at risk of running out of capacity?"

Agent Answer:
Based on the capacity forecast tool output, there are several resources at risk
of running out of capacity:

* Immediate action required:
  + web-server-prod-01 is already at 97.3% CPU utilization (threshold: 90%).
    Scale up or add capacity.
  + db-primary-01 is already at 92.0% disk utilization (threshold: 90%).
    Scale up or add capacity.
* CRITICAL — immediate action required:
  + app-server-prod-02 will reach 90% memory utilization in ~0 days
    (1.11%/day growth). Consider increasing instance memory, investigating
    memory leaks, or tuning application memory settings.
* Plan capacity expansion:
  + db-primary-01 will reach 90% CPU utilization in ~46 days (0.43%/day growth).
    Scale up the instance or add horizontal replicas.
  + app-server-prod-02 will reach 90% CPU utilization in ~62 days (0.52%/day growth).
    Scale up the instance, add horizontal replicas, or investigate high-CPU processes.
  + web-server-prod-01 will reach 90% memory utilization in ~102 days (0.26%/day growth)
    and disk utilization in ~169 days (0.26%/day growth). Plan for capacity expansion.

Please take immediate action on the resources marked as "CRITICAL" or plan for
capacity expansion on the other resources to prevent potential issues.
```

---

## 11. Flow 9: Search Knowledge Base (RAG)

### What is This?

The knowledge base uses **Retrieval-Augmented Generation (RAG)** to let the agent answer questions about operations runbooks and troubleshooting procedures. Instead of relying on the LLM's training data (which may be outdated or generic), RAG fetches relevant content from actual PDF documents and injects it into the LLM's context.

### How the Knowledge Base Was Added

The RAG pipeline was imported from the `pdfReaderAIAgent` project and adapted as a new tool for the OpsRamp agent. Here's what was created:

1. **`knowledge/knowledge.go`** — A single-file package containing the entire RAG pipeline:
   - **Embedder** — Calls Ollama's `/api/embeddings` endpoint with the `nomic-embed-text` model to convert text into 768-dimensional vectors
   - **VectorStore** — In-memory database that stores chunk embeddings and performs brute-force cosine similarity search
   - **ChunkText()** — Splits extracted PDF text into 500-character overlapping chunks (100-char overlap) with sentence boundary detection
   - **ExtractTextFromPDF()** — Uses `github.com/ledongthuc/pdf` to extract plain text from PDF files
   - **KnowledgeBase** — Orchestrator that ties the pipeline together: `LoadPDF()` for indexing, `Search()` for querying

2. **`runbooks/opsramp_operations_runbook.pdf`** — A realistic operations runbook covering 9 sections:
   - High CPU Usage, Disk Space Full, Memory Leak, Container CrashLoopBackOff
   - Network Connectivity, Database Performance, SSL Certificate Expiry
   - Alert Response Matrix (with response times), Escalation Contacts, General Best Practices

3. **`search_knowledge_base` tool** — Added as the 8th tool in `tools/tools.go` with a `query` parameter. When the LLM calls this tool, it embeds the query, searches the vector store, and returns the top-3 most relevant text chunks with relevance scores.

4. **Wiring** — `main.go` loads all PDFs from `runbooks/` at startup, embeds them, and attaches the knowledge base to the agent via `SetKnowledgeBase()`. The `Agent` struct passes the KB to `tools.Execute()` on every tool call.

### The RAG Pipeline (Step by Step)

```
                    INDEXING (at startup)
                    ═══════════════════

  ┌──────────┐    ┌───────────┐    ┌───────────┐    ┌──────────────┐
  │ Load PDF │───>│ Extract   │───>│ Chunk     │───>│ Embed Each   │
  │ file     │    │ Text      │    │ Text      │    │ Chunk via    │
  │          │    │ (all pages│    │ (500 char, │    │ Ollama       │
  │          │    │  → string)│    │  100 over- │    │ nomic-embed  │
  │          │    │           │    │  lap)      │    │ -text        │
  └──────────┘    └───────────┘    └───────────┘    └──────┬───────┘
                                                          │
                                                          ▼
                                                   ┌──────────────┐
                                                   │ Vector Store  │
                                                   │ (in-memory)   │
                                                   │ chunk_0: [...]│
                                                   │ chunk_1: [...]│
                                                   │ chunk_N: [...]│
                                                   └──────────────┘

                    QUERYING (when tool is called)
                    ═════════════════════════════

  ┌───────────┐    ┌───────────┐    ┌──────────────┐    ┌───────────┐
  │ User asks │───>│ Embed the │───>│ Cosine       │───>│ Return    │
  │ "How to   │    │  query    │    │ Similarity   │    │ Top-3     │
  │  fix high │    │  (same    │    │ Search vs    │    │ Chunks    │
  │  CPU?"    │    │  model)   │    │ all stored   │    │ as JSON   │
  └───────────┘    └───────────┘    │ embeddings   │    └───────────┘
                                    └──────────────┘
```

**Key design decision:** The knowledge base is a *tool*, not a separate system. The LLM decides whether to call `search_knowledge_base` based on the user's question, just like it decides whether to call `search_alerts` or `predict_capacity`. This means:
- Runbook questions go through the same tool-calling pipeline as everything else
- The LLM sees the retrieved text chunks as tool results and summarizes them naturally
- No special prompting or separate RAG endpoint is needed

### User Query Examples
- "What is the runbook for high CPU usage?"
- "How do I troubleshoot disk space full?"
- "What are the escalation contacts?"
- "How to fix a container CrashLoopBackOff?"
- "What's the procedure for SSL certificate expiry?"

### Flow Diagram

```
User: "What is the runbook for high CPU usage?"
  │
  ▼
agent.Ask() → callLLM()
  │
  ▼
Ollama receives: system prompt + user message + 11 tool schemas
Ollama decides: call search_knowledge_base with {"query": "high CPU usage runbook"}
  │
  ▼
agent.executeTool() → tools.Execute()
  │
  ▼
executeSearchKnowledgeBase(kb, {"query": "high CPU usage runbook"})
  → kb.Search("high CPU usage runbook")
    → embedder.Embed("high CPU usage runbook")  // → 768-dim vector via Ollama
    → vectorStore.Search(queryEmbedding, topK=3)
      → computes cosine similarity against all stored chunk embeddings
      → sorts by score (highest first)
      → returns top 3 matches
  → marshals to JSON: {results: [...], count: 3, source: "OpsRamp Operations Runbook"}
  │
  ▼
Tool result appended to history as role="tool"
callLLM() called again with history including retrieved chunks
  │
  ▼
Ollama reads the runbook text chunks, generates a clear answer with steps
Returns role="assistant" with summarized runbook content
```

### What the Tool Returns

```json
{
  "results": [
    {
      "chunk_id": "chunk_3",
      "relevance_score": 0.8742,
      "content": "1.2 Immediate Triage Steps\nStep 1: Identify the process consuming the most CPU.\n• SSH into the affected server and run: top -bn1 | head -20\n• Alternatively, use htop for a more interactive view.\n• Check if the high CPU process is a known application process..."
    },
    {
      "chunk_id": "chunk_2",
      "relevance_score": 0.8531,
      "content": "1. High CPU Usage Runbook\n1.1 Overview\nHigh CPU utilization (above 85%) on any production server triggers a Warning alert, and above 95% triggers a Critical alert in OpsRamp. Sustained high CPU can lead to application slowdowns..."
    },
    {
      "chunk_id": "chunk_4",
      "relevance_score": 0.8215,
      "content": "1.3 Resolution Steps\n• If a runaway process is identified, consider restarting the service: systemctl restart <service-name>\n• If a Java application, check for GC pressure: jstat -gc <pid>\n• If a web server (Nginx/Apache)..."
    }
  ],
  "count": 3,
  "source": "OpsRamp Operations Runbook"
}
```

### What the LLM Does With It

The LLM receives these text chunks as a tool result and synthesizes them into a clear, actionable response — pulling out the specific steps, commands, and escalation procedures from the runbook content. It does NOT make up procedures — everything comes from the actual PDF.

### Prerequisites

The knowledge base requires the `nomic-embed-text` embedding model:
```bash
ollama pull nomic-embed-text
```

### Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `EMBEDDING_MODEL` | `nomic-embed-text` | Ollama model for generating embeddings |
| `RUNBOOK_DIR` | `runbooks` | Directory containing PDF runbook files |

---

## 12. Flow 10: Meta Questions (No Tool Call)

### User Query Examples
- "What can you do?"
- "Who are you?"
- "How can you help me?"

### Flow
The LLM answers directly from the system prompt without calling any tools. The system prompt contains a full IDENTITY section that lists all capabilities including the knowledge base.

---

## 13. Flow 11: MCP Server Mode

### What is MCP Mode?

When started with `--mcp` or `--mcp-http`, the agent runs as a **Model Context Protocol server** instead of a conversational chatbot. In this mode:

- **No local LLM is used** — the MCP client (Claude Desktop, VS Code Copilot, etc.) provides its own LLM
- **Tools are called directly** — no agent orchestration loop, no conversation history
- **Protocol-standard transport** — stdio (for local clients) or Streamable HTTP (for remote clients)

### Architecture

```
┌──────────────────────────┐
│   MCP Client             │
│   (Claude Desktop /      │
│    VS Code / Cursor)     │
│                          │
│   Client's LLM decides   │
│   which tool to call     │
└──────────┬───────────────┘
           │ MCP tools/call request (JSON-RPC over stdio or HTTP)
           ▼
┌──────────────────────────┐
│   mcpserver/server.go    │
│                          │
│   1. Receives CallTool   │
│   2. Converts MCP args   │
│      to internal ToolCall│
│   3. tools.Execute()     │
│   4. Returns JSON result │
│      as CallToolResult   │
└──────────┬───────────────┘
           │
           ▼
┌──────────────────────────┐
│   tools/tools.go         │
│   Execute() dispatcher   │
│   + OpsRamp mock client  │
│   + Knowledge base       │
└──────────────────────────┘
```

### How Tools Are Registered

At startup, `mcpserver.newMCPServer()` reads all 11 tool definitions from `tools.GetToolDefinitions()` and converts each Ollama-format schema to an MCP tool:

```go
for _, def := range tools.GetToolDefinitions() {
    mcpTool := convertTool(def)   // Ollama JSON Schema → mcp.NewTool()
    s.AddTool(mcpTool, makeHandler(client, kb, def.Function.Name))
}
```

The `convertTool()` function maps:
- `def.Function.Name` → `mcp.NewTool(name, ...)`
- `def.Function.Description` → `mcp.WithDescription()`
- `def.Function.Parameters.Properties` → `mcp.WithString(name, mcp.Description(), mcp.Enum(), mcp.Required())`

### Tool Call Flow (MCP)

```
MCP Client: tools/call {"name": "search_alerts", "arguments": {"state": "Critical"}}
  │
  ▼
mcpserver handler receives mcp.CallToolRequest
  → request.GetArguments() → map[string]interface{}{"state": "Critical"}
  → convert to map[string]string{"state": "Critical"}
  → build tools.ToolCall{Name: "search_alerts", Arguments: args}
  → tools.Execute(client, call, kb)
  → returns JSON string: {"results": [...], "count": 4}
  │
  ▼
mcpserver returns mcp.NewToolResultText(jsonString)
  → serialized as MCP CallToolResult to client
  │
  ▼
MCP Client's LLM reads JSON result and generates human-friendly answer
```

### Key Difference: Who Controls the LLM?

| Mode | LLM Location | Tool Orchestration | Transport |
|------|-------------|-------------------|----------|
| CLI / Web | Server-side (Ollama) | `agent.Agent` manages tool loop | stdin / HTTP REST |
| MCP | Client-side (Claude, GPT-4, etc.) | MCP client manages tool loop | stdio / Streamable HTTP (JSON-RPC) |

In MCP mode, the agent is a **pure tool server** — it exposes capabilities, the client's LLM decides what to call.

### Transports

- **Stdio** (`--mcp`): Agent reads JSON-RPC from stdin, writes to stdout. Used by Claude Desktop, VS Code Copilot.
- **Streamable HTTP** (`--mcp-http`): Agent listens on `:8081` (configurable via `--mcp-port`). Used by remote MCP clients.

### Client Configuration Examples

**Claude Desktop** (`~/Library/Application Support/Claude/claude_desktop_config.json`):
```json
{
  "mcpServers": {
    "opsramp": {
      "command": "/path/to/opsramp-agent",
      "args": ["--mcp"]
    }
  }
}
```

**VS Code** (`.vscode/mcp.json`):
```json
{
  "servers": {
    "opsramp": {
      "type": "stdio",
      "command": "/path/to/opsramp-agent",
      "args": ["--mcp"]
    }
  }
}
```

---

## 14. Flow 12: Correlate Network (Juniper)

### User Query Examples
- "Correlate network for k8s-node-04"
- "Is the network causing latency on web-server-prod-01?"
- "Check the switch port for k8s-node-04"

### Flow Diagram

```
User: "Correlate network for k8s-node-04"
  │
  ▼
agent.Ask() → callLLM()
  │
  ▼
Ollama receives: system prompt + user message + 11 tool schemas
Ollama decides: call correlate_network with {"resource_name": "k8s-node-04"}
  │
  ▼
agent.executeTool() → tools.ExecuteWithOptions()
  │
  ▼
executeCorrelateNetwork(juniperClient, {"resource_name": "k8s-node-04"})
  → juniperClient.CorrelateNetwork("k8s-node-04")
    → findMapping("k8s-node-04") → {SwitchID: "sw-dc-east-04", PortID: "ge-0/0/5"}
    → getSwitchByID("sw-dc-east-04") → sw-dc-east-04 (Juniper EX4300-48T)
    → getSwitchPort(sw, "ge-0/0/5") → port with 8.3% packet loss, 156789 rx_errors
    → analyzePortIssues(port, mapping) → 5 issues detected:
        • packet_loss: 8.3% (critical, threshold: 1.0%)
        • rx_errors: 156789 (critical)
        • link_flap: last flapped 12 minutes ago (warning)
        • speed_downgrade: 100Mbps (critical, expected: 1000Mbps)
        • duplex_mismatch: half-duplex (critical)
    → verdict: "NETWORK IS THE LIKELY ROOT CAUSE"
  → returns JSON with full port telemetry + issues + verdict
  │
  ▼
Tool result appended to history as role="tool"
callLLM() called again
  │
  ▼
Ollama reads JSON, generates human summary:
  "Network correlation reveals significant issues on sw-dc-east-04 port ge-0/0/5:
   - 8.3% packet loss (critical)
   - 156,789 RX errors
   - Link flapping (last flap 12 minutes ago)
   - Speed downgraded to 100Mbps (should be 1Gbps)
   - Half-duplex mismatch
   Network is the likely root cause of the performance issues."
```

### Key Data Flow
```
Resource name → findMapping() → {switch_id, port_id}
                    ↓
              getSwitchByID() → SwitchStats (from mockdata/network.go)
                    ↓
              getSwitchPort() → SwitchPort (port-level telemetry)
                    ↓
              analyzePortIssues() → []NetworkIssue
                    ↓
              buildRecommendation() → verdict + recommendation string
                    ↓
              NetworkCorrelation result → JSON for LLM
```

---

## 15. Flow 13: Blast Radius Analysis

### User Query Examples
- "What's the blast radius for k8s-node-04?"
- "How many users are affected by the k8s-node-04 issue?"
- "Show me the impact of the network problem on k8s-node-04"

### Flow Diagram

```
User: "What's the blast radius for k8s-node-04?"
  │
  ▼
agent.Ask() → callLLM()
  │
  ▼
Ollama receives: system prompt + user message + 11 tool schemas
Ollama decides: call blast_radius with {"resource_name": "k8s-node-04"}
  │
  ▼
agent.executeTool() → tools.ExecuteWithOptions()
  │
  ▼
executeBlastRadius(juniperClient, {"resource_name": "k8s-node-04"})
  → juniperClient.AnalyzeBlastRadius("k8s-node-04")
    → Find start node: k8s-node-04 (server, compute layer)
    → Find switch mapping: sw-dc-east-04 ge-0/0/5
    → BFS traversal of dependency graph:
        k8s-node-04 ──hosts──→ greenlake-portal (critical)
        k8s-node-04 ──hosts──→ aruba-central (high)
        k8s-node-04 ──hosts──→ dscc-console (high)
        dscc-console  ──depends_on── greenlake-portal (already visited)
        dscc-console  ──depends_on── aruba-central (already visited)
        greenlake-portal ──serves──→ greenlake-tenants (3000 users)
        greenlake-portal ──serves──→ aruba-wifi-users (2000 users)
        aruba-central ──serves──→ greenlake-tenants (already visited)
        dscc-console ──serves──→ greenlake-tenants (already visited)
    → Result:
        Affected applications: 3 (greenlake-portal, aruba-central, dscc-console)
        Affected users: 5000 (greenlake-tenants: 3000 + aruba-wifi-users: 2000)
        Severity: critical
        Business impact: "3 application(s) and approximately 5000 users affected.
          This is a HIGH-IMPACT incident affecting customer-facing services.
          Revenue-generating services are impacted — immediate escalation recommended."
        Critical path: k8s-node-04 → greenlake-portal → greenlake-tenants
  │
  ▼
Tool result appended to history as role="tool"
callLLM() called again
  │
  ▼
Ollama reads JSON, generates human summary:
  "Blast radius analysis for k8s-node-04 reveals a HIGH-IMPACT incident:
   - Root cause: Network issues on sw-dc-east-04 port ge-0/0/5
   - 3 applications affected: greenlake-portal, aruba-central, dscc-console
   - ~5,000 end users impacted (3,000 GreenLake tenants + 2,000 Aruba WiFi users)
   - Revenue-generating services are at risk
   - Critical path: k8s-node-04 → greenlake-portal → greenlake-tenants
   Immediate escalation recommended."
```

### Dependency Graph Traversal
```
             sw-dc-east-04 (ge-0/0/5)
                     │ connects_to
                     ▼
              ┌─ k8s-node-04 ─┐
              │    (server)    │
         hosts│               │hosts
              ▼               ▼
   ┌─────────────┐   ┌──────────────┐   ┌──────────┐
   │ greenlake-  │   │ aruba-       │   │ dscc-    │
   │ portal      │   │ central      │   │ console  │
   │ (critical)  │   │ (high)       │   │ (high)   │
   └──────┬──────┘   └──────┬───────┘   └──────┬────┘
          │ serves           │ serves           │ serves
          ▼                  ▼                  ▼
   ┌──────────────┐  ┌──────────────┐
   │ greenlake-   │  │ aruba-wifi-  │
   │ tenants      │  │ users        │
   │ (3000 users) │  │ (2000 users) │
   └──────────────┘  └──────────────┘
         Total affected: ~5,000 users
```

---

## 16. Flow 14: Guided Remediation

### User Query Examples
- "Give me a remediation plan for k8s-node-04"
- "How do I fix the network issue on k8s-node-04?"
- "What steps should I take to resolve the packet loss?"

### Flow Diagram

```
User: "Give me a remediation plan for k8s-node-04"
  │
  ▼
agent.Ask() → callLLM()
  │
  ▼
Ollama receives: system prompt + user message + 11 tool schemas
Ollama decides: call get_remediation_plan with {"resource_name": "k8s-node-04"}
  │
  ▼
agent.executeTool() → tools.ExecuteWithOptions()
  │
  ▼
executeGetRemediationPlan(juniperClient, {"resource_name": "k8s-node-04"})
  → juniperClient.GetRemediationPlan("k8s-node-04")
    → CorrelateNetwork("k8s-node-04") → 5 issues (packet_loss, rx_errors, link_flap, speed_downgrade, duplex_mismatch)
    → findMapping("k8s-node-04") → {SwitchID: "sw-dc-east-04", PortID: "ge-0/0/5"}
    → Generate remediation steps:
        Step 1: [diagnostic] show interfaces ge-0/0/5 extensive
        Step 2: [diagnostic] show interfaces diagnostics optics ge-0/0/5
        Step 3: [mitigation] Bounce interface (disable/enable) ⚠️ REQUIRES APPROVAL
        Step 4: [diagnostic] Clear error counters
        Step 5: [resolution] Reseat physical cable ⚠️ REQUIRES APPROVAL
        Step 6: [resolution] Force 1Gbps full-duplex ⚠️ REQUIRES APPROVAL
        Step 7: [verification] Check port status
        Step 8: [verification] Check application health
    → PlanID: REM-sw-dc-east-04-ge-0/0/5
    → Urgency: immediate
    → Risk: medium
    → Rollback available: yes
  │
  ▼
Tool result appended to history as role="tool"
callLLM() called again
  │
  ▼
Ollama presents step-by-step plan to user with approval gates highlighted
```

### Remediation Step Categories

| Category | Risk | Description |
|----------|------|-------------|
| `diagnostic` | none | Read-only commands to gather information |
| `mitigation` | medium | Quick fixes with potential brief outage (interface bounce) |
| `resolution` | low-medium | Permanent fixes (speed/duplex config, cable reseat) |
| `verification` | none | Post-remediation checks to confirm fix |

### Example Remediation Output (k8s-node-04)

```
Plan: REM-sw-dc-east-04-ge-0/0/5
Title: Remediate network issues on sw-dc-east-04 port ge-0/0/5 (connected to k8s-node-04)
Urgency: IMMEDIATE | Risk: MEDIUM | Approval Required: YES

Step 1 [diagnostic] — 10s
  Action: Run diagnostics on the affected switch port
  Command: show interfaces ge-0/0/5 extensive
  Target: sw-dc-east-04

Step 2 [diagnostic] — 10s
  Action: Check interface error counters and optics
  Command: show interfaces diagnostics optics ge-0/0/5
  Target: sw-dc-east-04

Step 3 [mitigation] — 30s ⚠️ APPROVAL REQUIRED
  Action: Bounce the interface to clear link flap state
  Command: set interfaces ge-0/0/5 disable → commit → delete → commit
  Target: sw-dc-east-04
  Risk: MEDIUM

Step 4 [diagnostic] — 5s
  Action: Clear interface error counters
  Command: clear interfaces statistics ge-0/0/5
  Target: sw-dc-east-04

Step 5 [resolution] — 2min ⚠️ APPROVAL REQUIRED
  Action: Check physical cable connectivity
  Target: sw-dc-east-04 ↔ k8s-node-04

Step 6 [resolution] — 30s ⚠️ APPROVAL REQUIRED
  Action: Force port speed and duplex to 1000Mbps full-duplex
  Command: set interfaces ge-0/0/5 speed 1g + link-mode full-duplex → commit
  Target: sw-dc-east-04

Step 7 [verification] — 10s
  Action: Verify remediation
  Command: show interfaces ge-0/0/5 extensive | match "errors|loss|flap|speed|duplex"
  Target: sw-dc-east-04

Step 8 [verification] — 10s
  Action: Verify application health on k8s-node-04
  Command: curl -s https://10.128.0.14:8443/health
  Target: k8s-node-04

Rollback: If remediation fails, restore port ge-0/0/5 config from backup.
```

---

## 17. Flow 15: End-to-End — "Why is the GreenLake portal slow?"

### The Showcase Scenario

This is the flagship demo flow showing autonomous multi-tool reasoning. The agent chains **6 tool calls** across different domains to go from a vague user question to an actionable remediation plan.

### User Query
"Why is the GreenLake portal slow?"

### Flow Diagram (6 Rounds of Autonomous Reasoning)

```
User: "Why is the GreenLake portal slow?"
  │
  ▼
═══════════════════════════════════════════════════════════════
ROUND 1: Search for alerts related to the GreenLake portal
═══════════════════════════════════════════════════════════════
  │
  ▼
LLM decides: call search_alerts(query="greenlake")
  → Finds: ALR-20260219-009 — "HTTP response latency exceeding thresholds"
     - Resource: k8s-node-04, P1 priority, Critical state
     - Metric: http_response_time_ms Avg:2847ms P99:4200ms
  │
  ▼
═══════════════════════════════════════════════════════════════
ROUND 2: Investigate the underlying server
═══════════════════════════════════════════════════════════════
  │
  ▼
LLM decides: call investigate_resource(resource_name="k8s-node-04")
  → Server metrics look NORMAL:
     - CPU: 72.5% (elevated but not critical)
     - Memory: 81.2%
     - Disk: 55.8%
     - Network: 0.8 Mbps in / 1.2 Mbps out
  → 1 active alert (the HTTP latency one)
  → LLM thinks: "Metrics are normal but app is slow... could be network?"
  │
  ▼
═══════════════════════════════════════════════════════════════
ROUND 3: Correlate with network telemetry (Juniper switch)
═══════════════════════════════════════════════════════════════
  │
  ▼
LLM decides: call correlate_network(resource_name="k8s-node-04")
  → Switch: sw-dc-east-04, Port: ge-0/0/5
  → NETWORK IS THE LIKELY ROOT CAUSE:
     - 8.3% packet loss (critical)
     - 156,789 RX errors
     - Link flapping (12 minutes ago)
     - Speed: 100Mbps (should be 1Gbps)
     - Half-duplex mismatch
  │
  ▼
═══════════════════════════════════════════════════════════════
ROUND 4: Map the blast radius — who's affected?
═══════════════════════════════════════════════════════════════
  │
  ▼
LLM decides: call blast_radius(resource_name="k8s-node-04")
  → 3 applications affected:
     - greenlake-portal (critical)
     - aruba-central (high)
     - dscc-console (high)
  → ~5,000 users affected:
     - greenlake-tenants: 3,000
     - aruba-wifi-users: 2,000
  → Business impact: HIGH — revenue-generating services impacted
  │
  ▼
═══════════════════════════════════════════════════════════════
ROUND 5: Search knowledge base for remediation procedures
═══════════════════════════════════════════════════════════════
  │
  ▼
LLM decides: call search_knowledge_base(query="network packet loss switch port remediation")
  → Finds runbook sections on:
     - Network troubleshooting procedures
     - Switch port diagnostics
     - Escalation procedures
  │
  ▼
═══════════════════════════════════════════════════════════════
ROUND 6: Generate guided remediation plan
═══════════════════════════════════════════════════════════════
  │
  ▼
LLM decides: call get_remediation_plan(resource_name="k8s-node-04")
  → 8-step remediation plan generated:
     - Diagnostics → Interface bounce → Cable check → Force speed/duplex → Verify
     - Urgency: IMMEDIATE
     - 3 steps require operator approval
     - Rollback plan included
  │
  ▼
═══════════════════════════════════════════════════════════════
FINAL ANSWER: Agent synthesizes everything into a comprehensive response
═══════════════════════════════════════════════════════════════

Agent response (synthesized from 6 tool calls):

"The GreenLake portal slowness is caused by a NETWORK issue, not a server problem.

**Root Cause:**
The Juniper switch sw-dc-east-04, port ge-0/0/5 connected to k8s-node-04 has:
- 8.3% packet loss
- 156,789 RX errors
- Speed downgraded to 100Mbps (should be 1Gbps)
- Half-duplex mismatch + link flapping

**Blast Radius:**
- 3 applications affected: greenlake-portal, aruba-central, dscc-console
- ~5,000 end users impacted (3,000 GreenLake tenants + 2,000 Aruba WiFi users)
- Revenue-generating services at risk — immediate action needed

**Recommended Fix (Plan: REM-sw-dc-east-04-ge-0/0/5):**
1. Run diagnostics: show interfaces ge-0/0/5 extensive
2. Bounce the interface (⚠️ requires approval — 30sec outage)
3. Reseat cable between sw-dc-east-04 and k8s-node-04
4. Force 1Gbps full-duplex configuration
5. Verify port status and application health

Shall I proceed with step 1 (diagnostics)?"
```

### Why This Flow Matters

| Aspect | Value |
|--------|-------|
| **Rounds** | 6 autonomous tool calls with zero manual intervention |
| **Cross-domain** | Spans alerts → compute → network → dependencies → knowledge → remediation |
| **Root cause** | Correctly identifies network as root cause despite normal server metrics |
| **Business context** | Quantifies impact: 3 apps, 5000 users, revenue at risk |
| **Actionable** | Produces specific CLI commands with approval gates |
| **Rollback** | Includes fallback plan if remediation fails |

---

## 18. Ollama API Request/Response Format

### Request (POST /api/chat)

```json
{
  "model": "llama3.1",              // or "mistral" — set via LLM_MODEL env var
  "messages": [
    {"role": "system", "content": "You are an HPE Operations Assistant..."},
    {"role": "user", "content": "Show me all critical alerts"}
  ],
  "tools": [
    {
      "type": "function",
      "function": {
        "name": "search_alerts",
        "description": "Search and filter OpsRamp alerts...",
        "parameters": {
          "type": "object",
          "properties": {
            "state": {"type": "string", "enum": ["Critical", "Warning", "Ok", "Info"]},
            "priority": {"type": "string", "enum": ["P0", "P1", "P2", "P3", "P4", "P5"]},
            "resource_name": {"type": "string"},
            "query": {"type": "string"}
          }
        }
      }
    }
    // ... 10 more tool definitions (11 total)
  ],
  "stream": false
}
```

### Response — Tool Call
```json
{
  "model": "llama3.1",
  "created_at": "2026-02-20T18:02:08.245184Z",
  "message": {
    "role": "assistant",
    "content": "",
    "tool_calls": [{
      "id": "call_ocl8vw75",
      "function": {
        "index": 0,
        "name": "search_alerts",
        "arguments": {"state": "Critical"}
      }
    }]
  },
  "done": true,
  "done_reason": "stop",
  "total_duration": 8003994333,
  "load_duration": 2604105583,
  "prompt_eval_count": 1801,
  "prompt_eval_duration": 4956284375,
  "eval_count": 18,
  "eval_duration": 426537420
}
```

### Response — Final Answer (after tool result)
```json
{
  "model": "llama3.1",
  "message": {
    "role": "assistant",
    "content": "Based on the tool output, here are all critical alerts:\n\n• CPU utilization exceeded 95%..."
  },
  "done": true,
  "prompt_eval_count": 1801,
  "eval_count": 18
}
```

### Multi-turn Conversation History
After tool execution, the history sent to Ollama looks like:
```json
[
  {"role": "system", "content": "You are an OpsRamp..."},
  {"role": "user", "content": "Show me all critical alerts"},
  {"role": "assistant", "content": "", "tool_calls": [{"function": {"name": "search_alerts", "arguments": {"state": "Critical"}}}]},
  {"role": "tool", "content": "{\"results\":[{\"id\":\"ALR-001\",\"subject\":\"CPU high\",...}],\"count\":4}"},
  // → Ollama now generates the final human-readable answer
]
```

---

## 19. Testing Reference — All Captured Responses

All responses below were captured on **February 20, 2026** using **llama3.1** via Ollama.
These can be used as baseline expectations for regression testing.

> **Note:** The default model is now configurable via `LLM_MODEL` (default: `llama3.1`).
> The native `make setup` pulls `mistral`. Captured responses may differ with different models.

### API Endpoint
```
POST http://localhost:8080/api/chat
Content-Type: application/json
Body: {"message": "<question>"}

POST http://localhost:8080/api/clear   (resets conversation history)
```

### Test Case 1: Critical Alerts
```bash
curl -s http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"Show me all critical alerts"}'
```
**Tool called:** `search_alerts(state="Critical")`
**Expected:** 4 critical alerts (web-server-prod-01 CPU, db-primary-01 disk, monitoring-agent-staging-01 PING, greenlake-portal pod crash loop)

### Test Case 2: AWS Resources
```bash
curl -s http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"List all resources running in AWS"}'
```
**Tool called:** `search_resources(cloud="AWS")`
**Expected:** 10 AWS resources with CPU/memory/disk metrics

### Test Case 3: Resource Investigation
```bash
curl -s http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"Investigate web-server-prod-01"}'
```
**Tool called:** `investigate_resource(resource_name="web-server-prod-01")`
**Expected:** Resource details + 97.3% CPU + 1 active alert (ALR-20260219-001) + 1 incident (INC-20260219-001)

### Test Case 4: Open Incidents
```bash
curl -s http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"Show me all open incidents"}'
```
**Tool called:** `search_incidents(status="Open")`
**Expected:** 5 open incidents with SLA breach info

### Test Case 5: Environment Summary
```bash
curl -s http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"Give me an environment summary"}'
```
**Tool called:** `get_environment_summary()`
**Expected:** 22 resources, 8 alerts (4 critical, 4 warning), 5 open incidents, cloud breakdown (AWS:10, HPE GreenLake:5, Azure:3, OnPrem:4)

### Test Case 6: Single Resource Capacity Forecast
```bash
curl -s http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"Predict capacity for web-server-prod-01"}'
```
**Tool called:** `predict_capacity(resource_name="web-server-prod-01", threshold="90")`
**Expected:** 3 forecasts (CPU: already exceeded, Memory: ~102 days, Disk: ~169 days)

### Test Case 7: All Resources At-Risk Forecast
```bash
curl -s http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"Which resources are at risk of running out of capacity?"}'
```
**Tool called:** `predict_capacity(resource_name="", threshold="90")`
**Expected:** 15 forecasts across 5 resources, sorted by risk. web-server-prod-01 CPU and db-primary-01 disk already exceeded. app-server-prod-02 memory imminent.

### Test Case 8: Resource Details
```bash
curl -s http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"Show me details of db-primary-01"}'
```
**Tool called:** `get_resource_details(resource_name="db-primary-01")`
**Expected:** Full resource details including IP (10.0.2.10), OS (Ubuntu 22.04), instance size (r6g.4xlarge), all metrics

### Test Case 9: Meta Question
```bash
curl -s http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"What can you do?"}'
```
**Tool called:** None expected (may still call one — LLM limitation)
**Expected:** List of capabilities (alerts, resources, incidents, investigation, summary, forecasting, knowledge base)

---

### Test Case 10: Knowledge Base — Runbook Query
```bash
curl -s http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"What is the runbook for high CPU usage?"}'
```
**Tool called:** `search_knowledge_base` with `{"query": "high CPU usage runbook"}`
**Expected:** Steps from the runbook including SSH, `top`, `systemctl restart`, and escalation procedures

---

### Test Case 11: Knowledge Base — Procedure Lookup
```bash
curl -s http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"How do I troubleshoot disk space full?"}'
```
**Tool called:** `search_knowledge_base` with `{"query": "disk space full troubleshoot"}`
**Expected:** Steps including `df -h`, `du -sh`, `find` for large files, log rotation

---

### Test Case 12: Network Correlation
```bash
curl -s http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"Correlate network for k8s-node-04"}'
```
**Tool called:** `correlate_network(resource_name="k8s-node-04")`
**Expected:** sw-dc-east-04, ge-0/0/5, 8.3% packet loss, 156789 RX errors, "NETWORK IS THE LIKELY ROOT CAUSE"

---

### Test Case 13: Blast Radius
```bash
curl -s http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"What is the blast radius for k8s-node-04?"}'
```
**Tool called:** `blast_radius(resource_name="k8s-node-04")`
**Expected:** 3 affected applications (greenlake-portal, aruba-central, dscc-console), ~5000 users, critical severity

---

### Test Case 14: Guided Remediation
```bash
curl -s http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"Give me a remediation plan for k8s-node-04"}'
```
**Tool called:** `get_remediation_plan(resource_name="k8s-node-04")`
**Expected:** Multiple steps including interface bounce, cable reseat, force 1Gbps full-duplex, verification

---

### Test Case 15: End-to-End Multi-Tool
```bash
curl -s http://localhost:8080/api/chat \
  -H 'Content-Type: application/json' \
  -d '{"message":"Why is the GreenLake portal slow?"}'
```
**Tool called:** Multiple — search_alerts → investigate_resource → correlate_network → blast_radius → search_knowledge_base → get_remediation_plan (up to 6 rounds)
**Expected:** Identifies network root cause on sw-dc-east-04, 3 apps + 5000 users affected, remediation plan with Junos commands

---

### Automated Test Script

```bash
#!/bin/bash
# Run: chmod +x test_agent.sh && ./test_agent.sh
# Requires: agent running with --web on port 8080

BASE="http://localhost:8080"
PASS=0
FAIL=0

test_query() {
    local name="$1"
    local query="$2"
    local expected="$3"

    # Clear history
    curl -s -X POST "$BASE/api/clear" > /dev/null

    # Send query
    response=$(curl -s --max-time 120 "$BASE/api/chat" \
        -H 'Content-Type: application/json' \
        -d "{\"message\":\"$query\"}")

    answer=$(echo "$response" | python3 -c "import sys,json; print(json.load(sys.stdin).get('answer',''))" 2>/dev/null)

    if echo "$answer" | grep -qi "$expected"; then
        echo "✅ PASS: $name"
        ((PASS++))
    else
        echo "❌ FAIL: $name (expected to contain: '$expected')"
        echo "   Got: ${answer:0:200}..."
        ((FAIL++))
    fi
}

echo "🧪 OpsRamp Agent Test Suite"
echo "=========================="
echo ""

test_query "Critical Alerts" "Show me all critical alerts" "web-server-prod-01"
test_query "AWS Resources" "List all resources running in AWS" "db-primary-01"
test_query "Investigate Resource" "Investigate web-server-prod-01" "97.3"
test_query "Open Incidents" "Show me all open incidents" "INC-20260219-001"
test_query "Environment Summary" "Give me an environment summary" "22"
test_query "Capacity Forecast" "Predict capacity for web-server-prod-01" "Already"
test_query "At-Risk Resources" "Which resources are at risk?" "web-server-prod-01"
test_query "Resource Details" "Show me details of db-primary-01" "10.0.2.10"
test_query "Network Correlation" "Correlate network for k8s-node-04" "packet_loss"
test_query "Blast Radius" "What is the blast radius for k8s-node-04?" "greenlake-portal"
test_query "Remediation Plan" "Give me a remediation plan for k8s-node-04" "ge-0/0/5"

echo ""
echo "Results: $PASS passed, $FAIL failed out of $((PASS+FAIL)) tests"
```

# OpsRamp ChatBot

An AI-powered OpsRamp operations assistant that lets you query your infrastructure using natural language. Ask about alerts, resources, incidents, capacity forecasts, and operations runbooks — the agent figures out what to look up and responds with clear, actionable answers.

## Web UI

![OpsRamp Agent Web UI](screenshots/web-ui.png)

## What You Can Ask

- **Alerts** — "Show me all critical alerts" · "Any P0 alerts?"
- **Resources** — "List all AWS resources" · "Show servers in GCP"
- **Incidents** — "Show open incidents" · "Any urgent tickets?"
- **Investigation** — "Investigate web-server-prod-01" · "Why is the DB slow?"
- **Environment** — "Give me an environment summary"
- **Capacity Forecast** — "Predict capacity for db-primary-01" · "Which resources are at risk?"
- **Knowledge Base** — "What is the runbook for high CPU?" · "How to fix disk full?" · "Escalation contacts?"

## Capabilities
- **Search Alerts** - filter by state (Critical/Warning), priority, resource
- **Search Resources** - find servers across AWS, Azure, GCP, on-prem
- **Resource Details** — deep-dive into configuration, metrics, tags
- **Search Incidents** - filter tickets by status, priority, SLA
- **Investigate Resource** -correlated view of alerts + incidents + metrics for a resource
- **Environment Summary** -high-level infrastructure health dashboard
- **Capacity Forecasting** -linear regression on 30-day metric history to predict CPU/memory/disk exhaustion
- **Knowledge Base (RAG)** - retrieval-augmented generation over operations runbooks (PDF), using vector embeddings and cosine similarity search
- **MCP Server Mode** - exposes all 8 tools as a Model Context Protocol server (stdio + HTTP transport) for Claude Desktop, VS Code Copilot, and other MCP-compatible clients

## RoadMap Ideas

- [x] **Phase 1**: Mock data + CLI + basic tool-calling
- [x] **Phase 2**: Web UI with browser-based chat (go:embed)
- [x] **Phase 3**: Capacity forecasting with linear regression
- [x] **Phase 4**: Knowledge base — RAG over PDF runbooks (current)
- [ ] **Phase 5**: Real OpsRamp API integration (OAuth2 + tenant config)
- [ ] **Phase 6**: Proactive insights + recommendations
- [ ] **Phase 7**: Actionable operations (acknowledge alerts, create incidents)
- [ ] **Phase 8**: Slack/Teams integration
- [x] **Phase 9**: MCP server mode — dual-mode binary serves as both standalone chatbot AND MCP tool server
  - Stdio transport for Claude Desktop, VS Code Copilot (--mcp flag)
  - Streamable HTTP transport for remote MCP clients (--mcp-http flag)
  - All 8 tools auto-converted from Ollama format to MCP format via mcp-go SDK
- [ ] **Phase 10**: Multi-MCP agent architecture — generic agent orchestrator discovers and composes MCP servers dynamically
  - MCP Gateway (auth, rate-limiting, observability across servers)
  - Off-the-shelf MCP servers for Jira, PagerDuty, Slack (replace custom integrations)
  - Agent mesh — specialist agents (Ops, Dev, Security) coordinating via shared MCP infrastructure

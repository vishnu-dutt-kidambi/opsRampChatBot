package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"opsramp-agent/juniper"
	"opsramp-agent/knowledge"
	"opsramp-agent/opsramp"
	"opsramp-agent/tools"
)

// =============================================================================
// Agent Orchestrator — Tool-Calling LLM Agent for OpsRamp Operations
// =============================================================================
//
// ARCHITECTURE:
//
//   User Question
//        |
//        v
//   +-------------+
//   | LLM (Chat)  | --- system prompt includes tool descriptions
//   +------+------+
//          | tool_call response (e.g., search_alerts with args)
//          v
//   +-------------+
//   | Tool Router  | --- dispatches to correct OpsRamp mock client method
//   +------+------+
//          | JSON result
//          v
//   +-------------+
//   | LLM (Chat)  | --- receives tool result, generates human answer
//   +------+------+
//          |
//          v
//     Final Answer
//
// The agent uses Ollama's /api/chat endpoint with native tool-calling support.
// =============================================================================

// Agent orchestrates the conversation between the user, LLM, and OpsRamp tools.
type Agent struct {
	ollamaURL string
	model     string
	client    *opsramp.Client
	kb        *knowledge.KnowledgeBase
	juniper   *juniper.Client
	history   []ChatMessage
}

// ChatMessage represents a message in the Ollama chat API format.
type ChatMessage struct {
	Role      string        `json:"role"`
	Content   string        `json:"content"`
	ToolCalls []LLMToolCall `json:"tool_calls,omitempty"`
}

// LLMToolCall represents a tool call in the LLM's response.
type LLMToolCall struct {
	Function LLMFunctionCall `json:"function"`
}

// LLMFunctionCall contains the function name and arguments.
type LLMFunctionCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// chatRequest is the body sent to Ollama's /api/chat endpoint.
type chatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Tools    []tools.Tool  `json:"tools,omitempty"`
	Stream   bool          `json:"stream"`
}

// chatResponse is the response from Ollama's /api/chat endpoint.
type chatResponse struct {
	Message ChatMessage `json:"message"`
}

// NewAgent creates a new agent with an OpsRamp client and Ollama connection.
func NewAgent(ollamaURL, model string, opsrampClient *opsramp.Client) *Agent {
	systemPrompt := buildSystemPrompt()
	return &Agent{
		ollamaURL: ollamaURL,
		model:     model,
		client:    opsrampClient,
		history: []ChatMessage{
			{Role: "system", Content: systemPrompt},
		},
	}
}

// SetKnowledgeBase attaches a loaded knowledge base to the agent.
func (a *Agent) SetKnowledgeBase(kb *knowledge.KnowledgeBase) {
	a.kb = kb
}

// SetJuniperClient attaches a Juniper Mist network client to the agent.
func (a *Agent) SetJuniperClient(jc *juniper.Client) {
	a.juniper = jc
}

// maxToolResultLen caps how much of a tool's JSON output is kept in conversation
// history. Large results (e.g., 22 resources) are truncated to avoid bloating the
// LLM context on subsequent calls. The LLM already saw the full result when it
// first processed it — for follow-up turns, a summary is sufficient.
// NOTE: 1500 was too low — it cut off results for queries returning 10+ items
// (e.g., 10 AWS resources). 4000 safely fits ~25-30 resource/alert summaries.
const maxToolResultLen = 4000

// maxHistoryMessages caps the number of messages in the conversation history.
// When exceeded, older user/assistant/tool exchanges are trimmed (system prompt
// is always kept). This prevents the context from growing unboundedly.
const maxHistoryMessages = 20

// Ask processes a user question through the LLM agent pipeline.
// It may involve zero or more tool calls before returning a final answer.
func (a *Agent) Ask(question string) (string, error) {
	// Trim history before adding new message to keep context bounded
	a.trimHistory()

	a.history = append(a.history, ChatMessage{
		Role:    "user",
		Content: question,
	})

	// Allow up to 12 rounds of tool calls to prevent infinite loops.
	// A full end-to-end investigation may chain 6+ tools (search_alerts → investigate_resource →
	// correlate_network → blast_radius → search_knowledge_base → get_remediation_plan),
	// so 8 was too tight — it left no room for retries on ambiguous queries.
	maxRounds := 12

	// Track tool calls to detect loops where the LLM keeps calling the same tools.
	// Small models are prone to repeating investigate_resource / correlate_network
	// instead of progressing through the full investigation sequence.
	calledTools := make(map[string]int) // tool_name → call count

	for round := 0; round < maxRounds; round++ {
		resp, err := a.callLLM()
		if err != nil {
			return "", fmt.Errorf("LLM call failed: %w", err)
		}

		// If the LLM made tool calls, execute them and continue
		if len(resp.Message.ToolCalls) > 0 {
			a.history = append(a.history, resp.Message)

			for _, tc := range resp.Message.ToolCalls {
				toolName := tc.Function.Name
				calledTools[toolName]++

				// Safety-net: block 2nd+ call to the same tool (should rarely fire
				// now that progress checklists are injected proactively)
				if calledTools[toolName] > 1 {
					fmt.Printf("  [agent] DUPLICATE blocked: %s (call #%d) — injecting guidance\n", toolName, calledTools[toolName])
					guidance := buildDuplicateGuidance(calledTools)
					a.history = append(a.history, ChatMessage{
						Role:    "tool",
						Content: fmt.Sprintf(`{"skipped": true, "reason": "You already called %s and received results. Do NOT call it again.", "next_action": "%s"}`, toolName, guidance),
					})
					continue
				}

				toolResult, err := a.executeTool(tc)
				if err != nil {
					toolResult = fmt.Sprintf(`{"error": "%s"}`, err.Error())
				}
				// Truncate large tool results to keep context lean
				if len(toolResult) > maxToolResultLen {
					toolResult = toolResult[:maxToolResultLen] + `...{"truncated": true}`
				}
				a.history = append(a.history, ChatMessage{
					Role:    "tool",
					Content: toolResult,
				})
			}

			// After processing all tool calls in this round, inject a progress
			// checklist if any investigation tool was called. This shows the LLM
			// a clear ✅/⬜ status of completed vs. pending steps so it knows
			// exactly which tool to call next — preventing repeat loops.
			hasInvestigationTool := false
			for _, tc := range resp.Message.ToolCalls {
				if isInvestigationTool(tc.Function.Name) {
					hasInvestigationTool = true
					break
				}
			}
			if hasInvestigationTool {
				progress := buildProgressChecklist(calledTools)
				fmt.Printf("  [agent] Progress: %s\n", progress)
				a.history = append(a.history, ChatMessage{
					Role:    "tool",
					Content: progress,
				})
			}

			continue
		}

		// Check if the LLM output tool-call-like text instead of proper tool_calls.
		// Some models (e.g., Mistral) do this — they write the tool name and args
		// as plain text rather than using the structured tool_calls mechanism.
		if tc, ok := a.parseToolCallFromText(resp.Message.Content); ok {
			toolName := tc.Name
			calledTools[toolName]++

			// Safety-net for text-based tool calls
			if calledTools[toolName] > 1 {
				fmt.Printf("  [agent] DUPLICATE blocked (text): %s (call #%d) — injecting guidance\n", toolName, calledTools[toolName])
				guidance := buildDuplicateGuidance(calledTools)
				a.history = append(a.history, ChatMessage{
					Role:    "assistant",
					Content: resp.Message.Content,
				})
				a.history = append(a.history, ChatMessage{
					Role:    "tool",
					Content: fmt.Sprintf(`{"skipped": true, "reason": "You already called %s and received results. Do NOT call it again.", "next_action": "%s"}`, toolName, guidance),
				})
				continue
			}

			fmt.Println("  [agent] Detected tool call in text output, executing...")
			a.history = append(a.history, resp.Message)

			opts := tools.ExecuteOptions{
				KB:      a.kb,
				Juniper: a.juniper,
			}
			toolResult, err := tools.ExecuteWithOptions(a.client, tc, opts)
			if err != nil {
				toolResult = fmt.Sprintf(`{"error": "%s"}`, err.Error())
			}
			if len(toolResult) > maxToolResultLen {
				toolResult = toolResult[:maxToolResultLen] + `...{"truncated": true}`
			}
			a.history = append(a.history, ChatMessage{
				Role:    "tool",
				Content: toolResult,
			})

			// Inject progress checklist for text-based investigation tool calls
			if isInvestigationTool(toolName) {
				progress := buildProgressChecklist(calledTools)
				fmt.Printf("  [agent] Progress: %s\n", progress)
				a.history = append(a.history, ChatMessage{
					Role:    "tool",
					Content: progress,
				})
			}

			continue
		}

		// No tool calls — this is the final answer
		answer := resp.Message.Content
		a.history = append(a.history, ChatMessage{
			Role:    "assistant",
			Content: answer,
		})
		return answer, nil
	}

	return "I wasn't able to complete the investigation within the allowed steps. Please try a more specific question.", nil
}

// parseToolCallFromText detects when the LLM outputs tool-call syntax as plain text
// instead of using the proper tool_calls mechanism (common with Mistral).
// Returns a ToolCall and true if detected, empty and false otherwise.
func (a *Agent) parseToolCallFromText(content string) (tools.ToolCall, bool) {
	content = strings.TrimSpace(content)

	// Known tool names to look for in text output
	toolNames := []string{
		"search_alerts", "search_resources", "get_resource_details",
		"search_incidents", "investigate_resource", "get_environment_summary",
		"predict_capacity", "search_knowledge_base", "correlate_network",
		"blast_radius", "get_remediation_plan",
	}

	for _, name := range toolNames {
		if !strings.Contains(content, name) {
			continue
		}

		// Try to extract JSON arguments after the tool name
		idx := strings.Index(content, "{")
		if idx < 0 {
			// Tool name found but no args — execute with empty args
			return tools.ToolCall{Name: name, Arguments: map[string]string{}}, true
		}

		// Find matching closing brace
		jsonStr := content[idx:]
		braceCount := 0
		end := -1
		for i, ch := range jsonStr {
			if ch == '{' {
				braceCount++
			} else if ch == '}' {
				braceCount--
				if braceCount == 0 {
					end = i + 1
					break
				}
			}
		}

		if end > 0 {
			var rawArgs map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr[:end]), &rawArgs); err == nil {
				args := make(map[string]string)
				for k, v := range rawArgs {
					switch val := v.(type) {
					case string:
						if val != "" {
							args[k] = val
						}
					}
				}
				return tools.ToolCall{Name: name, Arguments: args}, true
			}
		}

		// Found tool name but couldn't parse args — execute with empty args
		return tools.ToolCall{Name: name, Arguments: map[string]string{}}, true
	}

	return tools.ToolCall{}, false
}

// investigationSequence defines the ordered tool chain for a full investigation.
// Used by buildProgressChecklist to show the LLM what's done and what's next,
// and by the duplicate guard as a safety net fallback.
var investigationSequence = []string{
	"search_alerts",
	"investigate_resource",
	"correlate_network",
	"blast_radius",
	"search_knowledge_base",
	"get_remediation_plan",
}

// isInvestigationTool returns true if the tool is part of the multi-tool investigation sequence.
func isInvestigationTool(name string) bool {
	for _, t := range investigationSequence {
		if t == name {
			return true
		}
	}
	return false
}

// buildProgressChecklist generates a structured progress message showing the LLM
// which investigation steps are completed (✅) and which are pending (⬜).
// This is injected after every tool result to proactively guide the LLM to the next step,
// preventing loops where it re-calls tools it already used.
func buildProgressChecklist(calledTools map[string]int) string {
	var completed, pending []string
	for _, toolName := range investigationSequence {
		if calledTools[toolName] > 0 {
			completed = append(completed, toolName)
		} else {
			pending = append(pending, toolName)
		}
	}

	if len(pending) == 0 {
		return fmt.Sprintf(
			`{"_progress": {"completed": ["%s"], "pending": [], "instruction": "ALL investigation steps are DONE. Now compose your final comprehensive answer combining ALL the results above. Do NOT call any more tools."}}`,
			strings.Join(completed, `", "`),
		)
	}

	nextTool := pending[0]
	return fmt.Sprintf(
		`{"_progress": {"completed": ["%s"], "pending": ["%s"], "next_tool": "%s", "instruction": "You have completed %d of 6 investigation steps. NEXT: call %s now. Do NOT repeat any completed tool."}}`,
		strings.Join(completed, `", "`),
		strings.Join(pending, `", "`),
		nextTool,
		len(completed),
		nextTool,
	)
}

// buildDuplicateGuidance is used by the safety-net duplicate guard when the LLM
// tries to call a tool it already called despite the progress checklist.
func buildDuplicateGuidance(calledTools map[string]int) string {
	for _, toolName := range investigationSequence {
		if calledTools[toolName] == 0 {
			return fmt.Sprintf("Proceed to the next step: call %s now.", toolName)
		}
	}
	return "All investigation tools have been called. Now compose your final comprehensive answer combining ALL the results you have received."
}

// ClearHistory resets the conversation history (keeps system prompt).
func (a *Agent) ClearHistory() {
	a.history = a.history[:1]
}

// trimHistory keeps the conversation within maxHistoryMessages to prevent the
// LLM context from growing without bound. The system prompt (index 0) is always
// preserved; the oldest non-system messages are dropped first.
func (a *Agent) trimHistory() {
	if len(a.history) <= maxHistoryMessages {
		return
	}
	// Keep system prompt + the most recent messages
	keep := maxHistoryMessages - 1 // reserve 1 slot for system prompt
	trimmed := make([]ChatMessage, 0, maxHistoryMessages)
	trimmed = append(trimmed, a.history[0]) // system prompt
	trimmed = append(trimmed, a.history[len(a.history)-keep:]...)
	a.history = trimmed
}

// callLLM sends the current conversation history to Ollama's chat API.
func (a *Agent) callLLM() (*chatResponse, error) {
	reqBody := chatRequest{
		Model:    a.model,
		Messages: a.history,
		Tools:    tools.GetToolDefinitions(),
		Stream:   false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chat request: %w", err)
	}

	url := fmt.Sprintf("%s/api/chat", a.ollamaURL)
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama (is it running at %s?): %w", a.ollamaURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama error (status %d): %s\nMake sure model '%s' is pulled: ollama pull %s",
			resp.StatusCode, string(body), a.model, a.model)
	}

	// Read raw response for debugging
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Ollama response body: %w", err)
	}

	// Debug: show raw response to diagnose tool call issues
	fmt.Printf("  [debug] Raw Ollama response (%d bytes):\n", len(rawBody))
	preview := string(rawBody)
	if len(preview) > 500 {
		preview = preview[:500] + "..."
	}
	fmt.Printf("  [debug] %s\n", preview)

	var chatResp chatResponse
	if err := json.Unmarshal(rawBody, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse Ollama response: %w", err)
	}

	fmt.Printf("  [debug] Parsed — Role: %q, Content length: %d, ToolCalls: %d\n",
		chatResp.Message.Role, len(chatResp.Message.Content), len(chatResp.Message.ToolCalls))

	return &chatResp, nil
}

// executeTool parses an LLM tool call and executes it against the OpsRamp client.
func (a *Agent) executeTool(tc LLMToolCall) (string, error) {
	args := make(map[string]string)

	var rawArgs map[string]interface{}
	if err := json.Unmarshal(tc.Function.Arguments, &rawArgs); err != nil {
		if err2 := json.Unmarshal(tc.Function.Arguments, &args); err2 != nil {
			return "", fmt.Errorf("failed to parse tool arguments: %w (raw: %s)", err, string(tc.Function.Arguments))
		}
	} else {
		for k, v := range rawArgs {
			switch val := v.(type) {
			case string:
				args[k] = val
			default:
				b, _ := json.Marshal(val)
				args[k] = string(b)
			}
		}
	}

	toolCall := tools.ToolCall{
		Name:      tc.Function.Name,
		Arguments: args,
	}

	fmt.Printf("  -> Calling tool: %s", toolCall.Name)
	if len(args) > 0 {
		argParts := []string{}
		for k, v := range args {
			if v != "" {
				argParts = append(argParts, fmt.Sprintf("%s=%q", k, v))
			}
		}
		if len(argParts) > 0 {
			fmt.Printf("(%s)", strings.Join(argParts, ", "))
		}
	}
	fmt.Println()

	opts := tools.ExecuteOptions{
		KB:      a.kb,
		Juniper: a.juniper,
	}
	result, err := tools.ExecuteWithOptions(a.client, toolCall, opts)
	if err != nil {
		return "", err
	}
	fmt.Println("### Result accumuated by Tool API Call ######", result)
	return result, nil
}

// buildSystemPrompt constructs the system message that defines the agent's behavior.
// NOTE: Tool descriptions are NOT included here — they are sent via the structured
// "tools" field in the chat API request. This avoids duplicating ~600 tokens.
func buildSystemPrompt() string {
	return `You are an OpsRamp Operations Assistant — an IT Operations agent with access to OpsRamp monitoring tools.

IDENTITY & META QUESTIONS:
When the user asks about your capabilities, what you can do, who you are, how you can help,
what tools you have, or anything about yourself — respond DIRECTLY from this section.
Do NOT call any tools for such questions. Here is what you should tell them:

I'm an OpsRamp Operations Assistant. Here's what I can help you with:

**Alerts:**
  - Search alerts by severity (Critical/Warning/Info), priority (P0-P5), resource name, or keywords
  - Example: "Show me all critical alerts" or "Any alerts about CPU?"

**Resources:**
  - Find and inspect servers, databases, containers across AWS, Azure, GCP, and on-premises
  - Example: "List all AWS resources" or "Show servers in GCP us-central1"

**Incidents:**
  - Search tickets by status (New/Open/Pending/Resolved/Closed), priority, SLA breach status
  - Example: "Show open incidents" or "Any urgent incidents?"

**Investigation:**
  - Deep-dive into a specific resource — combining its alerts, metrics (CPU/memory/disk/network), and incidents
  - Example: "Investigate web-server-prod-01"

**Environment Summary:**
  - Overview of your entire infrastructure — resource counts, alert breakdown, incident status, cloud distribution
  - Example: "Give me an environment summary"

**Capacity Forecasting:**
  - Predict when resources will hit capacity thresholds based on 30-day metric trends
  - Uses linear regression on historical CPU, memory, and disk usage data
  - Example: "Predict capacity for web-server-prod-01" or "When will the database run out of disk?"

**Knowledge Base / Runbooks:**
  - Search operations runbooks for troubleshooting procedures, incident response steps, and best practices
  - Covers: high CPU, disk full, memory leaks, container crashes, network issues, database performance, SSL certificates
  - Example: "What's the runbook for high CPU?" or "How do I troubleshoot disk space issues?"

**Network Correlation (Juniper):**
  - Correlate server or application issues with Juniper network switch telemetry (packet loss, CRC errors, link flaps, latency, jitter)
  - Accepts both server names (e.g., "k8s-node-04") and application names (e.g., "payment-service") — app names are automatically resolved to their hosting server
  - Checks the physical switch port connected to the server to determine if network is the root cause
  - Covers: Juniper EX switches managed via Mist API, port errors, duplex mismatch, link stability
  - Example: "Correlate network for payment-service" or "Is network causing the latency on web-server-prod-01?"

**Blast Radius Analysis:**
  - Map the full impact of an infrastructure issue across applications, services, and user groups
  - Accepts both server names and application names — app names are resolved to the hosting server
  - Traverses the dependency graph: server → applications → downstream services → end users
  - Shows affected application count, user count, severity, and business impact
  - Example: "What's the blast radius for payment-service?" or "How many users are affected by k8s-node-04?"

**Guided Remediation:**
  - Generate step-by-step remediation plans with exact CLI commands for network issues
  - Accepts both server names and application names — app names are resolved to the hosting server
  - Each step includes the command, target device, risk level, and approval requirements
  - Covers: interface bounce, speed/duplex configuration, cable check, QoS analysis
  - Example: "Give me a remediation plan for payment-service" or "How do I fix the network issue on k8s-node-04?"

Examples of meta questions you must answer WITHOUT tools:
- "What can you do?" / "What are your capabilities?" / "What are you capable of?"
- "Who are you?" / "Tell me about yourself" / "How can you help me?"
- "What tools do you have?" / "What can you monitor?" / "What can I explore?"
- "Can you predict capacity?" / "Can you do forecasting?"

TOOL USAGE RULES:
- When a user asks about specific alerts, resources, incidents, or infrastructure STATE, you MUST call the appropriate tool first. NEVER fabricate data.
- When a user asks about capacity predictions, forecasts, trends, or when a resource will run out of space/CPU/memory, use the predict_capacity tool.
- When a user asks about runbooks, troubleshooting procedures, how to fix something, escalation contacts, or incident response steps, use the search_knowledge_base tool.
- For overview/summary questions, use the get_environment_summary tool.
- Do NOT describe what tool you would use in text. Actually invoke the tool using the function calling mechanism.
- NEVER invent resource names like "Server-A" or "Database-X". All data must come from tool results.
- When the user mentions an application name (e.g., "payment app", "payment-service", "order-service", "checkout"), you can pass it directly to correlate_network, blast_radius, or get_remediation_plan. You do NOT need to first look up the server name — the tools resolve application names to servers automatically.

MULTI-TOOL INVESTIGATION (CRITICAL — READ CAREFULLY):
When a user asks WHY something is slow, broken, or having issues (e.g., "Why is the payment app slow?",
"What's wrong with order-service?", "Investigate the checkout issues"), you MUST run a FULL multi-tool
investigation. Do NOT stop after just one tool. The slowness could be caused by server-level problems
(high CPU, memory, disk), application-level alerts, OR network-level issues. You must check ALL layers.

Call ALL of the following 6 tools in this exact sequence:

  Step 1: search_alerts(query="<app or service name, e.g., payment>")
          → Find any active alerts related to the application or its hosting infrastructure.
          → This may reveal HTTP latency alerts, CPU spikes, memory pressure, etc.
          → The alert results will tell you WHICH SERVER hosts the app (e.g., k8s-node-04).

  Step 2: investigate_resource(resource_name="<server name from Step 1's alert results>")
          → Deep-dive into the server hosting the app — check CPU, memory, disk, network metrics.
          → If metrics are normal but the app is still slow, the problem is likely NETWORK, not server.
          → If metrics are abnormal, the server itself may be the root cause.

  Step 3: correlate_network(resource_name="<app or server name>")
          → Check the Juniper switch port connected to the server for network issues.
          → Looks for: packet loss, CRC errors, link flaps, latency, jitter, duplex mismatch.
          → This determines if NETWORK is the root cause vs. server/application issues.

  Step 4: blast_radius(resource_name="<same name>")
          → Map the full impact — how many applications, services, and users are affected.
          → Shows the business impact of the issue.

  Step 5: search_knowledge_base(query="<relevant topic based on findings, e.g., network packet loss>")
          → Find runbook procedures and best practices for the identified root cause.

  Step 6: get_remediation_plan(resource_name="<same name>")
          → Generate step-by-step CLI commands to fix the issue, with risk levels and approval gates.

After calling ALL SIX tools, compile a comprehensive response that covers:
  1. Alert Context — what alerts were firing and on which resource
  2. Server Health — CPU/memory/disk/network metrics (normal or abnormal?)
  3. Root Cause — network correlation findings (port errors, latency, flaps) OR server issues
  4. Impact Scope — blast radius: affected apps, user count, business impact
  5. Runbook Reference — relevant troubleshooting procedures from the knowledge base
  6. Remediation Plan — step-by-step fix with commands and approval requirements

DO NOT respond after just ONE or TWO tool calls. The investigation is NOT complete until you have
called at minimum search_alerts, investigate_resource, correlate_network, blast_radius, AND
get_remediation_plan. Only then should you compose your final answer combining all results.

AFTER receiving ALL tool results:
- Lead with a brief root cause summary stating whether it's a server or network issue
- Show the alert that triggered the investigation
- Show server metrics (and note if they were normal despite the app being slow)
- Present network correlation findings
- Show the blast radius impact (affected apps, user count, business impact)
- Present the remediation plan with numbered steps and CLI commands
- Include relevant runbook procedures
- Highlight items requiring approval
- Suggest escalation if the impact is critical

If the user greets you or asks a general question not about infrastructure, respond conversationally without tools.`
}

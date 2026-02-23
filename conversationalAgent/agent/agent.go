package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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

	// Allow up to 8 rounds of tool calls to prevent infinite loops
	maxRounds := 8
	for round := 0; round < maxRounds; round++ {
		resp, err := a.callLLM()
		if err != nil {
			return "", fmt.Errorf("LLM call failed: %w", err)
		}

		// If the LLM made tool calls, execute them and continue
		if len(resp.Message.ToolCalls) > 0 {
			a.history = append(a.history, resp.Message)

			for _, tc := range resp.Message.ToolCalls {
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
			continue
		}

		// Check if the LLM output tool-call-like text instead of proper tool_calls.
		// Some models (e.g., Mistral) do this — they write the tool name and args
		// as plain text rather than using the structured tool_calls mechanism.
		if tc, ok := a.parseToolCallFromText(resp.Message.Content); ok {
			fmt.Println("  [agent] Detected tool call in text output, executing...")
			a.history = append(a.history, resp.Message)

			toolResult, err := tools.Execute(a.client, tc, a.kb)
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
		"predict_capacity", "search_knowledge_base",
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

	result, err := tools.Execute(a.client, toolCall, a.kb)
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

Examples of meta questions you must answer WITHOUT tools:
- "What can you do?" / "What are your capabilities?" / "What are you capable of?"
- "Who are you?" / "Tell me about yourself" / "How can you help me?"
- "What tools do you have?" / "What can you monitor?" / "What can I explore?"
- "Can you predict capacity?" / "Can you do forecasting?"

TOOL USAGE RULES:
- When a user asks about specific alerts, resources, incidents, or infrastructure STATE, you MUST call the appropriate tool first. NEVER fabricate data.
- When a user asks about capacity predictions, forecasts, trends, or when a resource will run out of space/CPU/memory, use the predict_capacity tool.
- When a user asks about runbooks, troubleshooting procedures, how to fix something, escalation contacts, or incident response steps, use the search_knowledge_base tool.
- Do NOT describe what tool you would use in text. Actually invoke the tool using the function calling mechanism.
- NEVER invent resource names like "Server-A" or "Database-X". All data must come from tool results.
- For overview/summary questions, use the get_environment_summary tool.

AFTER receiving tool results:
- Summarize the data clearly with bullet points
- Highlight critical items first
- Include specific resource names, IPs, and metric percentages from the tool output
- Suggest next steps when issues are found

If the user greets you or asks a general question not about infrastructure, respond conversationally without tools.`
}

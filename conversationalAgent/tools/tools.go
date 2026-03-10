package tools

import (
	"encoding/json"
	"fmt"
	"strings"

	"opsramp-agent/juniper"
	"opsramp-agent/knowledge"
	"opsramp-agent/opsramp"
)

// Tool represents an LLM-callable function with its Ollama-compatible schema.
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// ToolFunction describes the function name, description, and parameter schema.
type ToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  ToolParameters `json:"parameters"`
}

// ToolParameters is a JSON Schema-like definition for tool arguments.
type ToolParameters struct {
	Type       string                  `json:"type"`
	Properties map[string]ToolProperty `json:"properties"`
	Required   []string                `json:"required,omitempty"`
}

// ToolProperty describes a single parameter.
type ToolProperty struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}

// ToolCall represents a parsed tool call from the LLM response.
type ToolCall struct {
	Name      string            `json:"name"`
	Arguments map[string]string `json:"arguments"`
}

// GetToolDefinitions returns all available tools for the LLM system prompt.
func GetToolDefinitions() []Tool {
	return []Tool{
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "search_alerts",
				Description: "Search and filter OpsRamp alerts. Use this to find alerts by severity (Critical/Warning/Info), priority (P0/P1/P2/P3/P4/P5), resource name, or free-text search across alert subjects and descriptions.",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]ToolProperty{
						"state": {
							Type:        "string",
							Description: "Filter by alert state/severity",
							Enum:        []string{"Critical", "Warning", "Ok", "Info"},
						},
						"priority": {
							Type:        "string",
							Description: "Filter by alert priority",
							Enum:        []string{"P0", "P1", "P2", "P3", "P4", "P5"},
						},
						"resource_name": {
							Type:        "string",
							Description: "Filter by resource or device name (partial match supported)",
						},
						"query": {
							Type:        "string",
							Description: "Free-text search across alert subject, description, metric, and component",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "search_resources",
				Description: "Search and filter monitored infrastructure resources. Use this to find servers, databases, containers, cloud resources by cloud provider (AWS/Azure/GCP/OnPrem), region, resource type (Linux/Windows/Azure SQL Database), environment tag (production/staging), role tag (web/app/database/k8s-worker), or free-text search.",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]ToolProperty{
						"cloud": {
							Type:        "string",
							Description: "Filter by cloud provider",
							Enum:        []string{"AWS", "Azure", "GCP", "OnPrem"},
						},
						"region": {
							Type:        "string",
							Description: "Filter by region (e.g., us-east-1, eastus, us-central1, datacenter-east)",
						},
						"type": {
							Type:        "string",
							Description: "Filter by resource type (e.g., Linux, Windows, Azure SQL Database, VMware ESXi)",
						},
						"tag": {
							Type:        "string",
							Description: "Filter by tag value (e.g., production, staging, database, web, k8s-worker)",
						},
						"query": {
							Type:        "string",
							Description: "Free-text search across resource name, hostname, IP, OS, cloud, and tags",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_resource_details",
				Description: "Get detailed information about a specific resource including its configuration, current metrics (CPU, memory, disk, network), and tags. Use this when you need deep details about a particular server or resource.",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]ToolProperty{
						"resource_name": {
							Type:        "string",
							Description: "The name or hostname of the resource to look up",
						},
					},
					Required: []string{"resource_name"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "search_incidents",
				Description: "Search and filter OpsRamp incidents/tickets. Use this to find incidents by status (New/Open/Pending/Resolved/Closed), priority (Urgent/High/Normal/Low/Very Low), or free-text search across incident subjects, descriptions, categories, and assignees.",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]ToolProperty{
						"status": {
							Type:        "string",
							Description: "Filter by incident status",
							Enum:        []string{"New", "Open", "Pending", "Resolved", "Closed", "On Hold"},
						},
						"priority": {
							Type:        "string",
							Description: "Filter by incident priority",
							Enum:        []string{"Urgent", "High", "Normal", "Low", "Very Low"},
						},
						"query": {
							Type:        "string",
							Description: "Free-text search across incident subject, description, category, and assignee",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "investigate_resource",
				Description: "Run a comprehensive investigation on a resource. Returns resource details, all active alerts, related incidents, and current metrics in one call. Use this when a user asks to investigate, diagnose, or troubleshoot a specific resource.",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]ToolProperty{
						"resource_name": {
							Type:        "string",
							Description: "The name, hostname, or ID of the resource to investigate",
						},
					},
					Required: []string{"resource_name"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_environment_summary",
				Description: "Get a high-level summary of the entire monitored environment including total resource count, alert counts by severity, incident counts by status, and resource distribution across cloud providers. Use this when the user asks for an overview, dashboard, or status of the infrastructure.",
				Parameters: ToolParameters{
					Type:       "object",
					Properties: map[string]ToolProperty{},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "predict_capacity",
				Description: "Predict when resources will reach a capacity threshold based on historical metric trends. Uses 30-day linear regression to forecast CPU, memory, or disk exhaustion dates. If resource_name is provided, forecasts that specific resource. If resource_name is omitted or empty, forecasts ALL monitored resources and returns them sorted by risk (most critical first). Use this when the user asks about capacity planning, predictions, forecasts, trends, at-risk resources, when hardware will run out, whether to scale up, or when to add resources.",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]ToolProperty{
						"resource_name": {
							Type:        "string",
							Description: "The name, hostname, or ID of the resource to forecast. Leave empty to forecast ALL monitored resources.",
						},
						"metric": {
							Type:        "string",
							Description: "Specific metric to forecast (leave empty for all metrics)",
							Enum:        []string{"cpu", "memory", "disk"},
						},
						"threshold": {
							Type:        "string",
							Description: "Capacity threshold percentage to predict against (default: 90)",
						},
					},
					Required: []string{},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "search_knowledge_base",
				Description: "Search the operations knowledge base (runbooks, procedures, troubleshooting guides) for relevant information. Use this when the user asks about runbook procedures, troubleshooting steps, how to fix or resolve an issue, escalation procedures, best practices, or incident response procedures. This searches pre-loaded PDF documentation using semantic similarity.",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]ToolProperty{
						"query": {
							Type:        "string",
							Description: "The search query describing what procedure, runbook, or troubleshooting information you need",
						},
					},
					Required: []string{"query"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "correlate_network",
				Description: "Correlate a server or application issue with Juniper network switch telemetry. Given a server name, application name, or IP, this tool looks up the connected Juniper switch port and analyzes its telemetry (packet loss, CRC errors, link flaps, latency, jitter, duplex mismatch) to determine if the root cause is network-related. Accepts application names like 'payment-service' which are automatically resolved to the hosting server. Use this when investigating performance issues, latency problems, or connectivity issues. Returns switch port stats along with an analysis verdict.",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]ToolProperty{
						"resource_name": {
							Type:        "string",
							Description: "The name of the server, application, or IP address to correlate with network telemetry. Application names (e.g., 'payment-service') are automatically resolved to their hosting server.",
						},
					},
					Required: []string{"resource_name"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "blast_radius",
				Description: "Analyze the blast radius of an infrastructure issue. Given a server or application name, this tool traverses the dependency graph (server → applications → downstream services → user groups) to determine how many applications, services, and end-users are affected. Accepts application names like 'payment-service' which are automatically resolved to the hosting server. Use this after correlate_network confirms a network issue, or after investigate_resource reveals a problem, to understand the full scope and business impact of the incident.",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]ToolProperty{
						"resource_name": {
							Type:        "string",
							Description: "The name of the server, application, or IP address to analyze blast radius for. Application names (e.g., 'payment-service') are automatically resolved to their hosting server.",
						},
					},
					Required: []string{"resource_name"},
				},
			},
		},
		{
			Type: "function",
			Function: ToolFunction{
				Name:        "get_remediation_plan",
				Description: "Generate a step-by-step guided remediation plan for a resource with network issues. Given a server or application name, this tool correlates with the Juniper switch, identifies the specific issues, and produces an ordered list of remediation steps with exact CLI commands, risk levels, and approval requirements. Accepts application names like 'payment-service' which are automatically resolved to the hosting server. Use this when the root cause is confirmed and you need to propose a fix.",
				Parameters: ToolParameters{
					Type: "object",
					Properties: map[string]ToolProperty{
						"resource_name": {
							Type:        "string",
							Description: "The name of the server, application, or IP address to generate remediation steps for. Application names (e.g., 'payment-service') are automatically resolved to their hosting server.",
						},
					},
					Required: []string{"resource_name"},
				},
			},
		},
	}
}

// ExecuteOptions holds optional dependencies for tool execution.
type ExecuteOptions struct {
	KB      *knowledge.KnowledgeBase
	Juniper *juniper.Client
}

// Execute runs a tool call against the OpsRamp client and returns the JSON result.
// The optional kb parameter enables the search_knowledge_base tool.
func Execute(client *opsramp.Client, call ToolCall, kb ...*knowledge.KnowledgeBase) (string, error) {
	return ExecuteWithOptions(client, call, ExecuteOptions{
		KB: firstKB(kb),
	})
}

// ExecuteWithOptions runs a tool call with full options (including Juniper client).
func ExecuteWithOptions(client *opsramp.Client, call ToolCall, opts ExecuteOptions) (string, error) {
	switch call.Name {
	case "search_alerts":
		return executeSearchAlerts(client, call.Arguments)
	case "search_resources":
		return executeSearchResources(client, call.Arguments)
	case "get_resource_details":
		return executeGetResourceDetails(client, call.Arguments)
	case "search_incidents":
		return executeSearchIncidents(client, call.Arguments)
	case "investigate_resource":
		return executeInvestigateResource(client, call.Arguments)
	case "get_environment_summary":
		return executeGetSummary(client)
	case "predict_capacity":
		return executePredictCapacity(client, call.Arguments)
	case "search_knowledge_base":
		if opts.KB != nil {
			return executeSearchKnowledgeBase(opts.KB, call.Arguments)
		}
		return `{"error": "Knowledge base not loaded — no runbook documents available"}`, nil
	case "correlate_network":
		if opts.Juniper != nil {
			return executeCorrelateNetwork(opts.Juniper, call.Arguments)
		}
		return `{"error": "Juniper network client not configured — network correlation unavailable"}`, nil
	case "blast_radius":
		if opts.Juniper != nil {
			return executeBlastRadius(opts.Juniper, call.Arguments)
		}
		return `{"error": "Juniper network client not configured — blast radius analysis unavailable"}`, nil
	case "get_remediation_plan":
		if opts.Juniper != nil {
			return executeGetRemediationPlan(opts.Juniper, call.Arguments)
		}
		return `{"error": "Juniper network client not configured — remediation planning unavailable"}`, nil
	default:
		return "", fmt.Errorf("unknown tool: %s", call.Name)
	}
}

func firstKB(kbs []*knowledge.KnowledgeBase) *knowledge.KnowledgeBase {
	if len(kbs) > 0 {
		return kbs[0]
	}
	return nil
}

func executeSearchAlerts(client *opsramp.Client, args map[string]string) (string, error) {
	alerts := client.SearchAlerts(args["state"], args["priority"], args["resource_name"], args["query"])
	if len(alerts) == 0 {
		return `{"results": [], "count": 0, "message": "No alerts found matching the criteria"}`, nil
	}
	type alertSummary struct {
		ID       string `json:"id"`
		Subject  string `json:"subject"`
		State    string `json:"state"`
		Priority string `json:"priority"`
		Resource string `json:"resource"`
		Cloud    string `json:"cloud"`
		Elapsed  string `json:"elapsed"`
		Acked    bool   `json:"acknowledged"`
		Ticketed bool   `json:"ticketed"`
	}
	summaries := make([]alertSummary, len(alerts))
	for i, a := range alerts {
		summaries[i] = alertSummary{
			ID:       a.UniqueID,
			Subject:  a.Subject,
			State:    a.CurrentState,
			Priority: a.Priority,
			Resource: a.Resource.Name,
			Cloud:    a.Resource.Cloud,
			Elapsed:  a.ElapsedTime,
			Acked:    a.Acknowledged,
			Ticketed: a.Ticketed,
		}
	}
	result := map[string]interface{}{
		"results": summaries,
		"count":   len(summaries),
	}
	return toJSON(result)
}

func executeSearchResources(client *opsramp.Client, args map[string]string) (string, error) {
	resources := client.SearchResources(args["cloud"], args["region"], args["type"], args["state"], args["tag"], args["query"])
	if len(resources) == 0 {
		return `{"results": [], "count": 0, "message": "No resources found matching the criteria"}`, nil
	}
	type resourceSummary struct {
		ID           string  `json:"id"`
		Name         string  `json:"name"`
		Type         string  `json:"type"`
		Cloud        string  `json:"cloud"`
		Region       string  `json:"region"`
		InstanceSize string  `json:"instanceSize,omitempty"`
		State        string  `json:"state"`
		CPU          float64 `json:"cpuPercent"`
		Memory       float64 `json:"memoryPercent"`
		Disk         float64 `json:"diskPercent"`
	}
	summaries := make([]resourceSummary, len(resources))
	for i, r := range resources {
		summaries[i] = resourceSummary{
			ID:           r.ID,
			Name:         r.Name,
			Type:         r.Type,
			Cloud:        r.Cloud,
			Region:       r.Region,
			InstanceSize: r.InstanceSize,
			State:        r.State,
			CPU:          r.Metrics.CPUUtilization,
			Memory:       r.Metrics.MemoryUtilization,
			Disk:         r.Metrics.DiskUtilization,
		}
	}
	result := map[string]interface{}{
		"results": summaries,
		"count":   len(summaries),
	}
	return toJSON(result)
}

func executeGetResourceDetails(client *opsramp.Client, args map[string]string) (string, error) {
	name := args["resource_name"]
	if name == "" {
		return "", fmt.Errorf("resource_name is required")
	}
	r := client.GetResourceByName(name)
	if r == nil {
		return fmt.Sprintf(`{"error": "Resource '%s' not found"}`, name), nil
	}
	return toJSON(r)
}

func executeSearchIncidents(client *opsramp.Client, args map[string]string) (string, error) {
	incidents := client.SearchIncidents(args["status"], args["priority"], args["query"])
	if len(incidents) == 0 {
		return `{"results": [], "count": 0, "message": "No incidents found matching the criteria"}`, nil
	}
	type incidentSummary struct {
		ID            string `json:"id"`
		Subject       string `json:"subject"`
		Status        string `json:"status"`
		SubStatus     string `json:"subStatus"`
		Priority      string `json:"priority"`
		AssignedTo    string `json:"assignedTo"`
		Category      string `json:"category,omitempty"`
		SLAResBreach  bool   `json:"slaResolutionBreach"`
		SLARespBreach bool   `json:"slaResponseBreach"`
		CreatedDate   string `json:"createdDate"`
	}
	summaries := make([]incidentSummary, len(incidents))
	for i, inc := range incidents {
		s := incidentSummary{
			ID:          inc.ID,
			Subject:     inc.Subject,
			Status:      inc.Status,
			SubStatus:   inc.SubStatus,
			Priority:    inc.Priority,
			AssignedTo:  inc.AssignedTo.Name,
			CreatedDate: inc.CreatedDate,
		}
		if inc.Category != nil {
			s.Category = inc.Category.Name
		}
		if inc.SLADetails != nil {
			s.SLAResBreach = inc.SLADetails.ResolutionBreach
			s.SLARespBreach = inc.SLADetails.ResponseBreach
		}
		summaries[i] = s
	}
	result := map[string]interface{}{
		"results": summaries,
		"count":   len(summaries),
	}
	return toJSON(result)
}

func executeInvestigateResource(client *opsramp.Client, args map[string]string) (string, error) {
	name := args["resource_name"]
	if name == "" {
		return "", fmt.Errorf("resource_name is required")
	}
	report := client.InvestigateResource(name)
	if report == nil {
		return fmt.Sprintf(`{"error": "Resource '%s' not found"}`, name), nil
	}

	type alertInfo struct {
		ID       string `json:"id"`
		Subject  string `json:"subject"`
		State    string `json:"state"`
		Priority string `json:"priority"`
		Elapsed  string `json:"elapsed"`
	}
	type incidentInfo struct {
		ID            string `json:"id"`
		Subject       string `json:"subject"`
		Status        string `json:"status"`
		Priority      string `json:"priority"`
		AssignedTo    string `json:"assignedTo"`
		SLAResBreach  bool   `json:"slaResolutionBreach"`
		SLARespBreach bool   `json:"slaResponseBreach"`
	}
	type investigationResult struct {
		ResourceName string            `json:"resourceName"`
		IP           string            `json:"ip"`
		OS           string            `json:"os"`
		Cloud        string            `json:"cloud"`
		Region       string            `json:"region"`
		InstanceSize string            `json:"instanceSize"`
		Tags         map[string]string `json:"tags"`
		CPU          float64           `json:"cpuPercent"`
		Memory       float64           `json:"memoryPercent"`
		Disk         float64           `json:"diskPercent"`
		NetIn        float64           `json:"networkInMbps"`
		NetOut       float64           `json:"networkOutMbps"`
		ActiveAlerts []alertInfo       `json:"activeAlerts"`
		Incidents    []incidentInfo    `json:"relatedIncidents"`
		AlertCount   int               `json:"alertCount"`
		IncidentCnt  int               `json:"incidentCount"`
	}

	result := investigationResult{
		ResourceName: report.Resource.Name,
		IP:           report.Resource.IPAddress,
		OS:           report.Resource.OSName,
		Cloud:        report.Resource.Cloud,
		Region:       report.Resource.Region,
		InstanceSize: report.Resource.InstanceSize,
		Tags:         report.Resource.Tags,
		CPU:          report.Metrics.CPUUtilization,
		Memory:       report.Metrics.MemoryUtilization,
		Disk:         report.Metrics.DiskUtilization,
		NetIn:        report.Metrics.NetworkIn,
		NetOut:       report.Metrics.NetworkOut,
		AlertCount:   len(report.Alerts),
		IncidentCnt:  len(report.Incidents),
	}

	for _, a := range report.Alerts {
		result.ActiveAlerts = append(result.ActiveAlerts, alertInfo{
			ID: a.UniqueID, Subject: a.Subject, State: a.CurrentState,
			Priority: a.Priority, Elapsed: a.ElapsedTime,
		})
	}
	for _, inc := range report.Incidents {
		info := incidentInfo{
			ID: inc.ID, Subject: inc.Subject, Status: inc.Status,
			Priority: inc.Priority, AssignedTo: inc.AssignedTo.Name,
		}
		if inc.SLADetails != nil {
			info.SLAResBreach = inc.SLADetails.ResolutionBreach
			info.SLARespBreach = inc.SLADetails.ResponseBreach
		}
		result.Incidents = append(result.Incidents, info)
	}

	return toJSON(result)
}

func executeGetSummary(client *opsramp.Client) (string, error) {
	summary := client.GetSummary()
	return toJSON(summary)
}

func executePredictCapacity(client *opsramp.Client, args map[string]string) (string, error) {
	name := args["resource_name"]
	metric := args["metric"]
	var threshold float64
	if t := args["threshold"]; t != "" {
		fmt.Sscanf(t, "%f", &threshold)
	}

	// If no resource specified, forecast ALL monitored resources
	if name == "" {
		forecasts := client.PredictAllCapacity(metric, threshold)
		if len(forecasts) == 0 {
			return `{"results": [], "count": 0, "message": "No resources with historical metric data found"}`, nil
		}
		result := map[string]interface{}{
			"forecasts": forecasts,
			"count":     len(forecasts),
			"scope":     "all monitored resources",
		}
		return toJSON(result)
	}

	forecasts := client.PredictCapacity(name, metric, threshold)
	if forecasts == nil {
		return fmt.Sprintf(`{"error": "Resource '%s' not found or no historical metric data available"}`, name), nil
	}
	if len(forecasts) == 0 {
		return fmt.Sprintf(`{"error": "No metric history matching '%s' for resource '%s'"}`, metric, name), nil
	}

	result := map[string]interface{}{
		"forecasts": forecasts,
		"count":     len(forecasts),
	}
	return toJSON(result)
}

func executeSearchKnowledgeBase(kb *knowledge.KnowledgeBase, args map[string]string) (string, error) {
	query := args["query"]
	if query == "" {
		return "", fmt.Errorf("query is required")
	}

	results, err := kb.Search(query)
	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error()), nil
	}

	if len(results) == 0 {
		return `{"results": [], "count": 0, "message": "No relevant runbook content found for the query"}`, nil
	}

	type kbResult struct {
		ChunkID string  `json:"chunk_id"`
		Score   float64 `json:"relevance_score"`
		Content string  `json:"content"`
	}

	summaries := make([]kbResult, len(results))
	for i, r := range results {
		summaries[i] = kbResult{
			ChunkID: r.ID,
			Score:   r.Score,
			Content: r.Text,
		}
	}

	result := map[string]interface{}{
		"results": summaries,
		"count":   len(summaries),
		"source":  "OpsRamp Operations Runbook",
	}
	return toJSON(result)
}

func executeCorrelateNetwork(jClient *juniper.Client, args map[string]string) (string, error) {
	name := args["resource_name"]
	if name == "" {
		return "", fmt.Errorf("resource_name is required")
	}

	correlation := jClient.CorrelateNetwork(name)
	if correlation == nil {
		// No mapping found — resource may be cloud-only (no physical switch)
		return fmt.Sprintf(`{"error": "No network switch port mapping found for '%s'. This resource may be in a cloud environment without a managed Juniper switch, or the LLDP/DCIM mapping is not configured."}`, name), nil
	}

	// Build a structured summary for the LLM
	type issueSummary struct {
		Type        string `json:"type"`
		Severity    string `json:"severity"`
		Description string `json:"description"`
		Value       string `json:"value"`
		Threshold   string `json:"threshold"`
	}

	type correlationResult struct {
		ResourceName   string         `json:"resource_name"`
		ResourceIP     string         `json:"resource_ip"`
		SwitchName     string         `json:"switch_name"`
		SwitchModel    string         `json:"switch_model"`
		SwitchIP       string         `json:"switch_ip"`
		PortID         string         `json:"port_id"`
		PortStatus     string         `json:"port_status"`
		Speed          string         `json:"speed"`
		Duplex         string         `json:"duplex"`
		PacketLoss     string         `json:"packet_loss"`
		RxErrors       int            `json:"rx_errors"`
		TxErrors       int            `json:"tx_errors"`
		Latency        string         `json:"latency_ms"`
		Jitter         string         `json:"jitter_ms"`
		RxBps          string         `json:"rx_rate"`
		TxBps          string         `json:"tx_rate"`
		LastFlapped    string         `json:"last_flapped"`
		Issues         []issueSummary `json:"issues"`
		IssueCount     int            `json:"issue_count"`
		NetworkIsRoot  bool           `json:"network_is_likely_root_cause"`
		Verdict        string         `json:"verdict"`
		Recommendation string         `json:"recommendation,omitempty"`
	}

	duplex := "full-duplex"
	if !correlation.FullDuplex {
		duplex = "half-duplex"
	}

	issues := make([]issueSummary, len(correlation.Issues))
	for i, iss := range correlation.Issues {
		issues[i] = issueSummary{
			Type:        iss.Type,
			Severity:    iss.Severity,
			Description: iss.Description,
			Value:       iss.Value,
			Threshold:   iss.Threshold,
		}
	}

	result := correlationResult{
		ResourceName:   correlation.ResourceName,
		ResourceIP:     correlation.ResourceIP,
		SwitchName:     correlation.SwitchName,
		SwitchModel:    correlation.SwitchModel,
		SwitchIP:       correlation.SwitchIP,
		PortID:         correlation.PortID,
		PortStatus:     correlation.PortStatus,
		Speed:          fmt.Sprintf("%dMbps", correlation.Speed),
		Duplex:         duplex,
		PacketLoss:     fmt.Sprintf("%.1f%%", correlation.PacketLoss),
		RxErrors:       correlation.RxErrors,
		TxErrors:       correlation.TxErrors,
		Latency:        fmt.Sprintf("%.1f", correlation.Latency),
		Jitter:         fmt.Sprintf("%.1f", correlation.Jitter),
		RxBps:          formatBps(correlation.RxBps),
		TxBps:          formatBps(correlation.TxBps),
		LastFlapped:    correlation.LastFlapped,
		Issues:         issues,
		IssueCount:     correlation.IssueCount,
		NetworkIsRoot:  correlation.NetworkIsRoot,
		Verdict:        correlation.Verdict,
		Recommendation: correlation.Recommendation,
	}

	return toJSON(result)
}

func executeBlastRadius(jClient *juniper.Client, args map[string]string) (string, error) {
	name := args["resource_name"]
	if name == "" {
		return "", fmt.Errorf("resource_name is required")
	}

	result := jClient.AnalyzeBlastRadius(name)
	if result == nil {
		return fmt.Sprintf(`{"error": "No dependency graph data found for '%s'. This resource may not be in the infrastructure topology map."}`, name), nil
	}

	// Build a structured summary for the LLM
	type nodeSummary struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		Layer       string `json:"layer"`
		Criticality string `json:"criticality"`
		Impact      string `json:"impact"`
		Reason      string `json:"reason"`
	}

	type pathEntry struct {
		Name         string `json:"name"`
		Type         string `json:"type"`
		Relationship string `json:"relationship,omitempty"`
	}

	type blastResult struct {
		RootCause            string        `json:"root_cause"`
		RootCauseDesc        string        `json:"root_cause_description"`
		AffectedApplications int           `json:"affected_applications"`
		AffectedServers      int           `json:"affected_servers"`
		AffectedServices     int           `json:"affected_services"`
		AffectedUsers        int           `json:"affected_users"`
		TotalImpacted        int           `json:"total_impacted_nodes"`
		Severity             string        `json:"overall_severity"`
		BusinessImpact       string        `json:"business_impact"`
		ImpactedNodes        []nodeSummary `json:"impacted_nodes"`
		CriticalPath         []pathEntry   `json:"critical_path"`
	}

	nodes := make([]nodeSummary, len(result.ImpactedNodes))
	for i, n := range result.ImpactedNodes {
		nodes[i] = nodeSummary{
			Name:        n.Name,
			Type:        n.Type,
			Layer:       n.Layer,
			Criticality: n.Criticality,
			Impact:      n.Impact,
			Reason:      n.Reason,
		}
	}

	path := make([]pathEntry, len(result.CriticalPath))
	for i, p := range result.CriticalPath {
		path[i] = pathEntry{
			Name:         p.NodeName,
			Type:         p.NodeType,
			Relationship: p.Relationship,
		}
	}

	out := blastResult{
		RootCause:            result.RootCauseName,
		RootCauseDesc:        result.RootCauseDesc,
		AffectedApplications: result.AffectedApplications,
		AffectedServers:      result.AffectedServers,
		AffectedServices:     result.AffectedServices,
		AffectedUsers:        result.AffectedUsers,
		TotalImpacted:        result.TotalImpactedNodes,
		Severity:             result.OverallSeverity,
		BusinessImpact:       result.BusinessImpact,
		ImpactedNodes:        nodes,
		CriticalPath:         path,
	}

	return toJSON(out)
}

func executeGetRemediationPlan(jClient *juniper.Client, args map[string]string) (string, error) {
	name := args["resource_name"]
	if name == "" {
		return "", fmt.Errorf("resource_name is required")
	}

	plan := jClient.GetRemediationPlan(name)
	if plan == nil {
		return fmt.Sprintf(`{"error": "Cannot generate remediation plan for '%s'. No network issues detected or resource not mapped to a Juniper switch."}`, name), nil
	}

	// Build a structured summary for the LLM
	type stepSummary struct {
		Step             int    `json:"step"`
		Action           string `json:"action"`
		Command          string `json:"command,omitempty"`
		Target           string `json:"target"`
		Category         string `json:"category"`
		ExpectedOutcome  string `json:"expected_outcome"`
		RiskLevel        string `json:"risk_level"`
		RequiresApproval bool   `json:"requires_approval"`
		EstimatedTime    string `json:"estimated_time"`
	}

	type planResult struct {
		PlanID            string        `json:"plan_id"`
		Title             string        `json:"title"`
		Resource          string        `json:"resource"`
		Switch            string        `json:"switch"`
		Port              string        `json:"port"`
		RootCause         string        `json:"root_cause"`
		Urgency           string        `json:"urgency"`
		RiskLevel         string        `json:"risk_level"`
		EstimatedDowntime string        `json:"estimated_downtime"`
		RequiresApproval  bool          `json:"requires_approval"`
		ApprovalNote      string        `json:"approval_note,omitempty"`
		TotalSteps        int           `json:"total_steps"`
		Steps             []stepSummary `json:"steps"`
		RollbackAvailable bool          `json:"rollback_available"`
		RollbackPlan      string        `json:"rollback_plan"`
	}

	steps := make([]stepSummary, len(plan.Steps))
	for i, s := range plan.Steps {
		steps[i] = stepSummary{
			Step:             s.StepNumber,
			Action:           s.Action,
			Command:          s.Command,
			Target:           s.Target,
			Category:         s.Category,
			ExpectedOutcome:  s.ExpectedOutcome,
			RiskLevel:        s.RiskLevel,
			RequiresApproval: s.RequiresApproval,
			EstimatedTime:    s.EstimatedTime,
		}
	}

	out := planResult{
		PlanID:            plan.PlanID,
		Title:             plan.Title,
		Resource:          plan.ResourceName,
		Switch:            plan.SwitchName,
		Port:              plan.PortID,
		RootCause:         plan.RootCause,
		Urgency:           plan.Urgency,
		RiskLevel:         plan.RiskLevel,
		EstimatedDowntime: plan.EstimatedDowntime,
		RequiresApproval:  plan.RequiresApproval,
		ApprovalNote:      plan.ApprovalNote,
		TotalSteps:        plan.TotalSteps,
		Steps:             steps,
		RollbackAvailable: plan.RollbackAvailable,
		RollbackPlan:      plan.RollbackPlan,
	}

	return toJSON(out)
}

func formatBps(bps int64) string {
	switch {
	case bps >= 1000000000:
		return fmt.Sprintf("%.1f Gbps", float64(bps)/1000000000)
	case bps >= 1000000:
		return fmt.Sprintf("%.1f Mbps", float64(bps)/1000000)
	case bps >= 1000:
		return fmt.Sprintf("%.1f Kbps", float64(bps)/1000)
	default:
		return fmt.Sprintf("%d bps", bps)
	}
}

// FormatToolsForPrompt generates a human-readable tool description block for the system prompt.
func FormatToolsForPrompt() string {
	defs := GetToolDefinitions()
	var sb strings.Builder
	sb.WriteString("You have access to the following tools to query the OpsRamp monitoring platform:\n\n")
	for _, t := range defs {
		sb.WriteString(fmt.Sprintf("### %s\n", t.Function.Name))
		sb.WriteString(fmt.Sprintf("%s\n", t.Function.Description))
		if len(t.Function.Parameters.Properties) > 0 {
			sb.WriteString("Parameters:\n")
			for name, prop := range t.Function.Parameters.Properties {
				required := ""
				for _, r := range t.Function.Parameters.Required {
					if r == name {
						required = " (required)"
						break
					}
				}
				if len(prop.Enum) > 0 {
					sb.WriteString(fmt.Sprintf("  - %s (%s%s): %s. Values: %s\n", name, prop.Type, required, prop.Description, strings.Join(prop.Enum, ", ")))
				} else {
					sb.WriteString(fmt.Sprintf("  - %s (%s%s): %s\n", name, prop.Type, required, prop.Description))
				}
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func toJSON(v interface{}) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}
	return string(b), nil
}

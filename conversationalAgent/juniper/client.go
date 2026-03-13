package juniper

import (
	"fmt"
	"strings"
	"time"
)

// =============================================================================
// Juniper Mist Client — Network Telemetry & Correlation
// =============================================================================
//
// This client simulates the Juniper Mist API for network switch telemetry.
// In production, it would make real HTTP calls to:
//   https://api.mist.com/api/v1/sites/{site_id}/stats/devices?type=switch
//   https://api.mist.com/api/v1/sites/{site_id}/stats/switch_ports
//
// The key method is CorrelateNetwork, which takes a server resource name/IP
// and returns network telemetry from the connected switch port, along with
// an analysis of whether network issues are the likely root cause.
// =============================================================================

// Client provides access to Juniper Mist network switch data and dependency graph.
type Client struct {
	Switches []SwitchStats
	Mappings []PortMapping
	Nodes    []DependencyNode
	Edges    []DependencyEdge
}

// NewClient creates a Juniper Mist client pre-loaded with switch data and port mappings.
func NewClient(switches []SwitchStats, mappings []PortMapping) *Client {
	return &Client{
		Switches: switches,
		Mappings: mappings,
	}
}

// SetDependencyGraph loads the dependency graph for blast radius analysis.
func (c *Client) SetDependencyGraph(nodes []DependencyNode, edges []DependencyEdge) {
	c.Nodes = nodes
	c.Edges = edges
}

// CorrelateNetwork correlates a server resource with its network switch port telemetry.
// It finds the switch port connected to the given resource, analyzes the port stats,
// and determines if network issues are the likely root cause of server problems.
//
// Accepts both server names (e.g., "k8s-node-04") and application names
// (e.g., "greenlake-portal"). Application names are automatically resolved to
// their hosting server via the dependency graph.
//
// This is the primary tool for the "Correlate Network" feature in the agent pipeline:
//  1. Resolve input → server name (handles app names via dependency graph)
//  2. Resolve server → switch port mapping (via LLDP/DCIM/manual map)
//  3. Retrieve switch port telemetry (errors, loss, latency, flaps)
//  4. Analyze for issues against thresholds
//  5. Return verdict: is network the likely root cause?
func (c *Client) CorrelateNetwork(resourceNameOrIP string) *NetworkCorrelation {
	// Step 0: Resolve application names to server names
	resourceNameOrIP = c.resolveToServer(resourceNameOrIP)

	// Step 1: Find the port mapping for this resource
	mapping := c.findMapping(resourceNameOrIP)
	if mapping == nil {
		return nil
	}

	// Step 2: Find the switch and port
	sw := c.getSwitchByID(mapping.SwitchID)
	if sw == nil {
		return nil
	}

	port := c.getSwitchPort(sw, mapping.PortID)
	if port == nil {
		return nil
	}

	// Step 3: Build the correlation result with port telemetry
	result := &NetworkCorrelation{
		ResourceID:   mapping.ResourceID,
		ResourceName: mapping.ResourceName,
		ResourceIP:   mapping.ResourceIP,
		SwitchID:     sw.ID,
		SwitchName:   sw.Name,
		SwitchModel:  sw.Model,
		SwitchIP:     sw.IP,
		PortID:       port.PortID,
		Speed:        port.Speed,
		FullDuplex:   port.FullDuplex,
		RxErrors:     port.RxErrors,
		TxErrors:     port.TxErrors,
		PacketLoss:   port.Loss,
		Jitter:       port.Jitter,
		Latency:      port.Latency,
		RxBps:        port.RxBps,
		TxBps:        port.TxBps,
		FlappedEpoch: port.LastFlapped,
	}

	// Port status
	if !port.Up {
		result.PortStatus = "down"
	} else if port.Disabled {
		result.PortStatus = "disabled"
	} else {
		result.PortStatus = "up"
	}

	// Format last-flapped time
	if port.LastFlapped > 0 {
		flapTime := time.Unix(int64(port.LastFlapped), 0)
		elapsed := time.Since(flapTime)
		result.LastFlapped = fmt.Sprintf("%s (%s ago)", flapTime.Format("2006-01-02 15:04:05 MST"), formatDuration(elapsed))
	} else {
		result.LastFlapped = "never"
	}

	// Step 4: Analyze for issues
	result.Issues = analyzePortIssues(port)
	result.IssueCount = len(result.Issues)

	// Determine if network is likely root cause
	hasCritical := false
	hasWarning := false
	for _, issue := range result.Issues {
		if issue.Severity == "critical" {
			hasCritical = true
		}
		if issue.Severity == "warning" {
			hasWarning = true
		}
	}

	if hasCritical {
		result.NetworkIsRoot = true
		result.Verdict = "NETWORK IS LIKELY ROOT CAUSE — critical network issues detected on the switch port"
		result.Recommendation = buildRecommendation(result.Issues)
	} else if hasWarning {
		result.NetworkIsRoot = true
		result.Verdict = "NETWORK IS A CONTRIBUTING FACTOR — warning-level network issues detected"
		result.Recommendation = buildRecommendation(result.Issues)
	} else {
		result.NetworkIsRoot = false
		result.Verdict = "NETWORK IS HEALTHY — no significant issues detected on the switch port; root cause is likely elsewhere"
	}

	return result
}

// GetSwitchByName returns a switch by name or hostname.
func (c *Client) GetSwitchByName(name string) *SwitchStats {
	for _, sw := range c.Switches {
		if containsIgnoreCase(sw.Name, name) || containsIgnoreCase(sw.Hostname, name) {
			return &sw
		}
	}
	return nil
}

// GetAllSwitches returns all managed switches.
func (c *Client) GetAllSwitches() []SwitchStats {
	return c.Switches
}

// GetSwitchPortStats returns all port stats for a named switch.
func (c *Client) GetSwitchPortStats(switchName string) []SwitchPort {
	sw := c.GetSwitchByName(switchName)
	if sw == nil {
		return nil
	}
	return sw.Ports
}

// =============================================================================
// Blast Radius Analysis
// =============================================================================
//
// AnalyzeBlastRadius traverses the dependency graph starting from a resource
// (server) to map all affected applications, services, and user groups.
//
// Traversal: resource → applications (hosts) → downstream apps (depends_on) → user groups (serves)
// =============================================================================

// AnalyzeBlastRadius computes the blast radius from a resource name or switch port.
// It finds all downstream dependencies and affected user groups.
// Accepts both server names (e.g., "k8s-node-04") and application names
// (e.g., "greenlake-portal"). Application names are resolved to their hosting server.
func (c *Client) AnalyzeBlastRadius(resourceNameOrIP string) *BlastRadiusResult {
	if len(c.Nodes) == 0 || len(c.Edges) == 0 {
		return nil
	}

	// Resolve application names to server names
	resourceNameOrIP = c.resolveToServer(resourceNameOrIP)

	// Build lookup maps
	nodeMap := make(map[string]DependencyNode)
	for _, n := range c.Nodes {
		nodeMap[n.ID] = n
	}

	// Find the starting resource node
	var startNode *DependencyNode
	for _, n := range c.Nodes {
		if n.Type == "server" &&
			(containsIgnoreCase(n.Name, resourceNameOrIP) ||
				containsIgnoreCase(n.ID, resourceNameOrIP)) {
			node := n
			startNode = &node
			break
		}
	}
	if startNode == nil {
		return nil
	}

	// Find the switch connected to this server
	var switchName string
	mapping := c.findMapping(resourceNameOrIP)
	if mapping != nil {
		switchName = mapping.SwitchName
	}

	// BFS to find all impacted nodes
	visited := make(map[string]bool)
	impacted := []ImpactedNode{}
	queue := []struct {
		id     string
		impact string // "direct" or "indirect"
		reason string
	}{{startNode.ID, "direct", "Root cause server"}}
	visited[startNode.ID] = true

	// Also mark the switch as visited (if found)
	if mapping != nil {
		visited[mapping.SwitchID] = true
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		node, ok := nodeMap[current.id]
		if !ok {
			continue
		}

		// Add to impacted list (skip the root server itself for cleaner output)
		if current.id != startNode.ID {
			impacted = append(impacted, ImpactedNode{
				ID: node.ID, Name: node.Name, Type: node.Type,
				Layer: node.Layer, Criticality: node.Criticality,
				Impact: current.impact, Reason: current.reason,
			})
		}

		// Find outgoing edges from this node
		for _, edge := range c.Edges {
			if edge.FromID != current.id {
				continue
			}
			if visited[edge.ToID] {
				continue
			}
			visited[edge.ToID] = true

			impact := "indirect"
			if edge.Relationship == "hosts" {
				impact = "direct"
			}
			reason := fmt.Sprintf("%s %s %s", node.Name, edge.Relationship, nodeMap[edge.ToID].Name)
			queue = append(queue, struct {
				id     string
				impact string
				reason string
			}{edge.ToID, impact, reason})
		}

		// For server nodes, also follow "depends_on" edges where this server hosts apps
		// that other apps depend on
		if node.Type == "application" {
			for _, edge := range c.Edges {
				if edge.ToID != current.id || edge.Relationship != "depends_on" {
					continue
				}
				if visited[edge.FromID] {
					continue
				}
				visited[edge.FromID] = true
				reason := fmt.Sprintf("%s depends_on %s (impacted)", nodeMap[edge.FromID].Name, node.Name)
				queue = append(queue, struct {
					id     string
					impact string
					reason string
				}{edge.FromID, "indirect", reason})
			}
		}
	}

	// Count by type
	appCount := 0
	serverCount := 0
	serviceCount := 0
	userCount := 0
	highestCriticality := "low"

	for _, n := range impacted {
		switch n.Type {
		case "application":
			appCount++
		case "server":
			serverCount++
		case "service":
			serviceCount++
		case "user_group":
			if meta, ok := nodeMap[n.ID]; ok {
				if users, exists := meta.Metadata["estimated_users"]; exists {
					var u int
					fmt.Sscanf(users, "%d", &u)
					userCount += u
				}
			}
		}
		if criticalityRank(n.Criticality) > criticalityRank(highestCriticality) {
			highestCriticality = n.Criticality
		}
	}

	// Build critical path (root server → most critical affected app → users)
	criticalPath := buildCriticalPath(startNode, impacted, c.Edges, nodeMap)

	// Business impact statement
	businessImpact := buildBusinessImpact(appCount, userCount, impacted)

	// Root cause description
	rootCauseDesc := fmt.Sprintf("Network issues on %s affecting %s", switchName, startNode.Name)
	if mapping != nil {
		rootCauseDesc = fmt.Sprintf("Network issues on %s port %s affecting %s (%s)",
			mapping.SwitchName, mapping.PortID, startNode.Name, startNode.ID)
	}

	return &BlastRadiusResult{
		RootCauseID:          startNode.ID,
		RootCauseName:        startNode.Name,
		RootCauseType:        "server",
		RootCauseDesc:        rootCauseDesc,
		AffectedApplications: appCount,
		AffectedServers:      serverCount,
		AffectedServices:     serviceCount,
		AffectedUsers:        userCount,
		TotalImpactedNodes:   len(impacted),
		OverallSeverity:      highestCriticality,
		BusinessImpact:       businessImpact,
		ImpactedNodes:        impacted,
		CriticalPath:         criticalPath,
	}
}

func criticalityRank(c string) int {
	switch c {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

func buildCriticalPath(start *DependencyNode, impacted []ImpactedNode, edges []DependencyEdge, nodeMap map[string]DependencyNode) []PathStep {
	path := []PathStep{
		{NodeID: start.ID, NodeName: start.Name, NodeType: start.Type, Relationship: "hosts"},
	}

	// Find the most critical directly impacted application
	var critApp *ImpactedNode
	for i, n := range impacted {
		if n.Type == "application" && n.Impact == "direct" {
			if critApp == nil || criticalityRank(n.Criticality) > criticalityRank(critApp.Criticality) {
				critApp = &impacted[i]
			}
		}
	}
	if critApp != nil {
		path = append(path, PathStep{
			NodeID: critApp.ID, NodeName: critApp.Name, NodeType: critApp.Type, Relationship: "serves",
		})

		// Find the most critical user group served by this app
		for _, edge := range edges {
			if edge.FromID == critApp.ID && edge.Relationship == "serves" {
				if ug, ok := nodeMap[edge.ToID]; ok && ug.Type == "user_group" {
					path = append(path, PathStep{
						NodeID: ug.ID, NodeName: ug.Name, NodeType: ug.Type, Relationship: "",
					})
					break
				}
			}
		}
	}

	return path
}

func buildBusinessImpact(appCount, userCount int, impacted []ImpactedNode) string {
	if appCount == 0 {
		return "No applications directly affected — impact is limited to the server layer."
	}

	// Find revenue data from user groups
	totalRevenue := ""
	for _, n := range impacted {
		if n.Type == "user_group" && n.Criticality == "critical" {
			totalRevenue = "significant revenue"
			break
		}
	}

	impact := fmt.Sprintf("%d application(s) and approximately %d users affected.", appCount, userCount)
	if userCount > 1000 {
		impact += " This is a HIGH-IMPACT incident affecting customer-facing services."
	}
	if totalRevenue != "" {
		impact += " Revenue-generating services are impacted — immediate escalation recommended."
	}
	return impact
}

// =============================================================================
// Guided Remediation
// =============================================================================
//
// GetRemediationPlan generates a step-by-step remediation plan based on the
// network correlation results. Each step includes the command, target device,
// expected outcome, and whether it requires operator approval.
// =============================================================================

// GetRemediationPlan generates a remediation plan for a resource with network issues.
// Accepts both server names (e.g., "k8s-node-04") and application names
// (e.g., "greenlake-portal"). Application names are resolved to their hosting server.
func (c *Client) GetRemediationPlan(resourceNameOrIP string) *RemediationPlan {
	// Resolve application names to server names (CorrelateNetwork also resolves,
	// but we need the resolved name for findMapping below)
	resourceNameOrIP = c.resolveToServer(resourceNameOrIP)

	// First, run correlation to get the issues
	correlation := c.CorrelateNetwork(resourceNameOrIP)
	if correlation == nil {
		return nil
	}
	if len(correlation.Issues) == 0 {
		return nil
	}

	mapping := c.findMapping(resourceNameOrIP)
	if mapping == nil {
		return nil
	}

	sw := c.getSwitchByID(mapping.SwitchID)
	if sw == nil {
		return nil
	}

	plan := &RemediationPlan{
		ResourceName:      correlation.ResourceName,
		ResourceIP:        correlation.ResourceIP,
		SwitchName:        correlation.SwitchName,
		PortID:            correlation.PortID,
		RootCause:         correlation.Verdict,
		PlanID:            fmt.Sprintf("REM-%s-%s", sw.Name, correlation.PortID),
		Title:             fmt.Sprintf("Remediate network issues on %s port %s (connected to %s)", sw.Name, correlation.PortID, correlation.ResourceName),
		RollbackAvailable: true,
		RollbackPlan:      fmt.Sprintf("If remediation fails, restore port %s configuration from backup and engage network engineering team.", correlation.PortID),
	}

	stepNum := 0
	seen := make(map[string]bool)

	// Always start with diagnostics
	stepNum++
	plan.Steps = append(plan.Steps, RemediationStep{
		StepNumber:       stepNum,
		Action:           "Run diagnostics on the affected switch port",
		Command:          fmt.Sprintf("show interfaces %s extensive", correlation.PortID),
		Target:           sw.Name,
		Category:         "diagnostic",
		ExpectedOutcome:  "Confirm current port status, error counters, and link state",
		RiskLevel:        "none",
		RequiresApproval: false,
		EstimatedTime:    "10 seconds",
	})

	stepNum++
	plan.Steps = append(plan.Steps, RemediationStep{
		StepNumber:       stepNum,
		Action:           "Check interface error counters and optics",
		Command:          fmt.Sprintf("show interfaces diagnostics optics %s", correlation.PortID),
		Target:           sw.Name,
		Category:         "diagnostic",
		ExpectedOutcome:  "Identify if transceiver/optics are reporting errors or degraded signal",
		RiskLevel:        "none",
		RequiresApproval: false,
		EstimatedTime:    "10 seconds",
	})

	// Issue-specific remediation steps
	for _, issue := range correlation.Issues {
		switch issue.Type {
		case "link_flap":
			if !seen["link_flap"] {
				stepNum++
				plan.Steps = append(plan.Steps, RemediationStep{
					StepNumber:       stepNum,
					Action:           "Bounce the interface to clear link flap state",
					Command:          fmt.Sprintf("set interfaces %s disable\ncommit\ndelete interfaces %s disable\ncommit", correlation.PortID, correlation.PortID),
					Target:           sw.Name,
					Category:         "mitigation",
					ExpectedOutcome:  "Interface restarts cleanly, link flapping stops, port negotiates at full speed",
					RiskLevel:        "medium",
					RequiresApproval: true,
					EstimatedTime:    "30 seconds",
				})
				seen["link_flap"] = true
			}

		case "packet_loss", "rx_errors", "tx_errors":
			if !seen["errors"] {
				stepNum++
				plan.Steps = append(plan.Steps, RemediationStep{
					StepNumber:       stepNum,
					Action:           "Clear interface error counters to establish fresh baseline",
					Command:          fmt.Sprintf("clear interfaces statistics %s", correlation.PortID),
					Target:           sw.Name,
					Category:         "diagnostic",
					ExpectedOutcome:  "Error counters reset to zero — monitor for new errors over next 5 minutes",
					RiskLevel:        "none",
					RequiresApproval: false,
					EstimatedTime:    "5 seconds",
				})

				stepNum++
				plan.Steps = append(plan.Steps, RemediationStep{
					StepNumber:       stepNum,
					Action:           "Check physical cable connectivity — reseat cable on both switch and server NIC",
					Command:          "",
					Target:           fmt.Sprintf("%s ↔ %s", sw.Name, correlation.ResourceName),
					Category:         "resolution",
					ExpectedOutcome:  "Physical cable properly seated, link renegotiates at correct speed and duplex",
					RiskLevel:        "low",
					RequiresApproval: true,
					EstimatedTime:    "2 minutes",
				})
				seen["errors"] = true
			}

		case "duplex_mismatch", "speed_downgrade":
			if !seen["duplex"] {
				stepNum++
				plan.Steps = append(plan.Steps, RemediationStep{
					StepNumber:       stepNum,
					Action:           "Force port speed and duplex to 1000Mbps full-duplex",
					Command:          fmt.Sprintf("set interfaces %s speed 1g\nset interfaces %s link-mode full-duplex\ncommit", correlation.PortID, correlation.PortID),
					Target:           sw.Name,
					Category:         "resolution",
					ExpectedOutcome:  "Port negotiates at 1Gbps full-duplex, eliminating half-duplex collisions",
					RiskLevel:        "medium",
					RequiresApproval: true,
					EstimatedTime:    "30 seconds",
				})
				seen["duplex"] = true
			}

		case "high_latency", "high_jitter":
			if !seen["latency"] {
				stepNum++
				plan.Steps = append(plan.Steps, RemediationStep{
					StepNumber:       stepNum,
					Action:           "Check for congestion on uplink and QoS policies",
					Command:          fmt.Sprintf("show class-of-service interface %s\nshow interfaces queue %s", correlation.PortID, correlation.PortID),
					Target:           sw.Name,
					Category:         "diagnostic",
					ExpectedOutcome:  "Identify if QoS policies are throttling traffic or if uplink is congested",
					RiskLevel:        "none",
					RequiresApproval: false,
					EstimatedTime:    "15 seconds",
				})
				seen["latency"] = true
			}

		case "port_down":
			if !seen["port_down"] {
				stepNum++
				plan.Steps = append(plan.Steps, RemediationStep{
					StepNumber:       stepNum,
					Action:           "Enable the port if administratively disabled",
					Command:          fmt.Sprintf("delete interfaces %s disable\ncommit", correlation.PortID),
					Target:           sw.Name,
					Category:         "resolution",
					ExpectedOutcome:  "Port comes up and establishes link with the connected server",
					RiskLevel:        "medium",
					RequiresApproval: true,
					EstimatedTime:    "15 seconds",
				})
				seen["port_down"] = true
			}
		}
	}

	// Always end with verification
	stepNum++
	plan.Steps = append(plan.Steps, RemediationStep{
		StepNumber:       stepNum,
		Action:           "Verify remediation — check port status and error counters",
		Command:          fmt.Sprintf("show interfaces %s extensive | match \"errors|loss|flap|speed|duplex\"", correlation.PortID),
		Target:           sw.Name,
		Category:         "verification",
		ExpectedOutcome:  "Port is up, 1Gbps full-duplex, zero new errors, no packet loss",
		RiskLevel:        "none",
		RequiresApproval: false,
		EstimatedTime:    "10 seconds",
	})

	stepNum++
	plan.Steps = append(plan.Steps, RemediationStep{
		StepNumber:       stepNum,
		Action:           fmt.Sprintf("Verify application health on %s", correlation.ResourceName),
		Command:          fmt.Sprintf("curl -s -o /dev/null -w '%%{http_code} %%{time_total}s' https://%s:8443/health", correlation.ResourceIP),
		Target:           correlation.ResourceName,
		Category:         "verification",
		ExpectedOutcome:  "Application responds with HTTP 200, response time under 1 second",
		RiskLevel:        "none",
		RequiresApproval: false,
		EstimatedTime:    "10 seconds",
	})

	plan.TotalSteps = len(plan.Steps)

	// Set urgency and risk based on issue severity
	hasCritical := false
	for _, issue := range correlation.Issues {
		if issue.Severity == "critical" {
			hasCritical = true
			break
		}
	}

	if hasCritical {
		plan.Urgency = "immediate"
		plan.RiskLevel = "medium"
		plan.RequiresApproval = true
		plan.ApprovalNote = "This plan includes service-impacting steps (interface bounce/reconfiguration). Approval from on-call network engineer required."
		plan.EstimatedDowntime = "30-60 seconds during interface bounce"
	} else {
		plan.Urgency = "urgent"
		plan.RiskLevel = "low"
		plan.RequiresApproval = true
		plan.ApprovalNote = "This plan includes diagnostic and minor remediation steps. Approval recommended before cable reseat."
		plan.EstimatedDowntime = "Minimal — brief link interruption during cable reseat"
	}

	return plan
}

// =============================================================================
// Internal helpers
// =============================================================================

// resolveToServer resolves any name (server, application, or IP) to a server name.
// If the input matches a server directly (by name, ID, or IP), it returns the input.
// If it matches an application, it follows the 'hosts' edge backwards to find the
// server that runs that application and returns the server's name.
// This allows correlate_network, blast_radius, and get_remediation_plan to accept
// both server names and application names (e.g., "greenlake-portal" → "k8s-node-04").
func (c *Client) resolveToServer(nameOrIP string) string {
	// 1. Check if it directly matches a server in port mappings
	if m := c.findMapping(nameOrIP); m != nil {
		return nameOrIP // already a server name/IP
	}

	// 2. Check if it matches an application node, then find the hosting server.
	//    Two-pass approach: exact match first, then fuzzy match.
	//    This prevents "greenlake-ops-portal" from fuzzy-matching "greenlake-portal"
	//    when an exact match exists.

	// resolveAppToServer finds the hosting server for a matched application node.
	resolveAppToServer := func(node DependencyNode) string {
		for _, edge := range c.Edges {
			if edge.ToID == node.ID && edge.Relationship == "hosts" {
				for _, sn := range c.Nodes {
					if sn.ID == edge.FromID && sn.Type == "server" {
						fmt.Printf("  [resolve] Resolved application %q → server %q\n", nameOrIP, sn.Name)
						return sn.Name
					}
				}
			}
		}
		return ""
	}

	// Pass 1: Exact match (normalized names are equal)
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(nameOrIP, "-", " "), "_", " "))
	for _, node := range c.Nodes {
		if node.Type != "application" {
			continue
		}
		nn := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(node.Name, "-", " "), "_", " "))
		ni := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(node.ID, "-", " "), "_", " "))
		if nn == normalized || ni == normalized {
			if server := resolveAppToServer(node); server != "" {
				return server
			}
		}
	}

	// Pass 2: Fuzzy match (fallback — substring/word matching)
	for _, node := range c.Nodes {
		if node.Type != "application" {
			continue
		}
		if !fuzzyMatchName(node.Name, nameOrIP) && !fuzzyMatchName(node.ID, nameOrIP) {
			continue
		}
		if server := resolveAppToServer(node); server != "" {
			return server
		}
	}

	// 3. Check if it matches a server node by name/ID (may not be in port mappings)
	for _, node := range c.Nodes {
		if node.Type == "server" &&
			(containsIgnoreCase(node.Name, nameOrIP) || containsIgnoreCase(node.ID, nameOrIP)) {
			return node.Name
		}
	}

	// No match — return as-is and let the caller handle the nil result
	return nameOrIP
}

// findMapping looks up the PortMapping entry that correlates an OpsRamp
// resource to its Juniper switch port. It searches by IP address (primary
// real-world key), hostname, or resource ID — whichever the caller provides.
//
// In production, the most reliable correlation attribute is the IP address
// because both OpsRamp (via agent/discovery) and Juniper (via ARP/LLDP on
// the switch port) independently observe the same IP for a given server.
// The mapping data itself comes from GetNetworkPortMappings() in
// mockdata/network.go, where each ResourceIP is deliberately kept in sync
// with the IPAddress in mockdata/resources.go to mirror this behaviour.
func (c *Client) findMapping(resourceNameOrIP string) *PortMapping {
	q := strings.ToLower(resourceNameOrIP)
	for _, m := range c.Mappings {
		if containsIgnoreCase(m.ResourceName, q) ||
			containsIgnoreCase(m.ResourceIP, q) ||
			containsIgnoreCase(m.ResourceID, q) {
			return &m
		}
	}
	return nil
}

func (c *Client) getSwitchByID(id string) *SwitchStats {
	for _, sw := range c.Switches {
		if strings.EqualFold(sw.ID, id) {
			return &sw
		}
	}
	return nil
}

func (c *Client) getSwitchPort(sw *SwitchStats, portID string) *SwitchPort {
	for _, p := range sw.Ports {
		if strings.EqualFold(p.PortID, portID) {
			return &p
		}
	}
	return nil
}

// analyzePortIssues checks port telemetry against operational thresholds
// and returns a list of detected issues.
//
// Thresholds are based on common network operations best practices:
//
//	Packet loss: >0.1% warning, >1% critical
//	RX/TX errors: >100 warning, >10000 critical
//	Latency: >10ms warning, >50ms critical (for datacenter)
//	Jitter: >5ms warning, >20ms critical
//	Link flap: within last 24h is warning, within last 1h is critical
//	Half-duplex: always warning (should be full-duplex in datacenter)
//	Port down: always critical
func analyzePortIssues(port *SwitchPort) []NetworkIssue {
	var issues []NetworkIssue

	// Port down
	if !port.Up && !port.Disabled {
		issues = append(issues, NetworkIssue{
			Type:        "port_down",
			Severity:    "critical",
			Description: fmt.Sprintf("Port %s is DOWN — link is not operational", port.PortID),
			Value:       "down",
			Threshold:   "up",
		})
	}

	// Packet loss
	if port.Loss >= 1.0 {
		issues = append(issues, NetworkIssue{
			Type:        "packet_loss",
			Severity:    "critical",
			Description: fmt.Sprintf("High packet loss of %.1f%% on port %s — packets are being dropped", port.Loss, port.PortID),
			Value:       fmt.Sprintf("%.1f%%", port.Loss),
			Threshold:   ">1.0%",
		})
	} else if port.Loss >= 0.1 {
		issues = append(issues, NetworkIssue{
			Type:        "packet_loss",
			Severity:    "warning",
			Description: fmt.Sprintf("Elevated packet loss of %.1f%% on port %s", port.Loss, port.PortID),
			Value:       fmt.Sprintf("%.1f%%", port.Loss),
			Threshold:   ">0.1%",
		})
	}

	// RX errors
	if port.RxErrors >= 10000 {
		issues = append(issues, NetworkIssue{
			Type:        "rx_errors",
			Severity:    "critical",
			Description: fmt.Sprintf("High receive errors (%d) on port %s — likely CRC/FCS errors from bad cable, transceiver, or EMI", port.RxErrors, port.PortID),
			Value:       fmt.Sprintf("%d", port.RxErrors),
			Threshold:   ">10000",
		})
	} else if port.RxErrors >= 100 {
		issues = append(issues, NetworkIssue{
			Type:        "rx_errors",
			Severity:    "warning",
			Description: fmt.Sprintf("Elevated receive errors (%d) on port %s", port.RxErrors, port.PortID),
			Value:       fmt.Sprintf("%d", port.RxErrors),
			Threshold:   ">100",
		})
	}

	// TX errors
	if port.TxErrors >= 10000 {
		issues = append(issues, NetworkIssue{
			Type:        "tx_errors",
			Severity:    "critical",
			Description: fmt.Sprintf("High transmit errors (%d) on port %s — possible congestion or hardware fault", port.TxErrors, port.PortID),
			Value:       fmt.Sprintf("%d", port.TxErrors),
			Threshold:   ">10000",
		})
	} else if port.TxErrors >= 100 {
		issues = append(issues, NetworkIssue{
			Type:        "tx_errors",
			Severity:    "warning",
			Description: fmt.Sprintf("Elevated transmit errors (%d) on port %s", port.TxErrors, port.PortID),
			Value:       fmt.Sprintf("%d", port.TxErrors),
			Threshold:   ">100",
		})
	}

	// Latency (datacenter thresholds)
	if port.Latency >= 50.0 {
		issues = append(issues, NetworkIssue{
			Type:        "high_latency",
			Severity:    "critical",
			Description: fmt.Sprintf("Very high port latency of %.1fms on %s — significantly impacting application performance", port.Latency, port.PortID),
			Value:       fmt.Sprintf("%.1fms", port.Latency),
			Threshold:   ">50ms",
		})
	} else if port.Latency >= 10.0 {
		issues = append(issues, NetworkIssue{
			Type:        "high_latency",
			Severity:    "warning",
			Description: fmt.Sprintf("Elevated port latency of %.1fms on %s", port.Latency, port.PortID),
			Value:       fmt.Sprintf("%.1fms", port.Latency),
			Threshold:   ">10ms",
		})
	}

	// Jitter
	if port.Jitter >= 20.0 {
		issues = append(issues, NetworkIssue{
			Type:        "high_jitter",
			Severity:    "critical",
			Description: fmt.Sprintf("Very high jitter of %.1fms on port %s — causes inconsistent latency and retransmissions", port.Jitter, port.PortID),
			Value:       fmt.Sprintf("%.1fms", port.Jitter),
			Threshold:   ">20ms",
		})
	} else if port.Jitter >= 5.0 {
		issues = append(issues, NetworkIssue{
			Type:        "high_jitter",
			Severity:    "warning",
			Description: fmt.Sprintf("Elevated jitter of %.1fms on port %s", port.Jitter, port.PortID),
			Value:       fmt.Sprintf("%.1fms", port.Jitter),
			Threshold:   ">5ms",
		})
	}

	// Link flaps
	if port.LastFlapped > 0 {
		flapTime := time.Unix(int64(port.LastFlapped), 0)
		elapsed := time.Since(flapTime)

		if elapsed < 1*time.Hour {
			issues = append(issues, NetworkIssue{
				Type:        "link_flap",
				Severity:    "critical",
				Description: fmt.Sprintf("Port %s link flapped %s ago — very recent instability indicates active hardware or cable fault", port.PortID, formatDuration(elapsed)),
				Value:       formatDuration(elapsed) + " ago",
				Threshold:   "<1 hour",
			})
		} else if elapsed < 24*time.Hour {
			issues = append(issues, NetworkIssue{
				Type:        "link_flap",
				Severity:    "warning",
				Description: fmt.Sprintf("Port %s link flapped %s ago — recent instability may indicate intermittent connectivity", port.PortID, formatDuration(elapsed)),
				Value:       formatDuration(elapsed) + " ago",
				Threshold:   "<24 hours",
			})
		}
	}

	// Half-duplex mismatch
	if port.Up && !port.FullDuplex && port.PortUsage == "lan" {
		issues = append(issues, NetworkIssue{
			Type:        "duplex_mismatch",
			Severity:    "warning",
			Description: fmt.Sprintf("Port %s is running HALF-DUPLEX at %dMbps — typical of auto-negotiation failure or bad cable", port.PortID, port.Speed),
			Value:       fmt.Sprintf("half-duplex %dMbps", port.Speed),
			Threshold:   "full-duplex 1000Mbps",
		})
	}

	// Speed downgrade (expected 1000, got 100 or 10)
	if port.Up && port.Speed < 1000 && port.PortUsage == "lan" {
		issues = append(issues, NetworkIssue{
			Type:        "speed_downgrade",
			Severity:    "warning",
			Description: fmt.Sprintf("Port %s negotiated at %dMbps instead of expected 1Gbps — check cable and NIC", port.PortID, port.Speed),
			Value:       fmt.Sprintf("%dMbps", port.Speed),
			Threshold:   "1000Mbps",
		})
	}

	return issues
}

// buildRecommendation generates actionable remediation steps based on detected issues.
func buildRecommendation(issues []NetworkIssue) string {
	var steps []string
	seen := make(map[string]bool)

	for _, issue := range issues {
		switch issue.Type {
		case "packet_loss":
			if !seen["packet_loss"] {
				steps = append(steps, "Check physical cable connections and replace if damaged")
				steps = append(steps, "Inspect transceiver/SFP module for errors ('show interfaces diagnostics optics')")
				steps = append(steps, "Check for congestion on the uplink port")
				seen["packet_loss"] = true
			}
		case "rx_errors":
			if !seen["rx_errors"] {
				steps = append(steps, "Replace the cable between server and switch port (likely CRC/FCS errors)")
				steps = append(steps, "Check for electromagnetic interference near the cable run")
				steps = append(steps, "Inspect the server NIC for hardware faults")
				seen["rx_errors"] = true
			}
		case "tx_errors":
			if !seen["tx_errors"] {
				steps = append(steps, "Check switch port for micro-bursts or queue drops ('show interfaces queue')")
				steps = append(steps, "Consider enabling flow control or increasing buffer allocation")
				seen["tx_errors"] = true
			}
		case "link_flap":
			if !seen["link_flap"] {
				steps = append(steps, "Reseat the cable on both ends (switch port and server NIC)")
				steps = append(steps, "Replace the patch cable if flapping persists")
				steps = append(steps, "Check for STP topology changes ('show spanning-tree interface')")
				seen["link_flap"] = true
			}
		case "duplex_mismatch", "speed_downgrade":
			if !seen["duplex"] {
				steps = append(steps, "Force speed and duplex to 1000Mbps full-duplex on both switch port and server NIC")
				steps = append(steps, "Replace the cable — half-duplex/100Mbps typically indicates a bad cable or connector")
				seen["duplex"] = true
			}
		case "high_latency":
			if !seen["latency"] {
				steps = append(steps, "Check for congestion on intermediate links and uplink ports")
				steps = append(steps, "Verify QoS policies are not de-prioritizing this traffic")
				seen["latency"] = true
			}
		case "high_jitter":
			if !seen["jitter"] {
				steps = append(steps, "Investigate buffer utilization on the switch ('show class-of-service interface')")
				steps = append(steps, "Check for other high-bandwidth flows causing contention")
				seen["jitter"] = true
			}
		case "port_down":
			if !seen["port_down"] {
				steps = append(steps, "Check physical cable connectivity — port is completely down")
				steps = append(steps, "Verify the port is not administratively disabled ('show interfaces terse')")
				steps = append(steps, "Try a different switch port to isolate the fault")
				seen["port_down"] = true
			}
		}
	}

	if len(steps) == 0 {
		return ""
	}

	return "Recommended actions:\n" + numberedList(steps)
}

func numberedList(items []string) string {
	var sb strings.Builder
	for i, item := range items {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, item))
	}
	return sb.String()
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		mins := int(d.Minutes()) % 60
		if mins > 0 {
			return fmt.Sprintf("%dh %dm", hours, mins)
		}
		return fmt.Sprintf("%d hours", hours)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	if hours > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	return fmt.Sprintf("%d days", days)
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// fuzzyMatchName does a flexible name match between a node name and a user query.
// It handles variations like "greenlake portal" matching "greenlake-portal",
// "aruba-central" matching "aruba central", "dscc" matching "dscc-console", etc.
//
// Matching strategy:
//  1. Direct substring match (after normalizing separators)
//  2. Word-level match: any significant query word (4+ chars) found in the node name
func fuzzyMatchName(nodeName, query string) bool {
	// Normalize: lowercase, replace all separators with spaces
	normalize := func(s string) string {
		s = strings.ToLower(s)
		s = strings.ReplaceAll(s, "-", " ")
		s = strings.ReplaceAll(s, "_", " ")
		return strings.TrimSpace(s)
	}

	nn := normalize(nodeName)
	qq := normalize(query)

	// Direct substring match on normalized forms
	if strings.Contains(nn, qq) || strings.Contains(qq, nn) {
		return true
	}

	// Word-level match: check if any significant query word appears in the node name
	// "greenlake portal" → words ["greenlake", "portal"] → "greenlake" (9 chars) matches "greenlake portal"
	for _, word := range strings.Fields(qq) {
		if len(word) >= 4 && strings.Contains(nn, word) {
			return true
		}
	}

	return false
}

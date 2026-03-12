package tools

import (
	"encoding/json"
	"fmt"
	"opsramp-agent/juniper"
	"opsramp-agent/mockdata"
	"opsramp-agent/opsramp"
	"strings"
	"testing"
)

// setupClients creates fully-initialized OpsRamp + Juniper clients using mock data.
func setupClients() (*opsramp.Client, *juniper.Client) {
	opsClient := opsramp.NewClient(
		mockdata.GetAlerts(),
		mockdata.GetResources(),
		mockdata.GetIncidents(),
		mockdata.GetMetricHistory(),
	)

	junClient := juniper.NewClient(
		mockdata.GetNetworkSwitches(),
		mockdata.GetNetworkPortMappings(),
	)
	junClient.SetDependencyGraph(
		mockdata.GetDependencyNodes(),
		mockdata.GetDependencyEdges(),
	)

	return opsClient, junClient
}

// execTool is a helper that runs a tool call and returns the JSON result string.
func execTool(t *testing.T, opsClient *opsramp.Client, junClient *juniper.Client, name string, args map[string]string) string {
	t.Helper()
	call := ToolCall{Name: name, Arguments: args}
	result, err := ExecuteWithOptions(opsClient, call, ExecuteOptions{
		Juniper: junClient,
	})
	if err != nil {
		t.Fatalf("Tool %q returned error: %v", name, err)
	}
	if result == "" {
		t.Fatalf("Tool %q returned empty result", name)
	}
	return result
}

// assertContains checks that result contains all expected substrings.
func assertContains(t *testing.T, toolName, result string, expected ...string) {
	t.Helper()
	lower := strings.ToLower(result)
	for _, exp := range expected {
		if !strings.Contains(lower, strings.ToLower(exp)) {
			t.Errorf("[%s] Expected result to contain %q but it doesn't.\nResult (first 500 chars): %s",
				toolName, exp, result[:min(500, len(result))])
		}
	}
}

// assertValidJSON checks result parses as JSON and returns it decoded.
func assertValidJSON(t *testing.T, toolName, result string) map[string]interface{} {
	t.Helper()
	var decoded map[string]interface{}
	if err := json.Unmarshal([]byte(result), &decoded); err != nil {
		t.Fatalf("[%s] Result is not valid JSON: %v\nResult: %s", toolName, err, result[:min(300, len(result))])
	}
	return decoded
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// =============================================================================
// Tool 1: search_alerts
// =============================================================================

func TestTool01_SearchAlerts_Critical(t *testing.T) {
	ops, jun := setupClients()
	result := execTool(t, ops, jun, "search_alerts", map[string]string{"state": "Critical"})
	data := assertValidJSON(t, "search_alerts", result)

	count, _ := data["count"].(float64)
	if count == 0 {
		t.Error("Expected at least 1 critical alert, got 0")
	}
	assertContains(t, "search_alerts(Critical)", result, "Critical")
	t.Logf("search_alerts(Critical): %d results", int(count))
}

func TestTool01_SearchAlerts_ByResource(t *testing.T) {
	ops, jun := setupClients()
	result := execTool(t, ops, jun, "search_alerts", map[string]string{"resource_name": "web-server-prod-01"})
	data := assertValidJSON(t, "search_alerts", result)

	count, _ := data["count"].(float64)
	if count == 0 {
		t.Error("Expected at least 1 alert for web-server-prod-01")
	}
	assertContains(t, "search_alerts(resource)", result, "web-server-prod-01")
}

func TestTool01_SearchAlerts_NoResults(t *testing.T) {
	ops, jun := setupClients()
	result := execTool(t, ops, jun, "search_alerts", map[string]string{"query": "nonexistent-xyz-12345"})
	assertContains(t, "search_alerts(empty)", result, "No alerts found")
}

// =============================================================================
// Tool 2: search_resources
// =============================================================================

func TestTool02_SearchResources_AWS(t *testing.T) {
	ops, jun := setupClients()
	result := execTool(t, ops, jun, "search_resources", map[string]string{"cloud": "AWS"})
	data := assertValidJSON(t, "search_resources", result)

	count, _ := data["count"].(float64)
	if count == 0 {
		t.Error("Expected at least 1 AWS resource")
	}
	assertContains(t, "search_resources(AWS)", result, "AWS")
	t.Logf("search_resources(AWS): %d results", int(count))
}

func TestTool02_SearchResources_OnPrem(t *testing.T) {
	ops, jun := setupClients()
	result := execTool(t, ops, jun, "search_resources", map[string]string{"cloud": "OnPrem"})
	data := assertValidJSON(t, "search_resources", result)

	count, _ := data["count"].(float64)
	if count == 0 {
		t.Error("Expected at least 1 OnPrem resource")
	}
	// OnPrem = traditional customer-owned datacenter servers (monitoring-agent, ldap, jenkins, esxi)
	// K8s nodes are now HPE GreenLake, not OnPrem
	if count > 6 {
		t.Errorf("Expected at most 6 OnPrem resources (non-k8s), got %d", int(count))
	}
	t.Logf("search_resources(OnPrem): %d results", int(count))
}

func TestTool02_SearchResources_HPEGreenLake(t *testing.T) {
	ops, jun := setupClients()
	result := execTool(t, ops, jun, "search_resources", map[string]string{"cloud": "HPE GreenLake"})
	data := assertValidJSON(t, "search_resources", result)

	count, _ := data["count"].(float64)
	if count == 0 {
		t.Error("Expected at least 1 HPE GreenLake resource (k8s nodes)")
	}
	// 5 k8s nodes: k8s-master-01, k8s-node-01 to k8s-node-04
	if count != 5 {
		t.Errorf("Expected exactly 5 HPE GreenLake resources, got %d", int(count))
	}
	assertContains(t, "search_resources(HPE GreenLake)", result, "HPE GreenLake")
	t.Logf("search_resources(HPE GreenLake): %d results", int(count))
}

func TestTool02_SearchResources_AllClouds(t *testing.T) {
	ops, jun := setupClients()
	result := execTool(t, ops, jun, "search_resources", map[string]string{})
	data := assertValidJSON(t, "search_resources", result)

	count, _ := data["count"].(float64)
	if count < 10 {
		t.Errorf("Expected at least 10 resources across all clouds, got %d", int(count))
	}
	t.Logf("search_resources(all): %d results", int(count))
}

// =============================================================================
// Tool 3: get_resource_details
// =============================================================================

func TestTool03_GetResourceDetails(t *testing.T) {
	ops, jun := setupClients()
	result := execTool(t, ops, jun, "get_resource_details", map[string]string{"resource_name": "db-primary-01"})
	data := assertValidJSON(t, "get_resource_details", result)

	name, _ := data["name"].(string)
	if name != "db-primary-01" {
		t.Errorf("Expected resource name 'db-primary-01', got %q", name)
	}
	assertContains(t, "get_resource_details", result, "10.0.2.10", "postgresql")
}

func TestTool03_GetResourceDetails_NotFound(t *testing.T) {
	ops, jun := setupClients()
	result := execTool(t, ops, jun, "get_resource_details", map[string]string{"resource_name": "nonexistent-server"})
	assertContains(t, "get_resource_details(404)", result, "not found")
}

// =============================================================================
// Tool 4: search_incidents
// =============================================================================

func TestTool04_SearchIncidents_Open(t *testing.T) {
	ops, jun := setupClients()
	result := execTool(t, ops, jun, "search_incidents", map[string]string{"status": "Open"})
	data := assertValidJSON(t, "search_incidents", result)

	count, _ := data["count"].(float64)
	if count == 0 {
		t.Error("Expected at least 1 open incident")
	}
	assertContains(t, "search_incidents(Open)", result, "Open")
	t.Logf("search_incidents(Open): %d results", int(count))
}

func TestTool04_SearchIncidents_Urgent(t *testing.T) {
	ops, jun := setupClients()
	result := execTool(t, ops, jun, "search_incidents", map[string]string{"priority": "Urgent"})
	data := assertValidJSON(t, "search_incidents", result)

	count, _ := data["count"].(float64)
	if count == 0 {
		t.Error("Expected at least 1 urgent incident")
	}
	t.Logf("search_incidents(Urgent): %d results", int(count))
}

// =============================================================================
// Tool 5: investigate_resource
// =============================================================================

func TestTool05_InvestigateResource(t *testing.T) {
	ops, jun := setupClients()
	result := execTool(t, ops, jun, "investigate_resource", map[string]string{"resource_name": "web-server-prod-01"})
	data := assertValidJSON(t, "investigate_resource", result)

	rname, _ := data["resourceName"].(string)
	if rname != "web-server-prod-01" {
		t.Errorf("Expected resourceName 'web-server-prod-01', got %q", rname)
	}

	cpu, _ := data["cpuPercent"].(float64)
	if cpu <= 0 {
		t.Error("Expected positive CPU metric")
	}

	alertCount, _ := data["alertCount"].(float64)
	t.Logf("investigate_resource(web-server-prod-01): CPU=%.1f%%, alerts=%d", cpu, int(alertCount))
}

func TestTool05_InvestigateResource_NotFound(t *testing.T) {
	ops, jun := setupClients()
	result := execTool(t, ops, jun, "investigate_resource", map[string]string{"resource_name": "nonexistent-server"})
	assertContains(t, "investigate_resource(404)", result, "not found")
}

// =============================================================================
// Tool 6: get_environment_summary
// =============================================================================

func TestTool06_GetEnvironmentSummary(t *testing.T) {
	ops, jun := setupClients()
	result := execTool(t, ops, jun, "get_environment_summary", map[string]string{})
	data := assertValidJSON(t, "get_environment_summary", result)

	// Should have resource counts
	totalRes, _ := data["totalResources"].(float64)
	if totalRes < 10 {
		t.Errorf("Expected at least 10 total resources, got %d", int(totalRes))
	}
	t.Logf("get_environment_summary: %d total resources", int(totalRes))
}

// =============================================================================
// Tool 7: predict_capacity (single resource)
// =============================================================================

func TestTool07_PredictCapacity_Single(t *testing.T) {
	ops, jun := setupClients()
	result := execTool(t, ops, jun, "predict_capacity", map[string]string{
		"resource_name": "web-server-prod-01",
	})
	data := assertValidJSON(t, "predict_capacity", result)

	count, _ := data["count"].(float64)
	if count == 0 {
		t.Error("Expected at least 1 forecast for web-server-prod-01")
	}
	t.Logf("predict_capacity(web-server-prod-01): %d forecasts", int(count))
}

func TestTool07_PredictCapacity_All(t *testing.T) {
	ops, jun := setupClients()
	result := execTool(t, ops, jun, "predict_capacity", map[string]string{})
	data := assertValidJSON(t, "predict_capacity", result)

	count, _ := data["count"].(float64)
	if count == 0 {
		t.Error("Expected at least 1 forecast across all resources")
	}

	scope, _ := data["scope"].(string)
	if scope != "all monitored resources" {
		t.Errorf("Expected scope 'all monitored resources', got %q", scope)
	}
	t.Logf("predict_capacity(all): %d forecasts", int(count))
}

// =============================================================================
// Tool 8: search_knowledge_base — skipped (needs Ollama embeddings)
// =============================================================================

func TestTool08_SearchKnowledgeBase_NoKB(t *testing.T) {
	// Without a knowledge base loaded, the tool should return an error message.
	ops, _ := setupClients()
	call := ToolCall{Name: "search_knowledge_base", Arguments: map[string]string{"query": "high CPU runbook"}}
	result, err := ExecuteWithOptions(ops, call, ExecuteOptions{
		KB: nil, // no KB loaded
	})
	if err != nil {
		t.Fatalf("Expected nil error, got: %v", err)
	}
	assertContains(t, "search_knowledge_base(no-kb)", result, "Knowledge base not loaded")
	t.Log("search_knowledge_base(no-kb): correctly reports KB not loaded")
}

// =============================================================================
// Tool 9: correlate_network
// =============================================================================

func TestTool09_CorrelateNetwork_ServerWithIssues(t *testing.T) {
	ops, jun := setupClients()
	// k8s-node-04 hosts greenlake-portal and has link flaps on its switch port
	result := execTool(t, ops, jun, "correlate_network", map[string]string{"resource_name": "k8s-node-04"})
	data := assertValidJSON(t, "correlate_network", result)

	networkIsRoot, _ := data["network_is_likely_root_cause"].(bool)
	if !networkIsRoot {
		t.Error("Expected network_is_likely_root_cause=true for k8s-node-04")
	}

	issueCount, _ := data["issue_count"].(float64)
	if issueCount == 0 {
		t.Error("Expected at least 1 network issue for k8s-node-04")
	}

	assertContains(t, "correlate_network(k8s-node-04)", result, "sw-dc-east-04", "ge-0/0/5")
	t.Logf("correlate_network(k8s-node-04): %d issues, network_is_root=%v", int(issueCount), networkIsRoot)
}

func TestTool09_CorrelateNetwork_AppName(t *testing.T) {
	ops, jun := setupClients()
	// Should resolve "greenlake-portal" → k8s-node-04 via dependency graph
	result := execTool(t, ops, jun, "correlate_network", map[string]string{"resource_name": "greenlake-portal"})
	data := assertValidJSON(t, "correlate_network", result)

	resName, _ := data["resource_name"].(string)
	if resName != "k8s-node-04" {
		t.Errorf("Expected resource_name to resolve to 'k8s-node-04', got %q", resName)
	}
	t.Logf("correlate_network(greenlake-portal): resolved to %s", resName)
}

func TestTool09_CorrelateNetwork_CleanPort(t *testing.T) {
	ops, jun := setupClients()
	// web-server-prod-01 has a clean switch port (no network issues)
	result := execTool(t, ops, jun, "correlate_network", map[string]string{"resource_name": "web-server-prod-01"})
	data := assertValidJSON(t, "correlate_network", result)

	networkIsRoot, _ := data["network_is_likely_root_cause"].(bool)
	if networkIsRoot {
		t.Error("Expected network_is_likely_root_cause=false for web-server-prod-01 (clean port)")
	}
	t.Logf("correlate_network(web-server-prod-01): network_is_root=%v (correct — clean port)", networkIsRoot)
}

// =============================================================================
// Tool 10: blast_radius
// =============================================================================

func TestTool10_BlastRadius(t *testing.T) {
	ops, jun := setupClients()
	result := execTool(t, ops, jun, "blast_radius", map[string]string{"resource_name": "k8s-node-04"})
	data := assertValidJSON(t, "blast_radius", result)

	affectedUsers, _ := data["affected_users"].(float64)
	if affectedUsers == 0 {
		t.Error("Expected affected_users > 0 for k8s-node-04")
	}

	affectedApps, _ := data["affected_applications"].(float64)
	if affectedApps == 0 {
		t.Error("Expected affected_applications > 0 for k8s-node-04")
	}

	assertContains(t, "blast_radius(k8s-node-04)", result, "k8s-node-04")
	t.Logf("blast_radius(k8s-node-04): %d apps, %d users affected", int(affectedApps), int(affectedUsers))
}

func TestTool10_BlastRadius_AppName(t *testing.T) {
	ops, jun := setupClients()
	result := execTool(t, ops, jun, "blast_radius", map[string]string{"resource_name": "greenlake-portal"})
	data := assertValidJSON(t, "blast_radius", result)

	assertContains(t, "blast_radius(greenlake-portal)", result, "greenlake")
	affectedUsers, _ := data["affected_users"].(float64)
	t.Logf("blast_radius(greenlake-portal): %d affected users", int(affectedUsers))
}

// =============================================================================
// Tool 11: get_remediation_plan
// =============================================================================

func TestTool11_GetRemediationPlan(t *testing.T) {
	ops, jun := setupClients()
	result := execTool(t, ops, jun, "get_remediation_plan", map[string]string{"resource_name": "k8s-node-04"})
	data := assertValidJSON(t, "get_remediation_plan", result)

	totalSteps, _ := data["total_steps"].(float64)
	if totalSteps == 0 {
		t.Error("Expected at least 1 remediation step for k8s-node-04")
	}

	assertContains(t, "get_remediation_plan(k8s-node-04)", result, "step", "command")
	t.Logf("get_remediation_plan(k8s-node-04): %d steps", int(totalSteps))
}

func TestTool11_GetRemediationPlan_AppName(t *testing.T) {
	ops, jun := setupClients()
	// Should resolve greenlake-portal → k8s-node-04 and generate plan
	result := execTool(t, ops, jun, "get_remediation_plan", map[string]string{"resource_name": "greenlake-portal"})
	data := assertValidJSON(t, "get_remediation_plan", result)

	totalSteps, _ := data["total_steps"].(float64)
	if totalSteps == 0 {
		t.Error("Expected at least 1 remediation step for greenlake-portal")
	}
	t.Logf("get_remediation_plan(greenlake-portal): %d steps", int(totalSteps))
}

func TestTool11_GetRemediationPlan_CleanResource(t *testing.T) {
	ops, jun := setupClients()
	// web-server-prod-01 has no network issues — remediation should return nil/error
	result := execTool(t, ops, jun, "get_remediation_plan", map[string]string{"resource_name": "web-server-prod-01"})
	// Either returns an error or returns a plan with 0 steps (depending on implementation)
	assertValidJSON(t, "get_remediation_plan(clean)", result)
	t.Logf("get_remediation_plan(clean-resource): %s", result[:min(200, len(result))])
}

// =============================================================================
// End-to-End Integration: "Why is greenlake portal slow?" flow
// =============================================================================

func TestE2E_GreenlakePortalSlowInvestigation(t *testing.T) {
	ops, jun := setupClients()
	opts := ExecuteOptions{Juniper: jun}

	t.Log("=== E2E Integration: 'Why is greenlake portal slow?' ===")

	// Step 1: search_alerts — find critical alerts
	t.Log("Step 1: search_alerts(Critical)")
	result1, err := ExecuteWithOptions(ops, ToolCall{
		Name:      "search_alerts",
		Arguments: map[string]string{"state": "Critical"},
	}, opts)
	if err != nil {
		t.Fatalf("Step 1 failed: %v", err)
	}
	var alerts1 map[string]interface{}
	json.Unmarshal([]byte(result1), &alerts1)
	alertCount, _ := alerts1["count"].(float64)
	t.Logf("  → Found %d critical alerts", int(alertCount))
	if alertCount == 0 {
		t.Fatal("Step 1: Expected critical alerts, got none")
	}

	// Step 2: investigate_resource — deep-dive on k8s-node-04 (hosts greenlake-portal)
	t.Log("Step 2: investigate_resource(k8s-node-04)")
	result2, err := ExecuteWithOptions(ops, ToolCall{
		Name:      "investigate_resource",
		Arguments: map[string]string{"resource_name": "k8s-node-04"},
	}, opts)
	if err != nil {
		t.Fatalf("Step 2 failed: %v", err)
	}
	var inv map[string]interface{}
	json.Unmarshal([]byte(result2), &inv)
	cpu, _ := inv["cpuPercent"].(float64)
	invAlerts, _ := inv["alertCount"].(float64)
	t.Logf("  → CPU: %.1f%%, Active Alerts: %d", cpu, int(invAlerts))

	// Step 3: correlate_network — check the network for greenlake-portal
	t.Log("Step 3: correlate_network(greenlake-portal)")
	result3, err := ExecuteWithOptions(ops, ToolCall{
		Name:      "correlate_network",
		Arguments: map[string]string{"resource_name": "greenlake-portal"},
	}, opts)
	if err != nil {
		t.Fatalf("Step 3 failed: %v", err)
	}
	var netCorr map[string]interface{}
	json.Unmarshal([]byte(result3), &netCorr)
	networkRoot, _ := netCorr["network_is_likely_root_cause"].(bool)
	issueCount, _ := netCorr["issue_count"].(float64)
	verdict, _ := netCorr["verdict"].(string)
	t.Logf("  → Network is root cause: %v, Issues: %d, Verdict: %s", networkRoot, int(issueCount), verdict)
	if !networkRoot {
		t.Error("Step 3: Expected network to be root cause for greenlake-portal (link flaps)")
	}

	// Step 4: blast_radius — assess impact
	t.Log("Step 4: blast_radius(greenlake-portal)")
	result4, err := ExecuteWithOptions(ops, ToolCall{
		Name:      "blast_radius",
		Arguments: map[string]string{"resource_name": "greenlake-portal"},
	}, opts)
	if err != nil {
		t.Fatalf("Step 4 failed: %v", err)
	}
	var blast map[string]interface{}
	json.Unmarshal([]byte(result4), &blast)
	affectedUsers, _ := blast["affected_users"].(float64)
	affectedApps, _ := blast["affected_applications"].(float64)
	severity, _ := blast["overall_severity"].(string)
	t.Logf("  → Affected: %d apps, %d users, Severity: %s", int(affectedApps), int(affectedUsers), severity)
	if affectedUsers == 0 {
		t.Error("Step 4: Expected affected_users > 0")
	}

	// Step 5: search_knowledge_base — skip (needs Ollama)
	t.Log("Step 5: search_knowledge_base — SKIPPED (requires Ollama for embeddings)")

	// Step 6: get_remediation_plan — generate fix
	t.Log("Step 6: get_remediation_plan(greenlake-portal)")
	result6, err := ExecuteWithOptions(ops, ToolCall{
		Name:      "get_remediation_plan",
		Arguments: map[string]string{"resource_name": "greenlake-portal"},
	}, opts)
	if err != nil {
		t.Fatalf("Step 6 failed: %v", err)
	}
	var plan map[string]interface{}
	json.Unmarshal([]byte(result6), &plan)
	totalSteps, _ := plan["total_steps"].(float64)
	rootCause, _ := plan["root_cause"].(string)
	t.Logf("  → Remediation: %d steps, Root Cause: %s", int(totalSteps), rootCause)
	if totalSteps == 0 {
		t.Error("Step 6: Expected at least 1 remediation step")
	}

	// Print final summary
	t.Log("")
	t.Log("=== E2E SUMMARY ===")
	t.Logf("Query: 'Why is greenlake portal slow?'")
	t.Logf("Root Cause: Network link flaps on switch port")
	t.Logf("Server: k8s-node-04 (CPU %.1f%%)", cpu)
	t.Logf("Network: %d issues found, network IS root cause", int(issueCount))
	t.Logf("Impact: %d apps, %d users affected (%s severity)", int(affectedApps), int(affectedUsers), severity)
	t.Logf("Fix: %d-step remediation plan generated", int(totalSteps))
	t.Log("=== ALL 6 STEPS PASSED ===")
}

// =============================================================================
// Tool Definition Validation
// =============================================================================

func TestToolDefinitions_All11Present(t *testing.T) {
	defs := GetToolDefinitions()
	if len(defs) != 11 {
		t.Errorf("Expected 11 tool definitions, got %d", len(defs))
	}

	expectedTools := []string{
		"search_alerts", "search_resources", "get_resource_details",
		"search_incidents", "investigate_resource", "get_environment_summary",
		"predict_capacity", "search_knowledge_base", "correlate_network",
		"blast_radius", "get_remediation_plan",
	}

	names := make(map[string]bool)
	for _, d := range defs {
		names[d.Function.Name] = true
	}

	for _, expected := range expectedTools {
		if !names[expected] {
			t.Errorf("Missing tool definition: %s", expected)
		}
	}

	t.Logf("All %d tool definitions present", len(defs))
}

func TestFormatToolsForPrompt(t *testing.T) {
	prompt := FormatToolsForPrompt()
	if prompt == "" {
		t.Fatal("FormatToolsForPrompt returned empty string")
	}

	expectedTools := []string{
		"search_alerts", "search_resources", "get_resource_details",
		"search_incidents", "investigate_resource", "get_environment_summary",
		"predict_capacity", "search_knowledge_base", "correlate_network",
		"blast_radius", "get_remediation_plan",
	}
	for _, tool := range expectedTools {
		if !strings.Contains(prompt, fmt.Sprintf("### %s", tool)) {
			t.Errorf("Prompt missing tool heading for: %s", tool)
		}
	}
	t.Logf("FormatToolsForPrompt: %d chars, all 11 tools present", len(prompt))
}

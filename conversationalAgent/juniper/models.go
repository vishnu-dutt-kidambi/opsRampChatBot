package juniper

// =============================================================================
// Juniper Mist API Data Models — Network Switch Telemetry
// =============================================================================
//
// These structs mirror the real Juniper Mist API v1 response schemas from
// https://api.mist.com/api/v1/sites/{site_id}/stats/devices
//
// Field names, JSON tags, and enums are aligned with the actual Mist API so
// swapping from mock data to real API calls requires minimal changes.
//
// Schema sources (Mist API / mistapi-go):
//
//   Switch Stats:     GET /api/v1/sites/{site_id}/stats/devices?type=switch
//   Switch Port Stats: Embedded in switch stats as "ports" array
//   Interface Stats:  Embedded in switch stats as "if_stat" map
//
// References:
//   - https://github.com/tmunzer/mistapi-go/blob/main/doc/models/stats-switch.md
//   - https://github.com/tmunzer/mistapi-go/blob/main/doc/models/stats-switch-port.md
//   - https://github.com/tmunzer/mistapi-go/blob/main/doc/models/if-stat-property.md
// =============================================================================

// SwitchStats represents telemetry for a Juniper Mist-managed network switch.
// API: GET /api/v1/sites/{site_id}/stats/devices?type=switch
type SwitchStats struct {
	ID       string `json:"id"`
	OrgID    string `json:"org_id"`
	SiteID   string `json:"site_id"`
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	Mac      string `json:"mac"`
	Model    string `json:"model"`   // e.g., "EX4600", "EX4300-48T"
	Serial   string `json:"serial"`  // e.g., "TC3714190003"
	Version  string `json:"version"` // Junos version, e.g., "21.4R3-S5"
	Status   string `json:"status"`  // "connected", "disconnected"
	Uptime   int64  `json:"uptime"`  // seconds
	Type     string `json:"type"`    // always "switch"
	LastSeen int64  `json:"last_seen"`

	// Switch resource utilization
	CpuStat    CpuStat    `json:"cpu_stat,omitempty"`
	MemoryStat MemoryStat `json:"memory_stat,omitempty"`

	// Port-level telemetry (the key data for network correlation)
	Ports []SwitchPort `json:"ports"`

	// Interface-level stats (keyed by interface name, e.g., "ge-0/0/0")
	IfStat map[string]IfStatProperty `json:"if_stat,omitempty"`

	// Location / topology context
	Location    string `json:"location,omitempty"`
	Description string `json:"description,omitempty"`
}

// CpuStat represents switch CPU utilization.
// Mist API: cpu_stat object within switch stats.
type CpuStat struct {
	Idle   float64 `json:"idle"`
	System float64 `json:"system"`
	User   float64 `json:"user"`
}

// MemoryStat represents switch memory utilization.
// Mist API: memory_stat object within switch stats.
type MemoryStat struct {
	Usage float64 `json:"usage"` // percentage 0-100
}

// SwitchPort represents per-port telemetry on a Juniper switch.
// API: Embedded in switch stats as the "ports" array.
// Schema: https://github.com/tmunzer/mistapi-go/blob/main/doc/models/stats-switch-port.md
type SwitchPort struct {
	PortID    string `json:"port_id"`    // e.g., "ge-0/0/0", "xe-0/0/0"
	PortMac   string `json:"port_mac"`   // MAC address of the port itself
	PortUsage string `json:"port_usage"` // "lan", "uplink", "trunk", etc.

	// Link state
	Active     bool `json:"active"`
	Up         bool `json:"up"`
	Disabled   bool `json:"disabled"`
	FullDuplex bool `json:"full_duplex"`
	Speed      int  `json:"speed"` // Mbps (e.g., 1000, 10000)

	// Traffic counters (cumulative since link up)
	RxBytes     int64 `json:"rx_bytes"`
	RxPkts      int64 `json:"rx_pkts"`
	RxBps       int64 `json:"rx_bps"` // current rate, bits/sec
	RxErrors    int   `json:"rx_errors"`
	RxBcastPkts int   `json:"rx_bcast_pkts"`
	RxMcastPkts int   `json:"rx_mcast_pkts"`
	TxBytes     int64 `json:"tx_bytes"`
	TxPkts      int64 `json:"tx_pkts"`
	TxBps       int64 `json:"tx_bps"` // current rate, bits/sec
	TxErrors    int   `json:"tx_errors"`
	TxBcastPkts int   `json:"tx_bcast_pkts"`
	TxMcastPkts int   `json:"tx_mcast_pkts"`

	// Quality metrics (key for network correlation)
	Jitter  float64 `json:"jitter"`  // ms
	Latency float64 `json:"latency"` // ms
	Loss    float64 `json:"loss"`    // packet loss percentage 0-100

	// Flap detection
	LastFlapped float64 `json:"last_flapped"` // epoch timestamp of last link flap

	// MAC table
	MacCount int `json:"mac_count"`
	MacLimit int `json:"mac_limit"`

	// LLDP / CDP neighbor discovery (maps port → connected device)
	NeighborMac        string `json:"neighbor_mac"`
	NeighborPortDesc   string `json:"neighbor_port_desc"`
	NeighborSystemName string `json:"neighbor_system_name"`

	// STP state
	StpRole  string `json:"stp_role"`  // "designated", "root", "alternate", etc.
	StpState string `json:"stp_state"` // "forwarding", "blocking", "learning", etc.

	// Authentication state
	AuthState string `json:"auth_state,omitempty"` // "authenticated", "init", etc.

	// Transceiver info
	XcvrModel      string `json:"xcvr_model,omitempty"`
	XcvrPartNumber string `json:"xcvr_part_number,omitempty"`
	XcvrSerial     string `json:"xcvr_serial,omitempty"`
}

// IfStatProperty represents interface-level stats from the Mist API.
// API: Embedded in switch stats as "if_stat" map keyed by interface name.
type IfStatProperty struct {
	PortID      string   `json:"port_id"`
	PortUsage   string   `json:"port_usage"`
	NetworkName string   `json:"network_name,omitempty"`
	AddressMode string   `json:"address_mode,omitempty"`
	Ips         []string `json:"ips,omitempty"`
	RxBytes     int64    `json:"rx_bytes"`
	RxPkts      int64    `json:"rx_pkts"`
	TxBytes     int64    `json:"tx_bytes"`
	TxPkts      int64    `json:"tx_pkts"`
	Up          bool     `json:"up"`
	Vlan        int      `json:"vlan,omitempty"`
}

// =============================================================================
// Network-to-Server Port Mapping
// =============================================================================
//
// This mapping connects OpsRamp-monitored servers to their physical switch ports.
// In production, this would come from LLDP/CDP neighbor discovery, DCIM (e.g.,
// NetBox), or the Mist API's neighbor_system_name field on switch ports.
//
// For the mock, we explicitly map resource → switch port.
// =============================================================================

// PortMapping connects an OpsRamp resource (server) to a Juniper switch port.
type PortMapping struct {
	ResourceID   string `json:"resource_id"`
	ResourceName string `json:"resource_name"`
	ResourceIP   string `json:"resource_ip"`
	SwitchID     string `json:"switch_id"`
	SwitchName   string `json:"switch_name"`
	PortID       string `json:"port_id"` // e.g., "ge-0/0/1"
}

// =============================================================================
// Network Correlation Result
// =============================================================================

// NetworkCorrelation is the result of correlating a server issue with network telemetry.
type NetworkCorrelation struct {
	// Server context
	ResourceID   string `json:"resource_id"`
	ResourceName string `json:"resource_name"`
	ResourceIP   string `json:"resource_ip"`

	// Switch context
	SwitchID    string `json:"switch_id"`
	SwitchName  string `json:"switch_name"`
	SwitchModel string `json:"switch_model"`
	SwitchIP    string `json:"switch_ip"`

	// Port telemetry
	PortID       string  `json:"port_id"`
	PortStatus   string  `json:"port_status"` // "up", "down", "disabled"
	Speed        int     `json:"speed_mbps"`
	FullDuplex   bool    `json:"full_duplex"`
	RxErrors     int     `json:"rx_errors"`
	TxErrors     int     `json:"tx_errors"`
	PacketLoss   float64 `json:"packet_loss_pct"`
	Jitter       float64 `json:"jitter_ms"`
	Latency      float64 `json:"latency_ms"`
	RxBps        int64   `json:"rx_bps"`
	TxBps        int64   `json:"tx_bps"`
	LastFlapped  string  `json:"last_flapped"` // human-readable time
	FlappedEpoch float64 `json:"last_flapped_epoch,omitempty"`

	// Analysis
	Issues         []NetworkIssue `json:"issues"`
	IssueCount     int            `json:"issue_count"`
	NetworkIsRoot  bool           `json:"network_is_likely_root_cause"`
	Verdict        string         `json:"verdict"`
	Recommendation string         `json:"recommendation,omitempty"`
}

// NetworkIssue describes a detected network problem on the port.
type NetworkIssue struct {
	Type        string `json:"type"`     // "packet_loss", "rx_errors", "tx_errors", "link_flap", "half_duplex", "port_down"
	Severity    string `json:"severity"` // "critical", "warning", "info"
	Description string `json:"description"`
	Value       string `json:"value"`
	Threshold   string `json:"threshold"`
}

// =============================================================================
// Blast Radius Analysis — Impact & Dependency Mapping
// =============================================================================
//
// When a network issue is detected on a switch port, Blast Radius analysis
// maps the full scope of impact by traversing a dependency graph:
//   Switch Port → Server → Applications → Users
//
// This answers: "What's affected?" — how many applications, services, and
// end users are impacted by an infrastructure issue.
// =============================================================================

// DependencyNode represents a node in the infrastructure dependency graph.
// Nodes can be: switch, server, application, service, or user_group.
type DependencyNode struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`  // "switch", "server", "application", "service", "user_group"
	Layer       string            `json:"layer"` // "network", "compute", "application", "user"
	Cloud       string            `json:"cloud,omitempty"`
	Region      string            `json:"region,omitempty"`
	Criticality string            `json:"criticality"` // "critical", "high", "medium", "low"
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// DependencyEdge represents a directed relationship between two nodes.
type DependencyEdge struct {
	FromID       string `json:"from_id"`
	ToID         string `json:"to_id"`
	Relationship string `json:"relationship"` // "connects_to", "hosts", "runs_on", "depends_on", "serves"
	Description  string `json:"description,omitempty"`
}

// BlastRadiusResult is the output of a blast radius analysis.
type BlastRadiusResult struct {
	// Root cause
	RootCauseID   string `json:"root_cause_id"`
	RootCauseName string `json:"root_cause_name"`
	RootCauseType string `json:"root_cause_type"` // "switch_port", "server", "switch"
	RootCauseDesc string `json:"root_cause_description"`

	// Impact summary
	AffectedApplications int `json:"affected_applications"`
	AffectedServers      int `json:"affected_servers"`
	AffectedServices     int `json:"affected_services"`
	AffectedUsers        int `json:"affected_users"`
	TotalImpactedNodes   int `json:"total_impacted_nodes"`

	// Severity assessment
	OverallSeverity string `json:"overall_severity"` // "critical", "high", "medium", "low"
	BusinessImpact  string `json:"business_impact"`  // human-readable business impact statement

	// Detailed impact per layer
	ImpactedNodes []ImpactedNode `json:"impacted_nodes"`

	// Dependency path (root cause → most critical affected component)
	CriticalPath []PathStep `json:"critical_path"`
}

// ImpactedNode is a node affected by the blast radius.
type ImpactedNode struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Layer       string `json:"layer"`
	Criticality string `json:"criticality"`
	Impact      string `json:"impact"` // "direct", "indirect"
	Reason      string `json:"reason"` // why this node is affected
}

// PathStep is one step in the critical dependency path.
type PathStep struct {
	NodeID       string `json:"node_id"`
	NodeName     string `json:"node_name"`
	NodeType     string `json:"node_type"`
	Relationship string `json:"relationship"` // edge type to next step
}

// =============================================================================
// Guided Remediation — Actionable Fix Proposals
// =============================================================================
//
// After identifying root cause (network correlation) and blast radius,
// Guided Remediation generates specific, actionable fix steps that the agent
// proposes to the operator for approval before execution.
//
// Steps are ordered by priority and each includes:
//   - The command or action to take
//   - Expected outcome
//   - Risk level
//   - Whether it requires approval (destructive actions always do)
// =============================================================================

// RemediationPlan is the output of the guided remediation tool.
type RemediationPlan struct {
	// Context
	ResourceName string `json:"resource_name"`
	ResourceIP   string `json:"resource_ip"`
	SwitchName   string `json:"switch_name"`
	PortID       string `json:"port_id"`
	RootCause    string `json:"root_cause"` // summary of the detected issue

	// Plan details
	PlanID     string            `json:"plan_id"`
	Title      string            `json:"title"`
	Urgency    string            `json:"urgency"` // "immediate", "urgent", "scheduled"
	Steps      []RemediationStep `json:"steps"`
	TotalSteps int               `json:"total_steps"`

	// Impact assessment
	EstimatedDowntime string `json:"estimated_downtime"`
	RiskLevel         string `json:"risk_level"` // "low", "medium", "high"
	RequiresApproval  bool   `json:"requires_approval"`
	ApprovalNote      string `json:"approval_note,omitempty"`

	// Rollback
	RollbackAvailable bool   `json:"rollback_available"`
	RollbackPlan      string `json:"rollback_plan,omitempty"`
}

// RemediationStep is a single actionable step in a remediation plan.
type RemediationStep struct {
	StepNumber       int    `json:"step_number"`
	Action           string `json:"action"`            // human-readable description
	Command          string `json:"command,omitempty"` // actual CLI command (Junos, Linux, etc.)
	Target           string `json:"target"`            // device/server where action is performed
	Category         string `json:"category"`          // "diagnostic", "mitigation", "resolution", "verification"
	ExpectedOutcome  string `json:"expected_outcome"`
	RiskLevel        string `json:"risk_level"` // "none", "low", "medium", "high"
	RequiresApproval bool   `json:"requires_approval"`
	EstimatedTime    string `json:"estimated_time"` // e.g., "30 seconds", "2 minutes"
}

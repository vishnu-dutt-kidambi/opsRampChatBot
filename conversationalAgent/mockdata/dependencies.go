package mockdata

import (
	"opsramp-agent/juniper"
)

// =============================================================================
// Infrastructure Dependency Graph — Mock Data
// =============================================================================
//
// This dependency graph maps the relationships between:
//   Network switches → Servers → Applications → Services → User groups
//
// The graph enables Blast Radius analysis — when a network issue is detected
// on a switch port, we traverse the graph upstream to find all affected
// applications, services, and users.
//
// Dependency Layers:
//   1. Network  — Juniper switches and ports
//   2. Compute  — Servers, VMs, Kubernetes nodes
//   3. Application — Applications and microservices running on compute
//   4. User     — User groups / customer segments consuming applications
//
// Relationships:
//   switch  --connects_to-->  server
//   server  --hosts-->        application
//   application --depends_on-->  application  (inter-service dependencies)
//   application --serves-->    user_group
//
// Key Scenario (GreenLake Portal Slowdown):
//   sw-dc-east-04:ge-0/0/5 (link flaps, packet loss)
//     └─ k8s-node-04 (on-prem HPE ProLiant server)
//         ├─ greenlake-portal (critical app)
//         │   └─ depends_on: db-primary-01:postgresql
//         │   └─ depends_on: redis-cache-01:redis-sessions
//         │   └─ serves: greenlake-tenants (3000 users)
//         │   └─ serves: aruba-wifi-users (2000 users)
//         └─ aruba-central (high app)
//             └─ depends_on: greenlake-portal
//             └─ depends_on: rabbitmq-prod-01:rabbitmq
//             └─ serves: greenlake-tenants
//
//   Total blast radius from ge-0/0/5 issues:
//     3 applications (greenlake-portal, aruba-central, dscc-console)
//     5,000 users (3000 greenlake-tenants + 2000 aruba-wifi-users)
// =============================================================================

// GetDependencyNodes returns all nodes in the infrastructure dependency graph.
func GetDependencyNodes() []juniper.DependencyNode {
	return []juniper.DependencyNode{
		// ── Network Layer (Switches) ────────────────────────────────────
		{
			ID: "sw-001", Name: "sw-dc-east-01", Type: "switch",
			Layer: "network", Criticality: "critical",
			Metadata: map[string]string{"model": "EX4600", "site": "dc-east", "role": "web-app-tor"},
		},
		{
			ID: "sw-002", Name: "sw-dc-east-02", Type: "switch",
			Layer: "network", Criticality: "critical",
			Metadata: map[string]string{"model": "EX4600", "site": "dc-east", "role": "database-tor"},
		},
		{
			ID: "sw-003", Name: "sw-dc-east-03", Type: "switch",
			Layer: "network", Criticality: "high",
			Metadata: map[string]string{"model": "EX4300-48T", "site": "dc-east", "role": "infra-tor"},
		},
		{
			ID: "sw-004", Name: "sw-dc-east-04", Type: "switch",
			Layer: "network", Criticality: "critical",
			Metadata: map[string]string{"model": "EX4300-48T", "site": "dc-east", "role": "k8s-greenlake-tor"},
		},

		// ── Compute Layer (Servers) ─────────────────────────────────────
		{
			ID: "res-001", Name: "web-server-prod-01", Type: "server",
			Layer: "compute", Cloud: "AWS", Region: "us-east-1", Criticality: "critical",
			Metadata: map[string]string{"role": "web", "tier": "frontend"},
		},
		{
			ID: "res-002", Name: "app-server-prod-01", Type: "server",
			Layer: "compute", Cloud: "AWS", Region: "us-east-1", Criticality: "high",
			Metadata: map[string]string{"role": "app", "tier": "backend"},
		},
		{
			ID: "res-003", Name: "app-server-prod-02", Type: "server",
			Layer: "compute", Cloud: "AWS", Region: "us-east-1", Criticality: "high",
			Metadata: map[string]string{"role": "app", "tier": "backend"},
		},
		{
			ID: "res-004", Name: "web-server-prod-02", Type: "server",
			Layer: "compute", Cloud: "AWS", Region: "us-east-1", Criticality: "high",
			Metadata: map[string]string{"role": "web", "tier": "frontend"},
		},
		{
			ID: "res-005", Name: "db-primary-01", Type: "server",
			Layer: "compute", Cloud: "AWS", Region: "us-east-1", Criticality: "critical",
			Metadata: map[string]string{"role": "database", "engine": "postgresql"},
		},
		{
			ID: "res-006", Name: "db-replica-01", Type: "server",
			Layer: "compute", Cloud: "AWS", Region: "us-east-1", Criticality: "high",
			Metadata: map[string]string{"role": "database", "engine": "postgresql-replica"},
		},
		{
			ID: "res-007", Name: "redis-cache-01", Type: "server",
			Layer: "compute", Cloud: "AWS", Region: "us-east-1", Criticality: "high",
			Metadata: map[string]string{"role": "cache", "engine": "redis"},
		},
		{
			ID: "res-008", Name: "rabbitmq-prod-01", Type: "server",
			Layer: "compute", Cloud: "AWS", Region: "us-east-1", Criticality: "high",
			Metadata: map[string]string{"role": "messaging", "engine": "rabbitmq"},
		},
		{
			ID: "res-009", Name: "elasticsearch-prod-01", Type: "server",
			Layer: "compute", Cloud: "AWS", Region: "us-east-1", Criticality: "medium",
			Metadata: map[string]string{"role": "search", "engine": "elasticsearch"},
		},
		{
			ID: "res-010", Name: "api-gateway-prod", Type: "server",
			Layer: "compute", Cloud: "AWS", Region: "us-east-1", Criticality: "critical",
			Metadata: map[string]string{"role": "gateway", "tier": "edge"},
		},
		{
			ID: "res-011", Name: "k8s-master-01", Type: "server",
			Layer: "compute", Cloud: "HPE GreenLake", Region: "hpe-dc-east", Criticality: "critical",
			Metadata: map[string]string{"role": "k8s-master", "hardware": "HPE ProLiant DL360"},
		},
		{
			ID: "res-013", Name: "k8s-node-01", Type: "server",
			Layer: "compute", Cloud: "HPE GreenLake", Region: "hpe-dc-east", Criticality: "high",
			Metadata: map[string]string{"role": "k8s-worker", "hardware": "HPE ProLiant DL380"},
		},
		{
			ID: "res-014", Name: "k8s-node-02", Type: "server",
			Layer: "compute", Cloud: "HPE GreenLake", Region: "hpe-dc-east", Criticality: "high",
			Metadata: map[string]string{"role": "k8s-worker", "hardware": "HPE ProLiant DL380"},
		},
		{
			ID: "res-015", Name: "k8s-node-03", Type: "server",
			Layer: "compute", Cloud: "HPE GreenLake", Region: "hpe-dc-east", Criticality: "high",
			Metadata: map[string]string{"role": "k8s-worker", "hardware": "HPE ProLiant DL380"},
		},
		{
			ID: "res-016", Name: "k8s-node-04", Type: "server",
			Layer: "compute", Cloud: "HPE GreenLake", Region: "hpe-dc-east", Criticality: "critical",
			Metadata: map[string]string{"role": "k8s-worker", "hardware": "HPE ProLiant DL380", "note": "hosts greenlake-portal pods"},
		},

		// ── Application Layer ───────────────────────────────────────────
		{
			ID: "app-greenlake", Name: "greenlake-portal", Type: "application",
			Layer: "application", Criticality: "critical",
			Metadata: map[string]string{
				"port": "8443", "protocol": "HTTPS",
				"team": "greenlake", "pagerduty": "greenlake-oncall",
				"description": "HPE GreenLake cloud platform portal — tenant provisioning, resource management, and billing",
			},
		},
		{
			ID: "app-aruba", Name: "aruba-central", Type: "application",
			Layer: "application", Criticality: "critical",
			Metadata: map[string]string{
				"port": "8080", "protocol": "HTTP",
				"team": "networking", "pagerduty": "aruba-oncall",
				"description": "Aruba Central — network management, AP monitoring, SD-WAN orchestration",
			},
		},
		{
			ID: "app-dscc", Name: "dscc-console", Type: "application",
			Layer: "application", Criticality: "critical",
			Metadata: map[string]string{
				"port": "3000", "protocol": "HTTPS",
				"team": "storage", "pagerduty": "storage-oncall",
				"description": "HPE Data Services Cloud Console — storage provisioning, data protection, and cloud volumes",
			},
		},
		{
			ID: "app-oneview", Name: "oneview-api", Type: "application",
			Layer: "application", Criticality: "high",
			Metadata: map[string]string{
				"port": "8081", "protocol": "HTTP",
				"team":        "infrastructure",
				"description": "HPE OneView — server hardware management, firmware updates, and template deployment",
			},
		},
		{
			ID: "app-user-auth", Name: "user-auth-service", Type: "application",
			Layer: "application", Criticality: "critical",
			Metadata: map[string]string{
				"port": "8443", "protocol": "HTTPS",
				"team":        "platform",
				"description": "User authentication, session management, and OAuth flows",
			},
		},
		{
			ID: "app-notification", Name: "notification-service", Type: "application",
			Layer: "application", Criticality: "medium",
			Metadata: map[string]string{
				"port": "8082", "protocol": "HTTP",
				"team":        "platform",
				"description": "Email, SMS, and push notifications for order status and alerts",
			},
		},
		{
			ID: "app-search", Name: "search-service", Type: "application",
			Layer: "application", Criticality: "high",
			Metadata: map[string]string{
				"port": "9200", "protocol": "HTTP",
				"team":        "infrastructure",
				"description": "Full-text search across infrastructure inventory powered by Elasticsearch",
			},
		},
		{
			ID: "app-api-gw", Name: "api-gateway", Type: "application",
			Layer: "application", Criticality: "critical",
			Metadata: map[string]string{
				"port": "443", "protocol": "HTTPS",
				"team":        "platform",
				"description": "Edge gateway — rate limiting, auth, routing for all public APIs",
			},
		},
		{
			ID: "app-web-frontend", Name: "web-frontend", Type: "application",
			Layer: "application", Criticality: "critical",
			Metadata: map[string]string{
				"port": "443", "protocol": "HTTPS",
				"team":        "frontend",
				"description": "Main HPE management console — dashboard, resource browsing, account management",
			},
		},
		{
			ID: "app-monitoring", Name: "monitoring-stack", Type: "application",
			Layer: "application", Criticality: "high",
			Metadata: map[string]string{
				"port": "9090", "protocol": "HTTP",
				"team":        "sre",
				"description": "Prometheus + Grafana monitoring and alerting stack",
			},
		},

		// ── User Layer (User Groups / Customer Segments) ────────────────
		{
			ID: "ug-greenlake-tenants", Name: "greenlake-tenants", Type: "user_group",
			Layer: "user", Criticality: "critical",
			Metadata: map[string]string{
				"estimated_users": "3000",
				"description":     "HPE GreenLake cloud tenants managing compute, storage, and networking resources",
				"revenue_impact":  "$45,000/hour",
			},
		},
		{
			ID: "ug-aruba-wifi-users", Name: "aruba-wifi-users", Type: "user_group",
			Layer: "user", Criticality: "critical",
			Metadata: map[string]string{
				"estimated_users": "2000",
				"description":     "Aruba wireless network users — enterprise Wi-Fi, SD-WAN, and branch connectivity",
				"revenue_impact":  "$30,000/hour",
			},
		},
		{
			ID: "ug-api-integrations", Name: "api-integrations", Type: "user_group",
			Layer: "user", Criticality: "high",
			Metadata: map[string]string{
				"estimated_users": "150",
				"description":     "Third-party API integrations — MSP partners, ITSM connectors, and automation workflows",
				"revenue_impact":  "$12,000/hour",
			},
		},
		{
			ID: "ug-hpe-ops-team", Name: "hpe-ops-team", Type: "user_group",
			Layer: "user", Criticality: "medium",
			Metadata: map[string]string{
				"estimated_users": "50",
				"description":     "HPE internal operations staff — SRE, platform engineering, and support",
			},
		},
	}
}

// GetDependencyEdges returns all edges (relationships) in the dependency graph.
func GetDependencyEdges() []juniper.DependencyEdge {
	return []juniper.DependencyEdge{
		// ── Network → Compute (switch connects_to server) ───────────────
		// sw-dc-east-01 (Web/App Tier)
		{FromID: "sw-001", ToID: "res-001", Relationship: "connects_to", Description: "ge-0/0/1 → web-server-prod-01"},
		{FromID: "sw-001", ToID: "res-002", Relationship: "connects_to", Description: "ge-0/0/2 → app-server-prod-01"},
		{FromID: "sw-001", ToID: "res-003", Relationship: "connects_to", Description: "ge-0/0/3 → app-server-prod-02"},
		{FromID: "sw-001", ToID: "res-004", Relationship: "connects_to", Description: "ge-0/0/4 → web-server-prod-02"},

		// sw-dc-east-02 (Database Tier)
		{FromID: "sw-002", ToID: "res-005", Relationship: "connects_to", Description: "ge-0/0/1 → db-primary-01"},
		{FromID: "sw-002", ToID: "res-006", Relationship: "connects_to", Description: "ge-0/0/2 → db-replica-01"},
		{FromID: "sw-002", ToID: "res-007", Relationship: "connects_to", Description: "ge-0/0/3 → redis-cache-01"},

		// sw-dc-east-03 (Infrastructure)
		{FromID: "sw-003", ToID: "res-008", Relationship: "connects_to", Description: "ge-0/0/1 → rabbitmq-prod-01"},
		{FromID: "sw-003", ToID: "res-009", Relationship: "connects_to", Description: "ge-0/0/2 → elasticsearch-prod-01"},
		{FromID: "sw-003", ToID: "res-010", Relationship: "connects_to", Description: "ge-0/0/3 → api-gateway-prod"},

		// sw-dc-east-04 (On-Prem K8s)
		{FromID: "sw-004", ToID: "res-011", Relationship: "connects_to", Description: "ge-0/0/1 → k8s-master-01"},
		{FromID: "sw-004", ToID: "res-013", Relationship: "connects_to", Description: "ge-0/0/2 → k8s-node-01"},
		{FromID: "sw-004", ToID: "res-014", Relationship: "connects_to", Description: "ge-0/0/3 → k8s-node-02"},
		{FromID: "sw-004", ToID: "res-015", Relationship: "connects_to", Description: "ge-0/0/4 → k8s-node-03"},
		{FromID: "sw-004", ToID: "res-016", Relationship: "connects_to", Description: "ge-0/0/5 → k8s-node-04"},

		// ── Compute → Application (server hosts application) ────────────
		// k8s-node-04 hosts HPE cloud management services (KEY for demo scenario)
		{FromID: "res-016", ToID: "app-greenlake", Relationship: "hosts", Description: "k8s-node-04 runs greenlake-portal pods"},
		{FromID: "res-016", ToID: "app-aruba", Relationship: "hosts", Description: "k8s-node-04 runs aruba-central pods"},
		{FromID: "res-016", ToID: "app-dscc", Relationship: "hosts", Description: "k8s-node-04 runs dscc-console pods"},

		// k8s-node-03 hosts oneview & search
		{FromID: "res-015", ToID: "app-oneview", Relationship: "hosts", Description: "k8s-node-03 runs oneview-api pods"},
		{FromID: "res-015", ToID: "app-search", Relationship: "hosts", Description: "k8s-node-03 runs search-service pods"},

		// k8s-node-01 hosts auth & notification
		{FromID: "res-013", ToID: "app-user-auth", Relationship: "hosts", Description: "k8s-node-01 runs user-auth-service pods"},
		{FromID: "res-013", ToID: "app-notification", Relationship: "hosts", Description: "k8s-node-01 runs notification-service pods"},

		// k8s-node-02 hosts monitoring
		{FromID: "res-014", ToID: "app-monitoring", Relationship: "hosts", Description: "k8s-node-02 runs monitoring stack"},

		// Web servers host frontend
		{FromID: "res-001", ToID: "app-web-frontend", Relationship: "hosts", Description: "web-server-prod-01 serves web-frontend"},
		{FromID: "res-004", ToID: "app-web-frontend", Relationship: "hosts", Description: "web-server-prod-02 serves web-frontend (HA)"},

		// API gateway server hosts api-gateway app
		{FromID: "res-010", ToID: "app-api-gw", Relationship: "hosts", Description: "api-gateway-prod runs api-gateway"},

		// ── Application → Application (inter-service dependencies) ──────
		// GreenLake portal dependencies
		{FromID: "app-greenlake", ToID: "app-user-auth", Relationship: "depends_on", Description: "greenlake-portal validates auth tokens"},

		// Aruba Central dependencies
		{FromID: "app-aruba", ToID: "app-greenlake", Relationship: "depends_on", Description: "aruba-central calls greenlake-portal for tenant validation"},
		{FromID: "app-aruba", ToID: "app-notification", Relationship: "depends_on", Description: "aruba-central triggers network event notifications"},

		// DSCC Console dependencies
		{FromID: "app-dscc", ToID: "app-greenlake", Relationship: "depends_on", Description: "dscc-console authenticates via greenlake-portal"},
		{FromID: "app-dscc", ToID: "app-aruba", Relationship: "depends_on", Description: "dscc-console queries network topology from aruba-central"},
		{FromID: "app-dscc", ToID: "app-oneview", Relationship: "depends_on", Description: "dscc-console fetches server inventory from oneview-api"},

		// Web frontend dependencies
		{FromID: "app-web-frontend", ToID: "app-api-gw", Relationship: "depends_on", Description: "web-frontend routes all API calls through gateway"},

		// API gateway dependencies
		{FromID: "app-api-gw", ToID: "app-dscc", Relationship: "depends_on", Description: "api-gateway routes /storage to dscc-console"},
		{FromID: "app-api-gw", ToID: "app-oneview", Relationship: "depends_on", Description: "api-gateway routes /servers to oneview-api"},
		{FromID: "app-api-gw", ToID: "app-user-auth", Relationship: "depends_on", Description: "api-gateway validates all requests via user-auth"},

		// Search service dependencies
		{FromID: "app-search", ToID: "app-oneview", Relationship: "depends_on", Description: "search indexes infrastructure inventory data"},

		// ── Application → User Group (application serves users) ─────────
		// GreenLake portal consumers
		{FromID: "app-greenlake", ToID: "ug-greenlake-tenants", Relationship: "serves", Description: "greenlake-portal serves cloud tenant management"},
		{FromID: "app-greenlake", ToID: "ug-aruba-wifi-users", Relationship: "serves", Description: "greenlake-portal provides SSO for Aruba users"},

		// Aruba Central consumers
		{FromID: "app-aruba", ToID: "ug-greenlake-tenants", Relationship: "serves", Description: "aruba-central manages tenant network infrastructure"},
		{FromID: "app-aruba", ToID: "ug-aruba-wifi-users", Relationship: "serves", Description: "aruba-central provides Wi-Fi and SD-WAN management"},

		// DSCC Console consumers
		{FromID: "app-dscc", ToID: "ug-greenlake-tenants", Relationship: "serves", Description: "dscc-console provides storage management interface"},

		// Web frontend consumers
		{FromID: "app-web-frontend", ToID: "ug-greenlake-tenants", Relationship: "serves", Description: "web-frontend serves the main management console"},
		{FromID: "app-web-frontend", ToID: "ug-hpe-ops-team", Relationship: "serves", Description: "web-frontend includes internal ops dashboard"},

		// API gateway consumers
		{FromID: "app-api-gw", ToID: "ug-api-integrations", Relationship: "serves", Description: "api-gateway exposes public APIs to integration partners"},
		{FromID: "app-api-gw", ToID: "ug-aruba-wifi-users", Relationship: "serves", Description: "api-gateway serves Aruba network API calls"},

		// OneView consumers
		{FromID: "app-oneview", ToID: "ug-greenlake-tenants", Relationship: "serves", Description: "oneview-api provides server hardware management"},
		{FromID: "app-oneview", ToID: "ug-aruba-wifi-users", Relationship: "serves", Description: "oneview-api serves infrastructure inventory to network users"},

		// Monitoring consumers
		{FromID: "app-monitoring", ToID: "ug-hpe-ops-team", Relationship: "serves", Description: "monitoring-stack provides dashboards to HPE SRE/ops team"},
	}
}

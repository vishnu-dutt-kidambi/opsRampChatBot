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
// Key Scenario (Payment App Slowdown):
//   sw-gcp-central-01:ge-0/0/5 (link flaps, packet loss)
//     └─ k8s-node-04 (server)
//         ├─ payment-service (critical app)
//         │   └─ depends_on: db-primary-01:postgresql
//         │   └─ depends_on: redis-cache-01:redis-sessions
//         │   └─ serves: online-shoppers (3000 users)
//         │   └─ serves: mobile-app-users (2000 users)
//         └─ order-service (high app)
//             └─ depends_on: payment-service
//             └─ depends_on: rabbitmq-prod-01:rabbitmq
//             └─ serves: online-shoppers
//
//   Total blast radius from ge-0/0/5 issues:
//     3 applications (payment-service, order-service, checkout-ui)
//     5,000 users (3000 online-shoppers + 2000 mobile-app-users)
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
			ID: "sw-004", Name: "sw-gcp-central-01", Type: "switch",
			Layer: "network", Criticality: "critical",
			Metadata: map[string]string{"model": "EX4300-48T", "site": "gcp-central", "role": "k8s-interconnect"},
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
			Layer: "compute", Cloud: "GCP", Region: "us-central1", Criticality: "critical",
			Metadata: map[string]string{"role": "k8s-master"},
		},
		{
			ID: "res-013", Name: "k8s-node-01", Type: "server",
			Layer: "compute", Cloud: "GCP", Region: "us-central1", Criticality: "high",
			Metadata: map[string]string{"role": "k8s-worker"},
		},
		{
			ID: "res-014", Name: "k8s-node-02", Type: "server",
			Layer: "compute", Cloud: "GCP", Region: "us-central1", Criticality: "high",
			Metadata: map[string]string{"role": "k8s-worker"},
		},
		{
			ID: "res-015", Name: "k8s-node-03", Type: "server",
			Layer: "compute", Cloud: "GCP", Region: "us-central1", Criticality: "high",
			Metadata: map[string]string{"role": "k8s-worker"},
		},
		{
			ID: "res-016", Name: "k8s-node-04", Type: "server",
			Layer: "compute", Cloud: "GCP", Region: "us-central1", Criticality: "critical",
			Metadata: map[string]string{"role": "k8s-worker", "note": "hosts payment-service pods"},
		},

		// ── Application Layer ───────────────────────────────────────────
		{
			ID: "app-payment", Name: "payment-service", Type: "application",
			Layer: "application", Criticality: "critical",
			Metadata: map[string]string{
				"port": "8443", "protocol": "HTTPS",
				"team": "payments", "pagerduty": "payments-oncall",
				"description": "Handles all payment processing, credit card tokenization, and transaction settlements",
			},
		},
		{
			ID: "app-order", Name: "order-service", Type: "application",
			Layer: "application", Criticality: "critical",
			Metadata: map[string]string{
				"port": "8080", "protocol": "HTTP",
				"team": "commerce", "pagerduty": "commerce-oncall",
				"description": "Order lifecycle management — create, update, fulfill, cancel orders",
			},
		},
		{
			ID: "app-checkout", Name: "checkout-ui", Type: "application",
			Layer: "application", Criticality: "critical",
			Metadata: map[string]string{
				"port": "3000", "protocol": "HTTPS",
				"team": "frontend", "pagerduty": "frontend-oncall",
				"description": "Customer-facing checkout experience — cart, payment form, order confirmation",
			},
		},
		{
			ID: "app-catalog", Name: "product-catalog", Type: "application",
			Layer: "application", Criticality: "high",
			Metadata: map[string]string{
				"port": "8081", "protocol": "HTTP",
				"team":        "commerce",
				"description": "Product listings, search, and inventory management",
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
				"team":        "commerce",
				"description": "Full-text product search powered by Elasticsearch",
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
				"description": "Main website — product browsing, account management, order tracking",
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
			ID: "ug-online-shoppers", Name: "online-shoppers", Type: "user_group",
			Layer: "user", Criticality: "critical",
			Metadata: map[string]string{
				"estimated_users": "3000",
				"description":     "Web browser customers actively shopping and completing purchases",
				"revenue_impact":  "$45,000/hour",
			},
		},
		{
			ID: "ug-mobile-users", Name: "mobile-app-users", Type: "user_group",
			Layer: "user", Criticality: "critical",
			Metadata: map[string]string{
				"estimated_users": "2000",
				"description":     "iOS and Android app users — browsing, ordering, and tracking deliveries",
				"revenue_impact":  "$30,000/hour",
			},
		},
		{
			ID: "ug-api-partners", Name: "api-partners", Type: "user_group",
			Layer: "user", Criticality: "high",
			Metadata: map[string]string{
				"estimated_users": "150",
				"description":     "Third-party integrators using the public API (marketplace sellers, shipping providers)",
				"revenue_impact":  "$12,000/hour",
			},
		},
		{
			ID: "ug-internal-ops", Name: "internal-ops-team", Type: "user_group",
			Layer: "user", Criticality: "medium",
			Metadata: map[string]string{
				"estimated_users": "50",
				"description":     "Internal operations staff — order management, customer support, warehouse",
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

		// sw-gcp-central-01 (GCP K8s)
		{FromID: "sw-004", ToID: "res-011", Relationship: "connects_to", Description: "ge-0/0/1 → k8s-master-01"},
		{FromID: "sw-004", ToID: "res-013", Relationship: "connects_to", Description: "ge-0/0/2 → k8s-node-01"},
		{FromID: "sw-004", ToID: "res-014", Relationship: "connects_to", Description: "ge-0/0/3 → k8s-node-02"},
		{FromID: "sw-004", ToID: "res-015", Relationship: "connects_to", Description: "ge-0/0/4 → k8s-node-03"},
		{FromID: "sw-004", ToID: "res-016", Relationship: "connects_to", Description: "ge-0/0/5 → k8s-node-04"},

		// ── Compute → Application (server hosts application) ────────────
		// k8s-node-04 hosts payment-related services (KEY for demo scenario)
		{FromID: "res-016", ToID: "app-payment", Relationship: "hosts", Description: "k8s-node-04 runs payment-service pods"},
		{FromID: "res-016", ToID: "app-order", Relationship: "hosts", Description: "k8s-node-04 runs order-service pods"},
		{FromID: "res-016", ToID: "app-checkout", Relationship: "hosts", Description: "k8s-node-04 runs checkout-ui pods"},

		// k8s-node-03 hosts catalog & search
		{FromID: "res-015", ToID: "app-catalog", Relationship: "hosts", Description: "k8s-node-03 runs product-catalog pods"},
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
		// Payment service dependencies
		{FromID: "app-payment", ToID: "app-user-auth", Relationship: "depends_on", Description: "payment-service validates auth tokens"},

		// Order service dependencies
		{FromID: "app-order", ToID: "app-payment", Relationship: "depends_on", Description: "order-service calls payment-service for charge/refund"},
		{FromID: "app-order", ToID: "app-notification", Relationship: "depends_on", Description: "order-service triggers order confirmation notifications"},

		// Checkout UI dependencies
		{FromID: "app-checkout", ToID: "app-payment", Relationship: "depends_on", Description: "checkout-ui submits payments to payment-service"},
		{FromID: "app-checkout", ToID: "app-order", Relationship: "depends_on", Description: "checkout-ui creates orders via order-service"},
		{FromID: "app-checkout", ToID: "app-catalog", Relationship: "depends_on", Description: "checkout-ui fetches product details from catalog"},

		// Web frontend dependencies
		{FromID: "app-web-frontend", ToID: "app-api-gw", Relationship: "depends_on", Description: "web-frontend routes all API calls through gateway"},

		// API gateway dependencies
		{FromID: "app-api-gw", ToID: "app-checkout", Relationship: "depends_on", Description: "api-gateway routes /checkout to checkout-ui"},
		{FromID: "app-api-gw", ToID: "app-catalog", Relationship: "depends_on", Description: "api-gateway routes /products to product-catalog"},
		{FromID: "app-api-gw", ToID: "app-user-auth", Relationship: "depends_on", Description: "api-gateway validates all requests via user-auth"},

		// Search service dependencies
		{FromID: "app-search", ToID: "app-catalog", Relationship: "depends_on", Description: "search indexes product catalog data"},

		// ── Application → User Group (application serves users) ─────────
		// Payment service consumers
		{FromID: "app-payment", ToID: "ug-online-shoppers", Relationship: "serves", Description: "payment-service processes web checkout payments"},
		{FromID: "app-payment", ToID: "ug-mobile-users", Relationship: "serves", Description: "payment-service processes mobile app payments"},

		// Order service consumers
		{FromID: "app-order", ToID: "ug-online-shoppers", Relationship: "serves", Description: "order-service manages web orders"},
		{FromID: "app-order", ToID: "ug-mobile-users", Relationship: "serves", Description: "order-service manages mobile orders"},

		// Checkout UI consumers
		{FromID: "app-checkout", ToID: "ug-online-shoppers", Relationship: "serves", Description: "checkout-ui is the web checkout experience"},

		// Web frontend consumers
		{FromID: "app-web-frontend", ToID: "ug-online-shoppers", Relationship: "serves", Description: "web-frontend serves the main website"},
		{FromID: "app-web-frontend", ToID: "ug-internal-ops", Relationship: "serves", Description: "web-frontend includes internal ops dashboard"},

		// API gateway consumers
		{FromID: "app-api-gw", ToID: "ug-api-partners", Relationship: "serves", Description: "api-gateway exposes public APIs to partners"},
		{FromID: "app-api-gw", ToID: "ug-mobile-users", Relationship: "serves", Description: "api-gateway serves mobile app API calls"},

		// Catalog consumers
		{FromID: "app-catalog", ToID: "ug-online-shoppers", Relationship: "serves", Description: "product-catalog powers product browsing"},
		{FromID: "app-catalog", ToID: "ug-mobile-users", Relationship: "serves", Description: "product-catalog serves mobile product listings"},

		// Monitoring consumers
		{FromID: "app-monitoring", ToID: "ug-internal-ops", Relationship: "serves", Description: "monitoring-stack provides dashboards to SRE/ops team"},
	}
}

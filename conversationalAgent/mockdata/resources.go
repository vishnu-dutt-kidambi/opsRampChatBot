package mockdata

import (
	"opsramp-agent/opsramp"
)

// GetResources returns realistic multi-cloud infrastructure resource data.
func GetResources() []opsramp.Resource {
	return []opsramp.Resource{
		// ── AWS US-EAST-1 (Production Web/App Tier) ───────────────────────
		{
			ID: "res-001", Name: "web-server-prod-01",
			HostName: "web-server-prod-01", IPAddress: "10.0.1.101",
			Type: "Linux", State: "active", OSName: "Ubuntu 22.04",
			AgentInstalled: true, Status: "managed", ResourceType: "Linux",
			Cloud: "AWS", Region: "us-east-1", InstanceSize: "c5.2xlarge",
			Tags:    map[string]string{"env": "production", "team": "platform", "role": "web", "tier": "frontend"},
			Metrics: getWebServerMetrics(),
		},
		{
			ID: "res-002", Name: "app-server-prod-01",
			HostName: "app-server-prod-01", IPAddress: "10.0.1.102",
			Type: "Linux", State: "active", OSName: "Ubuntu 22.04",
			AgentInstalled: true, Status: "managed", ResourceType: "Linux",
			Cloud: "AWS", Region: "us-east-1", InstanceSize: "m5.xlarge",
			Tags:    map[string]string{"env": "production", "team": "platform", "role": "app", "tier": "backend"},
			Metrics: getHealthyAppMetrics(),
		},
		{
			ID: "res-003", Name: "app-server-prod-02",
			HostName: "app-server-prod-02", IPAddress: "10.0.1.103",
			Type: "Linux", State: "active", OSName: "Ubuntu 22.04",
			AgentInstalled: true, Status: "managed", ResourceType: "Linux",
			Cloud: "AWS", Region: "us-east-1", InstanceSize: "m5.xlarge",
			Tags:    map[string]string{"env": "production", "team": "platform", "role": "app", "tier": "backend"},
			Metrics: getHighMemAppMetrics(),
		},
		{
			ID: "res-004", Name: "web-server-prod-02",
			HostName: "web-server-prod-02", IPAddress: "10.0.1.104",
			Type: "Linux", State: "active", OSName: "Ubuntu 22.04",
			AgentInstalled: true, Status: "managed", ResourceType: "Linux",
			Cloud: "AWS", Region: "us-east-1", InstanceSize: "c5.2xlarge",
			Tags:    map[string]string{"env": "production", "team": "platform", "role": "web", "tier": "frontend"},
			Metrics: getHealthyWebMetrics(),
		},

		// ── AWS US-EAST-1 (Database Tier) ────────────────────────────────
		{
			ID: "res-005", Name: "db-primary-01",
			HostName: "db-primary-01", IPAddress: "10.0.2.10",
			Type: "Linux", State: "active", OSName: "Ubuntu 22.04",
			AgentInstalled: true, Status: "managed", ResourceType: "Linux",
			Cloud: "AWS", Region: "us-east-1", InstanceSize: "r6g.4xlarge",
			Tags:    map[string]string{"env": "production", "team": "data", "role": "database", "engine": "postgresql"},
			Metrics: getDbPrimaryMetrics(),
		},
		{
			ID: "res-006", Name: "db-replica-01",
			HostName: "db-replica-01", IPAddress: "10.0.2.11",
			Type: "Linux", State: "active", OSName: "Ubuntu 22.04",
			AgentInstalled: true, Status: "managed", ResourceType: "Linux",
			Cloud: "AWS", Region: "us-east-1", InstanceSize: "r6g.2xlarge",
			Tags:    map[string]string{"env": "production", "team": "data", "role": "database-replica", "engine": "postgresql"},
			Metrics: getDbReplicaMetrics(),
		},
		{
			ID: "res-007", Name: "redis-cache-01",
			HostName: "redis-cache-01", IPAddress: "10.0.2.20",
			Type: "Linux", State: "active", OSName: "Amazon Linux 2023",
			AgentInstalled: true, Status: "managed", ResourceType: "Linux",
			Cloud: "AWS", Region: "us-east-1", InstanceSize: "r6g.xlarge",
			Tags:    map[string]string{"env": "production", "team": "data", "role": "cache", "engine": "redis"},
			Metrics: getCacheMetrics(),
		},

		// ── AWS US-EAST-1 (Supporting Infrastructure) ────────────────────
		{
			ID: "res-008", Name: "rabbitmq-prod-01",
			HostName: "rabbitmq-prod-01", IPAddress: "10.0.3.10",
			Type: "Linux", State: "active", OSName: "Ubuntu 22.04",
			AgentInstalled: true, Status: "managed", ResourceType: "Linux",
			Cloud: "AWS", Region: "us-east-1", InstanceSize: "m5.large",
			Tags:    map[string]string{"env": "production", "team": "platform", "role": "message-queue"},
			Metrics: getQueueMetrics(),
		},
		{
			ID: "res-009", Name: "elasticsearch-prod-01",
			HostName: "elasticsearch-prod-01", IPAddress: "10.0.3.30",
			Type: "Linux", State: "active", OSName: "Ubuntu 22.04",
			AgentInstalled: true, Status: "managed", ResourceType: "Linux",
			Cloud: "AWS", Region: "us-east-1", InstanceSize: "r5.2xlarge",
			Tags:    map[string]string{"env": "production", "team": "platform", "role": "search"},
			Metrics: getSearchMetrics(),
		},
		{
			ID: "res-010", Name: "api-gateway-prod",
			HostName: "api-gateway-prod", IPAddress: "10.0.3.50",
			Type: "Linux", State: "active", OSName: "Amazon Linux 2023",
			AgentInstalled: true, Status: "managed", ResourceType: "Linux",
			Cloud: "AWS", Region: "us-east-1", InstanceSize: "t3.medium",
			Tags:    map[string]string{"env": "production", "team": "platform", "role": "api-gateway"},
			Metrics: getGatewayMetrics(),
		},

		// ── GCP US-CENTRAL1 (Kubernetes Cluster) ─────────────────────────
		{
			ID: "res-011", Name: "k8s-master-01",
			HostName: "k8s-master-01", IPAddress: "10.0.4.10",
			Type: "Linux", State: "active", OSName: "Ubuntu 22.04",
			AgentInstalled: true, Status: "managed", ResourceType: "Linux",
			Cloud: "GCP", Region: "us-central1", InstanceSize: "e2-standard-4",
			Tags:    map[string]string{"env": "production", "team": "platform", "role": "k8s-master", "cluster": "prod-central"},
			Metrics: getK8sMasterMetrics(),
		},
		{
			ID: "res-013", Name: "k8s-node-01",
			HostName: "k8s-node-01", IPAddress: "10.0.4.21",
			Type: "Linux", State: "active", OSName: "Ubuntu 22.04",
			AgentInstalled: true, Status: "managed", ResourceType: "Linux",
			Cloud: "GCP", Region: "us-central1", InstanceSize: "e2-standard-8",
			Tags:    map[string]string{"env": "production", "team": "platform", "role": "k8s-worker", "cluster": "prod-central"},
			Metrics: getK8sNodeMetrics(),
		},
		{
			ID: "res-014", Name: "k8s-node-02",
			HostName: "k8s-node-02", IPAddress: "10.0.4.22",
			Type: "Linux", State: "active", OSName: "Ubuntu 22.04",
			AgentInstalled: true, Status: "managed", ResourceType: "Linux",
			Cloud: "GCP", Region: "us-central1", InstanceSize: "e2-standard-8",
			Tags:    map[string]string{"env": "production", "team": "platform", "role": "k8s-worker", "cluster": "prod-central"},
			Metrics: getK8sNodeMetrics(),
		},
		{
			ID: "res-015", Name: "k8s-node-03",
			HostName: "k8s-node-03", IPAddress: "10.0.4.23",
			Type: "Linux", State: "active", OSName: "Ubuntu 22.04",
			AgentInstalled: true, Status: "managed", ResourceType: "Linux",
			Cloud: "GCP", Region: "us-central1", InstanceSize: "e2-standard-8",
			Tags:    map[string]string{"env": "production", "team": "platform", "role": "k8s-worker", "cluster": "prod-central"},
			Metrics: getHighNetworkK8sMetrics(),
		},
		{
			ID: "res-016", Name: "k8s-node-04",
			HostName: "k8s-node-04", IPAddress: "10.0.4.24",
			Type: "Linux", State: "active", OSName: "Ubuntu 22.04",
			AgentInstalled: true, Status: "managed", ResourceType: "Linux",
			Cloud: "GCP", Region: "us-central1", InstanceSize: "e2-standard-8",
			Tags:    map[string]string{"env": "production", "team": "platform", "role": "k8s-worker", "cluster": "prod-central"},
			Metrics: getK8sNodeMetrics(),
		},

		// ── Azure East US (Cloud Services) ───────────────────────────────
		{
			ID: "res-020", Name: "azure-sql-prod-01",
			HostName: "azure-sql-prod-01.database.windows.net", IPAddress: "",
			Type: "Azure SQL Database", State: "active",
			AgentInstalled: false, Status: "managed", ResourceType: "Azure SQL Database",
			Cloud: "Azure", Region: "eastus", InstanceSize: "Standard S3",
			Tags:    map[string]string{"env": "production", "team": "data", "role": "database", "engine": "mssql"},
			Metrics: getAzureSqlMetrics(),
		},
		{
			ID: "res-021", Name: "azure-vm-analytics-01",
			HostName: "azure-vm-analytics-01", IPAddress: "10.1.1.10",
			Type: "Windows", State: "active", OSName: "Windows Server 2022",
			AgentInstalled: true, Status: "managed", ResourceType: "Windows",
			Cloud: "Azure", Region: "eastus", InstanceSize: "Standard_D4s_v3",
			Tags:    map[string]string{"env": "production", "team": "analytics", "role": "compute"},
			Metrics: getWindowsMetrics(),
		},
		{
			ID: "res-022", Name: "azure-func-notifier",
			HostName: "azure-func-notifier.azurewebsites.net", IPAddress: "",
			Type: "Azure Function", State: "active",
			AgentInstalled: false, Status: "managed", ResourceType: "Azure Function",
			Cloud: "Azure", Region: "eastus", InstanceSize: "Consumption",
			Tags:    map[string]string{"env": "production", "team": "platform", "role": "serverless"},
			Metrics: getLowUtilMetrics(),
		},

		// ── On-Premises (VMware Datacenter) ──────────────────────────────
		{
			ID: "res-025", Name: "monitoring-agent-staging-01",
			HostName: "monitoring-agent-staging-01", IPAddress: "172.16.0.50",
			Type: "Linux", State: "active", OSName: "CentOS 8",
			AgentInstalled: true, Status: "managed", ResourceType: "Linux",
			Cloud: "OnPrem", Region: "datacenter-east",
			Tags:    map[string]string{"env": "staging", "team": "ops", "role": "monitoring"},
			Metrics: getLowUtilMetrics(),
		},
		{
			ID: "res-026", Name: "ldap-server-01",
			HostName: "ldap-server-01", IPAddress: "172.16.0.10",
			Type: "Linux", State: "active", OSName: "RHEL 9",
			AgentInstalled: true, Status: "managed", ResourceType: "Linux",
			Cloud: "OnPrem", Region: "datacenter-east",
			Tags:    map[string]string{"env": "production", "team": "security", "role": "identity"},
			Metrics: getLowUtilMetrics(),
		},
		{
			ID: "res-027", Name: "jenkins-build-01",
			HostName: "jenkins-build-01", IPAddress: "172.16.0.20",
			Type: "Linux", State: "active", OSName: "Ubuntu 22.04",
			AgentInstalled: true, Status: "managed", ResourceType: "Linux",
			Cloud: "OnPrem", Region: "datacenter-east",
			Tags:    map[string]string{"env": "production", "team": "devops", "role": "ci-cd"},
			Metrics: getCIMetrics(),
		},
		{
			ID: "res-028", Name: "esxi-host-01",
			HostName: "esxi-host-01", IPAddress: "172.16.0.5",
			Type: "VMware ESXi", State: "active", OSName: "ESXi 8.0",
			AgentInstalled: false, Status: "managed", ResourceType: "VMware ESXi",
			Cloud: "OnPrem", Region: "datacenter-east",
			Tags:    map[string]string{"env": "production", "team": "infra", "role": "hypervisor"},
			Metrics: getEsxiMetrics(),
		},
	}
}

// =============================================================================
// Metric helper functions (return realistic metric snapshots)
// =============================================================================

func getWebServerMetrics() opsramp.ResourceMetrics {
	return opsramp.ResourceMetrics{
		CPUUtilization:    97.3,
		MemoryUtilization: 62.1,
		DiskUtilization:   45.0,
		NetworkIn:         850.5,
		NetworkOut:        1200.3,
	}
}

func getHealthyWebMetrics() opsramp.ResourceMetrics {
	return opsramp.ResourceMetrics{
		CPUUtilization:    35.2,
		MemoryUtilization: 50.4,
		DiskUtilization:   42.0,
		NetworkIn:         720.0,
		NetworkOut:        1050.0,
	}
}

func getHealthyAppMetrics() opsramp.ResourceMetrics {
	return opsramp.ResourceMetrics{
		CPUUtilization:    42.5,
		MemoryUtilization: 58.3,
		DiskUtilization:   38.2,
		NetworkIn:         320.0,
		NetworkOut:        180.0,
	}
}

func getHighMemAppMetrics() opsramp.ResourceMetrics {
	return opsramp.ResourceMetrics{
		CPUUtilization:    55.0,
		MemoryUtilization: 88.0,
		DiskUtilization:   41.5,
		NetworkIn:         290.0,
		NetworkOut:        160.0,
	}
}

func getDbPrimaryMetrics() opsramp.ResourceMetrics {
	return opsramp.ResourceMetrics{
		CPUUtilization:    68.2,
		MemoryUtilization: 75.0,
		DiskUtilization:   92.0,
		NetworkIn:         500.0,
		NetworkOut:        450.0,
	}
}

func getDbReplicaMetrics() opsramp.ResourceMetrics {
	return opsramp.ResourceMetrics{
		CPUUtilization:    30.5,
		MemoryUtilization: 55.0,
		DiskUtilization:   78.0,
		NetworkIn:         400.0,
		NetworkOut:        50.0,
	}
}

func getCacheMetrics() opsramp.ResourceMetrics {
	return opsramp.ResourceMetrics{
		CPUUtilization:    15.0,
		MemoryUtilization: 72.0,
		DiskUtilization:   10.0,
		NetworkIn:         1200.0,
		NetworkOut:        1100.0,
	}
}

func getQueueMetrics() opsramp.ResourceMetrics {
	return opsramp.ResourceMetrics{
		CPUUtilization:    22.0,
		MemoryUtilization: 45.0,
		DiskUtilization:   30.0,
		NetworkIn:         600.0,
		NetworkOut:        580.0,
	}
}

func getSearchMetrics() opsramp.ResourceMetrics {
	return opsramp.ResourceMetrics{
		CPUUtilization:    55.0,
		MemoryUtilization: 80.0,
		DiskUtilization:   70.0,
		NetworkIn:         400.0,
		NetworkOut:        350.0,
	}
}

func getGatewayMetrics() opsramp.ResourceMetrics {
	return opsramp.ResourceMetrics{
		CPUUtilization:    28.0,
		MemoryUtilization: 35.0,
		DiskUtilization:   20.0,
		NetworkIn:         2000.0,
		NetworkOut:        1800.0,
	}
}

func getK8sMasterMetrics() opsramp.ResourceMetrics {
	return opsramp.ResourceMetrics{
		CPUUtilization:    40.0,
		MemoryUtilization: 55.0,
		DiskUtilization:   35.0,
		NetworkIn:         300.0,
		NetworkOut:        280.0,
	}
}

func getK8sNodeMetrics() opsramp.ResourceMetrics {
	return opsramp.ResourceMetrics{
		CPUUtilization:    60.0,
		MemoryUtilization: 65.0,
		DiskUtilization:   45.0,
		NetworkIn:         500.0,
		NetworkOut:        450.0,
	}
}

func getHighNetworkK8sMetrics() opsramp.ResourceMetrics {
	return opsramp.ResourceMetrics{
		CPUUtilization:    58.0,
		MemoryUtilization: 62.0,
		DiskUtilization:   50.0,
		NetworkIn:         3500.0,
		NetworkOut:        3200.0,
	}
}

func getAzureSqlMetrics() opsramp.ResourceMetrics {
	return opsramp.ResourceMetrics{
		CPUUtilization:    65.0,
		MemoryUtilization: 0, // not applicable for managed SQL
		DiskUtilization:   85.0,
		NetworkIn:         200.0,
		NetworkOut:        150.0,
	}
}

func getWindowsMetrics() opsramp.ResourceMetrics {
	return opsramp.ResourceMetrics{
		CPUUtilization:    45.0,
		MemoryUtilization: 70.0,
		DiskUtilization:   55.0,
		NetworkIn:         100.0,
		NetworkOut:        80.0,
	}
}

func getLowUtilMetrics() opsramp.ResourceMetrics {
	return opsramp.ResourceMetrics{
		CPUUtilization:    8.0,
		MemoryUtilization: 25.0,
		DiskUtilization:   15.0,
		NetworkIn:         10.0,
		NetworkOut:        5.0,
	}
}

func getCIMetrics() opsramp.ResourceMetrics {
	return opsramp.ResourceMetrics{
		CPUUtilization:    75.0,
		MemoryUtilization: 60.0,
		DiskUtilization:   65.0,
		NetworkIn:         200.0,
		NetworkOut:        400.0,
	}
}

func getEsxiMetrics() opsramp.ResourceMetrics {
	return opsramp.ResourceMetrics{
		CPUUtilization:    52.0,
		MemoryUtilization: 68.0,
		DiskUtilization:   40.0,
		NetworkIn:         800.0,
		NetworkOut:        750.0,
	}
}

package mockdata

import (
	"opsramp-agent/juniper"
)

// =============================================================================
// Juniper Network Switch Mock Data
// =============================================================================
//
// Realistic Juniper Mist-managed switch telemetry aligned with the Juniper
// Mist API v1 schema (GET /api/v1/sites/{site_id}/stats/devices?type=switch).
//
// Network topology:
//
//   ┌──────────────────────────────────────────────────────────────────────┐
//   │                    Datacenter East (HPE GreenLake + AWS)            │
//   │                                                                    │
//   │  sw-dc-east-01 (EX4600) — Web/App Tier (HPE ProLiant + AWS)      │
//   │    ge-0/0/1 → web-server-prod-01   (10.0.1.101) ⚠ RX ERRORS+LOSS│
//   │    ge-0/0/2 → app-server-prod-01   (10.0.1.102)                  │
//   │    ge-0/0/3 → app-server-prod-02   (10.0.1.103) ⚠ RX ERRORS     │
//   │    ge-0/0/4 → web-server-prod-02   (10.0.1.104)                  │
//   │    xe-0/0/0 → uplink to core                                      │
//   │                                                                    │
//   │  sw-dc-east-02 (EX4600) — Database Tier                          │
//   │    ge-0/0/1 → db-primary-01        (10.0.2.10)                   │
//   │    ge-0/0/2 → db-replica-01        (10.0.2.11)                   │
//   │    ge-0/0/3 → redis-cache-01       (10.0.2.20)                   │
//   │    xe-0/0/0 → uplink to core                                      │
//   │                                                                    │
//   │  sw-dc-east-03 (EX4300-48T) — Infra / On-Prem                   │
//   │    ge-0/0/1 → rabbitmq-prod-01     (10.0.3.10)                   │
//   │    ge-0/0/2 → elasticsearch-prod-01(10.0.3.30)                   │
//   │    ge-0/0/3 → api-gateway-prod     (10.0.3.50)                   │
//   │    ge-0/0/10 → esxi-host-01        (172.16.0.5)                  │
//   │    ge-0/0/11 → ldap-server-01      (172.16.0.10)                 │
//   │    ge-0/0/12 → jenkins-build-01    (172.16.0.20)                 │
//   │    ge-0/0/13 → monitoring-staging   (172.16.0.50)                │
//   │    xe-0/0/0 → uplink to core                                      │
//   │                                                                    │
//   │  sw-dc-east-04 (EX4300-48T) — HPE GreenLake K8s (HPE ProLiant)  │
//   │    ge-0/0/1 → k8s-master-01        (10.0.4.10)                   │
//   │    ge-0/0/2 → k8s-node-01          (10.0.4.21)                   │
//   │    ge-0/0/3 → k8s-node-02          (10.0.4.22)                   │
//   │    ge-0/0/4 → k8s-node-03          (10.0.4.23)  ⚠ PACKET LOSS   │
//   │    ge-0/0/5 → k8s-node-04          (10.0.4.24)  ⚠ LINK FLAPS   │
//   │    xe-0/0/0 → uplink to core                                      │
//   └──────────────────────────────────────────────────────────────────────┘
//
// Key scenarios designed for agent correlation:
//   1. k8s-node-03 has a latency alert → switch port shows 4.7% packet loss
//   2. k8s-node-04 runs greenlake-portal → switch port has link flaps
//   3. app-server-prod-02 has memory alert → switch port shows RX errors
//   4. web-server-prod-01 (HPE GreenLake) has CPU alert → switch port has RX errors + packet loss (compound incident)
// =============================================================================

// GetNetworkSwitches returns mock Juniper switch telemetry data.
func GetNetworkSwitches() []juniper.SwitchStats {
	return []juniper.SwitchStats{
		// ── sw-dc-east-01: Web/App Tier Switch ──────────────────────────
		{
			ID:       "sw-001",
			OrgID:    "org-acme-001",
			SiteID:   "site-dc-east",
			Name:     "sw-dc-east-01",
			Hostname: "sw-dc-east-01",
			IP:       "10.0.0.1",
			Mac:      "5c:45:27:a9:65:80",
			Model:    "EX4600",
			Serial:   "TC3714190003",
			Version:  "21.4R3-S5",
			Status:   "connected",
			Uptime:   8640000, // 100 days
			Type:     "switch",
			LastSeen: 1741478400, // 2025-03-09
			CpuStat: juniper.CpuStat{
				Idle: 85.0, System: 10.0, User: 5.0,
			},
			MemoryStat:  juniper.MemoryStat{Usage: 42.0},
			Location:    "Datacenter East, Rack A1, U20",
			Description: "Top-of-rack switch for HPE ProLiant web server and AWS application servers",
			Ports: []juniper.SwitchPort{
				// ge-0/0/1 → web-server-prod-01: ⚠ RX ERRORS + PACKET LOSS
				// Compound incident: CRC errors from a degrading cable cause
				// retransmissions which spike CPU. Network IS a contributing factor.
				{
					PortID: "ge-0/0/1", PortMac: "5c:45:27:a9:65:81",
					PortUsage: "lan", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 1000,
					RxBytes: 8515104416, RxPkts: 57770567, RxBps: 850000000,
					RxErrors: 14872, RxBcastPkts: 1200, RxMcastPkts: 450,
					TxBytes: 12021738968, TxPkts: 81220406, TxBps: 1200000000,
					TxErrors: 342, TxBcastPkts: 800, TxMcastPkts: 300,
					Jitter: 6.8, Latency: 14.2, Loss: 1.9,
					LastFlapped: 1741464600, // ~5 hours ago
					MacCount:    1, MacLimit: 4096,
					NeighborMac: "aa:bb:cc:01:01:01", NeighborPortDesc: "eth0",
					NeighborSystemName: "web-server-prod-01",
					StpRole:            "designated", StpState: "forwarding",
				},
				// ge-0/0/2 → app-server-prod-01: CLEAN
				{
					PortID: "ge-0/0/2", PortMac: "5c:45:27:a9:65:82",
					PortUsage: "lan", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 1000,
					RxBytes: 3200000000, RxPkts: 22000000, RxBps: 320000000,
					RxErrors: 0, RxBcastPkts: 900, RxMcastPkts: 350,
					TxBytes: 1800000000, TxPkts: 12500000, TxBps: 180000000,
					TxErrors: 0, TxBcastPkts: 600, TxMcastPkts: 250,
					Jitter: 0.2, Latency: 0.4, Loss: 0.0,
					LastFlapped: 0,
					MacCount:    1, MacLimit: 4096,
					NeighborMac: "aa:bb:cc:01:01:02", NeighborPortDesc: "eth0",
					NeighborSystemName: "app-server-prod-01",
					StpRole:            "designated", StpState: "forwarding",
				},
				// ge-0/0/3 → app-server-prod-02: ⚠ RX ERRORS (CRC errors from bad cable/transceiver)
				{
					PortID: "ge-0/0/3", PortMac: "5c:45:27:a9:65:83",
					PortUsage: "lan", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 1000,
					RxBytes: 2900000000, RxPkts: 20000000, RxBps: 290000000,
					RxErrors: 48752, RxBcastPkts: 950, RxMcastPkts: 380,
					TxBytes: 1600000000, TxPkts: 11000000, TxBps: 160000000,
					TxErrors: 12, TxBcastPkts: 550, TxMcastPkts: 220,
					Jitter: 1.8, Latency: 2.5, Loss: 0.8,
					LastFlapped: 1741392000, // ~1 day ago
					MacCount:    1, MacLimit: 4096,
					NeighborMac: "aa:bb:cc:01:01:03", NeighborPortDesc: "eth0",
					NeighborSystemName: "app-server-prod-02",
					StpRole:            "designated", StpState: "forwarding",
				},
				// ge-0/0/4 → web-server-prod-02: CLEAN
				{
					PortID: "ge-0/0/4", PortMac: "5c:45:27:a9:65:84",
					PortUsage: "lan", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 1000,
					RxBytes: 7200000000, RxPkts: 49000000, RxBps: 720000000,
					RxErrors: 0, RxBcastPkts: 1100, RxMcastPkts: 420,
					TxBytes: 10500000000, TxPkts: 71000000, TxBps: 1050000000,
					TxErrors: 0, TxBcastPkts: 750, TxMcastPkts: 290,
					Jitter: 0.2, Latency: 0.4, Loss: 0.0,
					LastFlapped: 0,
					MacCount:    1, MacLimit: 4096,
					NeighborMac: "aa:bb:cc:01:01:04", NeighborPortDesc: "eth0",
					NeighborSystemName: "web-server-prod-02",
					StpRole:            "designated", StpState: "forwarding",
				},
				// xe-0/0/0 → uplink to core switch
				{
					PortID: "xe-0/0/0", PortMac: "5c:45:27:a9:65:90",
					PortUsage: "uplink", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 10000,
					RxBytes: 50000000000, RxPkts: 340000000, RxBps: 5000000000,
					RxErrors: 0, RxBcastPkts: 5000, RxMcastPkts: 2000,
					TxBytes: 45000000000, TxPkts: 310000000, TxBps: 4500000000,
					TxErrors: 0, TxBcastPkts: 4500, TxMcastPkts: 1800,
					Jitter: 0.1, Latency: 0.2, Loss: 0.0,
					LastFlapped: 0,
					MacCount:    128, MacLimit: 16384,
					NeighborMac: "aa:bb:cc:00:00:01", NeighborPortDesc: "xe-0/0/1",
					NeighborSystemName: "core-sw-01",
					StpRole:            "root", StpState: "forwarding",
					XcvrModel: "SFP+-10G-SR", XcvrPartNumber: "740-021487", XcvrSerial: "N6AA9HT",
				},
			},
		},

		// ── sw-dc-east-02: Database Tier Switch ─────────────────────────
		{
			ID:       "sw-002",
			OrgID:    "org-acme-001",
			SiteID:   "site-dc-east",
			Name:     "sw-dc-east-02",
			Hostname: "sw-dc-east-02",
			IP:       "10.0.0.2",
			Mac:      "5c:45:27:b1:22:00",
			Model:    "EX4600",
			Serial:   "TC3714190004",
			Version:  "21.4R3-S5",
			Status:   "connected",
			Uptime:   7776000, // 90 days
			Type:     "switch",
			LastSeen: 1741478400,
			CpuStat: juniper.CpuStat{
				Idle: 90.0, System: 7.0, User: 3.0,
			},
			MemoryStat:  juniper.MemoryStat{Usage: 35.0},
			Location:    "Datacenter East, Rack B1, U20",
			Description: "Top-of-rack switch for database servers",
			Ports: []juniper.SwitchPort{
				// ge-0/0/1 → db-primary-01: CLEAN (disk issue is NOT network)
				{
					PortID: "ge-0/0/1", PortMac: "5c:45:27:b1:22:01",
					PortUsage: "lan", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 1000,
					RxBytes: 5000000000, RxPkts: 35000000, RxBps: 500000000,
					RxErrors: 0, RxBcastPkts: 600, RxMcastPkts: 200,
					TxBytes: 4500000000, TxPkts: 31000000, TxBps: 450000000,
					TxErrors: 0, TxBcastPkts: 500, TxMcastPkts: 180,
					Jitter: 0.1, Latency: 0.3, Loss: 0.0,
					LastFlapped: 0,
					MacCount:    1, MacLimit: 4096,
					NeighborMac: "aa:bb:cc:02:02:01", NeighborPortDesc: "eth0",
					NeighborSystemName: "db-primary-01",
					StpRole:            "designated", StpState: "forwarding",
				},
				// ge-0/0/2 → db-replica-01: CLEAN
				{
					PortID: "ge-0/0/2", PortMac: "5c:45:27:b1:22:02",
					PortUsage: "lan", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 1000,
					RxBytes: 4000000000, RxPkts: 28000000, RxBps: 400000000,
					RxErrors: 0, RxBcastPkts: 500, RxMcastPkts: 180,
					TxBytes: 500000000, TxPkts: 3500000, TxBps: 50000000,
					TxErrors: 0, TxBcastPkts: 400, TxMcastPkts: 150,
					Jitter: 0.1, Latency: 0.3, Loss: 0.0,
					LastFlapped: 0,
					MacCount:    1, MacLimit: 4096,
					NeighborMac: "aa:bb:cc:02:02:02", NeighborPortDesc: "eth0",
					NeighborSystemName: "db-replica-01",
					StpRole:            "designated", StpState: "forwarding",
				},
				// ge-0/0/3 → redis-cache-01: CLEAN
				{
					PortID: "ge-0/0/3", PortMac: "5c:45:27:b1:22:03",
					PortUsage: "lan", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 1000,
					RxBytes: 12000000000, RxPkts: 82000000, RxBps: 1200000000,
					RxErrors: 0, RxBcastPkts: 300, RxMcastPkts: 100,
					TxBytes: 11000000000, TxPkts: 75000000, TxBps: 1100000000,
					TxErrors: 0, TxBcastPkts: 280, TxMcastPkts: 90,
					Jitter: 0.1, Latency: 0.2, Loss: 0.0,
					LastFlapped: 0,
					MacCount:    1, MacLimit: 4096,
					NeighborMac: "aa:bb:cc:02:02:03", NeighborPortDesc: "eth0",
					NeighborSystemName: "redis-cache-01",
					StpRole:            "designated", StpState: "forwarding",
				},
				// xe-0/0/0 → uplink
				{
					PortID: "xe-0/0/0", PortMac: "5c:45:27:b1:22:90",
					PortUsage: "uplink", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 10000,
					RxBytes: 30000000000, RxPkts: 200000000, RxBps: 3000000000,
					RxErrors: 0, RxBcastPkts: 3000, RxMcastPkts: 1200,
					TxBytes: 28000000000, TxPkts: 190000000, TxBps: 2800000000,
					TxErrors: 0, TxBcastPkts: 2800, TxMcastPkts: 1100,
					Jitter: 0.1, Latency: 0.2, Loss: 0.0,
					LastFlapped: 0,
					MacCount:    64, MacLimit: 16384,
					NeighborMac: "aa:bb:cc:00:00:01", NeighborPortDesc: "xe-0/0/2",
					NeighborSystemName: "core-sw-01",
					StpRole:            "root", StpState: "forwarding",
					XcvrModel: "SFP+-10G-SR", XcvrPartNumber: "740-021487", XcvrSerial: "N6BB2KT",
				},
			},
		},

		// ── sw-dc-east-03: Infrastructure / On-Prem Switch ──────────────
		{
			ID:       "sw-003",
			OrgID:    "org-acme-001",
			SiteID:   "site-dc-east",
			Name:     "sw-dc-east-03",
			Hostname: "sw-dc-east-03",
			IP:       "10.0.0.3",
			Mac:      "5c:45:27:c3:44:00",
			Model:    "EX4300-48T",
			Serial:   "TC3714190005",
			Version:  "21.4R3-S5",
			Status:   "connected",
			Uptime:   6912000, // 80 days
			Type:     "switch",
			LastSeen: 1741478400,
			CpuStat: juniper.CpuStat{
				Idle: 92.0, System: 5.0, User: 3.0,
			},
			MemoryStat:  juniper.MemoryStat{Usage: 28.0},
			Location:    "Datacenter East, Rack C1, U20",
			Description: "Access switch for supporting infrastructure and on-premises servers",
			Ports: []juniper.SwitchPort{
				// ge-0/0/1 → rabbitmq-prod-01: CLEAN
				{
					PortID: "ge-0/0/1", PortMac: "5c:45:27:c3:44:01",
					PortUsage: "lan", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 1000,
					RxBytes: 6000000000, RxPkts: 41000000, RxBps: 600000000,
					RxErrors: 0, RxBcastPkts: 500, RxMcastPkts: 200,
					TxBytes: 5800000000, TxPkts: 40000000, TxBps: 580000000,
					TxErrors: 0, TxBcastPkts: 480, TxMcastPkts: 190,
					Jitter: 0.2, Latency: 0.3, Loss: 0.0,
					LastFlapped: 0,
					MacCount:    1, MacLimit: 4096,
					NeighborMac: "aa:bb:cc:03:03:01", NeighborPortDesc: "eth0",
					NeighborSystemName: "rabbitmq-prod-01",
					StpRole:            "designated", StpState: "forwarding",
				},
				// ge-0/0/2 → elasticsearch-prod-01: CLEAN
				{
					PortID: "ge-0/0/2", PortMac: "5c:45:27:c3:44:02",
					PortUsage: "lan", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 1000,
					RxBytes: 4000000000, RxPkts: 28000000, RxBps: 400000000,
					RxErrors: 0, RxBcastPkts: 400, RxMcastPkts: 160,
					TxBytes: 3500000000, TxPkts: 24000000, TxBps: 350000000,
					TxErrors: 0, TxBcastPkts: 380, TxMcastPkts: 150,
					Jitter: 0.3, Latency: 0.4, Loss: 0.0,
					LastFlapped: 0,
					MacCount:    1, MacLimit: 4096,
					NeighborMac: "aa:bb:cc:03:03:02", NeighborPortDesc: "eth0",
					NeighborSystemName: "elasticsearch-prod-01",
					StpRole:            "designated", StpState: "forwarding",
				},
				// ge-0/0/3 → api-gateway-prod: CLEAN
				{
					PortID: "ge-0/0/3", PortMac: "5c:45:27:c3:44:03",
					PortUsage: "lan", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 1000,
					RxBytes: 20000000000, RxPkts: 140000000, RxBps: 2000000000,
					RxErrors: 0, RxBcastPkts: 800, RxMcastPkts: 300,
					TxBytes: 18000000000, TxPkts: 125000000, TxBps: 1800000000,
					TxErrors: 0, TxBcastPkts: 750, TxMcastPkts: 280,
					Jitter: 0.2, Latency: 0.3, Loss: 0.0,
					LastFlapped: 0,
					MacCount:    1, MacLimit: 4096,
					NeighborMac: "aa:bb:cc:03:03:03", NeighborPortDesc: "eth0",
					NeighborSystemName: "api-gateway-prod",
					StpRole:            "designated", StpState: "forwarding",
				},
				// ge-0/0/10 → esxi-host-01: CLEAN
				{
					PortID: "ge-0/0/10", PortMac: "5c:45:27:c3:44:0a",
					PortUsage: "lan", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 1000,
					RxBytes: 8000000000, RxPkts: 55000000, RxBps: 800000000,
					RxErrors: 0, RxBcastPkts: 600, RxMcastPkts: 250,
					TxBytes: 7500000000, TxPkts: 51000000, TxBps: 750000000,
					TxErrors: 0, TxBcastPkts: 580, TxMcastPkts: 230,
					Jitter: 0.2, Latency: 0.3, Loss: 0.0,
					LastFlapped: 0,
					MacCount:    8, MacLimit: 4096,
					NeighborMac: "aa:bb:cc:03:03:10", NeighborPortDesc: "vmnic0",
					NeighborSystemName: "esxi-host-01",
					StpRole:            "designated", StpState: "forwarding",
				},
				// ge-0/0/11 → ldap-server-01: CLEAN
				{
					PortID: "ge-0/0/11", PortMac: "5c:45:27:c3:44:0b",
					PortUsage: "lan", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 1000,
					RxBytes: 100000000, RxPkts: 700000, RxBps: 10000000,
					RxErrors: 0, RxBcastPkts: 200, RxMcastPkts: 80,
					TxBytes: 50000000, TxPkts: 350000, TxBps: 5000000,
					TxErrors: 0, TxBcastPkts: 180, TxMcastPkts: 70,
					Jitter: 0.1, Latency: 0.2, Loss: 0.0,
					LastFlapped: 0,
					MacCount:    1, MacLimit: 4096,
					NeighborMac: "aa:bb:cc:03:03:11", NeighborPortDesc: "eth0",
					NeighborSystemName: "ldap-server-01",
					StpRole:            "designated", StpState: "forwarding",
				},
				// ge-0/0/12 → jenkins-build-01: CLEAN
				{
					PortID: "ge-0/0/12", PortMac: "5c:45:27:c3:44:0c",
					PortUsage: "lan", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 1000,
					RxBytes: 2000000000, RxPkts: 14000000, RxBps: 200000000,
					RxErrors: 0, RxBcastPkts: 350, RxMcastPkts: 140,
					TxBytes: 4000000000, TxPkts: 28000000, TxBps: 400000000,
					TxErrors: 0, TxBcastPkts: 320, TxMcastPkts: 130,
					Jitter: 0.2, Latency: 0.3, Loss: 0.0,
					LastFlapped: 0,
					MacCount:    1, MacLimit: 4096,
					NeighborMac: "aa:bb:cc:03:03:12", NeighborPortDesc: "eth0",
					NeighborSystemName: "jenkins-build-01",
					StpRole:            "designated", StpState: "forwarding",
				},
				// ge-0/0/13 → monitoring-agent-staging-01: CLEAN
				{
					PortID: "ge-0/0/13", PortMac: "5c:45:27:c3:44:0d",
					PortUsage: "lan", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 1000,
					RxBytes: 100000000, RxPkts: 700000, RxBps: 10000000,
					RxErrors: 0, RxBcastPkts: 150, RxMcastPkts: 60,
					TxBytes: 50000000, TxPkts: 350000, TxBps: 5000000,
					TxErrors: 0, TxBcastPkts: 130, TxMcastPkts: 50,
					Jitter: 0.1, Latency: 0.2, Loss: 0.0,
					LastFlapped: 0,
					MacCount:    1, MacLimit: 4096,
					NeighborMac: "aa:bb:cc:03:03:13", NeighborPortDesc: "eth0",
					NeighborSystemName: "monitoring-agent-staging-01",
					StpRole:            "designated", StpState: "forwarding",
				},
				// xe-0/0/0 → uplink
				{
					PortID: "xe-0/0/0", PortMac: "5c:45:27:c3:44:90",
					PortUsage: "uplink", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 10000,
					RxBytes: 25000000000, RxPkts: 170000000, RxBps: 2500000000,
					RxErrors: 0, RxBcastPkts: 2500, RxMcastPkts: 1000,
					TxBytes: 23000000000, TxPkts: 160000000, TxBps: 2300000000,
					TxErrors: 0, TxBcastPkts: 2300, TxMcastPkts: 950,
					Jitter: 0.1, Latency: 0.2, Loss: 0.0,
					LastFlapped: 0,
					MacCount:    32, MacLimit: 16384,
					NeighborMac: "aa:bb:cc:00:00:01", NeighborPortDesc: "xe-0/0/3",
					NeighborSystemName: "core-sw-01",
					StpRole:            "root", StpState: "forwarding",
					XcvrModel: "SFP+-10G-SR", XcvrPartNumber: "740-021487", XcvrSerial: "N6CC3LT",
				},
			},
		},

		// ── sw-dc-east-04: On-Prem K8s Cluster Switch (HPE ProLiant) ────
		// This is the KEY switch for the greenlake-portal correlation scenario.
		// These are on-prem HPE ProLiant servers running Kubernetes, managed by
		// a Juniper EX4300 switch with OpsRamp collector installed on-site.
		// k8s-node-03 (latency alert) and k8s-node-04 (greenlake-portal pods)
		// both connect here and have port-level issues.
		{
			ID:       "sw-004",
			OrgID:    "org-acme-001",
			SiteID:   "site-dc-east",
			Name:     "sw-dc-east-04",
			Hostname: "sw-dc-east-04",
			IP:       "10.0.4.1",
			Mac:      "5c:45:27:d4:88:00",
			Model:    "EX4300-48T",
			Serial:   "TC3714190006",
			Version:  "21.4R3-S5",
			Status:   "connected",
			Uptime:   5184000, // 60 days
			Type:     "switch",
			LastSeen: 1741478400,
			CpuStat: juniper.CpuStat{
				Idle: 78.0, System: 15.0, User: 7.0,
			},
			MemoryStat:  juniper.MemoryStat{Usage: 55.0},
			Location:    "Datacenter East, Rack C1, U10",
			Description: "Top-of-rack switch for HPE GreenLake Kubernetes cluster running on HPE ProLiant servers (OpsRamp collector on-site)",
			Ports: []juniper.SwitchPort{
				// ge-0/0/1 → k8s-master-01: CLEAN
				{
					PortID: "ge-0/0/1", PortMac: "5c:45:27:d4:88:01",
					PortUsage: "lan", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 1000,
					RxBytes: 3000000000, RxPkts: 21000000, RxBps: 300000000,
					RxErrors: 0, RxBcastPkts: 500, RxMcastPkts: 200,
					TxBytes: 2800000000, TxPkts: 19000000, TxBps: 280000000,
					TxErrors: 0, TxBcastPkts: 480, TxMcastPkts: 190,
					Jitter: 0.3, Latency: 0.5, Loss: 0.0,
					LastFlapped: 0,
					MacCount:    1, MacLimit: 4096,
					NeighborMac: "aa:bb:cc:04:04:01", NeighborPortDesc: "eth0",
					NeighborSystemName: "k8s-master-01",
					StpRole:            "designated", StpState: "forwarding",
				},
				// ge-0/0/2 → k8s-node-01: CLEAN
				{
					PortID: "ge-0/0/2", PortMac: "5c:45:27:d4:88:02",
					PortUsage: "lan", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 1000,
					RxBytes: 5000000000, RxPkts: 35000000, RxBps: 500000000,
					RxErrors: 0, RxBcastPkts: 700, RxMcastPkts: 300,
					TxBytes: 4500000000, TxPkts: 31000000, TxBps: 450000000,
					TxErrors: 0, TxBcastPkts: 680, TxMcastPkts: 280,
					Jitter: 0.3, Latency: 0.5, Loss: 0.0,
					LastFlapped: 0,
					MacCount:    1, MacLimit: 4096,
					NeighborMac: "aa:bb:cc:04:04:02", NeighborPortDesc: "eth0",
					NeighborSystemName: "k8s-node-01",
					StpRole:            "designated", StpState: "forwarding",
				},
				// ge-0/0/3 → k8s-node-02: CLEAN
				{
					PortID: "ge-0/0/3", PortMac: "5c:45:27:d4:88:03",
					PortUsage: "lan", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 1000,
					RxBytes: 5000000000, RxPkts: 35000000, RxBps: 500000000,
					RxErrors: 0, RxBcastPkts: 700, RxMcastPkts: 300,
					TxBytes: 4500000000, TxPkts: 31000000, TxBps: 450000000,
					TxErrors: 0, TxBcastPkts: 680, TxMcastPkts: 280,
					Jitter: 0.3, Latency: 0.5, Loss: 0.0,
					LastFlapped: 0,
					MacCount:    1, MacLimit: 4096,
					NeighborMac: "aa:bb:cc:04:04:03", NeighborPortDesc: "eth0",
					NeighborSystemName: "k8s-node-02",
					StpRole:            "designated", StpState: "forwarding",
				},
				// ge-0/0/4 → k8s-node-03: ⚠ PACKET LOSS (4.7%) + HIGH LATENCY
				// This correlates with ALR-20260219-005 "Network latency spike on k8s-node-03"
				{
					PortID: "ge-0/0/4", PortMac: "5c:45:27:d4:88:04",
					PortUsage: "lan", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 1000,
					RxBytes: 35000000000, RxPkts: 240000000, RxBps: 3500000000,
					RxErrors: 23841, RxBcastPkts: 1200, RxMcastPkts: 500,
					TxBytes: 32000000000, TxPkts: 220000000, TxBps: 3200000000,
					TxErrors: 8452, TxBcastPkts: 1100, TxMcastPkts: 460,
					Jitter: 12.5, Latency: 45.0, Loss: 4.7,
					LastFlapped: 1741435200, // ~12 hours ago
					MacCount:    1, MacLimit: 4096,
					NeighborMac: "aa:bb:cc:04:04:04", NeighborPortDesc: "eth0",
					NeighborSystemName: "k8s-node-03",
					StpRole:            "designated", StpState: "forwarding",
				},
				// ge-0/0/5 → k8s-node-04: ⚠ LINK FLAPS (greenlake-portal host)
				// This correlates with ALR-20260219-008 "Container restart loop on greenlake-portal"
				// The link has been flapping causing intermittent connectivity
				{
					PortID: "ge-0/0/5", PortMac: "5c:45:27:d4:88:05",
					PortUsage: "lan", Active: true, Up: true, Disabled: false,
					FullDuplex: false, Speed: 100, // ⚠ Negotiated at 100Mbps half-duplex (bad!)
					RxBytes: 5000000000, RxPkts: 35000000, RxBps: 500000000,
					RxErrors: 156789, RxBcastPkts: 2500, RxMcastPkts: 1000,
					TxBytes: 4500000000, TxPkts: 31000000, TxBps: 450000000,
					TxErrors: 89234, TxBcastPkts: 2200, TxMcastPkts: 900,
					Jitter: 25.0, Latency: 120.0, Loss: 8.3,
					LastFlapped: 1741474800, // ~1 hour ago (very recent!)
					MacCount:    1, MacLimit: 4096,
					NeighborMac: "aa:bb:cc:04:04:05", NeighborPortDesc: "eth0",
					NeighborSystemName: "k8s-node-04",
					StpRole:            "designated", StpState: "forwarding",
				},
				// xe-0/0/0 → uplink to core
				{
					PortID: "xe-0/0/0", PortMac: "5c:45:27:d4:88:90",
					PortUsage: "uplink", Active: true, Up: true, Disabled: false,
					FullDuplex: true, Speed: 10000,
					RxBytes: 80000000000, RxPkts: 550000000, RxBps: 8000000000,
					RxErrors: 0, RxBcastPkts: 4000, RxMcastPkts: 1600,
					TxBytes: 75000000000, TxPkts: 510000000, TxBps: 7500000000,
					TxErrors: 0, TxBcastPkts: 3800, TxMcastPkts: 1500,
					Jitter: 0.1, Latency: 0.3, Loss: 0.0,
					LastFlapped: 0,
					MacCount:    256, MacLimit: 16384,
					NeighborMac: "aa:bb:cc:00:00:02", NeighborPortDesc: "xe-0/0/1",
					NeighborSystemName: "core-sw-02",
					StpRole:            "root", StpState: "forwarding",
					XcvrModel: "SFP+-10G-LR", XcvrPartNumber: "740-021488", XcvrSerial: "N6DD4MT",
				},
			},
		},
	}
}

// GetNetworkPortMappings returns the mapping from OpsRamp resources to their
// connected Juniper switch ports.
//
// Correlation key: IP address.
// Each ResourceIP here is deliberately identical to the IPAddress field of the
// same resource in resources.go. This mirrors how production correlation works:
//
//	Production flow:
//	  1. OpsRamp discovers a server and records its IP (e.g., 10.0.4.24).
//	  2. Juniper switch learns the same IP via ARP table / LLDP neighbor on
//	     a specific port (e.g., ge-0/0/5 on sw-dc-east-04).
//	  3. A correlation service (DCIM like NetBox, or LLDP/CDP neighbor
//	     discovery) joins these two records by matching IP address, producing
//	     a mapping: OpsRamp resource ↔ Juniper switch:port.
//
//	Mock shortcut:
//	  We hardcode that mapping table here. The IP addresses are kept
//	  consistent with resources.go so the mock faithfully represents
//	  the production join-by-IP pattern.
//
// The juniper.Client.findMapping() function (juniper/client.go) looks up
// entries in this table by ResourceIP, ResourceName, or ResourceID, with
// IP being the primary real-world correlation key.
func GetNetworkPortMappings() []juniper.PortMapping {
	return []juniper.PortMapping{
		// sw-dc-east-01 (Web/App Tier)
		{ResourceID: "res-001", ResourceName: "web-server-prod-01", ResourceIP: "10.0.1.101", SwitchID: "sw-001", SwitchName: "sw-dc-east-01", PortID: "ge-0/0/1"},
		{ResourceID: "res-002", ResourceName: "app-server-prod-01", ResourceIP: "10.0.1.102", SwitchID: "sw-001", SwitchName: "sw-dc-east-01", PortID: "ge-0/0/2"},
		{ResourceID: "res-003", ResourceName: "app-server-prod-02", ResourceIP: "10.0.1.103", SwitchID: "sw-001", SwitchName: "sw-dc-east-01", PortID: "ge-0/0/3"},
		{ResourceID: "res-004", ResourceName: "web-server-prod-02", ResourceIP: "10.0.1.104", SwitchID: "sw-001", SwitchName: "sw-dc-east-01", PortID: "ge-0/0/4"},

		// sw-dc-east-02 (Database Tier)
		{ResourceID: "res-005", ResourceName: "db-primary-01", ResourceIP: "10.0.2.10", SwitchID: "sw-002", SwitchName: "sw-dc-east-02", PortID: "ge-0/0/1"},
		{ResourceID: "res-006", ResourceName: "db-replica-01", ResourceIP: "10.0.2.11", SwitchID: "sw-002", SwitchName: "sw-dc-east-02", PortID: "ge-0/0/2"},
		{ResourceID: "res-007", ResourceName: "redis-cache-01", ResourceIP: "10.0.2.20", SwitchID: "sw-002", SwitchName: "sw-dc-east-02", PortID: "ge-0/0/3"},

		// sw-dc-east-03 (Infrastructure / On-Prem)
		{ResourceID: "res-008", ResourceName: "rabbitmq-prod-01", ResourceIP: "10.0.3.10", SwitchID: "sw-003", SwitchName: "sw-dc-east-03", PortID: "ge-0/0/1"},
		{ResourceID: "res-009", ResourceName: "elasticsearch-prod-01", ResourceIP: "10.0.3.30", SwitchID: "sw-003", SwitchName: "sw-dc-east-03", PortID: "ge-0/0/2"},
		{ResourceID: "res-010", ResourceName: "api-gateway-prod", ResourceIP: "10.0.3.50", SwitchID: "sw-003", SwitchName: "sw-dc-east-03", PortID: "ge-0/0/3"},
		{ResourceID: "res-028", ResourceName: "esxi-host-01", ResourceIP: "172.16.0.5", SwitchID: "sw-003", SwitchName: "sw-dc-east-03", PortID: "ge-0/0/10"},
		{ResourceID: "res-026", ResourceName: "ldap-server-01", ResourceIP: "172.16.0.10", SwitchID: "sw-003", SwitchName: "sw-dc-east-03", PortID: "ge-0/0/11"},
		{ResourceID: "res-027", ResourceName: "jenkins-build-01", ResourceIP: "172.16.0.20", SwitchID: "sw-003", SwitchName: "sw-dc-east-03", PortID: "ge-0/0/12"},
		{ResourceID: "res-025", ResourceName: "monitoring-agent-staging-01", ResourceIP: "172.16.0.50", SwitchID: "sw-003", SwitchName: "sw-dc-east-03", PortID: "ge-0/0/13"},

		// sw-dc-east-04 (On-Prem K8s Cluster)
		{ResourceID: "res-011", ResourceName: "k8s-master-01", ResourceIP: "10.0.4.10", SwitchID: "sw-004", SwitchName: "sw-dc-east-04", PortID: "ge-0/0/1"},
		{ResourceID: "res-013", ResourceName: "k8s-node-01", ResourceIP: "10.0.4.21", SwitchID: "sw-004", SwitchName: "sw-dc-east-04", PortID: "ge-0/0/2"},
		{ResourceID: "res-014", ResourceName: "k8s-node-02", ResourceIP: "10.0.4.22", SwitchID: "sw-004", SwitchName: "sw-dc-east-04", PortID: "ge-0/0/3"},
		{ResourceID: "res-015", ResourceName: "k8s-node-03", ResourceIP: "10.0.4.23", SwitchID: "sw-004", SwitchName: "sw-dc-east-04", PortID: "ge-0/0/4"},
		{ResourceID: "res-016", ResourceName: "k8s-node-04", ResourceIP: "10.0.4.24", SwitchID: "sw-004", SwitchName: "sw-dc-east-04", PortID: "ge-0/0/5"},
	}
}

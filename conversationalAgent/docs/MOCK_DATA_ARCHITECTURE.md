# Mock Data Architecture — How OpsRamp & Juniper Data Are Connected

## The Key Idea (Read This First)

OpsRamp and Juniper are **two completely separate systems** with their own databases, their own IDs, and no shared state. They never talk to each other directly.

- **OpsRamp** monitors servers — it knows server names, IPs, CPU/memory metrics, alerts, incidents etc..
- **Juniper** manages network switches — it knows switch ports, packet loss, cable errors, link status etc..

The question this mock data answers is: **How do you connect an OpsRamp server to the Juniper switch port it's physically plugged into?**

### Answer: IP Address (via ARP / MAC Tables on the Switch)

Both systems independently observe the same IP address for a given server:

```
OpsRamp discovers a server:                 Juniper sees on its switch port:
────────────────────────────                ─────────────────────────────────
Name: k8s-node-04                           Port: ge-0/0/5
IP:   10.0.4.24          ◀── SAME IP ──▶   ARP entry: 10.0.4.24
OpsRamp ID: "res-016"                       Switch: sw-dc-east-04
(OpsRamp's own internal ID)                 (Juniper's own switch identifier)
```

**OpsRamp generates its own resource ID (`res-016`). Juniper has its own switch/port IDs (`sw-004`, `ge-0/0/5`). These IDs are NOT the same and cannot be shared.** The only thing both systems independently know about the same machine is its **IP address** (and hostname via LLDP/DNS).

#### How Juniper Knows Which IP Lives on Which Port

Juniper switches maintain two standard Layer 2/3 tables that make this correlation possible:

| Table | What It Stores | How It's Built | Junos CLI |
|-------|---------------|---------------|----------|
| **MAC table** (L2 forwarding table) | Maps a device's MAC address to the physical switch port it was learned on. | The switch passively learns MAC addresses from incoming Ethernet frames — every frame that arrives on a port teaches the switch "this MAC lives here." | `show ethernet-switching table` |
| **ARP table** (L3 neighbor cache) | Maps an IP address to the MAC address that owns it. | When traffic is routed through the switch's Layer 3 interface (VLAN IRB/SVI), the switch sends ARP requests and caches the responses: "IP 10.0.4.24 belongs to MAC `aa:bb:cc:dd:ee:ff`." | `show arp` |

By joining these two tables, the switch can answer: **"IP 10.0.4.24 → MAC `aa:bb:cc:dd:ee:ff` → learned on port `ge-0/0/5`"**. This is how we go from an IP address (which OpsRamp also knows) to a specific physical port (which only Juniper knows).

```
Juniper's internal join:
┌──────────────┐       ┌──────────────────┐       ┌──────────────┐
│  ARP table   │       │   MAC table      │       │ Port health  │
│ IP → MAC     │──MAC──▶│ MAC → Port       │──Port─▶│ ge-0/0/5     │
│ 10.0.4.24 →  │       │ aa:bb:.. →       │       │ Loss: 8.3%   │
│ aa:bb:cc:..  │       │ ge-0/0/5         │       │ Errors: 156K │
└──────────────┘       └──────────────────┘       └──────────────┘
```

> **Key point:** OpsRamp never sees the MAC or port — it only provides the IP. Juniper never sees the OpsRamp resource ID. The IP address is the **sole common key** between the two systems.

In production, a correlation service (like NetBox/DCIM or LLDP neighbor discovery) queries both OpsRamp's resource inventory and Juniper's ARP/MAC tables, then joins the records by matching IP:

```
OpsRamp resource "res-016" (IP: 10.0.4.24)
        │
        ├── matched by IP address 10.0.4.24
        │
Juniper ARP table: 10.0.4.24 → MAC aa:bb:cc:dd:ee:ff
Juniper MAC table: aa:bb:cc:dd:ee:ff → port ge-0/0/5 on sw-dc-east-04
```

**In our mock data, `GetNetworkPortMappings()` in `network.go` IS that correlation service** — it's a hardcoded lookup table where each row says "this OpsRamp resource IP maps to this Juniper switch:port", skipping the ARP/MAC resolution that would happen in production. The `ResourceIP` in the mapping table is deliberately identical to the `IPAddress` in `resources.go`.

---

## Table of Contents

1. [The Five Source Files](#1-the-five-source-files)
2. [How IDs Connect the Files](#2-how-ids-connect-the-files)
3. [The Port Mapping Table — Bridging OpsRamp ↔ Juniper](#3-the-port-mapping-table--bridging-opsramp--juniper)
4. [The Dependency Graph — Who Gets Hurt?](#4-the-dependency-graph--who-gets-hurt)
5. [Walk-Through: "Why is GreenLake portal slow?"](#5-walk-through-why-is-greenlake-portal-slow)
6. [Source File Reference](#6-source-file-reference)
7. [Scope: Physical Juniper Switch Connectivity — Not Limited to On-Prem](#7-scope-physical-juniper-switch-connectivity--not-limited-to-on-prem)

---

## 1. The Five Source Files

Think of each file as a spreadsheet on your desk:

| File | What's in it | Analogy |
|------|-------------|---------|
| `resources.go` | 20+ servers with IPs, metrics (CPU, memory, disk) | **Asset inventory** — list of all your servers |
| `alerts.go` | 9 monitoring alerts, each pointing to a server | **Alarm log** — "server X has a problem" |
| `incidents.go` | 7 incident tickets, linking alerts to servers | **Trouble tickets** — "we opened a ticket for alarm Y" |
| `network.go` | 4 Juniper switches with per-port telemetry + 20 port mappings | **Network wiring diagram** + cable health |
| `dependencies.go` | 34 nodes + 51 edges: switches → servers → apps → users | **Floor plan** — what's connected to what |

---

## 2. How IDs Connect the Files

### Within OpsRamp files: Resource ID (`res-016`)

Inside the OpsRamp world, the same `res-016` string connects a server across alerts and incidents:

```
resources.go          alerts.go               incidents.go
────────────          ─────────               ────────────
ID: "res-016"         Device.ID: "res-016"    ResourceIDs: ["res-016"]
Name: k8s-node-04     ALR-008: pod crash      INC-003: GreenLake down
IP: 10.0.4.24         ALR-009: HTTP slow       └─ AlertIDs: [ALR-008]
```

This is straightforward — OpsRamp owns all three files, so it uses its own ID everywhere.

### Between OpsRamp and Juniper: IP Address (`10.0.4.24`)

This is the critical bridge. OpsRamp and Juniper **cannot** share resource IDs — they're different products from different vendors. What they share is the **IP address** of a physical server:

```
resources.go (OpsRamp)              network.go port mapping            network.go switch telemetry (Juniper)
──────────────────────              ──────────────────────             ──────────────────────────────────────
ID: "res-016"                       ResourceIP: "10.0.4.24"  ◀─────▶  Port ge-0/0/5:
Name: "k8s-node-04"                 ResourceName: "k8s-node-04"         Loss: 8.3%
IPAddress: "10.0.4.24" ──SAME IP──▶ SwitchID: "sw-004"                   RX Errors: 156,789
                                    PortID: "ge-0/0/5"                   Speed: 100Mbps (should be 1000!)
```

The `findMapping()` function in `juniper/client.go` does the lookup — it searches the port mapping table by **IP address** (primary key), hostname, or resource ID (convenience for mock).

---

## 3. The Port Mapping Table — Bridging OpsRamp ↔ Juniper

**Source:** `network.go` → `GetNetworkPortMappings()`

This is the **most important table** in the mock data. It represents what a production DCIM/LLDP system would build dynamically by matching IPs:

```
OpsRamp Side (from resources.go)        Juniper Side (from switch telemetry)
─────────────────────────────────        ────────────────────────────────────
ResourceIP      ResourceName             SwitchName          PortID
──────────      ────────────             ──────────          ──────
10.0.1.101      web-server-prod-01       sw-dc-east-01       ge-0/0/1
10.0.1.102      app-server-prod-01       sw-dc-east-01       ge-0/0/2
10.0.1.103      app-server-prod-02       sw-dc-east-01       ge-0/0/3
10.0.1.104      web-server-prod-02       sw-dc-east-01       ge-0/0/4
10.0.2.10       db-primary-01            sw-dc-east-02       ge-0/0/1
10.0.2.11       db-replica-01            sw-dc-east-02       ge-0/0/2
10.0.2.20       redis-cache-01           sw-dc-east-02       ge-0/0/3
10.0.3.10       rabbitmq-prod-01         sw-dc-east-03       ge-0/0/1
10.0.3.30       elasticsearch-prod-01    sw-dc-east-03       ge-0/0/2
10.0.3.50       api-gateway-prod         sw-dc-east-03       ge-0/0/3
10.0.4.10       k8s-master-01            sw-dc-east-04   ge-0/0/1
10.0.4.21       k8s-node-01              sw-dc-east-04   ge-0/0/2
10.0.4.22       k8s-node-02              sw-dc-east-04   ge-0/0/3
10.0.4.23       k8s-node-03              sw-dc-east-04   ge-0/0/4
10.0.4.24       k8s-node-04              sw-dc-east-04   ge-0/0/5  ◀ KEY
```

### How production builds this table vs. how we mock it

```
PRODUCTION:
  1. OpsRamp agent starts on a server → reports IP 10.0.4.24 to OpsRamp API.
  2. Server sends traffic through Juniper switch port ge-0/0/5.
     - Switch learns MAC from the Ethernet frame → MAC table: aa:bb:.. → ge-0/0/5
     - Switch resolves IP via ARP → ARP table: 10.0.4.24 → aa:bb:..
     Combined: 10.0.4.24 → ge-0/0/5 (the switch now knows the IP-to-port binding).
  3. Correlation service (NetBox / LLDP discovery) queries OpsRamp for resource IPs
     and Juniper for ARP+MAC entries, joins by IP, and produces:
     "OpsRamp resource res-016 (10.0.4.24) ↔ sw-dc-east-04:ge-0/0/5"

OUR MOCK:
  We skip the ARP/MAC resolution and hardcode the final result in
  GetNetworkPortMappings(). The ResourceIP values are manually kept in sync
  with IPAddress in resources.go to faithfully represent the join-by-IP pattern.
```

---

## 4. The Dependency Graph — Who Gets Hurt?

**Source:** `dependencies.go` → `GetDependencyNodes()` + `GetDependencyEdges()`

Once the agent knows which server has a problem, the dependency graph answers: **"If this server goes down, what apps break and how many users are affected?"**

The graph has 4 layers connected by directed edges:

```
LAYER 1: SWITCHES ──connects_to──▶ LAYER 2: SERVERS ──hosts──▶ LAYER 3: APPS ──serves──▶ LAYER 4: USERS
                                                                      │
                                                                depends_on
                                                                      │
                                                                 other APPS
```

### Example: Blast radius of k8s-node-04

```
sw-dc-east-04 (switch)
    │
    │ connects_to (ge-0/0/5 has 8.3% packet loss)
    ▼
k8s-node-04 (server, res-016)
    │
    ├── hosts ──▶ greenlake-portal ──serves──▶ greenlake-tenants (3,000 users)
    │                               ──serves──▶ aruba-wifi-users  (2,000 users)
    │
    ├── hosts ──▶ aruba-central    ──serves──▶ greenlake-tenants
    │                               ──serves──▶ aruba-wifi-users
    │
    └── hosts ──▶ dscc-console     ──serves──▶ greenlake-tenants

TOTAL BLAST RADIUS: 3 apps, 5,000 users, $75K/hr revenue impact
```

### All 4 edge types

| Relationship | From → To | Count | Meaning |
|---|---|---|---|
| `connects_to` | switch → server | 16 | Server is physically cabled to this switch port |
| `hosts` | server → application | 12 | Server runs this application's pods/processes |
| `depends_on` | application → application | 11 | App calls another app (e.g., aruba-central calls greenlake-portal) |
| `serves` | application → user_group | 12 | Users who depend on this app |

### Full graph visualisation

```
                     SWITCHES (Layer 1)
  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐ ┌──────────────────┐
  │sw-dc-east-01 │ │sw-dc-east-02 │ │sw-dc-east-03 │ │sw-dc-east-04 │
  │  (Web/App)   │ │  (Database)  │ │  (Infra)     │ │  (K8s)           │
  └──┬──┬──┬──┬──┘ └──┬──┬──┬────┘ └──┬──┬──┬────┘ └──┬──┬──┬──┬──┬──┘
     │  │  │  │       │  │  │         │  │  │         │  │  │  │  │
     ▼  ▼  ▼  ▼       ▼  ▼  ▼         ▼  ▼  ▼         ▼  ▼  ▼  ▼  ▼
                     SERVERS (Layer 2)
  web01 app01 app02 web02  db-pri db-rep redis  rmq  es   gw   mst n01 n02 n03 n04
    │                │                                     │       │   │   │   │
    ▼    hosts       ▼                          hosts      ▼       ▼   ▼   ▼   ▼
                     APPLICATIONS (Layer 3)
  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐
  │  web-    │  │  user-   │  │ notifi-  │  │monitoring│  │ oneview- │
  │ frontend │  │  auth    │  │ cation   │  │  stack   │  │   api    │
  └────┬─────┘  └──────────┘  └──────────┘  └──────────┘  └────┬─────┘
       │                                                        │
       │  ┌──────────┐         ┌──────────┐    ┌──────────┐     │
       └─▶│   api-   │────────▶│  dscc-   │◀───│  aruba-  │◀────┘
           │ gateway  │         │ console  │    │ central  │
           └────┬─────┘         └────┬─────┘    └────┬─────┘
                │                    │               │
                │        ┌──────────┴──────┐         │
                │        │   greenlake-    │◀────────┘
                │        │    portal       │ depends_on
                │        └────────┬────────┘
                │    serves       │
                ▼                 ▼
                     USERS (Layer 4)
  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
  │  greenlake-  │  │ aruba-wifi-  │  │    api-      │  │  hpe-ops-    │
  │   tenants    │  │    users     │  │ integrations │  │    team      │
  │  3,000 users │  │  2,000 users │  │  150 partners│  │   50 staff   │
  │  $45K/hr     │  │  $30K/hr     │  │  $12K/hr     │  │              │
  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘
```

---

## 5. Walk-Through: "Why is GreenLake portal slow?"

This shows exactly how the agent chains tools and how data flows file-to-file:

```
STEP 1: search_alerts("greenlake")
─────────────────────────────────────────────────────────────────────
alerts.go:  "greenlake" matches ALR-008 (pod crash) and ALR-009 (HTTP slow)
            Both alerts point to Device.ID = "res-016"
            Agent now knows: the problem server is res-016 (k8s-node-04)

STEP 2: investigate_resource("k8s-node-04")
─────────────────────────────────────────────────────────────────────
resources.go:  res-016 → k8s-node-04, IP: 10.0.4.24
               CPU: 45%, Memory: 62% → server metrics look OK
               Agent thinks: "Server is fine, maybe it's a network issue."

STEP 3: correlate_network("k8s-node-04")
─────────────────────────────────────────────────────────────────────
network.go (port mapping):
    findMapping() searches by name "k8s-node-04"
    Match found → ResourceIP: 10.0.4.24, Switch: sw-dc-east-04, Port: ge-0/0/5

network.go (switch telemetry):
    ge-0/0/5 stats → Loss: 8.3%, RX Errors: 156K, Speed: 100Mbps (half-duplex!)
    🔴 ROOT CAUSE FOUND: bad cable or port config causing packet loss

STEP 4: blast_radius("k8s-node-04")
─────────────────────────────────────────────────────────────────────
dependencies.go:
    BFS traversal starting from node "res-016":
      res-016 ──hosts──▶ greenlake-portal ──serves──▶ 3,000 tenants
      res-016 ──hosts──▶ aruba-central    ──serves──▶ 2,000 wifi users
      res-016 ──hosts──▶ dscc-console     ──serves──▶ 3,000 tenants (overlap)

    BLAST RADIUS: 3 apps, 5,000 unique users, $75K/hr

STEP 5: get_remediation_plan("k8s-node-04")
─────────────────────────────────────────────────────────────────────
juniper/client.go:
    Uses same port mapping to generate Junos CLI commands:
      1. show interfaces ge-0/0/5 extensive
      2. request system reboot interface ge-0/0/5 (with approval gate)
      3. set interfaces ge-0/0/5 speed 1g, duplex full
      4. verify counters cleared
```

### Data flow summary

```
alerts.go ──"greenlake"──▶ ALR-008 ──Device.ID──▶ res-016
                                                      │
resources.go ◀───────── res-016 ──IPAddress──▶ 10.0.4.24
                                                      │
network.go port mapping ◀──── IP match ──────▶ sw-004 : ge-0/0/5
                                                      │
network.go switch telemetry ◀────────────────▶ ge-0/0/5: 8.3% loss 🔴
                                                      │
dependencies.go ◀──── BFS from res-016 ──────▶ 3 apps, 5000 users
                                                      │
juniper/client.go ◀── sw-004 : ge-0/0/5 ────▶ Junos CLI remediation
```

---

## 6. Source File Reference

### Deliberately Unhealthy Ports (Demo Scenarios)

| Switch Port | Server | Symptoms | Demo Purpose |
|---|---|---|---|
| sw-dc-east-04 `ge-0/0/5` | k8s-node-04 | 8.3% loss, 156K errors, link flaps, 100Mbps half-duplex | **Primary demo** — network root cause for GreenLake slowness |
| sw-dc-east-04 `ge-0/0/4` | k8s-node-03 | 4.7% packet loss | Correlates with latency alert ALR-005 |
| sw-dc-east-01 `ge-0/0/3` | app-server-prod-02 | 48K RX errors | Memory alert ALR-003 — NOT the network root cause (tests agent reasoning) |
| sw-dc-east-01 `ge-0/0/1` | web-server-prod-01 | Clean port | CPU alert ALR-001 — NOT a network issue (tests agent doesn't blame network) |

### Alert → Resource Mapping

| Alert | Subject | Resource |
|---|---|---|
| ALR-001 | CPU utilization exceeded 95% | res-001 (web-server-prod-01) |
| ALR-002 | Disk usage at 92% | res-005 (db-primary-01) |
| ALR-003 | Memory utilization at 88% | res-003 (app-server-prod-02) |
| ALR-004 | SSL certificate expires in 7 days | res-010 (api-gateway-prod) |
| ALR-005 | Network latency spike 250ms | res-015 (k8s-node-03) |
| ALR-006 | Azure SQL DTU at 85% | res-020 (azure-sql-prod-01) |
| ALR-007 | PING failed | res-025 (monitoring-staging) |
| ALR-008 | Container restart loop (greenlake-portal) | **res-016 (k8s-node-04)** |
| ALR-009 | HTTP response time exceeded 5s | **res-016 (k8s-node-04)** |

### Incident → Alert → Resource Chain

| Incident | Subject | Alerts | Resources |
|---|---|---|---|
| INC-001 | Web tier degraded | ALR-001 | res-001 |
| INC-002 | Database disk critical | ALR-002 | res-005 |
| **INC-003** | **GreenLake portal pod crash loop** | **ALR-008** | **res-016** |
| INC-004 | Staging server unreachable | ALR-007 | res-025 |
| INC-005 | SSL certificate renewal | ALR-004 | res-010 |

### Resource IP Addresses (Same in resources.go and network.go)

| Resource ID | Name | IP Address | Switch Port |
|---|---|---|---|
| res-001 | web-server-prod-01 | 10.0.1.101 | sw-dc-east-01:ge-0/0/1 |
| res-002 | app-server-prod-01 | 10.0.1.102 | sw-dc-east-01:ge-0/0/2 |
| res-003 | app-server-prod-02 | 10.0.1.103 | sw-dc-east-01:ge-0/0/3 |
| res-004 | web-server-prod-02 | 10.0.1.104 | sw-dc-east-01:ge-0/0/4 |
| res-005 | db-primary-01 | 10.0.2.10 | sw-dc-east-02:ge-0/0/1 |
| res-006 | db-replica-01 | 10.0.2.11 | sw-dc-east-02:ge-0/0/2 |
| res-007 | redis-cache-01 | 10.0.2.20 | sw-dc-east-02:ge-0/0/3 |
| res-008 | rabbitmq-prod-01 | 10.0.3.10 | sw-dc-east-03:ge-0/0/1 |
| res-009 | elasticsearch-prod-01 | 10.0.3.30 | sw-dc-east-03:ge-0/0/2 |
| res-010 | api-gateway-prod | 10.0.3.50 | sw-dc-east-03:ge-0/0/3 |
| res-011 | k8s-master-01 | 10.0.4.10 | sw-dc-east-04:ge-0/0/1 |
| res-013 | k8s-node-01 | 10.0.4.21 | sw-dc-east-04:ge-0/0/2 |
| res-014 | k8s-node-02 | 10.0.4.22 | sw-dc-east-04:ge-0/0/3 |
| res-015 | k8s-node-03 | 10.0.4.23 | sw-dc-east-04:ge-0/0/4 |
| **res-016** | **k8s-node-04** | **10.0.4.24** | **sw-dc-east-04:ge-0/0/5** |

---

## 7. Scope: Physical Juniper Switch Connectivity — Not Limited to On-Prem

> **Current status:** The mock data and correlation logic cover **any environment where physical Juniper switches provide network connectivity and an OpsRamp collector is deployed**. This includes HPE GreenLake (HPE-managed cloud), customer-owned datacenters, colocation facilities, and edge sites — not just traditional "on-prem."

### Where Juniper correlation works

The IP-to-port correlation described in this document depends on a direct physical relationship:

```
Physical server NIC ──cable──▶ Juniper switch port
       │                              │
       └── IP: 10.0.4.24              └── ARP/MAC tables resolve to ge-0/0/5
```

This works in **any environment** where:
- Each server has a **static, private IP** on a known subnet.
- The server is **physically cabled** to a Juniper switch port.
- Juniper's ARP and MAC tables reliably map that IP to that port.
- The IP is the **same IP** OpsRamp discovers when the agent reports in.

**Supported environments:**

| Environment | Example | Juniper Correlation? | Why |
|------------|---------|---------------------|-----|
| **HPE GreenLake** | HPE ProLiant DL380 in HPE-managed DC | ✅ Yes | Physical servers on Juniper switches, OpsRamp collector deployed |
| **Customer Datacenter** | On-prem servers in customer rack | ✅ Yes | Same physical topology |
| **Colocation** | Leased rack in Equinix/CyrusOne | ✅ Yes | Customer owns/manages the switches |
| **Edge Site** | Retail store or branch office | ✅ Yes | HPE Aruba/Juniper switching on-site |
| **AWS / Azure / GCP** | EC2 / Azure VM / GCE instance | ❌ No | No physical switch access — virtual network overlays |

### Why pure public cloud doesn't work

Cloud-hosted resources (AWS EC2, Azure VMs, GCP Compute instances) break the physical correlation model:

| Challenge | Physical Infra (On-Prem/GreenLake/Colo) | Pure Public Cloud | Impact |
|-----------|---------|-------|--------|
| **IP stability** | Static private IPs, rarely change | Ephemeral IPs; IPs reassigned on stop/start | IP-based joins become unreliable |
| **Physical port mapping** | Server → switch port is 1:1 | No physical switch port — traffic flows through virtual network overlays (VPC, VNet) | ARP/MAC table correlation doesn't apply |
| **Network telemetry source** | Juniper switch provides per-port packet loss, errors, CRC counts | Cloud providers expose different metrics (VPC Flow Logs, NSG flow data, CloudWatch NetworkIn/Out) | Different data model, different APIs |
| **Network topology** | Known: switch → server → app | Abstracted: VPC → subnet → ENI → instance; load balancers, NAT gateways add hops | Dependency graph structure changes |
| **Blast radius calculation** | BFS through physical switch → server → app graph | Must account for availability zones, auto-scaling groups, managed services | Graph nodes and edge types differ |

### What's needed for pure cloud support

Extending the agent to pure public cloud environments requires exploring the following data sources and mapping strategies:

1. **Cloud provider resource APIs** — Discover instances, their VPC/subnet placement, attached ENIs, and security groups. These replace the physical switch-port mapping.
2. **Cloud network telemetry** — VPC Flow Logs (AWS), NSG Flow Logs (Azure), or VPC Flow Logs (GCP) replace Juniper per-port telemetry. Metrics are per-flow rather than per-port.
3. **Instance identity correlation** — Instead of IP-based joins, use cloud-native identifiers (instance ID, resource ARN) that OpsRamp already collects for cloud-monitored resources.
4. **Managed service dependencies** — Cloud apps often depend on managed services (RDS, Cloud SQL, Azure Service Bus) that have no equivalent in the physical switch/server model. The dependency graph needs new node types.
5. **Dynamic topology** — Auto-scaling, spot instances, and container orchestration (EKS/AKS/GKE) mean the topology changes constantly. Static mock data won't suffice; real-time API queries will be needed.

> **Bottom line:** The Juniper correlation architecture works for **any physical infrastructure with Juniper switches and OpsRamp collectors** — including HPE GreenLake, customer datacenters, and colocation facilities. It is NOT limited to traditional on-prem. Pure public cloud integration is a separate workstream that requires cloud-provider APIs and a different correlation model.

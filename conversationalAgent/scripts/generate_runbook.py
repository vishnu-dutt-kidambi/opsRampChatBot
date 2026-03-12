#!/usr/bin/env python3
"""Generate a sample OpsRamp Operations Runbook PDF for the knowledge base."""

from reportlab.lib.pagesizes import letter
from reportlab.lib.styles import getSampleStyleSheet, ParagraphStyle
from reportlab.lib.enums import TA_LEFT
from reportlab.platypus import SimpleDocTemplate, Paragraph, Spacer, PageBreak
from reportlab.lib.units import inch

def create_runbook():
    doc = SimpleDocTemplate("runbooks/opsramp_operations_runbook.pdf", pagesize=letter,
                            topMargin=0.75*inch, bottomMargin=0.75*inch)
    styles = getSampleStyleSheet()
    title_style = styles['Title']
    h1 = styles['Heading1']
    h2 = styles['Heading2']
    h3 = styles['Heading3']
    body = styles['BodyText']
    body.spaceAfter = 6

    bullet = ParagraphStyle('Bullet', parent=body, leftIndent=20, bulletIndent=10,
                            spaceBefore=2, spaceAfter=2)

    story = []

    # Title
    story.append(Paragraph("OpsRamp Operations Runbook", title_style))
    story.append(Paragraph("Infrastructure Incident Response & Troubleshooting Guide", styles['Heading2']))
    story.append(Spacer(1, 0.3*inch))
    story.append(Paragraph("Version 2.1 — Last Updated: January 2025", body))
    story.append(Spacer(1, 0.5*inch))

    # Section 1: High CPU Usage
    story.append(Paragraph("1. High CPU Usage Runbook", h1))
    story.append(Paragraph("1.1 Overview", h2))
    story.append(Paragraph(
        "High CPU utilization (above 85%) on any production server triggers a Warning alert, "
        "and above 95% triggers a Critical alert in OpsRamp. Sustained high CPU can lead to "
        "application slowdowns, request timeouts, and service degradation.", body))

    story.append(Paragraph("1.2 Immediate Triage Steps", h2))
    story.append(Paragraph("Step 1: Identify the process consuming the most CPU.", body))
    story.append(Paragraph("• SSH into the affected server and run: top -bn1 | head -20", bullet))
    story.append(Paragraph("• Alternatively, use htop for a more interactive view.", bullet))
    story.append(Paragraph("• Check if the high CPU process is a known application process or something unexpected.", bullet))

    story.append(Paragraph("Step 2: Check if it is a known scheduled job.", body))
    story.append(Paragraph("• Review crontab -l for scheduled tasks.", bullet))
    story.append(Paragraph("• Check if a backup job, log rotation, or batch processing task is running.", bullet))
    story.append(Paragraph("• Scheduled jobs like nightly ETL or database backups may spike CPU temporarily.", bullet))

    story.append(Paragraph("Step 3: Check application logs for errors.", body))
    story.append(Paragraph("• Review /var/log/syslog and application-specific logs.", bullet))
    story.append(Paragraph("• Look for stack traces, OOM errors, or infinite loop patterns.", bullet))
    story.append(Paragraph("• Check if a recent deployment may have introduced a regression.", bullet))

    story.append(Paragraph("1.3 Resolution Steps", h2))
    story.append(Paragraph("• If a runaway process is identified, consider restarting the service: systemctl restart <service-name>", bullet))
    story.append(Paragraph("• If a Java application, check for GC pressure: jstat -gc <pid>", bullet))
    story.append(Paragraph("• If a web server (Nginx/Apache), check active connections: netstat -an | grep ESTABLISHED | wc -l", bullet))
    story.append(Paragraph("• For containerized workloads, check container resource limits: docker stats or kubectl top pods", bullet))
    story.append(Paragraph("• If the issue persists, consider scaling horizontally by adding more instances behind the load balancer.", bullet))

    story.append(Paragraph("1.4 Escalation", h2))
    story.append(Paragraph(
        "If CPU remains above 95% for more than 15 minutes after initial triage, escalate to the "
        "Application Engineering team. Include the output of top, recent deployment history, and "
        "any correlated alerts from OpsRamp.", body))

    story.append(PageBreak())

    # Section 2: Disk Space Full
    story.append(Paragraph("2. Disk Space Full Runbook", h1))
    story.append(Paragraph("2.1 Overview", h2))
    story.append(Paragraph(
        "Disk utilization above 80% triggers a Warning alert, and above 90% triggers a Critical alert. "
        "When a disk is full, applications cannot write logs, databases cannot process transactions, "
        "and the system may become unresponsive.", body))

    story.append(Paragraph("2.2 Immediate Triage Steps", h2))
    story.append(Paragraph("Step 1: Identify what is consuming disk space.", body))
    story.append(Paragraph("• Run: df -h to see overall disk usage per mount point.", bullet))
    story.append(Paragraph("• Run: du -sh /var/log/* | sort -rh | head -20 to find largest log files.", bullet))
    story.append(Paragraph("• Check /tmp and /var/tmp for stale temporary files.", bullet))

    story.append(Paragraph("Step 2: Check for log file growth.", body))
    story.append(Paragraph("• Verify log rotation is configured: ls -la /etc/logrotate.d/", bullet))
    story.append(Paragraph("• Check if application logs are being rotated properly.", bullet))
    story.append(Paragraph("• Look for core dumps: find / -name 'core.*' -type f 2>/dev/null", bullet))

    story.append(Paragraph("Step 3: Check for large database files or backups.", body))
    story.append(Paragraph("• Database WAL/binlog files can grow unbounded if replication is lagging.", bullet))
    story.append(Paragraph("• Old database backups may not be cleaned up: ls -lh /backup/", bullet))

    story.append(Paragraph("2.3 Resolution Steps", h2))
    story.append(Paragraph("• Clean up old logs: find /var/log -name '*.gz' -mtime +30 -delete", bullet))
    story.append(Paragraph("• Remove old package cache: apt-get clean or yum clean all", bullet))
    story.append(Paragraph("• Truncate large active log files (if safe): > /var/log/application.log", bullet))
    story.append(Paragraph("• For database servers, purge old binary logs: PURGE BINARY LOGS BEFORE '2025-01-01'", bullet))
    story.append(Paragraph("• If on cloud, consider expanding the EBS volume or attaching additional storage.", bullet))

    story.append(Paragraph("2.4 Prevention", h2))
    story.append(Paragraph(
        "Ensure all servers have log rotation configured with appropriate retention (7-14 days for "
        "application logs, 30 days for system logs). Set up OpsRamp alerts at 70% disk to allow "
        "proactive action before reaching critical levels.", body))

    story.append(PageBreak())

    # Section 3: Memory Leak
    story.append(Paragraph("3. Memory Leak / High Memory Usage Runbook", h1))
    story.append(Paragraph("3.1 Overview", h2))
    story.append(Paragraph(
        "Memory utilization above 85% triggers a Warning, and above 95% triggers Critical. "
        "Memory leaks cause gradual increase in memory consumption over time, eventually leading "
        "to OOM (Out of Memory) kills by the Linux kernel.", body))

    story.append(Paragraph("3.2 Identification", h2))
    story.append(Paragraph("Step 1: Check current memory usage.", body))
    story.append(Paragraph("• Run: free -h to see total, used, and available memory.", bullet))
    story.append(Paragraph("• Run: ps aux --sort=-%mem | head -20 to find top memory consumers.", bullet))
    story.append(Paragraph("• Check /var/log/kern.log for OOM killer messages.", bullet))

    story.append(Paragraph("Step 2: Determine if it is a leak or expected usage.", body))
    story.append(Paragraph("• Compare current memory with OpsRamp 30-day trend data.", bullet))
    story.append(Paragraph("• A steady upward trend without corresponding load increase suggests a leak.", bullet))
    story.append(Paragraph("• Check if the application was recently updated — new versions may have different memory profiles.", bullet))

    story.append(Paragraph("3.3 Resolution Steps", h2))
    story.append(Paragraph("• For Java applications: check heap usage with jmap -heap <pid>", bullet))
    story.append(Paragraph("• Trigger a heap dump for analysis: jmap -dump:format=b,file=heap.hprof <pid>", bullet))
    story.append(Paragraph("• For Node.js: use --max-old-space-size flag and check for event listener leaks.", bullet))
    story.append(Paragraph("• For Go applications: use pprof endpoint if available: go tool pprof http://localhost:6060/debug/pprof/heap", bullet))
    story.append(Paragraph("• As a short-term fix, restart the application to reclaim memory.", bullet))
    story.append(Paragraph("• For persistent issues, engage the development team with memory profiling data.", bullet))

    story.append(Paragraph("3.4 Escalation", h2))
    story.append(Paragraph(
        "If the application is consistently using more than 90% memory and restarts only provide "
        "temporary relief (memory climbs back within hours), this indicates a genuine memory leak. "
        "Create a P2 incident and assign to the application development team.", body))

    story.append(PageBreak())

    # Section 4: Container/K8s Restart Loop
    story.append(Paragraph("4. Container Restart Loop (CrashLoopBackOff) Runbook", h1))
    story.append(Paragraph("4.1 Overview", h2))
    story.append(Paragraph(
        "When a Kubernetes pod enters CrashLoopBackOff state, the container repeatedly starts "
        "and crashes. OpsRamp monitors container restart counts and alerts when a pod exceeds "
        "5 restarts in 10 minutes.", body))

    story.append(Paragraph("4.2 Diagnosis", h2))
    story.append(Paragraph("Step 1: Check pod status and events.", body))
    story.append(Paragraph("• kubectl get pods -n <namespace> to see pod status.", bullet))
    story.append(Paragraph("• kubectl describe pod <pod-name> -n <namespace> for detailed events.", bullet))
    story.append(Paragraph("• Look for 'OOMKilled', 'Error', or 'CrashLoopBackOff' in the status.", bullet))

    story.append(Paragraph("Step 2: Check container logs.", body))
    story.append(Paragraph("• kubectl logs <pod-name> -n <namespace> for current logs.", bullet))
    story.append(Paragraph("• kubectl logs <pod-name> -n <namespace> --previous for logs from the crashed instance.", bullet))
    story.append(Paragraph("• Common causes: missing config/secrets, database connection failures, port conflicts.", bullet))

    story.append(Paragraph("Step 3: Check resource limits.", body))
    story.append(Paragraph("• kubectl get pod <pod-name> -o yaml | grep -A5 resources", bullet))
    story.append(Paragraph("• If the pod is OOMKilled, increase memory limits in the deployment spec.", bullet))
    story.append(Paragraph("• Verify CPU limits are not too restrictive causing throttling.", bullet))

    story.append(Paragraph("4.3 Common Fixes", h2))
    story.append(Paragraph("• Missing ConfigMap/Secret: Verify all referenced configs exist: kubectl get configmap,secret -n <namespace>", bullet))
    story.append(Paragraph("• Image pull failure: Check image name and registry credentials: kubectl get events -n <namespace>", bullet))
    story.append(Paragraph("• Health check failure: Review liveness/readiness probe settings — initial delay may be too short.", bullet))
    story.append(Paragraph("• Resource limits: Increase memory/CPU limits: kubectl edit deployment <name>", bullet))
    story.append(Paragraph("• Rollback a bad deployment: kubectl rollout undo deployment/<name> -n <namespace>", bullet))

    story.append(PageBreak())

    # Section 5: Network Connectivity Issues
    story.append(Paragraph("5. Network Connectivity Issues Runbook", h1))
    story.append(Paragraph("5.1 Overview", h2))
    story.append(Paragraph(
        "Network issues manifest as high latency, packet loss, or complete connectivity loss between "
        "services. OpsRamp monitors network bandwidth (inbound/outbound Mbps) and alerts when "
        "anomalies are detected.", body))

    story.append(Paragraph("5.2 Diagnosis", h2))
    story.append(Paragraph("Step 1: Verify basic connectivity.", body))
    story.append(Paragraph("• Ping the target host: ping -c 5 <target-ip>", bullet))
    story.append(Paragraph("• Check DNS resolution: nslookup <hostname> or dig <hostname>", bullet))
    story.append(Paragraph("• Trace the network path: traceroute <target-ip>", bullet))

    story.append(Paragraph("Step 2: Check for port-level connectivity.", body))
    story.append(Paragraph("• Test specific ports: telnet <host> <port> or nc -zv <host> <port>", bullet))
    story.append(Paragraph("• Verify firewall rules: iptables -L -n or check cloud security groups.", bullet))
    story.append(Paragraph("• For AWS: check VPC security groups and NACLs.", bullet))
    story.append(Paragraph("• For Azure: check NSG rules and Azure Firewall.", bullet))

    story.append(Paragraph("Step 3: Check for bandwidth saturation.", body))
    story.append(Paragraph("• Use OpsRamp network metrics to check if bandwidth is near capacity.", bullet))
    story.append(Paragraph("• Run: iftop or nethogs to see real-time bandwidth usage by process.", bullet))
    story.append(Paragraph("• Check for large file transfers or backup jobs consuming bandwidth.", bullet))

    story.append(Paragraph("5.3 Resolution", h2))
    story.append(Paragraph("• DNS issues: Check /etc/resolv.conf and DNS server health.", bullet))
    story.append(Paragraph("• Firewall blocks: Update security group rules or iptables as needed.", bullet))
    story.append(Paragraph("• Bandwidth saturation: Implement QoS rules or schedule large transfers during off-peak.", bullet))
    story.append(Paragraph("• For persistent issues, engage the Network Engineering team.", bullet))

    story.append(PageBreak())

    # Section 5A: Switch Port & Physical Layer Troubleshooting (Juniper)
    story.append(Paragraph("5A. Switch Port &amp; Physical Layer Troubleshooting (Juniper)", h1))

    story.append(Paragraph("5A.1 Overview", h2))
    story.append(Paragraph(
        "When OpsRamp server metrics (CPU, memory, disk) look healthy but applications are slow or "
        "experiencing timeouts, the root cause is often at the physical network layer — the switch port "
        "connecting the server to the network. Common culprits include degraded cables, failing "
        "transceivers, duplex mismatches, and link instability (flapping). This section covers how to "
        "diagnose and fix these issues using Juniper Junos CLI commands on EX-series switches.", body))

    story.append(Paragraph("5A.2 When to Suspect the Physical/Switch Layer", h2))
    story.append(Paragraph(
        "If the OpsRamp investigate_resource tool shows normal server metrics but the correlate_network "
        "tool reports issues on the connected switch port, the problem is in the network layer between "
        "the server NIC and the switch. Key indicators:", body))
    story.append(Paragraph("• Packet loss above 0.1% on the switch port (critical if above 1%).", bullet))
    story.append(Paragraph("• High RX (receive) errors — indicates CRC/FCS errors from a bad cable, damaged transceiver, or electromagnetic interference (EMI).", bullet))
    story.append(Paragraph("• High TX (transmit) errors — indicates congestion, buffer overruns, or hardware faults on the switch ASIC.", bullet))
    story.append(Paragraph("• Link flaps (port going up/down repeatedly) — the link is physically unstable; usually a loose cable or failing SFP.", bullet))
    story.append(Paragraph("• Speed downgrade (e.g., negotiated at 100Mbps instead of 1000Mbps) — auto-negotiation failure or cable quality issue.", bullet))
    story.append(Paragraph("• Half-duplex mode — should always be full-duplex in a datacenter; half-duplex causes collisions and retransmissions.", bullet))
    story.append(Paragraph("• Elevated latency (above 10ms) or jitter (above 5ms) on a datacenter LAN port — unusual for direct-attached servers.", bullet))

    story.append(Paragraph("5A.3 Diagnosis — Junos CLI Commands", h2))

    story.append(Paragraph("Step 1: Inspect the port in detail.", body))
    story.append(Paragraph("• show interfaces ge-0/0/X extensive — full port stats including error counters, speed, duplex, flap history, and traffic rates.", bullet))
    story.append(Paragraph("• show interfaces ge-0/0/X | match \"errors|loss|flap|speed|duplex\" — filtered view of key health indicators.", bullet))
    story.append(Paragraph("• show interfaces diagnostics optics ge-0/0/X — check SFP/transceiver optical power levels and temperature.", bullet))

    story.append(Paragraph("Step 2: Check ARP and MAC tables to confirm the server is reachable.", body))
    story.append(Paragraph(
        "Juniper switches maintain two key tables that map a server's IP to its physical port. "
        "The ARP table (Layer 3) maps IP addresses to MAC addresses via ARP request/response. "
        "The MAC table (Layer 2) maps MAC addresses to the physical port where they were learned "
        "from incoming Ethernet frames. Together: IP → MAC (ARP) → Port (MAC table).", body))
    story.append(Paragraph("• show arp — verify the server's IP appears with a valid MAC address. A missing entry means the switch cannot reach the server at Layer 3.", bullet))
    story.append(Paragraph("• show ethernet-switching table — confirm the server's MAC is learned on the expected port. If the MAC appears on a different port, there may be a cabling error.", bullet))
    story.append(Paragraph("• show arp | match 10.0.4.24 — look up a specific server IP to confirm it's still active on the switch.", bullet))

    story.append(Paragraph("Step 3: Check for link flap history.", body))
    story.append(Paragraph("• show interfaces ge-0/0/X extensive | match flap — shows when the link last went down and came back up.", bullet))
    story.append(Paragraph("• A flap within the last hour is critical — the link is actively unstable.", bullet))
    story.append(Paragraph("• A flap within the last 24 hours is a warning — the link was recently disrupted.", bullet))

    story.append(Paragraph("Step 4: Check QoS and congestion.", body))
    story.append(Paragraph("• show class-of-service interface ge-0/0/X — see if QoS policies are throttling traffic on the port.", bullet))
    story.append(Paragraph("• show interfaces queue ge-0/0/X — view per-queue drop counts; high drops indicate congestion.", bullet))

    story.append(Paragraph("5A.4 Remediation Steps", h2))

    story.append(Paragraph("Fix 1: Clear error counters and monitor.", body))
    story.append(Paragraph("• clear interfaces statistics ge-0/0/X — resets all error counters to zero.", bullet))
    story.append(Paragraph("• Wait 5 minutes and re-check: show interfaces ge-0/0/X extensive.", bullet))
    story.append(Paragraph("• If errors reappear, the problem is active (not historical). Proceed to Fix 2.", bullet))

    story.append(Paragraph("Fix 2: Reseat the physical cable.", body))
    story.append(Paragraph("• Physically disconnect and reconnect the cable at both ends — the switch port and the server NIC.", bullet))
    story.append(Paragraph("• This resolves loose connections, oxidized contacts, and partially seated SFP transceivers.", bullet))
    story.append(Paragraph("• After reseating, verify the link comes up: show interfaces ge-0/0/X terse.", bullet))
    story.append(Paragraph("• IMPORTANT: This causes a brief link interruption (5-15 seconds). Coordinate with the application team.", bullet))

    story.append(Paragraph("Fix 3: Bounce the interface (disable/enable).", body))
    story.append(Paragraph("• set interfaces ge-0/0/X disable → commit → delete interfaces ge-0/0/X disable → commit", bullet))
    story.append(Paragraph("• This forces the port to fully renegotiate speed, duplex, and clear all internal state.", bullet))
    story.append(Paragraph("• Risk: medium — causes 15-30 seconds of downtime on that port. Requires approval from on-call network engineer.", bullet))

    story.append(Paragraph("Fix 4: Force speed and duplex (for mismatch issues).", body))
    story.append(Paragraph("• If the port negotiated at 100Mbps half-duplex instead of 1000Mbps full-duplex:", bullet))
    story.append(Paragraph("• set interfaces ge-0/0/X speed 1g", bullet))
    story.append(Paragraph("• set interfaces ge-0/0/X link-mode full-duplex", bullet))
    story.append(Paragraph("• commit", bullet))
    story.append(Paragraph("• Verify: show interfaces ge-0/0/X | match \"speed|duplex\"", bullet))
    story.append(Paragraph("• Risk: medium — may cause a brief link bounce. If the cable cannot support 1Gbps, the link will not come up.", bullet))

    story.append(Paragraph("Fix 5: Replace cable or transceiver.", body))
    story.append(Paragraph("• If errors persist after clearing counters and reseating, the cable or SFP is physically damaged.", bullet))
    story.append(Paragraph("• Check transceiver health: show interfaces diagnostics optics ge-0/0/X — look for low RX power or high temperature.", bullet))
    story.append(Paragraph("• Replace with a known-good cable/SFP and re-test.", bullet))

    story.append(Paragraph("5A.5 Correlating OpsRamp Alerts with Switch Telemetry", h2))
    story.append(Paragraph(
        "When troubleshooting an application issue, always cross-reference OpsRamp alerts with "
        "Juniper switch port data. The key correlation is via IP address — both systems independently "
        "know the server's IP, which is the common key between the OpsRamp resource inventory and "
        "the Juniper ARP/MAC tables.", body))
    story.append(Paragraph("• Step 1: Find the server from OpsRamp alerts — e.g., ALR-008 points to res-016 (k8s-node-04, IP 10.0.4.24).", bullet))
    story.append(Paragraph("• Step 2: Look up the IP in the Juniper ARP table: show arp | match 10.0.4.24 → gives you the MAC address.", bullet))
    story.append(Paragraph("• Step 3: Find the port in the MAC table: show ethernet-switching table | match &lt;mac-address&gt; → gives you the physical port (e.g., ge-0/0/5).", bullet))
    story.append(Paragraph("• Step 4: Inspect port health: show interfaces ge-0/0/5 extensive — check for errors, loss, flaps.", bullet))
    story.append(Paragraph("• Interpretation: If OpsRamp server metrics are normal (CPU/memory/disk OK) but the switch port shows errors, the network layer is the root cause.", bullet))
    story.append(Paragraph("• Interpretation: If both OpsRamp metrics AND switch port are healthy, the issue is at the application layer (code bug, config error, external dependency).", bullet))

    story.append(Paragraph("5A.6 Escalation", h2))
    story.append(Paragraph(
        "If switch port issues persist after interface bounce and cable reseat, escalate to the "
        "Network Engineering team (network-ops@company.com). Include: switch name, port ID, "
        "output of 'show interfaces extensive', optics diagnostics, and the OpsRamp correlation "
        "findings. For critical-impact issues affecting customer-facing services, page the on-call "
        "network engineer immediately via PagerDuty.", body))

    story.append(PageBreak())

    # Section 6: Database Performance
    story.append(Paragraph("6. Database Performance Degradation Runbook", h1))
    story.append(Paragraph("6.1 Overview", h2))
    story.append(Paragraph(
        "Database performance issues affect all applications that depend on the database. Common "
        "symptoms include slow query response times, connection pool exhaustion, replication lag, "
        "and lock contention.", body))

    story.append(Paragraph("6.2 Diagnosis", h2))
    story.append(Paragraph("Step 1: Check database health metrics.", body))
    story.append(Paragraph("• Use OpsRamp to review CPU, memory, disk I/O, and connection counts for the database server.", bullet))
    story.append(Paragraph("• For MySQL: SHOW PROCESSLIST; to see active queries.", bullet))
    story.append(Paragraph("• For PostgreSQL: SELECT * FROM pg_stat_activity WHERE state = 'active';", bullet))

    story.append(Paragraph("Step 2: Identify slow queries.", body))
    story.append(Paragraph("• Check slow query log: /var/log/mysql/slow-query.log", bullet))
    story.append(Paragraph("• For PostgreSQL: check pg_stat_statements extension for query statistics.", bullet))
    story.append(Paragraph("• Look for queries without proper indexing: EXPLAIN ANALYZE <query>", bullet))

    story.append(Paragraph("Step 3: Check replication status.", body))
    story.append(Paragraph("• For MySQL: SHOW SLAVE STATUS\\G — check Seconds_Behind_Master.", bullet))
    story.append(Paragraph("• For PostgreSQL: SELECT * FROM pg_stat_replication;", bullet))
    story.append(Paragraph("• Replication lag above 60 seconds is a concern for read-heavy applications.", bullet))

    story.append(Paragraph("6.3 Resolution Steps", h2))
    story.append(Paragraph("• Kill long-running queries if they are blocking: KILL <process_id>;", bullet))
    story.append(Paragraph("• Add missing indexes based on EXPLAIN output.", bullet))
    story.append(Paragraph("• Increase connection pool limits if connections are exhausted.", bullet))
    story.append(Paragraph("• For lock contention, identify and optimize the conflicting transactions.", bullet))
    story.append(Paragraph("• Consider read replicas for read-heavy workloads.", bullet))

    story.append(Paragraph("6.4 Escalation", h2))
    story.append(Paragraph(
        "If database performance does not improve after basic optimizations, escalate to the "
        "DBA team. Provide: current query patterns, slow query log, EXPLAIN plans for problematic "
        "queries, and OpsRamp metric trends for the past 24 hours.", body))

    story.append(PageBreak())

    # Section 7: SSL Certificate Expiry
    story.append(Paragraph("7. SSL/TLS Certificate Expiry Runbook", h1))
    story.append(Paragraph("7.1 Overview", h2))
    story.append(Paragraph(
        "SSL certificates must be renewed before expiry to avoid service disruption. OpsRamp "
        "monitors certificate expiry and alerts at 30 days (Warning) and 7 days (Critical) "
        "before expiration.", body))

    story.append(Paragraph("7.2 Verification", h2))
    story.append(Paragraph("Step 1: Check certificate details.", body))
    story.append(Paragraph("• openssl s_client -connect <host>:443 -servername <hostname> 2>/dev/null | openssl x509 -noout -dates", bullet))
    story.append(Paragraph("• Verify the issuer and Subject Alternative Names (SANs).", bullet))
    story.append(Paragraph("• Check if the certificate chain is complete: openssl s_client -connect <host>:443 -showcerts", bullet))

    story.append(Paragraph("7.3 Renewal Steps", h2))
    story.append(Paragraph("• For Let's Encrypt: certbot renew --dry-run then certbot renew", bullet))
    story.append(Paragraph("• For purchased certificates: submit CSR to certificate authority.", bullet))
    story.append(Paragraph("• After renewal, restart the web server: systemctl restart nginx", bullet))
    story.append(Paragraph("• Verify the new certificate: echo | openssl s_client -connect <host>:443 2>/dev/null | openssl x509 -noout -dates", bullet))
    story.append(Paragraph("• Update certificate in load balancer (AWS ALB, Azure App Gateway) if applicable.", bullet))

    story.append(PageBreak())

    # Section 8: Alert Response Matrix
    story.append(Paragraph("8. Alert Response Matrix", h1))
    story.append(Paragraph("This matrix maps OpsRamp alert types to the appropriate response procedure:", body))
    story.append(Spacer(1, 0.2*inch))

    story.append(Paragraph("CPU > 95% Critical — Follow Section 1 (High CPU Usage). Response time: 15 minutes.", body))
    story.append(Paragraph("CPU > 85% Warning — Monitor for 30 minutes. If sustained, follow Section 1.", body))
    story.append(Paragraph("Disk > 90% Critical — Follow Section 2 (Disk Space). Response time: 30 minutes.", body))
    story.append(Paragraph("Disk > 80% Warning — Schedule cleanup within 24 hours per Section 2.", body))
    story.append(Paragraph("Memory > 95% Critical — Follow Section 3 (Memory Leak). Response time: 15 minutes.", body))
    story.append(Paragraph("Memory > 85% Warning — Review trends. If increasing, follow Section 3.", body))
    story.append(Paragraph("Container CrashLoop — Follow Section 4 (Container Restart). Response time: 10 minutes.", body))
    story.append(Paragraph("Network Anomaly — Follow Section 5 (Network Issues). Response time: 15 minutes.", body))
    story.append(Paragraph("Switch Port Errors / Link Flaps — Follow Section 5A (Switch Port Troubleshooting). Response time: 15 minutes.", body))
    story.append(Paragraph("Packet Loss on Switch Port — Follow Section 5A (Switch Port Troubleshooting). Correlate with OpsRamp alerts. Response time: 10 minutes.", body))
    story.append(Paragraph("Duplex Mismatch / Speed Downgrade — Follow Section 5A, Fix 4 (Force Speed/Duplex). Response time: 15 minutes.", body))
    story.append(Paragraph("DB Slow Queries — Follow Section 6 (Database Performance). Response time: 30 minutes.", body))
    story.append(Paragraph("SSL Expiry < 7 days — Follow Section 7 (SSL Certificate). Response time: 4 hours.", body))

    story.append(Spacer(1, 0.3*inch))
    story.append(Paragraph("8.1 Escalation Contacts", h2))
    story.append(Paragraph("• Infrastructure Team: infra-ops@company.com — for server/VM issues.", bullet))
    story.append(Paragraph("• Application Engineering: app-eng@company.com — for application-level issues.", bullet))
    story.append(Paragraph("• DBA Team: dba-team@company.com — for database performance issues.", bullet))
    story.append(Paragraph("• Network Engineering: network-ops@company.com — for network-level issues.", bullet))
    story.append(Paragraph("• Security Team: security@company.com — for SSL/TLS and security-related issues.", bullet))
    story.append(Paragraph("• On-Call Manager: +1-555-OPS-CALL — for P1 escalation outside business hours.", bullet))

    story.append(PageBreak())

    # Section 9: General Best Practices
    story.append(Paragraph("9. General Troubleshooting Best Practices", h1))
    story.append(Paragraph("9.1 Before Making Changes", h2))
    story.append(Paragraph("• Always check OpsRamp for correlated alerts on other resources — the root cause may be elsewhere.", bullet))
    story.append(Paragraph("• Review recent change history — deployments, config changes, and infrastructure modifications.", bullet))
    story.append(Paragraph("• Document your findings before applying fixes.", bullet))
    story.append(Paragraph("• For production systems, follow the change management process.", bullet))

    story.append(Paragraph("9.2 During Incident Response", h2))
    story.append(Paragraph("• Communicate status updates every 15 minutes during active incidents.", bullet))
    story.append(Paragraph("• Use the OpsRamp incident timeline to log all actions taken.", bullet))
    story.append(Paragraph("• Prioritize service restoration over root cause analysis.", bullet))
    story.append(Paragraph("• If in doubt, escalate early rather than late.", bullet))

    story.append(Paragraph("9.3 After Resolution", h2))
    story.append(Paragraph("• Verify the alert is resolved in OpsRamp.", bullet))
    story.append(Paragraph("• Update the incident with resolution details.", bullet))
    story.append(Paragraph("• Schedule a post-incident review for P1/P2 incidents.", bullet))
    story.append(Paragraph("• Update this runbook if new procedures were discovered.", bullet))

    doc.build(story)
    print("Generated: runbooks/opsramp_operations_runbook.pdf")

if __name__ == "__main__":
    import os
    os.makedirs("runbooks", exist_ok=True)
    create_runbook()

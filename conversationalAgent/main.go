package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"opsramp-agent/agent"
	"opsramp-agent/juniper"
	"opsramp-agent/knowledge"
	"opsramp-agent/mcpserver"
	"opsramp-agent/mockdata"
	"opsramp-agent/opsramp"
)

func main() {
	ollamaHost := getEnv("OLLAMA_HOST", "http://localhost:11434")
	llmModel := getEnv("LLM_MODEL", "llama3.1")

	// Check for --mcp early — in stdio MCP mode, stdout IS the protocol transport,
	// so we must not print banners or status messages to stdout.
	quietMode := false
	for _, arg := range os.Args[1:] {
		if arg == "--mcp" || arg == "-mcp" {
			quietMode = true
			break
		}
	}

	// printf helper that respects quiet mode (prints to stderr in MCP mode)
	logf := func(format string, a ...interface{}) {
		if quietMode {
			fmt.Fprintf(os.Stderr, format, a...)
		} else {
			fmt.Printf(format, a...)
		}
	}
	logln := func(a ...interface{}) {
		if quietMode {
			fmt.Fprintln(os.Stderr, a...)
		} else {
			fmt.Println(a...)
		}
	}

	logln("╔══════════════════════════════════════════════════════════════╗")
	logln("║              HPE Autopilot                                  ║")
	logln("║         Powered by Ollama + Tool-Calling LLM               ║")
	logln("╠══════════════════════════════════════════════════════════════╣")
	logf("║  LLM:    %-49s ║\n", llmModel)
	logf("║  Ollama: %-49s ║\n", ollamaHost)
	logln("╚══════════════════════════════════════════════════════════════╝")
	logln()

	// Load mock data
	alerts := mockdata.GetAlerts()
	resources := mockdata.GetResources()
	incidents := mockdata.GetIncidents()
	metricHistory := mockdata.GetMetricHistory()
	networkSwitches := mockdata.GetNetworkSwitches()
	portMappings := mockdata.GetNetworkPortMappings()
	depNodes := mockdata.GetDependencyNodes()
	depEdges := mockdata.GetDependencyEdges()

	logf("  Loaded mock environment:\n")
	logf("    Resources: %d (AWS, Azure, GCP, On-Premise)\n", len(resources))

	criticalCount := 0
	warningCount := 0
	for _, a := range alerts {
		switch a.CurrentState {
		case "Critical":
			criticalCount++
		case "Warning":
			warningCount++
		}
	}
	logf("    Alerts:    %d (%d critical, %d warning)\n", len(alerts), criticalCount, warningCount)

	openCount := 0
	for _, i := range incidents {
		if i.Status == "Open" {
			openCount++
		}
	}
	logf("    Incidents: %d (%d open)\n", len(incidents), openCount)

	// Count unique resources with metric history
	metricResources := make(map[string]bool)
	for _, ms := range metricHistory {
		metricResources[ms.ResourceID] = true
	}
	logf("    Metric History: %d series across %d resources (30-day)\n", len(metricHistory), len(metricResources))

	// Count network stats
	totalPorts := 0
	for _, sw := range networkSwitches {
		totalPorts += len(sw.Ports)
	}
	logf("    Network Switches: %d (Juniper Mist) with %d ports mapped\n", len(networkSwitches), len(portMappings))
	logf("    Dependency Graph: %d nodes, %d edges (blast radius topology)\n", len(depNodes), len(depEdges))
	logln()

	// Create the OpsRamp client with mock data
	client := opsramp.NewClient(alerts, resources, incidents, metricHistory)

	// Create the Juniper Mist network client
	juniperClient := juniper.NewClient(networkSwitches, portMappings)
	juniperClient.SetDependencyGraph(depNodes, depEdges)

	// Create the agent
	opsAgent := agent.NewAgent(ollamaHost, llmModel, client)
	opsAgent.SetJuniperClient(juniperClient)

	// Load knowledge base (operations runbooks)
	embeddingModel := getEnv("EMBEDDING_MODEL", "nomic-embed-text")
	runbookDir := getEnv("RUNBOOK_DIR", "runbooks")
	kb := knowledge.NewKnowledgeBase(ollamaHost, embeddingModel)

	// In MCP stdio mode, the RAG package prints progress to stdout which would
	// corrupt the JSON-RPC stream. Temporarily redirect stdout to stderr.
	if quietMode {
		origStdout := os.Stdout
		os.Stdout = os.Stderr
		defer func() { os.Stdout = origStdout }()
		loadKB(kb, runbookDir, opsAgent, logf)
		os.Stdout = origStdout
	} else {
		loadKB(kb, runbookDir, opsAgent, logf)
	}
	logln()

	// Check for --web, --mcp, --mcp-http flags
	webMode := false
	mcpMode := false
	mcpHTTPMode := false
	webAddr := ":8080"
	mcpHTTPAddr := ":8081"
	for i, arg := range os.Args[1:] {
		switch arg {
		case "--web", "-web":
			webMode = true
		case "--mcp", "-mcp":
			mcpMode = true
		case "--mcp-http", "-mcp-http":
			mcpHTTPMode = true
		case "--port", "-port":
			if i+1 < len(os.Args[1:])-1 {
				webAddr = ":" + os.Args[i+2]
			}
		case "--mcp-port", "-mcp-port":
			if i+1 < len(os.Args[1:])-1 {
				mcpHTTPAddr = ":" + os.Args[i+2]
			}
		}
	}

	if mcpMode {
		// MCP server mode — stdio transport (for Claude Desktop, VS Code, etc.)
		if err := mcpserver.StartStdio(client, kb, juniperClient); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if mcpHTTPMode {
		// MCP server mode — HTTP transport (for remote MCP clients)
		fmt.Printf("  🔌 Starting MCP HTTP server on %s\n\n", mcpHTTPAddr)
		if err := mcpserver.StartHTTP(mcpHTTPAddr, client, kb, juniperClient); err != nil {
			fmt.Fprintf(os.Stderr, "MCP HTTP server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if webMode {
		startWebServer(webAddr, opsAgent)
		return
	}

	fmt.Println("  Type 'help' for example questions, 'clear' to reset, 'quit' to exit.")
	fmt.Println("  Tip: Run with --web to launch the browser-based chat UI.")
	fmt.Println("  Tip: Run with --mcp to start as an MCP server (stdio transport).")
	fmt.Println("  ─────────────────────────────────────────────────────────────────")
	fmt.Println()

	scanner := bufio.NewScanner(scanner_stdin())
	for {
		fmt.Print("You > ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		switch strings.ToLower(input) {
		case "quit", "exit", "q":
			fmt.Println("Goodbye!")
			return
		case "clear", "reset":
			opsAgent.ClearHistory()
			fmt.Println("  [Conversation history cleared]")
			fmt.Println()
			continue
		case "help", "?":
			printHelp()
			continue
		}

		fmt.Println()
		answer, err := opsAgent.Ask(input)
		if err != nil {
			fmt.Printf("  Error: %v\n\n", err)
			continue
		}

		fmt.Println()
		fmt.Println("Agent >", answer)
		fmt.Println()
	}
}

func scanner_stdin() *os.File {
	return os.Stdin
}

func printHelp() {
	fmt.Println()
	fmt.Println("  ╔═══════════════════════════════════════════════════════════╗")
	fmt.Println("  ║                    Example Questions                     ║")
	fmt.Println("  ╠═══════════════════════════════════════════════════════════╣")
	fmt.Println("  ║                                                         ║")
	fmt.Println("  ║  ALERTS:                                                ║")
	fmt.Println("  ║  • Show me all critical alerts                          ║")
	fmt.Println("  ║  • Are there any alerts about CPU usage?                ║")
	fmt.Println("  ║  • What warnings do we have right now?                  ║")
	fmt.Println("  ║                                                         ║")
	fmt.Println("  ║  RESOURCES:                                             ║")
	fmt.Println("  ║  • List all resources running in AWS                    ║")
	fmt.Println("  ║  • Show me servers in critical state                    ║")
	fmt.Println("  ║  • What do we have in GCP us-central1?                  ║")
	fmt.Println("  ║                                                         ║")
	fmt.Println("  ║  INVESTIGATION:                                         ║")
	fmt.Println("  ║  • Investigate web-server-prod-01                       ║")
	fmt.Println("  ║  • What's going on with the database server?            ║")
	fmt.Println("  ║  • Why is the production site slow?                     ║")
	fmt.Println("  ║                                                         ║")
	fmt.Println("  ║  INCIDENTS:                                             ║")
	fmt.Println("  ║  • Show me all open incidents                           ║")
	fmt.Println("  ║  • Any P1 incidents right now?                          ║")
	fmt.Println("  ║  • Which incidents have SLA breaches?                   ║")
	fmt.Println("  ║                                                         ║")
	fmt.Println("  ║  OVERVIEW:                                              ║")
	fmt.Println("  ║  • Give me an environment summary                       ║")
	fmt.Println("  ║  • What's the overall health of our infrastructure?     ║")
	fmt.Println("  ║                                                         ║")
	fmt.Println("  ║  FORECASTING:                                           ║")
	fmt.Println("  ║  • Predict capacity for web-server-prod-01              ║")
	fmt.Println("  ║  • When will db-primary-01 run out of disk?             ║")
	fmt.Println("  ║  • Show CPU forecast for all monitored resources        ║")
	fmt.Println("  ║  • Which servers are at risk of running out of capacity? ║")
	fmt.Println("  ║                                                         ║")
	fmt.Println("  ║  KNOWLEDGE BASE / RUNBOOKS:                             ║")
	fmt.Println("  ║  • What is the runbook for high CPU usage?              ║")
	fmt.Println("  ║  • How do I troubleshoot disk space full?               ║")
	fmt.Println("  ║  • What are the escalation contacts?                    ║")
	fmt.Println("  ║  • How to fix a container CrashLoopBackOff?             ║")
	fmt.Println("  ║                                                         ║")
	fmt.Println("  ║  NETWORK / BLAST RADIUS / REMEDIATION:                  ║")
	fmt.Println("  ║  • Correlate network for k8s-node-04                    ║")
	fmt.Println("  ║  • What is the blast radius for k8s-node-04?            ║")
	fmt.Println("  ║  • How many users are affected by the k8s-node-04 issue?║")
	fmt.Println("  ║  • Give me a remediation plan for k8s-node-04           ║")
	fmt.Println("  ║  • Why is the GreenLake portal slow?                    ║")
	fmt.Println("  ║                                                         ║")
	fmt.Println("  ╚═══════════════════════════════════════════════════════════╝")
	fmt.Println()
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}

// loadKB loads all PDF runbooks from the given directory into the knowledge base.
func loadKB(kb *knowledge.KnowledgeBase, runbookDir string, opsAgent *agent.Agent, logf func(string, ...interface{})) {
	entries, err := os.ReadDir(runbookDir)
	if err != nil {
		logf("  Knowledge Base: %s/ directory not found (skipping)\n", runbookDir)
		return
	}
	pdfCount := 0
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".pdf") {
			pdfPath := filepath.Join(runbookDir, entry.Name())
			if err := kb.LoadPDF(pdfPath); err != nil {
				logf("  Warning: failed to load %s: %v\n", entry.Name(), err)
			} else {
				pdfCount++
			}
		}
	}
	if pdfCount > 0 {
		opsAgent.SetKnowledgeBase(kb)
		logf("  Knowledge Base: %d runbook(s) loaded\n", pdfCount)
	} else {
		logf("  Knowledge Base: no runbook PDFs found in %s/\n", runbookDir)
	}
}

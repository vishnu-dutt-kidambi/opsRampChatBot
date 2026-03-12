package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"sync"

	"opsramp-agent/agent"
)

//go:embed web/index.html
var webContent embed.FS

type chatReqBody struct {
	Message string `json:"message"`
}

// startWebServer launches the HTTP server with a chat API and embedded UI.
func startWebServer(addr string, opsAgent *agent.Agent) {
	var mu sync.Mutex // serialize agent calls (single-user agent)

	// Serve the embedded HTML
	webFS, _ := fs.Sub(webContent, "web")
	http.Handle("/", http.FileServer(http.FS(webFS)))

	// Streaming chat endpoint — Server-Sent Events (SSE)
	http.HandleFunc("/api/chat/stream", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}

		var req chatReqBody
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON body", http.StatusBadRequest)
			return
		}
		if req.Message == "" {
			http.Error(w, "message is required", http.StatusBadRequest)
			return
		}

		// SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		mu.Lock()
		opsAgent.AskStream(req.Message, func(evt agent.StreamEvent) {
			data, _ := json.Marshal(evt)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		})
		mu.Unlock()
	})

	// Clear conversation endpoint
	http.HandleFunc("/api/clear", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		mu.Lock()
		opsAgent.ClearHistory()
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"cleared"}`))
	})

	fmt.Printf("  Web UI available at: http://localhost%s\n", addr)
	fmt.Printf("  API endpoint:        http://localhost%s/api/chat\n", addr)
	fmt.Println()
	fmt.Println("  Press Ctrl+C to stop the server.")

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

# =============================================================================
# Makefile - OpsRamp ChatBot
# =============================================================================
# Run 'make help' to see all available commands.
#
# Three modes:
#   CLI (default)   — Interactive terminal REPL
#   Web             — Browser-based chat UI on http://localhost:8080
#   MCP             — Model Context Protocol server (stdio or HTTP)
#
# Two environments:
#   Native (recommended on Mac) — Ollama runs directly on your machine
#   Docker — everything runs in containers
# =============================================================================

.PHONY: setup run build clean help setup-docker run-docker

# Default target
help: ## Show this help message
	@echo ""
	@echo "OpsRamp ChatBot - Available Commands:"
	@echo "=================================================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36mmake %-15s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Quick Start (Native — recommended for Mac):"
	@echo "  1. make setup     # Install Ollama + download Mistral model"
	@echo "  2. make run       # Start the agent in terminal mode"
	@echo "  3. make web       # Start the agent with browser chat UI"
	@echo "  4. make mcp       # Start as MCP server (stdio, for Claude Desktop/VS Code)"
	@echo "  5. make mcp-http  # Start as MCP server (HTTP, for remote clients)"
	@echo ""
	@echo "Quick Start (Docker):"
	@echo "  1. make setup-docker  # Start Ollama container + pull models"
	@echo "  2. make run-docker    # Run the agent in a container"
	@echo ""

# =============================================================================
# NATIVE MODE (recommended for Mac — uses Apple GPU)
# =============================================================================

setup: ## [FIRST TIME] Install Ollama natively and download models (~5GB total)
	@echo "=========================================="
	@echo "  Setting up OpsRamp ChatBot"
	@echo "=========================================="
	@echo ""
	@if command -v ollama > /dev/null 2>&1; then \
		echo "✅ Ollama is already installed: $$(ollama --version)"; \
	else \
		echo "📥 Installing Ollama via Homebrew..."; \
		brew install ollama; \
		echo "✅ Ollama installed!"; \
	fi
	@echo ""
	@echo "🚀 Starting Ollama server in the background..."
	@if curl -sf http://localhost:11434/api/tags > /dev/null 2>&1; then \
		echo "  ✅ Ollama is already running."; \
	else \
		ollama serve > /dev/null 2>&1 & \
		echo "  ⏳ Waiting for Ollama to start..."; \
		sleep 3; \
		until curl -sf http://localhost:11434/api/tags > /dev/null 2>&1; do \
			sleep 2; \
			echo "    still waiting..."; \
		done; \
		echo "  ✅ Ollama server is running!"; \
	fi
	@echo ""
	@echo "📥 Pulling LLM model: mistral (~4.1GB)..."
	@echo "   (This is the AI model that understands your questions and calls tools)"
	ollama pull mistral
	@echo ""
	@echo "📥 Pulling embedding model: nomic-embed-text (~274MB)..."
	@echo "   (This model generates embeddings for the knowledge base RAG pipeline)"
	ollama pull nomic-embed-text
	@echo ""
	@echo "============================================"
	@echo "✅ Setup complete!"
	@echo ""
	@echo "Next: make run"
	@echo "============================================"

run: build ## Run the conversational agent natively (terminal mode)
	@echo ""
	./opsramp-agent

web: build ## Run the agent with browser-based chat UI on http://localhost:8080
	@echo ""
	@echo "🌐 Starting web UI at http://localhost:8080"
	@echo "   Press Ctrl+C to stop."
	@echo ""
	./opsramp-agent --web

mcp: build ## Run as MCP server (stdio transport — for Claude Desktop, VS Code, etc.)
	@echo ""
	@echo "🔌 Starting MCP server (stdio transport)"
	@echo "   Connect via Claude Desktop or VS Code MCP config."
	@echo ""
	./opsramp-agent --mcp

mcp-http: build ## Run as MCP server (HTTP transport on http://localhost:8081)
	@echo ""
	@echo "🔌 Starting MCP HTTP server at http://localhost:8081"
	@echo "   Press Ctrl+C to stop."
	@echo ""
	./opsramp-agent --mcp-http

build: ## Build the Go binary
	@echo "🔨 Building opsramp-agent..."
	go build -o opsramp-agent .
	@echo "✅ Built: ./opsramp-agent"

clean: ## Remove build artifacts
	rm -f opsramp-agent
	@echo "✅ Cleaned."

# =============================================================================
# DOCKER MODE
# =============================================================================

setup-docker: ## Start Ollama in Docker and pull models
	@echo "🐳 Starting Ollama container..."
	docker compose up -d ollama
	@echo ""
	@echo "⏳ Waiting for Ollama to start..."
	@until curl -sf http://localhost:11434/api/tags > /dev/null 2>&1; do \
		sleep 2; \
		echo "    still waiting..."; \
	done
	@echo "✅ Ollama is running!"
	@echo ""
	@echo "📥 Pulling Mistral model..."
	docker compose exec ollama ollama pull mistral
	@echo ""
	@echo "📥 Pulling nomic-embed-text model..."
	docker compose exec ollama ollama pull nomic-embed-text
	@echo ""
	@echo "✅ Docker setup complete! Run: make run-docker"

run-docker: ## Run the agent in Docker
	docker compose up --build app

# =============================================================================
# DEVELOPMENT
# =============================================================================

test: ## Run any tests
	go test ./...

fmt: ## Format Go code
	go fmt ./...

vet: ## Run Go vet
	go vet ./...

lint: fmt vet ## Run all linters

# =============================================================================
# Makefile - OpsRamp ChatBot (Root — Docker orchestration)
# =============================================================================
# Manages the full dockerized stack: Ollama + Conversational Agent
#
# Two Docker modes:
#   Mac (recommended) — App in Docker, Ollama native on host (fast, GPU)
#   Full Docker       — Everything in containers (CPU-only, slow on Mac)
#
# Usage:
#   make docker-web-mac — Recommended for Mac (uses native Ollama + Apple GPU)
#   make docker-web     — Full Docker mode (CPU-only, ~60s per response)
#   make docker-setup   — Start dockerized Ollama and pull models
#   make docker-mcp     — MCP HTTP server on http://localhost:8081
#   make docker-down    — Stop all containers
#   make docker-clean   — Stop containers and remove images/volumes
#
# For native (non-Docker) usage, see conversationalAgent/Makefile
# =============================================================================

.PHONY: help docker-setup docker-web docker-web-mac docker-mcp docker-cli docker-down docker-clean docker-build docker-logs

help: ## Show available commands
	@echo ""
	@echo "OpsRamp ChatBot — Docker Commands"
	@echo "=================================================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36mmake %-18s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Quick Start (Mac — recommended, fast):"
	@echo "  1. Ensure Ollama is running natively: ollama serve"
	@echo "  2. make docker-web-mac  # Web UI using native Ollama (GPU)"
	@echo ""
	@echo "Quick Start (Full Docker — portable, slower on Mac):"
	@echo "  1. make docker-setup   # Pull Ollama + download models (~8GB)"
	@echo "  2. make docker-web     # Launch web UI at http://localhost:8080"
	@echo ""

# =============================================================================
# DOCKER
# =============================================================================

docker-setup: ## Start Ollama container and pull required LLM models
	@echo "🐳 Starting Ollama container..."
	docker compose up -d ollama
	@echo ""
	@echo "⏳ Waiting for Ollama to be healthy..."
	@until docker compose exec ollama ollama list > /dev/null 2>&1; do \
		sleep 3; \
		echo "    still waiting..."; \
	done
	@echo "✅ Ollama is running!"
	@echo ""
	@echo "📥 Pulling LLM model: llama3.1 (~4.7GB)..."
	docker compose exec ollama ollama pull llama3.1
	@echo ""
	@echo "📥 Pulling embedding model: nomic-embed-text (~274MB)..."
	docker compose exec ollama ollama pull nomic-embed-text
	@echo ""
	@echo "============================================"
	@echo "✅ Docker setup complete!"
	@echo ""
	@echo "Next: make docker-web"
	@echo "============================================"

docker-build: ## Build the Docker image
	@echo "🔨 Building OpsRamp ChatBot image..."
	docker compose build web
	@echo "✅ Image built."

docker-web: ## Run the Web UI in Docker — full Docker mode (http://localhost:8080)
	@echo "🌐 Starting OpsRamp ChatBot Web UI (full Docker — CPU-only)..."
	@echo "   ⚠️  On Mac this will be slow (~60s/response). Use 'make docker-web-mac' instead."
	@echo "   Open http://localhost:8080 in your browser."
	@echo "   Press Ctrl+C to stop."
	@echo ""
	docker compose up --build web

docker-web-mac: ## Run the Web UI in Docker using native Ollama on host (recommended for Mac)
	@echo "🌐 Starting OpsRamp ChatBot Web UI (using native Ollama)..."
	@echo "   Requires: Ollama running natively (ollama serve)"
	@echo ""
	@if ! curl -sf http://localhost:11434/api/tags > /dev/null 2>&1; then \
		echo "   ❌ Native Ollama not detected at localhost:11434"; \
		echo "   Please start it first: ollama serve"; \
		exit 1; \
	fi
	@echo "   ✅ Native Ollama detected — using Apple GPU acceleration"
	@echo "   Open http://localhost:8080 in your browser."
	@echo "   Press Ctrl+C to stop."
	@echo ""
	docker compose --profile mac up --build web-mac

docker-mcp: ## Run the MCP HTTP server in Docker (http://localhost:8081)
	@echo "🔌 Starting MCP HTTP server..."
	@echo "   Endpoint: http://localhost:8081"
	@echo ""
	docker compose --profile mcp up --build mcp

docker-cli: ## Run the interactive CLI in Docker
	@echo "💬 Starting interactive CLI..."
	docker compose --profile cli run --rm cli

docker-logs: ## Show logs for all running services
	docker compose logs -f

docker-down: ## Stop all containers
	docker compose --profile all down
	@echo "✅ All containers stopped."

docker-clean: ## Stop containers and remove images + volumes
	docker compose --profile all down --rmi local -v
	@echo "✅ Cleaned up containers, images, and volumes."

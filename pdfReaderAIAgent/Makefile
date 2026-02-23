# =============================================================================
# Makefile - Convenience commands for the PDF Q&A Agent
# =============================================================================
# Run 'make help' to see all available commands.
#
# Two modes:
#   Native (recommended on Mac) — Ollama runs directly on your machine
#   Docker — everything runs in containers (requires Docker Hub access)
# =============================================================================

.PHONY: setup setup-docker pull-models run run-docker run-local clean build help

# Default target
help: ## Show this help message
	@echo ""
	@echo "PDF Q&A Agent - Available Commands:"
	@echo "===================================="
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36mmake %-15s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Quick Start (Native — recommended for Mac):"
	@echo "  1. make setup              # Install Ollama + download AI models"
	@echo "  2. Put a PDF in ./pdfs/"
	@echo "  3. make run PDF=pdfs/yourfile.pdf"
	@echo ""
	@echo "Quick Start (Docker — if Docker Hub is accessible):"
	@echo "  1. make setup-docker       # Start Ollama container + download models"
	@echo "  2. Put a PDF in ./pdfs/"
	@echo "  3. make run-docker PDF=pdfs/yourfile.pdf"
	@echo ""

# =============================================================================
# NATIVE MODE (recommended for Mac — uses Apple GPU, no Docker Hub needed)
# =============================================================================

setup: ## [FIRST TIME] Install Ollama natively and download AI models (~4GB)
	@echo "=========================================="
	@echo "  Setting up Ollama (native macOS mode)"
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
	@echo "📥 Pulling embedding model: nomic-embed-text (~274MB)..."
	@echo "   (This converts text into numerical vectors for search)"
	ollama pull nomic-embed-text
	@echo ""
	@echo "📥 Pulling LLM model: mistral (~4.1GB)..."
	@echo "   (This is the AI brain that generates answers)"
	ollama pull mistral
	@echo ""
	@echo "============================================"
	@echo "✅ Setup complete! (native mode)"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Place a PDF file in the ./pdfs/ directory"
	@echo "  2. Run: make run PDF=pdfs/yourfile.pdf"
	@echo "============================================"

run: ## Run the agent natively (usage: make run PDF=pdfs/yourfile.pdf)
	@if [ -z "$(PDF)" ]; then \
		echo ""; \
		echo "❌ Please specify a PDF file:"; \
		echo "   make run PDF=pdfs/yourfile.pdf"; \
		echo ""; \
		echo "Available PDFs in ./pdfs/:"; \
		ls -1 pdfs/*.pdf 2>/dev/null || echo "   (none found - add a PDF to ./pdfs/)"; \
		echo ""; \
		exit 1; \
	fi
	@if ! curl -sf http://localhost:11434/api/tags > /dev/null 2>&1; then \
		echo "⚠️  Ollama not running. Starting it..."; \
		ollama serve > /dev/null 2>&1 & \
		sleep 3; \
	fi
	OLLAMA_HOST=http://localhost:11434 go run . $(PDF)

# =============================================================================
# DOCKER MODE (alternative — requires Docker Hub access)
# =============================================================================

setup-docker: ## [FIRST TIME] Start Ollama in Docker and download AI models
	@echo "🚀 Starting Ollama service in Docker..."
	docker compose up -d ollama
	@echo ""
	@echo "⏳ Waiting for Ollama to be ready..."
	@until curl -sf http://localhost:11434/api/tags > /dev/null 2>&1; do \
		sleep 2; \
		echo "  still waiting..."; \
	done
	@echo "✅ Ollama is ready!"
	@echo ""
	@echo "📥 Pulling embedding model: nomic-embed-text (~274MB)..."
	docker compose exec ollama ollama pull nomic-embed-text
	@echo ""
	@echo "📥 Pulling LLM model: mistral (~4.1GB)..."
	docker compose exec ollama ollama pull mistral
	@echo ""
	@echo "============================================"
	@echo "✅ Setup complete! (Docker mode)"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Place a PDF file in the ./pdfs/ directory"
	@echo "  2. Run: make run-docker PDF=pdfs/yourfile.pdf"
	@echo "============================================"

build: ## Build the Go application Docker image
	docker compose build app

run-docker: ## Run the agent in Docker (usage: make run-docker PDF=pdfs/yourfile.pdf)
	@if [ -z "$(PDF)" ]; then \
		echo ""; \
		echo "❌ Please specify a PDF file:"; \
		echo "   make run-docker PDF=pdfs/yourfile.pdf"; \
		echo ""; \
		echo "Available PDFs in ./pdfs/:"; \
		ls -1 pdfs/*.pdf 2>/dev/null || echo "   (none found - add a PDF to ./pdfs/)"; \
		echo ""; \
		exit 1; \
	fi
	@echo "🔨 Building the application..."
	docker compose build app
	@echo ""
	@echo "🚀 Starting PDF Q&A Agent..."
	@echo ""
	docker compose run --rm app $(PDF)

# =============================================================================
# SHARED COMMANDS
# =============================================================================

stop: ## Stop all running containers (Docker mode only)
	docker compose down
	@echo "✅ All containers stopped."

clean: ## Stop containers and remove all data (including downloaded models)
	docker compose down -v 2>/dev/null || true
	@echo "✅ Cleaned up! (Models will need to be re-downloaded on next setup)"

logs: ## Show Ollama logs (Docker mode only)
	docker compose logs -f ollama

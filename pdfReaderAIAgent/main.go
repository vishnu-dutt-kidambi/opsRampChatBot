package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"pdf-qa-agent/rag"
)

// =============================================================================
// PDF Q&A Agent - A RAG (Retrieval-Augmented Generation) Demo in Go
// =============================================================================
//
// This application demonstrates the core concepts of RAG:
//
//   1. DOCUMENT LOADING  - Extract text from a PDF file
//   2. CHUNKING          - Split text into smaller, overlapping pieces
//   3. EMBEDDING         - Convert text chunks into numerical vectors
//   4. VECTOR STORAGE    - Store embeddings for fast similarity search
//   5. RETRIEVAL         - Find the most relevant chunks for a question
//   6. GENERATION        - Use an LLM to answer based on retrieved context
//
// All powered by Ollama (free, local LLM) - no API keys needed!
// =============================================================================

func main() {
	// -------------------------------------------------------------------------
	// Parse command line arguments
	// -------------------------------------------------------------------------
	if len(os.Args) < 2 {
		fmt.Println("Usage: pdf-qa-agent <path-to-pdf>")
		fmt.Println("Example: pdf-qa-agent pdfs/sample.pdf")
		os.Exit(1)
	}

	pdfPath := os.Args[1]

	// -------------------------------------------------------------------------
	// Configuration from environment variables (with sensible defaults)
	// -------------------------------------------------------------------------
	ollamaHost := getEnv("OLLAMA_HOST", "http://localhost:11434")
	embeddingModel := getEnv("EMBEDDING_MODEL", "nomic-embed-text")
	llmModel := getEnv("LLM_MODEL", "mistral")

	fmt.Println("========================================")
	fmt.Println("  PDF Q&A Agent - RAG Demo (Go + Ollama)")
	fmt.Println("========================================")
	fmt.Printf("  Ollama:     %s\n", ollamaHost)
	fmt.Printf("  LLM Model:  %s\n", llmModel)
	fmt.Printf("  Embed Model: %s\n", embeddingModel)
	fmt.Printf("  PDF File:   %s\n", pdfPath)
	fmt.Println("========================================")
	fmt.Println()

	// -------------------------------------------------------------------------
	// Create the RAG Agent
	// -------------------------------------------------------------------------
	// The agent orchestrates the entire RAG pipeline:
	//   PDF → Chunks → Embeddings → Vector Store → Retrieval → LLM → Answer
	agent := rag.NewAgent(ollamaHost, llmModel, embeddingModel)

	// -------------------------------------------------------------------------
	// Check Ollama connectivity
	// -------------------------------------------------------------------------
	fmt.Println("[Preflight] Checking Ollama connectivity...")
	if err := agent.CheckOllama(); err != nil {
		fmt.Printf("  ✗ Cannot connect to Ollama: %v\n", err)
		fmt.Println()
		fmt.Println("  Make sure Ollama is running:")
		fmt.Println("    - Docker: make setup")
		fmt.Println("    - Local:  ollama serve")
		os.Exit(1)
	}
	fmt.Println("  ✓ Ollama is reachable")
	fmt.Println()

	// -------------------------------------------------------------------------
	// Load and process the PDF (Steps 1-4 of the RAG pipeline)
	// -------------------------------------------------------------------------
	if err := agent.LoadPDF(pdfPath); err != nil {
		fmt.Printf("\n✗ Error: %v\n", err)
		os.Exit(1)
	}

	// -------------------------------------------------------------------------
	// Interactive Q&A Loop (Steps 5-6 of the RAG pipeline)
	// -------------------------------------------------------------------------
	fmt.Println()
	fmt.Println("✅ Ready! Ask questions about your PDF.")
	fmt.Println("   Type 'quit' to exit.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("You: ")
		if !scanner.Scan() {
			break
		}

		question := strings.TrimSpace(scanner.Text())
		if question == "" {
			continue
		}
		if strings.ToLower(question) == "quit" || strings.ToLower(question) == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		// This triggers the retrieval + generation steps
		answer, err := agent.Ask(question)
		if err != nil {
			fmt.Printf("  ✗ Error: %v\n\n", err)
			continue
		}

		fmt.Printf("\nAgent: %s\n\n", answer)
	}
}

// getEnv returns the value of an environment variable, or a default if not set.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

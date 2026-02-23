package rag

import (
	"fmt"
	"net/http"
	"time"
)

// =============================================================================
// RAG Agent - Orchestrating the Pipeline
// =============================================================================
//
// WHAT IS A RAG AGENT?
//
// RAG = Retrieval-Augmented Generation
//
// It's a pattern that combines:
//   1. RETRIEVAL  - Finding relevant information from a knowledge base
//   2. AUGMENTED  - Injecting that information into the LLM's prompt
//   3. GENERATION - Using the LLM to generate an answer
//
// WHY RAG?
// LLMs are trained on general data and have a knowledge cutoff date.
// RAG lets you:
//   - Answer questions about YOUR specific documents
//   - Keep the LLM's answers grounded in real data
//   - Update knowledge without retraining the model
//   - Reduce hallucination (the LLM making things up)
//
// THE RAG PIPELINE:
//
//   ┌─────────┐    ┌──────────┐    ┌───────────┐    ┌──────────────┐
//   │  Load   │───>│  Chunk   │───>│  Embed    │───>│ Vector Store │
//   │  PDF    │    │  Text    │    │  Chunks   │    │  (Index)     │
//   └─────────┘    └──────────┘    └───────────┘    └──────┬───────┘
//                                                          │
//   ┌──────────┐    ┌──────────┐    ┌───────────┐    ┌─────┴────────┐
//   │  Answer  │<───│  LLM     │<───│  Build    │<───│  Similarity  │
//   │  User    │    │ Generate │    │  Prompt   │    │  Search      │
//   └──────────┘    └──────────┘    └───────────┘    └──────────────┘
//                                                          ▲
//                                                          │
//                                                   ┌──────┴───────┐
//                                                   │  Embed       │
//                                                   │  Question    │
//                                                   └──────────────┘
//
// =============================================================================

// Agent is the main RAG agent that orchestrates the entire pipeline.
type Agent struct {
	embedder     *OllamaEmbedder // Converts text → vectors
	llm          *OllamaLLM      // Generates answers from context
	vectorStore  *VectorStore    // Stores and searches embeddings
	chunkSize    int             // Characters per chunk (~500)
	chunkOverlap int             // Overlap between chunks (~100)
	topK         int             // Number of chunks to retrieve (3-5)
}

// NewAgent creates a new RAG agent with the given configuration.
//
// Parameters:
//   - ollamaHost:     URL of the Ollama server (e.g., "http://localhost:11434")
//   - llmModel:       Name of the LLM for generation (e.g., "mistral")
//   - embeddingModel: Name of the embedding model (e.g., "nomic-embed-text")
func NewAgent(ollamaHost, llmModel, embeddingModel string) *Agent {
	return &Agent{
		embedder:     NewOllamaEmbedder(ollamaHost, embeddingModel),
		llm:          NewOllamaLLM(ollamaHost, llmModel),
		vectorStore:  NewVectorStore(),
		chunkSize:    500, // Each chunk ~500 characters
		chunkOverlap: 100, // 100 chars overlap between chunks
		topK:         3,   // Retrieve top 3 most relevant chunks
	}
}

// CheckOllama verifies that the Ollama server is reachable.
func (a *Agent) CheckOllama() error {
	client := &http.Client{Timeout: 5 * time.Second}
	url := fmt.Sprintf("%s/api/tags", a.embedder.baseURL)

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("cannot reach Ollama at %s: %w", a.embedder.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama returned status %d", resp.StatusCode)
	}

	return nil
}

// LoadPDF processes a PDF file through the RAG indexing pipeline.
//
// This is the "offline" or "indexing" phase — it prepares the document
// so that questions can be answered quickly later.
//
// Steps:
//  1. Extract text from the PDF
//  2. Split text into overlapping chunks
//  3. Generate embeddings for each chunk
//  4. Store embeddings in the vector store
func (a *Agent) LoadPDF(path string) error {
	// =========================================================================
	// STEP 1: Extract text from the PDF
	// =========================================================================
	fmt.Println("[Step 1/4] Extracting text from PDF...")
	text, pageCount, err := ExtractTextFromPDF(path)
	if err != nil {
		return fmt.Errorf("PDF extraction failed: %w", err)
	}
	fmt.Printf("  ✓ Extracted %d characters from %d pages\n", len(text), pageCount)

	if len(text) == 0 {
		return fmt.Errorf("no text found in PDF — it might be a scanned/image-based PDF")
	}

	// =========================================================================
	// STEP 2: Split text into overlapping chunks
	// =========================================================================
	fmt.Println("[Step 2/4] Splitting text into chunks...")
	chunks := ChunkText(text, a.chunkSize, a.chunkOverlap)
	fmt.Printf("  ✓ Created %d chunks (size: %d chars, overlap: %d chars)\n",
		len(chunks), a.chunkSize, a.chunkOverlap)

	if len(chunks) == 0 {
		return fmt.Errorf("no chunks created from the text")
	}

	// =========================================================================
	// STEP 3: Generate embeddings for each chunk
	// =========================================================================
	fmt.Println("[Step 3/4] Generating embeddings (this may take a moment)...")
	for i, chunk := range chunks {
		fmt.Printf("\r  Embedding chunk %d/%d...", i+1, len(chunks))

		// Convert chunk text → vector using the embedding model
		embedding, err := a.embedder.Embed(chunk.Text)
		if err != nil {
			return fmt.Errorf("\nfailed to embed chunk %d: %w", i, err)
		}

		// =====================================================================
		// STEP 4: Store in vector database
		// =====================================================================
		a.vectorStore.Add(chunk.ID, embedding, chunk.Text, map[string]string{
			"chunk_index": fmt.Sprintf("%d", i),
		})
	}
	fmt.Printf("\r  ✓ Embedded %d chunks (dimension: %d)                  \n",
		len(chunks), a.embedder.Dimension())

	fmt.Printf("[Step 4/4] Vector store ready — %d entries indexed\n", a.vectorStore.Size())

	return nil
}

// Ask processes a user question through the RAG query pipeline.
//
// This is the "online" or "query" phase — it happens in real-time
// when the user asks a question.
//
// Steps:
//  1. Embed the question using the SAME embedding model
//  2. Search the vector store for similar chunks (retrieval)
//  3. Build a prompt with question + retrieved context (augmentation)
//  4. Send to the LLM for answer generation
func (a *Agent) Ask(question string) (string, error) {
	// =========================================================================
	// STEP 1: Embed the question
	// =========================================================================
	// We use the SAME embedding model that was used for the chunks.
	// This ensures the vectors are in the same "space" and comparable.
	fmt.Println("  🔎 Embedding your question...")
	questionEmbedding, err := a.embedder.Embed(question)
	if err != nil {
		return "", fmt.Errorf("failed to embed question: %w", err)
	}

	// =========================================================================
	// STEP 2: Retrieve the most relevant chunks (similarity search)
	// =========================================================================
	fmt.Printf("  🔎 Searching vector store (top %d matches)...\n", a.topK)
	results := a.vectorStore.Search(questionEmbedding, a.topK)

	if len(results) == 0 {
		return "I couldn't find any relevant information in the document.", nil
	}

	// Display which chunks were retrieved (educational — shows what RAG found)
	for i, r := range results {
		preview := r.Text
		if len(preview) > 80 {
			preview = preview[:80] + "..."
		}
		fmt.Printf("  📄 Match %d (score: %.4f): %s\n", i+1, r.Score, preview)
	}

	// =========================================================================
	// STEP 3: Build the augmented prompt (context + question)
	// =========================================================================
	var contexts []string
	for _, r := range results {
		contexts = append(contexts, r.Text)
	}

	// =========================================================================
	// STEP 4: Generate the answer using the LLM
	// =========================================================================
	fmt.Println("  📝 Generating answer with LLM...")
	answer, err := a.llm.GenerateRAGAnswer(question, contexts)
	if err != nil {
		return "", fmt.Errorf("failed to generate answer: %w", err)
	}

	return answer, nil
}

package knowledge

import (
	"fmt"

	"pdf-qa-agent/rag"
)

// =============================================================================
// Knowledge Base — RAG Pipeline for Operations Runbooks
// =============================================================================
//
// This package wraps the RAG pipeline from the pdfReaderAIAgent project
// (pdf-qa-agent/rag) and adapts it for use as a tool in the OpsRamp agent.
//
// Instead of duplicating the RAG code, we import the existing rag package
// which provides: OllamaEmbedder, VectorStore, ChunkText, ExtractTextFromPDF.
//
// Pipeline (delegated to rag package):
//   Load PDF → Extract Text → Chunk → Embed → Store in VectorStore
//   Query → Embed Question → Similarity Search → Return Context
//
// =============================================================================

// SearchResult contains a matched chunk and its similarity score.
// This re-exports the rag.SearchResult type so consumers don't need
// to import the rag package directly.
type SearchResult = rag.SearchResult

// KnowledgeBase manages the full RAG pipeline for operations runbooks.
// It wraps the rag package components (embedder, vector store, chunker, PDF extractor).
type KnowledgeBase struct {
	embedder     *rag.OllamaEmbedder
	vectorStore  *rag.VectorStore
	chunkSize    int
	chunkOverlap int
	topK         int
	loaded       bool
}

// NewKnowledgeBase creates a new knowledge base connected to Ollama for embeddings.
// Uses the rag package's OllamaEmbedder and VectorStore under the hood.
func NewKnowledgeBase(ollamaURL, embeddingModel string) *KnowledgeBase {
	return &KnowledgeBase{
		embedder:     rag.NewOllamaEmbedder(ollamaURL, embeddingModel),
		vectorStore:  rag.NewVectorStore(),
		chunkSize:    500,
		chunkOverlap: 100,
		topK:         3,
	}
}

// LoadPDF processes a PDF file through the RAG indexing pipeline:
// Extract text → Chunk → Embed → Store
// All steps delegate to the imported rag package.
func (kb *KnowledgeBase) LoadPDF(path string) error {
	fmt.Printf("  [knowledge] Loading PDF: %s\n", path)

	// Step 1: Extract text (using rag.ExtractTextFromPDF)
	text, pageCount, err := rag.ExtractTextFromPDF(path)
	if err != nil {
		return fmt.Errorf("PDF extraction failed: %w", err)
	}
	fmt.Printf("  [knowledge] Extracted %d characters from %d pages\n", len(text), pageCount)

	if len(text) == 0 {
		return fmt.Errorf("no text found in PDF (may be scanned/image-based)")
	}

	// Step 2: Chunk text (using rag.ChunkText)
	chunks := rag.ChunkText(text, kb.chunkSize, kb.chunkOverlap)
	fmt.Printf("  [knowledge] Created %d chunks (size: %d, overlap: %d)\n",
		len(chunks), kb.chunkSize, kb.chunkOverlap)

	// Step 3 & 4: Embed each chunk and store (using rag.OllamaEmbedder + rag.VectorStore)
	fmt.Printf("  [knowledge] Embedding chunks...")
	for i, chunk := range chunks {
		fmt.Printf("\r  [knowledge] Embedding chunk %d/%d...", i+1, len(chunks))

		embedding, err := kb.embedder.Embed(chunk.Text)
		if err != nil {
			return fmt.Errorf("failed to embed chunk %d: %w", i, err)
		}

		kb.vectorStore.Add(chunk.ID, embedding, chunk.Text, nil)
	}
	fmt.Printf("\r  [knowledge] Embedded %d chunks successfully                \n", len(chunks))

	kb.loaded = true
	fmt.Printf("  [knowledge] Vector store ready — %d entries indexed\n", kb.vectorStore.Size())
	return nil
}

// Search queries the knowledge base and returns the top-K most relevant text chunks.
// Delegates embedding and similarity search to the rag package.
func (kb *KnowledgeBase) Search(query string) ([]SearchResult, error) {
	if !kb.loaded || kb.vectorStore.Size() == 0 {
		return nil, fmt.Errorf("knowledge base is empty — no documents loaded")
	}

	// Embed the query (using rag.OllamaEmbedder)
	queryEmbedding, err := kb.embedder.Embed(query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Search for similar chunks (using rag.VectorStore)
	results := kb.vectorStore.Search(queryEmbedding, kb.topK)
	return results, nil
}

// IsLoaded returns whether any documents have been loaded.
func (kb *KnowledgeBase) IsLoaded() bool {
	return kb.loaded
}

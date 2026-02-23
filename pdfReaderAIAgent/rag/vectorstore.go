package rag

import (
	"math"
	"sort"
)

// =============================================================================
// Vector Store - In-Memory Vector Database
// =============================================================================
//
// WHAT IS A VECTOR DATABASE?
//
// A vector database stores vectors (embeddings) and enables "similarity search"
// — finding the vectors most similar to a given query vector.
//
// In our RAG pipeline:
//   1. Each document chunk is embedded (converted to a vector)
//   2. These vectors are stored in the vector database
//   3. When a user asks a question, the question is also embedded
//   4. We search for the chunk vectors most similar to the question vector
//   5. The matching chunks become the "context" for the LLM
//
// PRODUCTION VECTOR DATABASES:
//   - ChromaDB     (open source, easy to use)
//   - Pinecone     (cloud, managed)
//   - Weaviate     (open source, full-featured)
//   - Qdrant       (open source, Rust-based, fast)
//   - Milvus       (open source, scalable)
//   - pgvector     (PostgreSQL extension)
//
// WHY BUILD OUR OWN?
// This simple implementation lets you see EXACTLY how vector search works
// under the hood. The core algorithm (cosine similarity + brute force search)
// is the same concept used by all vector databases!
//
// =============================================================================

// VectorStore is a simple in-memory vector database.
// It stores embeddings and supports similarity search.
type VectorStore struct {
	entries []VectorEntry
}

// VectorEntry represents a single document chunk in the vector store.
type VectorEntry struct {
	ID        string            // Unique identifier (e.g., "chunk_0")
	Embedding []float64         // The vector representation of the text
	Text      string            // Original text (for displaying results)
	Metadata  map[string]string // Optional metadata (e.g., page number)
}

// SearchResult represents a single search result with its similarity score.
type SearchResult struct {
	ID    string  // Chunk identifier
	Text  string  // Original text content
	Score float64 // Cosine similarity (0.0 to 1.0, higher = more relevant)
}

// NewVectorStore creates a new empty vector store.
func NewVectorStore() *VectorStore {
	return &VectorStore{
		entries: make([]VectorEntry, 0),
	}
}

// Add inserts a new entry into the vector store.
//
// Parameters:
//   - id:        Unique identifier for this entry
//   - embedding: The vector representation (from the embedding model)
//   - text:      The original text (kept for display purposes)
//   - metadata:  Optional key-value pairs for filtering
func (vs *VectorStore) Add(id string, embedding []float64, text string, metadata map[string]string) {
	vs.entries = append(vs.entries, VectorEntry{
		ID:        id,
		Embedding: embedding,
		Text:      text,
		Metadata:  metadata,
	})
}

// Size returns the number of entries stored.
func (vs *VectorStore) Size() int {
	return len(vs.entries)
}

// Search finds the topK entries most similar to the query embedding.
//
// HOW VECTOR SEARCH WORKS (step by step):
//
//  1. Take the query embedding (e.g., the embedded question)
//  2. Compare it against EVERY stored embedding using cosine similarity
//  3. Sort all results by similarity score (highest first)
//  4. Return the top K most similar entries
//
// This is called "brute force" or "exact nearest neighbor" search.
// It checks every single entry, so it's O(n) where n = number of entries.
//
// For small datasets (< 100k entries), this is perfectly fine!
// For larger datasets, production vector databases use approximate methods:
//   - HNSW (Hierarchical Navigable Small World) - used by most modern DBs
//   - IVF (Inverted File Index) - good for very large datasets
//   - LSH (Locality Sensitive Hashing) - fast but less accurate
//
// These trade a tiny bit of accuracy for massive speed improvements.
func (vs *VectorStore) Search(queryEmbedding []float64, topK int) []SearchResult {
	if len(vs.entries) == 0 {
		return nil
	}

	// Step 1: Calculate similarity between query and EVERY stored embedding
	results := make([]SearchResult, len(vs.entries))
	for i, entry := range vs.entries {
		results[i] = SearchResult{
			ID:    entry.ID,
			Text:  entry.Text,
			Score: cosineSimilarity(queryEmbedding, entry.Embedding),
		}
	}

	// Step 2: Sort by similarity score (highest = most similar comes first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Step 3: Return only the top K results
	if topK > len(results) {
		topK = len(results)
	}
	return results[:topK]
}

// =============================================================================
// Cosine Similarity - The Heart of Vector Search
// =============================================================================
//
// COSINE SIMILARITY EXPLAINED:
//
// Cosine similarity measures the angle between two vectors, ignoring their
// length. It tells us how similar the DIRECTION of two vectors is.
//
// Score interpretation:
//   1.0  → Identical direction (same meaning)
//   0.5  → Somewhat similar
//   0.0  → Perpendicular (unrelated)
//  -1.0  → Opposite direction
//
// FORMULA:
//
//                    A · B           (dot product)
//   cos(θ) = ─────────────────── = ─────────────────
//              |A| × |B|          (product of magnitudes)
//
// Where:
//   A · B  = Σ(aᵢ × bᵢ)          Sum of element-wise products
//   |A|    = √(Σ aᵢ²)            Square root of sum of squares
//   |B|    = √(Σ bᵢ²)            Square root of sum of squares
//
// VISUAL EXAMPLE (in 2D, but our vectors have 768 dimensions!):
//
//   Vector A = [1, 1]    (pointing up-right at 45°)
//   Vector B = [1, 0]    (pointing right at 0°)
//
//   A · B = (1×1) + (1×0) = 1
//   |A|   = √(1² + 1²) = √2 ≈ 1.414
//   |B|   = √(1² + 0²) = 1
//
//   cos(θ) = 1 / (1.414 × 1) ≈ 0.707
//
//   They're somewhat similar (45° apart).
//
// =============================================================================

func cosineSimilarity(a, b []float64) float64 {
	// Vectors must be the same length
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct float64 // A · B  (numerator)
	var normA float64      // |A|² (will take sqrt later)
	var normB float64      // |B|² (will take sqrt later)

	for i := range a {
		dotProduct += a[i] * b[i] // Multiply corresponding elements
		normA += a[i] * a[i]      // Square each element of A
		normB += b[i] * b[i]      // Square each element of B
	}

	// Avoid division by zero (zero vector has no direction)
	if normA == 0 || normB == 0 {
		return 0
	}

	// Final formula: dotProduct / (sqrt(normA) * sqrt(normB))
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

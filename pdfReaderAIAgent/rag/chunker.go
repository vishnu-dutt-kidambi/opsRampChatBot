package rag

import (
	"fmt"
	"strings"
)

// =============================================================================
// Text Chunking
// =============================================================================
//
// WHY DO WE CHUNK TEXT?
//
// Imagine you have a 100-page PDF. If you try to embed the entire document
// as one vector, you lose detail — the embedding becomes a vague "average"
// of everything in the document.
//
// Instead, we split the text into smaller CHUNKS (typically 200-1000 chars).
// Each chunk gets its own embedding, representing that specific section.
//
// When a user asks a question, we find the chunks most similar to the question
// and pass only those to the LLM. This means:
//   1. More precise retrieval (find the exact relevant section)
//   2. Less noise (don't overwhelm the LLM with irrelevant text)
//   3. Fits within LLM context windows
//
// OVERLAP:
// We use overlapping chunks to avoid losing context at boundaries.
// If an important sentence spans two chunks, the overlap ensures
// it appears fully in at least one chunk.
//
//   Example (chunkSize=20, overlap=5):
//   Text:    "The quick brown fox jumps over the lazy dog"
//   Chunk 1: "The quick brown fox "   (chars 0-19)
//   Chunk 2: " fox jumps over the "   (chars 15-34, overlaps by 5)
//   Chunk 3: " the lazy dog"          (chars 30-43, overlaps by 5)
//
// =============================================================================

// Chunk represents a piece of the original document.
type Chunk struct {
	ID   string // Unique identifier (e.g., "chunk_0", "chunk_1")
	Text string // The text content of this chunk
}

// ChunkText splits a large text into overlapping chunks.
//
// Parameters:
//   - text:      The full document text to split
//   - chunkSize: Target size of each chunk in characters (~500 is typical)
//   - overlap:   Number of characters to overlap between chunks (~100 is typical)
//
// Returns a slice of Chunk structs, each with a unique ID and text content.
func ChunkText(text string, chunkSize, overlap int) []Chunk {
	// Clean up the text: normalize whitespace
	text = strings.TrimSpace(text)

	if len(text) == 0 {
		return nil
	}

	// Validate parameters
	if chunkSize <= 0 {
		chunkSize = 500
	}
	if overlap < 0 || overlap >= chunkSize {
		overlap = chunkSize / 5 // Default to 20% overlap
	}

	var chunks []Chunk
	start := 0
	chunkIndex := 0

	for start < len(text) {
		// Calculate the end position of this chunk
		end := start + chunkSize
		if end > len(text) {
			end = len(text)
		}

		// Extract the chunk text
		chunkText := text[start:end]

		// Try to break at a sentence boundary (period, newline) for cleaner chunks
		// Only do this if we're not at the end of the text
		if end < len(text) {
			// Look for the last sentence-ending punctuation in the chunk
			lastBreak := findLastSentenceBreak(chunkText)
			if lastBreak > chunkSize/2 { // Only break if we keep at least half the chunk
				end = start + lastBreak + 1
				chunkText = text[start:end]
			}
		}

		chunks = append(chunks, Chunk{
			ID:   fmt.Sprintf("chunk_%d", chunkIndex),
			Text: strings.TrimSpace(chunkText),
		})

		// Move the start position forward by (chunkSize - overlap)
		// This creates the overlap between consecutive chunks
		step := chunkSize - overlap
		if step <= 0 {
			step = 1 // Safety: always move forward
		}
		start += step
		chunkIndex++

		// Don't create tiny final chunks (less than overlap size)
		if len(text)-start < overlap && len(text)-start > 0 {
			// Append remaining text to the last chunk or create final chunk
			remaining := strings.TrimSpace(text[start:])
			if len(remaining) > 0 {
				chunks = append(chunks, Chunk{
					ID:   fmt.Sprintf("chunk_%d", chunkIndex),
					Text: remaining,
				})
			}
			break
		}
	}

	return chunks
}

// findLastSentenceBreak finds the last position in text where a sentence ends.
// Returns -1 if no good break point is found.
func findLastSentenceBreak(text string) int {
	lastBreak := -1
	for i := len(text) - 1; i >= 0; i-- {
		ch := text[i]
		if ch == '.' || ch == '!' || ch == '?' || ch == '\n' {
			lastBreak = i
			break
		}
	}
	return lastBreak
}

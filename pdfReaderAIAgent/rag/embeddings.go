package rag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// =============================================================================
// Embeddings - Converting Text to Vectors
// =============================================================================
//
// WHAT ARE EMBEDDINGS?
//
// An embedding is a list of numbers (a "vector") that captures the MEANING
// of a piece of text. Think of it as a coordinate in a high-dimensional space
// where similar meanings are close together.
//
//   "cat"    → [0.23, 0.87, 0.11, 0.45, ...]   (768 numbers)
//   "kitten" → [0.25, 0.85, 0.13, 0.44, ...]   ← VERY similar to "cat"
//   "car"    → [0.91, 0.12, 0.78, 0.03, ...]   ← VERY different from "cat"
//
// WHY DO WE NEED THEM?
// Computers can't understand text directly. By converting text to numbers,
// we can use math (cosine similarity) to find which chunks of our document
// are most relevant to a user's question.
//
// HOW ARE THEY GENERATED?
// A neural network (the embedding model) has been trained on billions of
// text examples to learn which texts have similar meanings. We use
// Ollama's "nomic-embed-text" model, which runs locally and is free.
//
// =============================================================================

// OllamaEmbedder generates text embeddings using Ollama's local API.
type OllamaEmbedder struct {
	baseURL   string // Ollama server URL (e.g., "http://localhost:11434")
	model     string // Embedding model name (e.g., "nomic-embed-text")
	dimension int    // Dimension of embeddings (set after first call)
}

// embeddingRequest is the JSON body sent to Ollama's /api/embeddings endpoint.
type embeddingRequest struct {
	Model  string `json:"model"`  // Which embedding model to use
	Prompt string `json:"prompt"` // The text to embed
}

// embeddingResponse is the JSON response from Ollama's /api/embeddings endpoint.
type embeddingResponse struct {
	Embedding []float64 `json:"embedding"` // The embedding vector
}

// NewOllamaEmbedder creates a new embedder connected to an Ollama instance.
func NewOllamaEmbedder(baseURL, model string) *OllamaEmbedder {
	return &OllamaEmbedder{
		baseURL: baseURL,
		model:   model,
	}
}

// Dimension returns the dimensionality of the embeddings (e.g., 768).
// Only available after the first Embed() call.
func (e *OllamaEmbedder) Dimension() int {
	return e.dimension
}

// Embed converts a text string into a numerical vector (embedding).
//
// This is a key step in RAG:
//   - During indexing: each document chunk is embedded and stored
//   - During querying: the user's question is embedded using the SAME model
//   - Then we compare the question embedding with chunk embeddings
//
// IMPORTANT: You must use the SAME embedding model for both indexing and
// querying, otherwise the vectors won't be comparable!
func (e *OllamaEmbedder) Embed(text string) ([]float64, error) {
	// Build the request payload
	reqBody := embeddingRequest{
		Model:  e.model,
		Prompt: text,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	// Send POST request to Ollama's embedding endpoint
	url := fmt.Sprintf("%s/api/embeddings", e.baseURL)
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to call Ollama (is it running at %s?): %w", e.baseURL, err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama returned error (status %d): %s\nMake sure the model '%s' is pulled: ollama pull %s",
			resp.StatusCode, string(body), e.model, e.model)
	}

	// Parse the JSON response
	var embResp embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, fmt.Errorf("failed to parse Ollama response: %w", err)
	}

	// Store the dimension for reference (e.g., 768 for nomic-embed-text)
	if e.dimension == 0 && len(embResp.Embedding) > 0 {
		e.dimension = len(embResp.Embedding)
	}

	return embResp.Embedding, nil
}

// =============================================================================
// LLM (Large Language Model) - Text Generation
// =============================================================================
//
// WHAT IS THE LLM'S ROLE IN RAG?
//
// The LLM is the "brain" that reads the retrieved context and formulates
// a human-readable answer. Without RAG, an LLM can only use its training
// data. With RAG, we inject relevant document chunks into the prompt,
// giving the LLM access to YOUR specific documents.
//
// THE RAG PROMPT PATTERN:
//   "Here is some context: [retrieved chunks]
//    Based on this context, answer: [user's question]"
//
// This is more reliable than just asking the LLM directly because:
//   1. The answer is grounded in your actual documents
//   2. You can verify the answer against the source
//   3. It reduces hallucination (making things up)
//
// =============================================================================

// OllamaLLM handles text generation using Ollama's local API.
type OllamaLLM struct {
	baseURL string // Ollama server URL
	model   string // LLM model name (e.g., "mistral")
}

// generateRequest is the JSON body sent to Ollama's /api/generate endpoint.
type generateRequest struct {
	Model  string `json:"model"`  // Which LLM model to use
	Prompt string `json:"prompt"` // The full prompt including context
	Stream bool   `json:"stream"` // false = get complete response at once
}

// generateResponse is the JSON response from Ollama's /api/generate endpoint.
type generateResponse struct {
	Response string `json:"response"` // The generated text
}

// NewOllamaLLM creates a new LLM client connected to an Ollama instance.
func NewOllamaLLM(baseURL, model string) *OllamaLLM {
	return &OllamaLLM{
		baseURL: baseURL,
		model:   model,
	}
}

// GenerateRAGAnswer creates an answer to a question using retrieved context chunks.
//
// This constructs a carefully designed prompt that:
//  1. Tells the LLM its role (helpful assistant)
//  2. Provides the retrieved context chunks
//  3. Presents the user's question
//  4. Instructs the LLM to ONLY use the provided context
//
// The "only use provided context" instruction is crucial — it prevents
// the LLM from making up answers based on its general training data.
func (l *OllamaLLM) GenerateRAGAnswer(question string, contexts []string) (string, error) {
	// Build the context section from retrieved chunks
	contextText := ""
	for i, ctx := range contexts {
		contextText += fmt.Sprintf("\n--- Context Chunk %d ---\n%s\n", i+1, ctx)
	}

	// Construct the RAG prompt
	// This prompt template is simple but effective. Production systems
	// often use more sophisticated templates with examples.
	prompt := fmt.Sprintf(`You are a helpful assistant that answers questions based ONLY on the provided context.

RULES:
1. Use ONLY the information from the context below to answer.
2. If the context doesn't contain enough information, say "I don't have enough information in the document to answer that."
3. Keep your answer concise and relevant.
4. If you quote from the context, mention which part you're referencing.

CONTEXT FROM DOCUMENT:
%s

USER'S QUESTION: %s

YOUR ANSWER:`, contextText, question)

	// Send the prompt to Ollama's generate API
	reqBody := generateRequest{
		Model:  l.model,
		Prompt: prompt,
		Stream: false, // We want the complete response, not streaming
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal generate request: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", l.baseURL)
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to call Ollama LLM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama LLM error (status %d): %s\nMake sure model '%s' is pulled: ollama pull %s",
			resp.StatusCode, string(body), l.model, l.model)
	}

	var genResp generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return "", fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return genResp.Response, nil
}

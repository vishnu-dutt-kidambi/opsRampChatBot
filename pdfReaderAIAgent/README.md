# PDF Q&A Agent — RAG Demo in Go

A simple **Retrieval-Augmented Generation (RAG)** agent written in Go that lets you
ask questions about any PDF document. Runs entirely locally using **Ollama** — 
**no API keys, no tokens, no cloud services, completely free**.

---

## What You'll Learn

This project demonstrates the core concepts of AI agent development:

| Concept | What It Does | Where in Code |
|---------|-------------|---------------|
| **Document Loading** | Extract text from PDF files | `rag/pdf.go` |
| **Chunking** | Split text into overlapping pieces | `rag/chunker.go` |
| **Embeddings** | Convert text → numerical vectors | `rag/embeddings.go` |
| **Vector Store** | Store & search vectors by similarity | `rag/vectorstore.go` |
| **RAG Pipeline** | Orchestrate retrieval + generation | `rag/agent.go` |
| **LLM Generation** | Generate answers from context | `rag/embeddings.go` |

Each file is heavily commented to explain the concepts as you read the code.

---

## Architecture

```
 ┌─────────────────── INDEXING PHASE (one-time) ───────────────────┐
 │                                                                  │
 │   PDF ──→ Extract Text ──→ Chunk Text ──→ Embed Chunks ──→ Store│
 │                                                           Vectors│
 └──────────────────────────────────────────────────────────────────┘

 ┌─────────────────── QUERY PHASE (per question) ─────────────────┐
 │                                                                  │
 │   Question ──→ Embed Question ──→ Search Similar ──→ Build     │
 │                                    Chunks            Prompt     │
 │                                                        │        │
 │                                                        ▼        │
 │                                    Answer  ◀──── LLM Generate   │
 └──────────────────────────────────────────────────────────────────┘
```

---

## Prerequisites

- **Docker** and **Docker Compose** (that's it!)
- ~5GB disk space for AI models
- No API keys or tokens needed

---

## Quick Start

### 1. Clone & enter the project
```bash
cd /path/to/firstTry
```

### 2. Setup — Start Ollama and download models (one-time, ~5 min)
```bash
make setup
```
This will:
- Start the Ollama container (local AI server)
- Download `nomic-embed-text` model (~274MB) — for converting text to vectors
- Download `mistral` model (~4.1GB) — the LLM brain that generates answers

### 3. Add a PDF
Place any PDF file in the `pdfs/` directory:
```bash
cp ~/Downloads/some-document.pdf pdfs/
```

### 4. Run the agent
```bash
make run PDF=pdfs/some-document.pdf
```

### 5. Ask questions!
```
========================================
  PDF Q&A Agent - RAG Demo (Go + Ollama)
========================================
  Ollama:      http://ollama:11434
  LLM Model:   mistral
  Embed Model:  nomic-embed-text
  PDF File:    pdfs/some-document.pdf
========================================

[Step 1/4] Extracting text from PDF...
  ✓ Extracted 15432 characters from 8 pages
[Step 2/4] Splitting text into chunks...
  ✓ Created 38 chunks (size: 500 chars, overlap: 100 chars)
[Step 3/4] Generating embeddings (this may take a moment)...
  ✓ Embedded 38 chunks (dimension: 768)
[Step 4/4] Vector store ready — 38 entries indexed

✅ Ready! Ask questions about your PDF.
   Type 'quit' to exit.

You: What is this document about?
  🔎 Embedding your question...
  🔎 Searching vector store (top 3 matches)...
  📄 Match 1 (score: 0.8234): This document provides an overview of...
  📄 Match 2 (score: 0.7891): The main topics covered include...
  📄 Match 3 (score: 0.7456): In summary, the document describes...
  📝 Generating answer with LLM...

Agent: This document is an overview of...

You: quit
Goodbye!
```

---

## Project Structure

```
firstTry/
├── main.go              # Entry point — CLI and interactive loop
├── rag/
│   ├── agent.go         # RAG Agent — orchestrates the full pipeline
│   ├── pdf.go           # PDF text extraction
│   ├── chunker.go       # Text chunking with overlap
│   ├── embeddings.go    # Ollama API client (embeddings + LLM)
│   └── vectorstore.go   # In-memory vector store + cosine similarity
├── pdfs/                # Put your PDF files here
├── docker-compose.yml   # Docker orchestration (Ollama + App)
├── Dockerfile           # Multi-stage Go build
├── Makefile             # Convenience commands
├── go.mod               # Go module definition
└── README.md            # This file
```

---

## Available Commands

| Command | Description |
|---------|-------------|
| `make setup` | Start Ollama & download AI models (run once) |
| `make run PDF=pdfs/file.pdf` | Run the Q&A agent with a PDF |
| `make run-local PDF=pdfs/file.pdf` | Run locally without Docker |
| `make stop` | Stop all containers |
| `make clean` | Stop containers & delete all data |
| `make logs` | View Ollama logs |
| `make help` | Show all commands |

---

## Running Without Docker (Local Development)

If you prefer running natively (better performance on Mac with Apple Silicon):

### 1. Install Ollama
```bash
# macOS
brew install ollama

# Or download from https://ollama.com
```

### 2. Start Ollama & pull models
```bash
ollama serve                    # Start the server (leave running)
ollama pull nomic-embed-text    # Embedding model
ollama pull mistral             # LLM model
```

### 3. Run the Go app directly
```bash
go mod tidy                                 # Install Go dependencies
make run-local PDF=pdfs/some-document.pdf   # Run!
```

---

## How It Works — Step by Step (Deep Dive)

### 1. PDF Text Extraction (`rag/pdf.go`)
Opens the PDF and extracts all text content. Uses the `ledongthuc/pdf` library.

### 2. Text Chunking (`rag/chunker.go`)

**The Problem:** If you embed an entire 10,000-character document as one single vector, that vector becomes a vague "average" of everything — it can't represent any specific section well. It's like summarizing an entire book into one word.

**The Solution:** Split the document into small, overlapping pieces (~500 characters each).

```
Original text (1200 chars):
"Mercury is the smallest planet... Venus is the second planet... Earth is the third..."

After chunking (chunkSize=500, overlap=100):

 ┌─────────────────────────────────┐
 │ Chunk 0 (chars 0-499)          │  "Mercury is the smallest planet..."
 │                        ┌───────┼────────────────────────────────┐
 └────────────────────────┼───────┘                                │
                          │ Chunk 1 (chars 400-899)                │
                          │           "...Venus is the second..."  │
                          │                        ┌───────────────┼──┐
                          └────────────────────────┼───────────────┘  │
                                                   │ Chunk 2 (chars 800-1200)
                                                   │  "...Earth is the third..."
                                                   └─────────────────┘
```

**Why Overlap?** Without it, important sentences at boundaries get split in half:
```
Without overlap:
  Chunk 0: "...Mercury has no moons and no ri"
  Chunk 1: "ngs. Venus is the second..."
  ❌ The sentence about rings is broken!

With overlap (100 chars):
  Chunk 0: "...Mercury has no moons and no rings."
  Chunk 1: "no moons and no rings. Venus is the second..."
  ✅ The complete sentence appears in at least one chunk!
```

The key logic in `rag/chunker.go`:
```go
step := chunkSize - overlap   // 500 - 100 = 400
start += step                 // Each new chunk starts 400 chars after the previous
```

### 3. Embedding Generation (`rag/embeddings.go`)

**The Problem:** Computers work with numbers, not words. To find which chunk is most relevant to a question, we need to convert text into numbers that capture **meaning**.

**What is an Embedding?** A list of 768 numbers (a "vector") that represents the semantic meaning of text:

```
"Mercury is the smallest planet"  →  [0.23, 0.87, 0.11, 0.45, ... 768 numbers]
"What is the tiniest planet?"     →  [0.25, 0.85, 0.13, 0.44, ... 768 numbers]
                                       ↑ Very similar numbers! (similar meaning)

"Saturn has beautiful rings"      →  [0.91, 0.12, 0.78, 0.03, ... 768 numbers]
                                       ↑ Very different numbers! (different topic)
```

**How It Works in Our Code:**

```
Our Go App                              Ollama (local AI server)
    │                                           │
    │  POST /api/embeddings                     │
    │  {                                        │
    │    "model": "nomic-embed-text",           │
    │    "prompt": "Mercury is smallest..."     │
    │  }                                        │
    │ ─────────────────────────────────────────> │
    │                                           │  Neural network processes
    │                                           │  text and maps it to a
    │                                           │  point in 768-D space
    │  {"embedding": [0.23, 0.87, 0.11, ...]}   │
    │ <───────────────────────────────────────── │
```

**Why 768 Dimensions?** Think of coordinates. In 2D, a point is (x, y). In 3D, (x, y, z). Embeddings use 768 coordinates to capture nuances — some dimensions encode "is this about planets?", others "is this about size?", others "is this a question or statement?"

### 4. Vector Storage & Similarity Search (`rag/vectorstore.go`)

Embeddings are stored in a vector database. When a question comes in, we use **cosine similarity** to find the most relevant chunks:

```
INDEXING (one-time, when loading the PDF):

  "Mercury is smallest..."  ──embed──>  [0.23, 0.87, ...] ──store──> Vector DB
  "Venus is the second..."  ──embed──>  [0.91, 0.12, ...] ──store──> Vector DB
  "Earth supports life..."  ──embed──>  [0.45, 0.33, ...] ──store──> Vector DB

QUERYING (when user asks a question):

  "What is the tiniest planet?"  ──embed──>  [0.25, 0.85, ...]
                                                    │
                                          Compare with ALL stored vectors
                                          using cosine similarity
                                                    │
                                         ┌──────────┼──────────┐
                                         ▼          ▼          ▼
                                      Mercury    Venus      Earth
                                      score:0.95 score:0.3  score:0.4
                                         │
                                    MOST SIMILAR!
                                         │
                                         ▼
                        Send Mercury chunk to LLM as context
                                         │
                                         ▼
                    LLM answers: "Mercury is the smallest planet..."
```

Cosine similarity compares the **angle** between two vectors — vectors pointing in similar directions have similar meaning. Score of 0.95 = "almost identical meaning", 0.3 = "not very related."

### 5. Question Answering (`rag/agent.go`)
When you ask a question:
1. Your question is embedded using the same model
2. The vector store finds the 3 most similar chunks
3. A prompt is built: "Given this context: [chunks], answer: [question]"
4. The LLM generates an answer grounded in your document

---

## How Real-World Production Agents Compare

This project follows the **exact same pattern** used in production AI agents. Here is what real companies use at each step, compared to our implementation:

### Side-by-Side: Our Demo vs. Production

| Step | Our Demo (Free/Local) | Production (Paid/Cloud) |
|------|----------------------|------------------------|
| **Document Loading** | `ledongthuc/pdf` (Go lib) | Apache Tika, Unstructured.io, AWS Textract, LangChain document loaders |
| **Chunking** | Custom Go function | LangChain `RecursiveCharacterTextSplitter`, LlamaIndex `SentenceSplitter`, semantic chunking |
| **Embedding Model** | Ollama + `nomic-embed-text` (local) | OpenAI `text-embedding-3-large`, Cohere Embed, Google Vertex AI, AWS Bedrock |
| **Vector Database** | In-memory (our `VectorStore`) | Pinecone, Weaviate, ChromaDB, Qdrant, Milvus, pgvector (PostgreSQL) |
| **Similarity Search** | Brute-force cosine similarity | HNSW algorithm, IVF index (approximate nearest neighbor — same concept, much faster) |
| **LLM** | Ollama + `mistral` (local) | OpenAI GPT-4, Anthropic Claude, Google Gemini, AWS Bedrock |
| **Orchestration** | Our `Agent` struct | LangChain, LlamaIndex, Semantic Kernel, Haystack |

### What's the Same (Core Concepts)

The fundamental pipeline is **identical** in production:

```
Document → Chunk → Embed → Store in Vector DB → Retrieve → Augment Prompt → Generate
```

This is the standard RAG pattern. Every production system (ChatGPT with file upload, Notion AI, GitHub Copilot for docs, customer support bots) uses this exact flow.

### What's Different (Scale & Features)

| Feature | Our Demo | Production |
|---------|----------|------------|
| **Scale** | Dozens of chunks | Millions/billions of chunks |
| **Search Speed** | O(n) brute force | O(log n) using HNSW/IVF indexes |
| **Persistence** | In-memory (lost on restart) | Durable database (Pinecone, pgvector) |
| **Chunking Strategy** | Fixed character size | Semantic chunking (by meaning), sentence-level, or recursive |
| **Embedding API** | Local (free, slower) | Cloud API (paid, faster, higher quality) |
| **Re-ranking** | None | Cohere Rerank, cross-encoder models (re-score top results for better accuracy) |
| **Hybrid Search** | Vector only | Vector + keyword search combined |
| **Multi-tenancy** | Single user | Isolated data per customer/user |
| **Observability** | Print statements | LangSmith, Weights & Biases, tracing |
| **Memory** | None (stateless) | Conversation history, user preferences |
| **Evaluation** | Manual testing | RAGAS framework, automated quality scoring |

### LangChain — The Industry Standard Orchestrator

Our `Agent` struct in `rag/agent.go` is a simplified version of what **LangChain** provides. LangChain (Python/JS) is the most popular framework and offers:

- **Chains** — compose multiple LLM calls together (like our `LoadPDF` + `Ask` flow)
- **Agents** — LLMs that can decide which tools to use (web search, calculator, database queries)
- **Retrievers** — pluggable vector store integrations (swap ChromaDB for Pinecone with one line)
- **Memory** — automatic conversation history management
- **Callbacks** — logging, tracing, streaming

Our project teaches you the **fundamentals that LangChain abstracts away**. Once you understand how chunking, embeddings, and vector search work from our raw implementation, LangChain's abstractions will make much more sense.

### The Progression Path

```
1. THIS PROJECT          → Understand the raw mechanics (you are here)
2. Add ChromaDB/pgvector → Learn real vector database integration
3. Use LangChain (Go)    → Learn orchestration frameworks
4. Add Tools/Agents      → LLM decides what actions to take
5. Add Memory            → Multi-turn conversations
6. Add Evaluation        → Measure answer quality automatically
7. Deploy to Cloud       → Production-ready system
```

---

## API Keys / Tokens

**None needed!** This project uses:
- **Ollama** — runs AI models locally on your machine, completely free
- **nomic-embed-text** — open-source embedding model, no license needed
- **mistral** — open-source LLM, no license needed

Everything runs on your machine. No data leaves your computer.

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| "Cannot connect to Ollama" | Run `make setup` or `docker compose up -d ollama` |
| "Model not found" | Run `docker compose exec ollama ollama pull mistral` |
| "No text extracted from PDF" | The PDF might be image-based (scanned). Try a text-based PDF |
| Slow on first run | Models are loaded into memory on first query. Subsequent queries are faster |
| Out of memory | Try a smaller LLM: set `LLM_MODEL=llama3.2:1b` in docker-compose.yml |

---

## Next Steps for Learning

Once you're comfortable with this project, try:

1. **Try different models** — Change `LLM_MODEL` to `llama3.2`, `phi3`, or `gemma2`
2. **Tune chunk size** — Try 200, 500, 1000 chars and see how answers change
3. **Add a web UI** — Build an HTTP server with a simple HTML frontend
4. **Use ChromaDB** — Replace the in-memory store with a real vector database
5. **Add memory** — Store conversation history for follow-up questions
6. **Multi-document** — Load multiple PDFs into the same vector store
7. **Add metadata filtering** — Filter search results by page number or section

---

## Key Concepts Glossary

| Term | Definition |
|------|-----------|
| **RAG** | Retrieval-Augmented Generation — combining search with LLM generation |
| **Embedding** | A vector (list of numbers) representing the meaning of text |
| **Vector Store** | A database optimized for storing and searching embeddings |
| **Cosine Similarity** | A math formula measuring how similar two vectors are (0-1) |
| **Chunk** | A small piece of a larger document |
| **LLM** | Large Language Model — an AI that generates human-like text |
| **Ollama** | A tool for running LLMs locally on your computer |
| **Context Window** | Maximum amount of text an LLM can process at once |
| **Hallucination** | When an LLM confidently generates incorrect information |
| **Semantic Search** | Finding content by meaning, not just keyword matching |

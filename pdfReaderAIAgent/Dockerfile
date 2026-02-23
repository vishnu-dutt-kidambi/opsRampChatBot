# =============================================================================
# Dockerfile - Multi-stage build for the PDF Q&A Agent
# =============================================================================
#
# WHAT IS A MULTI-STAGE BUILD?
# Stage 1 (builder): Uses a full Go image to compile the binary
# Stage 2 (runtime): Uses a tiny Alpine image to run the binary
# This keeps the final image small (~15MB instead of ~1GB)
# =============================================================================

# ------------------------------------
# Stage 1: Build the Go binary
# ------------------------------------
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /build

# Copy Go module files first (for Docker layer caching)
# If go.mod/go.sum haven't changed, Docker reuses the cached layer
COPY go.mod ./
COPY go.sum* ./

# Download dependencies (cached if go.mod unchanged)
RUN go mod download || true

# Copy ALL source code
COPY . .

# Ensure dependencies are resolved
RUN go mod tidy

# Build the binary
# CGO_ENABLED=0: Pure Go binary (no C dependencies)
# GOOS=linux: Target Linux (even if building on Mac)
RUN CGO_ENABLED=0 GOOS=linux go build -o pdf-qa-agent .

# ------------------------------------
# Stage 2: Minimal runtime image
# ------------------------------------
FROM alpine:3.19

# Install CA certificates (needed for HTTPS, just in case)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy ONLY the compiled binary from the builder stage
COPY --from=builder /build/pdf-qa-agent .

# Create directory for PDF files
RUN mkdir -p /app/pdfs

# Set the binary as the entry point
# Arguments passed to 'docker run' will be appended to this
ENTRYPOINT ["./pdf-qa-agent"]

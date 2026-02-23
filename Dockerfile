# =============================================================================
# Dockerfile - OpsRamp ChatBot (full workspace build)
# =============================================================================
# Build context must be the repository root so both Go modules
# (conversationalAgent + pdfReaderAIAgent) are available.
#
# Multi-stage build:
#   Stage 1 (builder) — compiles the Go binary
#   Stage 2 (runtime) — minimal Alpine image (~15MB)
# =============================================================================

# ------------------------------------
# Stage 1: Build
# ------------------------------------
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Copy go.work and both module directories' go.mod/go.sum first for layer caching
COPY go.work ./
COPY conversationalAgent/go.mod conversationalAgent/go.sum* ./conversationalAgent/
COPY pdfReaderAIAgent/go.mod pdfReaderAIAgent/go.sum* ./pdfReaderAIAgent/

# Download dependencies (cached if go.mod unchanged)
RUN cd conversationalAgent && go mod download

# Copy full source for both modules
COPY pdfReaderAIAgent/ ./pdfReaderAIAgent/
COPY conversationalAgent/ ./conversationalAgent/

# Build the conversational agent binary
RUN cd conversationalAgent && \
    CGO_ENABLED=0 GOOS=linux go build -o /out/opsramp-agent .

# ------------------------------------
# Stage 2: Minimal runtime
# ------------------------------------
FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary
COPY --from=builder /out/opsramp-agent .

# Copy static web assets
COPY --from=builder /src/conversationalAgent/web/ ./web/

# Copy runbooks (if any)
COPY conversationalAgent/runbooks/ ./runbooks/

# Default: web mode on port 8080
EXPOSE 8080
ENTRYPOINT ["./opsramp-agent"]
CMD ["--web"]

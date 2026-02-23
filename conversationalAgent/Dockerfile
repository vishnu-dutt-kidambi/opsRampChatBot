# Build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o opsramp-agent .

# Runtime stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/opsramp-agent .
ENTRYPOINT ["./opsramp-agent"]

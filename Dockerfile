# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build arguments for versioning
ARG VERSION=dev
ARG BUILD_TIME
ARG GIT_COMMIT

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME} -X main.gitCommit=${GIT_COMMIT}" \
    -o /kibana-exporter ./cmd/exporter

# Final stage - minimal image
FROM scratch

# Copy CA certificates for HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary
COPY --from=builder /kibana-exporter /kibana-exporter

# Use non-root user
USER 65534:65534

EXPOSE 9684

ENTRYPOINT ["/kibana-exporter"]

# Simplified Dockerfile for cagent

FROM golang:1.25.3-alpine AS builder

WORKDIR /src

# Install build tools if necessary (git often needed for go mod)
RUN apk add --no-cache git ca-certificates

# Copy custom CA cert if present
COPY ca_chain.crt /usr/local/share/ca-certificates/custom-ca.crt
RUN update-ca-certificates

# Environment settings for Go
ENV GOPROXY=direct
ENV GOSUMDB=off

# Copy dependencies
COPY go.mod go.sum ./
# Skipping explicit go mod download to see if go build works or gives better error
# RUN go mod download

# Copy source
COPY . .

# Build for local architecture
# Disable CGO for static binary
RUN CGO_ENABLED=0 go build -v -trimpath -ldflags "-s -w" -o /app/cagent .

# Final stage
FROM alpine:latest

# Runtime dependencies
RUN apk add --no-cache ca-certificates docker-cli

# Setup user
RUN addgroup -S cagent && adduser -S -G cagent cagent

ENV DOCKER_MCP_IN_CONTAINER=1
ENV TERM=xterm-256color

# Setup directories
RUN mkdir -p /data /work && \
    chmod 777 /data /work

# Copy mcp-gateway (assuming public image)
COPY --from=docker/mcp-gateway:v2 /docker-mcp /usr/local/lib/docker/cli-plugins/

# Copy binary
COPY --from=builder /app/cagent /cagent

USER cagent
WORKDIR /work
ENTRYPOINT ["/cagent"]

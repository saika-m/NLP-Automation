# Build stage
FROM golang:1.21-bullseye AS builder

# Set working directory
WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    libc6-dev \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o nlp-automation-backend main.go

# Final stage
FROM debian:bullseye-slim

# Install runtime dependencies including sudo
RUN apt-get update && apt-get install -y \
    ca-certificates \
    sqlite3 \
    wget \
    sudo \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Create user with sudo privileges
RUN groupadd -g 1001 nlp-automation && \
    useradd -r -u 1001 -g nlp-automation -d /app -s /bin/bash nlp-automation && \
    echo 'nlp-automation ALL=(ALL) NOPASSWD: ALL' >> /etc/sudoers

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/nlp-automation-backend .

# Create required directories and set permissions
RUN mkdir -p uploads outputs temp data && \
    chown -R nlp-automation:nlp-automation /app

# Create entrypoint script that runs as root first to fix permissions, then switches to nlp-automation
RUN echo '#!/bin/bash' > /entrypoint.sh && \
    echo 'set -e' >> /entrypoint.sh && \
    echo '# Create directories as root first' >> /entrypoint.sh && \
    echo 'mkdir -p /app/data /app/data/uploads /app/data/outputs /app/data/temp' >> /entrypoint.sh && \
    echo '# Fix ownership' >> /entrypoint.sh && \
    echo 'chown -R nlp-automation:nlp-automation /app/data /app/uploads /app/outputs /app/temp' >> /entrypoint.sh && \
    echo '# Switch to nlp-automation user and execute command' >> /entrypoint.sh && \
    echo 'exec su nlp-automation -c "$*"' >> /entrypoint.sh && \
    chmod +x /entrypoint.sh

# Set entrypoint
ENTRYPOINT ["/entrypoint.sh"]

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run as root initially to handle permissions, then switch to nlp-automation via entrypoint
CMD ["./nlp-automation-backend"]

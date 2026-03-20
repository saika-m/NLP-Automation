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
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o tashi-backend main.go

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
RUN groupadd -g 1001 tashi && \
    useradd -r -u 1001 -g tashi -d /app -s /bin/bash tashi && \
    echo 'tashi ALL=(ALL) NOPASSWD: ALL' >> /etc/sudoers

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/tashi-backend .

# Create required directories and set permissions
RUN mkdir -p uploads outputs temp data && \
    chown -R tashi:tashi /app

# Create entrypoint script that runs as root first to fix permissions, then switches to tashi
RUN echo '#!/bin/bash' > /entrypoint.sh && \
    echo 'set -e' >> /entrypoint.sh && \
    echo '# Create directories as root first' >> /entrypoint.sh && \
    echo 'mkdir -p /app/data /app/data/uploads /app/data/outputs /app/data/temp' >> /entrypoint.sh && \
    echo '# Fix ownership' >> /entrypoint.sh && \
    echo 'chown -R tashi:tashi /app/data /app/uploads /app/outputs /app/temp' >> /entrypoint.sh && \
    echo '# Switch to tashi user and execute command' >> /entrypoint.sh && \
    echo 'exec su tashi -c "$*"' >> /entrypoint.sh && \
    chmod +x /entrypoint.sh

# Set entrypoint
ENTRYPOINT ["/entrypoint.sh"]

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run as root initially to handle permissions, then switch to tashi via entrypoint
CMD ["./tashi-backend"]

# Stage 1: Download dependencies
FROM golang:1.24-bookworm AS deps

WORKDIR /app

# Copy go workspace files
COPY go.work go.work.sum ./

# Copy all go.mod and go.sum files for dependency caching
COPY store/go.mod store/go.sum ./store/
COPY middleware/go.mod middleware/go.sum ./middleware/
COPY analytics/go.mod analytics/go.sum ./analytics/
COPY retries/go.mod retries/go.sum ./retries/
COPY shortener/go.mod shortener/go.sum ./shortener/
COPY redirector/go.mod redirector/go.sum ./redirector/

# Download dependencies
RUN go mod download

# Stage 2: Build the application
FROM golang:1.24-bookworm AS builder

WORKDIR /app

# Copy cached dependencies
COPY --from=deps /go/pkg /go/pkg

# Copy source code
COPY go.work go.work.sum ./
COPY store/ ./store/
COPY middleware/ ./middleware/
COPY analytics/ ./analytics/
COPY retries/ ./retries/
COPY shortener/ ./shortener/
COPY redirector/ ./redirector/

# Build the redirector binary
WORKDIR /app/redirector
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o redirector .

# Final stage: Run the application
FROM debian:bookworm-slim

WORKDIR /app

# Install wget for health checks and create non-root user
RUN apt-get update && apt-get install -y --no-install-recommends \
    wget \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd -r redirector \
    && useradd -r -g redirector redirector

# Copy the built application
COPY --from=builder /app/redirector/redirector .

# Change ownership of the application binary
RUN chown redirector:redirector /app/redirector

# Switch to the non-root user
USER redirector

# Expose port
EXPOSE 3001

# Run the binary
CMD ["./redirector"]

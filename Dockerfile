# Build stage
FROM golang:1.26rc2-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bot ./cmd/bot

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates && \
    addgroup -g 1000 appuser && adduser -u 1000 -D -G appuser appuser

WORKDIR /home/appuser

# Copy the binary from builder
COPY --from=builder /app/bot .

# Copy migrations (needed for production deployments)
COPY --from=builder /app/migrations ./migrations

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./bot"]

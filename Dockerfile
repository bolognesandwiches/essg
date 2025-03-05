FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o essg-api ./cmd/api

# Create final image
FROM timescale/timescaledb:2.11.2-pg14 AS postgres

# Install PostGIS and its dependencies using Alpine package manager
RUN apk add --no-cache postgis postgresql14-contrib

# Ensure extensions are available
RUN echo "CREATE EXTENSION IF NOT EXISTS postgis;" > /docker-entrypoint-initdb.d/enable_postgis.sql \
    && echo "CREATE EXTENSION IF NOT EXISTS timescaledb;" >> /docker-entrypoint-initdb.d/enable_postgis.sql

FROM alpine:3.16

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/essg-api .

# Expose API port
EXPOSE 8080

# Run the application
CMD ["./essg-api"]

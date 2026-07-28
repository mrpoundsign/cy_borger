# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application (modernc.org/sqlite is CGO free, so CGO_ENABLED=0 works perfectly)
RUN CGO_ENABLED=0 GOOS=linux go build -o cy_borger ./cmd/cy_borger

# Run stage
FROM alpine:latest

WORKDIR /app

# Install tzdata for timezone support if needed by Go
RUN apk add --no-cache tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/cy_borger .

# Create a directory for the database to enable volume mounting
RUN mkdir -p /app/data

# Expose port
EXPOSE 8080

# Environment variables
ENV DB_PATH=/app/data/cy_borger.db
ENV PORT=8080

# Command to run
CMD ["./cy_borger"]

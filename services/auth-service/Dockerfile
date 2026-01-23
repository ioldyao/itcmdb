# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy shared code
COPY shared /app/shared

# Copy service code
COPY auth-service /app/service
WORKDIR /app/service

# Download dependencies
RUN go mod download
RUN go mod tidy

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service ./cmd/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/service/auth-service .

# Copy config
COPY auth-service/internal/config/config.yaml /etc/auth-service/

EXPOSE 5001

CMD ["./auth-service"]

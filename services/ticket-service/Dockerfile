FROM golang:1.21-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git

# Copy shared code
COPY shared /app/shared

# Copy service code
COPY ticket-service /app/service
WORKDIR /app/service

RUN go mod download
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o ticket-service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/service/ticket-service .
COPY ticket-service/internal/config/config.yaml /etc/ticket-service/
EXPOSE 5003
CMD ["./ticket-service"]

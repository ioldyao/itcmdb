FROM golang:1.21-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git

# Copy shared code
COPY shared /app/shared

# Copy service code
COPY alert-service /app/service
WORKDIR /app/service

RUN go mod download
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o alert-service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/service/alert-service .
COPY alert-service/internal/config/config.yaml /etc/alert-service/
EXPOSE 5004
CMD ["./alert-service"]

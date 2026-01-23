FROM golang:1.23-alpine AS builder
WORKDIR /app

# 设置Go代理为阿里云
ENV GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
ENV GO111MODULE=on

RUN apk add --no-cache git

# Copy shared code
COPY shared /app/shared

# Copy service code
COPY notification-service /app/service
WORKDIR /app/service

RUN go mod download
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o notification-service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/service/notification-service .
COPY notification-service/internal/config/config.yaml /etc/notification-service/
EXPOSE 5005
CMD ["./notification-service"]

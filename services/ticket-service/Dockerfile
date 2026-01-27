FROM golang:1.23-alpine AS builder

# 设置 Alpine 镜像源为阿里云
# RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

WORKDIR /app

# 设置Go代理为阿里云
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

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

# 设置 Alpine 镜像源为阿里云
# RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
#     apk --no-cache add ca-certificates tzdata
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/
COPY --from=builder /app/service/ticket-service .
COPY ticket-service/internal/config/config.yaml /etc/ticket-service/
EXPOSE 5003
CMD ["./ticket-service"]

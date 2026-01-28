FROM golang:1.23-alpine AS go-builder

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
COPY alert-service /app/service
WORKDIR /app/service

RUN go mod download
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o alert-service ./cmd/main.go

# Python stage for workflow engine
FROM python:3.11-alpine AS python-builder

WORKDIR /app

# Install build dependencies for pydantic-core (需要 Rust)
RUN apk add --no-cache gcc musl-dev libffi-dev rust cargo

# 安装 Python 依赖
COPY alert-service/python/requirements.txt /app/requirements.txt
RUN pip install --no-cache-dir -r /app/requirements.txt --target /app/python

# Final stage
FROM alpine:latest

# 设置 Alpine 镜像源为阿里云
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk --no-cache add ca-certificates tzdata python3 py3-pip

WORKDIR /root/

# Copy Go binary
COPY --from=go-builder /app/service/alert-service .

# Copy Python workflow engine
COPY --from=python-builder /app/python /usr/local/lib/python3.11/site-packages
COPY alert-service/python/workflow_engine /usr/local/lib/python3.11/site-packages/workflow_engine

# Copy config
COPY alert-service/internal/config/config.yaml /etc/alert-service/

EXPOSE 5004
CMD ["./alert-service"]

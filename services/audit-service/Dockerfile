FROM golang:1.23-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 设置Go代理
ENV GOPROXY=https://goproxy.cn,https://goproxy.io,direct

# 复制shared代码
COPY shared /app/shared

# 复制audit-service代码
COPY audit-service /app/service

WORKDIR /app/service

# 下载依赖并编译
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o audit-service ./cmd/main.go

# 运行时镜像
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root

# 复制编译好的二进制文件
COPY --from=builder /app/service/audit-service .

# 复制配置文件
COPY --from=builder /app/service/internal/config/config.yaml /etc/audit-service/

EXPOSE 5007

CMD ["./audit-service"]

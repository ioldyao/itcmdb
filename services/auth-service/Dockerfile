FROM golang:1.23-alpine AS builder

# 设置 Alpine 镜像源为阿里云
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

WORKDIR /app

# 设置Go代理为阿里云
ENV GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
ENV GO111MODULE=on

RUN apk add --no-cache git protobuf protobuf-dev

# 安装protoc-gen-go和protoc-gen-go-grpc
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.1
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

# Copy shared code
COPY shared /app/shared

# 重新生成proto代码
WORKDIR /app/shared
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/auth/auth.proto proto/cmdb/cmdb.proto proto/alert/alert.proto proto/ticket/ticket.proto

# Copy service code
COPY auth-service /app/service
WORKDIR /app/service

RUN go mod download
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service ./cmd/main.go

FROM alpine:latest

# 设置 Alpine 镜像源为阿里云
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/service/auth-service .
COPY auth-service/internal/config/config.yaml /etc/auth-service/
EXPOSE 5001
CMD ["./auth-service"]

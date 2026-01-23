#!/bin/bash
# 构建脚本 - 构建所有Go微服务

set -e

echo "开始构建所有服务..."

# 构建认证服务
echo "构建 auth-service..."
cd services/auth-service
go mod tidy
go build -o bin/auth-service ./cmd/main.go
cd ../..

# 构建CMDB服务
echo "构建 cmdb-service..."
cd services/cmdb-service
go mod tidy
go build -o bin/cmdb-service ./cmd/main.go
cd ../..

# 构建工单服务
echo "构建 ticket-service..."
cd services/ticket-service
go mod tidy
go build -o bin/ticket-service ./cmd/main.go
cd ../..

# 构建告警服务
echo "构建 alert-service..."
cd services/alert-service
go mod tidy
go build -o bin/alert-service ./cmd/main.go
cd ../..

# 构建通知服务
echo "构建 notification-service..."
cd services/notification-service
go mod tidy
go build -o bin/notification-service ./cmd/main.go
cd ../..

# 构建报表服务
echo "构建 report-service..."
cd services/report-service
go mod tidy
go build -o bin/report-service ./cmd/main.go
cd ../..

# 构建前端
echo "构建 frontend..."
cd frontend
npm install
npm run build
cd ..

echo "所有服务构建完成！"

#!/bin/bash
# 启动脚本 - 启动所有服务

set -e

echo "启动 ITCMDB 系统..."

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    echo "错误: Docker 未安装，请先安装 Docker"
    exit 1
fi

# 检查 Docker Compose v2
if ! docker compose version &> /dev/null; then
    echo "错误: Docker Compose v2 未安装，请先安装 Docker Compose v2"
    exit 1
fi

# 构建并启动所有服务
echo "构建 Docker 镜像..."
docker compose build 

echo "启动所有服务..."
docker compose up -d

# 等待数据库启动
echo "等待数据库启动..."
sleep 5

# 检查服务状态
echo "检查服务状态..."
docker compose ps

echo ""
echo "ITCMDB 系统启动完成！"
echo ""
echo "服务访问地址："
echo "  - 前端应用: http://localhost"
echo "  - API网关: http://localhost:8000"
echo "  - 认证服务: http://localhost:5001"
echo "  - CMDB服务: http://localhost:5002"
echo "  - 工单服务: http://localhost:5003"
echo "  - 告警服务: http://localhost:5004"
echo "  - 通知服务: http://localhost:5005"
echo "  - 报表服务: http://localhost:5006"
echo "  - Kafka UI: http://localhost:8080"
echo ""
echo "默认管理员账号: admin / admin123"
echo ""
echo "查看日志: docker compose logs -f [service-name]"
echo "停止服务: docker compose down"

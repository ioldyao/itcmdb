#!/bin/bash
# Docker构建脚本 - 构建Docker镜像并自动清理悬空镜像

set -e

echo "开始构建Docker镜像..."

# 构建指定的服务，如果没有参数则构建所有服务
if [ $# -eq 0 ]; then
    echo "构建所有服务..."
    docker compose build
else
    echo "构建服务: $@"
    docker compose build "$@"
fi

echo ""
echo "构建完成，开始清理悬空镜像..."

# 清理悬空镜像
PRUNED=$(docker image prune -f)
echo "$PRUNED"

# 统计清理结果
if echo "$PRUNED" | grep -q "Total reclaimed space: 0B"; then
    echo "没有悬空镜像需要清理"
else
    echo "清理完成！"
fi

#!/bin/bash
# 停止脚本 - 停止所有服务

set -e

echo "停止 ITCMDB 系统..."

docker compose down

echo "所有服务已停止"
echo ""
echo "如需同时删除数据卷，请运行: docker compose down -v"

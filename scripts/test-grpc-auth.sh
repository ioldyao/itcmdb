#!/bin/bash

# 测试gRPC认证是否生效的脚本
# 使用方法: ./scripts/test-grpc-auth.sh <token>

CMDB_HOST="${CMDB_HOST:-localhost}"
CMDB_PORT="${CMDB_PORT:-50002}"
TOKEN="${1:-test-token}"

echo "=========================================="
echo "测试CMDB gRPC认证功能"
echo "=========================================="
echo "服务器: $CMDB_HOST:$CMDB_PORT"
echo "Token: $TOKEN"
echo ""

# 检查grpcurl是否安装
if ! command -v grpcurl &> /dev/null; then
    echo "❌ grpcurl未安装"
    echo "请安装: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
    exit 1
fi

echo "1️⃣  测试不带Token的请求（应该失败）"
echo "------------------------------------------"
grpcurl -plaintext "$CMDB_HOST:$CMDB_PORT" cmdb.CMDBService/GetCITypes 2>&1
RESULT=$?
echo ""

if [ $RESULT -eq 0 ]; then
    echo "❌ 认证未生效！不带Token也能访问"
else
    echo "✅ 正确：不带Token被拒绝"
fi
echo ""

echo "2️⃣  测试带错误Token的请求（应该失败）"
echo "------------------------------------------"
grpcurl -plaintext \
    -authorization "Bearer wrong-token-12345" \
    "$CMDB_HOST:$CMDB_PORT" \
    cmdb.CMDBService/GetCITypes 2>&1
RESULT=$?
echo ""

if [ $RESULT -eq 0 ]; then
    echo "❌ 认证未生效！错误的Token也能访问"
else
    echo "✅ 正确：错误Token被拒绝"
fi
echo ""

echo "3️⃣  测试带正确Token的请求（应该成功）"
echo "------------------------------------------"
grpcurl -plaintext \
    -authorization "Bearer $TOKEN" \
    "$CMDB_HOST:$CMDB_PORT" \
    cmdb.CMDBService/GetCITypes 2>&1
RESULT=$?
echo ""

if [ $RESULT -eq 0 ]; then
    echo "✅ 正确：有效Token可以访问"
else
    echo "❌ 认证配置可能有问题，有效Token也被拒绝"
fi
echo ""

echo "4️⃣  测试HardwareService (Agent使用的服务)"
echo "------------------------------------------"
grpcurl -plaintext \
    -authorization "Bearer $TOKEN" \
    "$CMDB_HOST:$CMDB_PORT" \
    list 2>&1 | grep -i hardware
echo ""

echo "=========================================="
echo "测试完成"
echo "=========================================="
echo ""
echo "📝 说明："
echo "  - 如果测试1和2都通过了，说明认证已正确配置"
echo "  - 如果测试失败，说明需要重新构建CMDB服务："
echo "    docker-compose build cmdb-service"
echo "    docker-compose up -d cmdb-service"
echo ""

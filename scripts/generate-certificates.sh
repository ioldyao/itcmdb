#!/bin/bash

# 生成mTLS所需的证书
# 用于CMDB服务和Agent之间的双向认证

set -e

CERT_DIR="${CERT_DIR:-./certificates}"
DAYS="${DAYS:-365}"

echo "=========================================="
echo "生成mTLS证书"
echo "=========================================="
echo "证书目录: $CERT_DIR"
echo "有效期: $DAYS 天"
echo ""

# 创建证书目录
mkdir -p "$CERT_DIR"

cd "$CERT_DIR"

echo "1️⃣  生成CA私钥和证书..."
openssl genrsa -out ca_key.pem 4096
openssl req -x509 -new -nodes -key ca_key.pem -sha256 -days $DAYS \
    -out ca_cert.pem \
    -subj "/C=CN/ST=Beijing/L=Beijing/O=ITCMDB/CN=ITCMDB-CA"

echo "✅ CA证书生成完成: ca_cert.pem"
echo ""

echo "2️⃣  生成服务端证书和私钥..."
openssl genrsa -out server_key.pem 4096
openssl req -new -key server_key.pem -out server.csr \
    -subj "/C=CN/ST=Beijing/L=Beijing/O=ITCMDB/CN=cmdb-service"

# 使用CA签名服务端证书
openssl x509 -req -in server.csr -CA ca_cert.pem -CAkey ca_key.pem \
    -CAcreateserial -out server_cert.pem -days $DAYS -sha256 \
    -extfile <(echo "subjectAltName=DNS:cmdb-service,DNS:localhost,IP:127.0.0.1")

# 清理CSR文件
rm server.csr

echo "✅ 服务端证书生成完成: server_cert.pem, server_key.pem"
echo ""

echo "3️⃣  生成客户端证书和私钥（用于Agent）..."
openssl genrsa -out client_key.pem 4096
openssl req -new -key client_key.pem -out client.csr \
    -subj "/C=CN/ST=Beijing/L=Beijing/O=ITCMDB/CN=hardware-agent"

# 使用CA签名客户端证书
openssl x509 -req -in client.csr -CA ca_cert.pem -CAkey ca_key.pem \
    -CAcreateserial -out client_cert.pem -days $DAYS -sha256

# 清理CSR文件
rm client.csr

echo "✅ 客户端证书生成完成: client_cert.pem, client_key.pem"
echo ""

echo "4️⃣  验证证书..."
echo "验证服务端证书..."
openssl verify -CAfile ca_cert.pem server_cert.pem

echo "验证客户端证书..."
openssl verify -CAfile ca_cert.pem client_cert.pem

echo ""
echo "=========================================="
echo "✅ 证书生成完成！"
echo "=========================================="
echo ""
echo "生成的文件："
echo "  📁 CA证书: ca_cert.pem"
echo "  🔑 CA私钥: ca_key.pem"
echo "  📁 服务端证书: server_cert.pem"
echo "  🔑 服务端私钥: server_key.pem"
echo "  📁 客户端证书: client_cert.pem"
echo "  🔑 客户端私钥: client_key.pem"
echo ""
echo "⚠️  重要提示："
echo "  1. 请妥善保管 ca_key.pem（泄露后可签发任意证书）"
echo "  2. client_cert.pem 和 client_key.pem 需要部署到Agent"
echo "  3. server_cert.pem 和 server_key.pem 需要部署到CMDB服务"
echo "  4. ca_cert.pem 需要部署到服务端和客户端（用于验证对端证书）"
echo ""
echo "🔒 文件权限建议："
echo "  chmod 600 *_key.pem"
echo "  chmod 644 *.pem"
echo ""

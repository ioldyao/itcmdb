# CMDB gRPC mTLS双向认证部署指南

## 概述

本文档说明如何为CMDB服务和Agent配置mTLS（双向TLS）认证，确保：
- ✅ 服务端验证Agent身份（通过客户端证书）
- ✅ Agent验证服务端身份（通过服务端证书）
- ✅ 所有通信加密（TLS 1.2+）
- ✅ 防止中间人攻击
- ✅ 防止证书伪造（CA签名）

## 安全原理

### 传统方案（不安全 ❌）
```
Agent ───────> CMDB服务
  ↑
  └─ 只发送Token，服务端验证

问题：
- 中间人可以冒充服务端
- Agent不知道在跟谁说话
- Token可能被截获
```

### mTLS方案（安全 ✅）
```
Agent ←──TLS加密──> CMDB服务
  ↓                    ↑
客户端证书          服务端证书
  ↓                    ↑
  └────── CA签名 ──────┘

验证流程：
1. Agent验证服务端证书（由CA签名）
2. 服务端验证Agent证书（由CA签名）
3. 双方都信任同一个CA
```

## 快速开始

### 1. 生成证书

```bash
cd /root/itcmdb

# 运行证书生成脚本
./scripts/generate-certificates.sh

# 证书将生成在 ./certificates/ 目录
ls -la certificates/
# ca_cert.pem         - CA根证书
# ca_key.pem          - CA私钥（妥善保管！）
# server_cert.pem     - 服务端证书
# server_key.pem      - 服务端私钥
# client_cert.pem     - Agent客户端证书
# client_key.pem      - Agent客户端私钥
```

### 2. 部署证书

#### CMDB服务端

```bash
# 1. 复制证书到容器可访问的位置
mkdir -p /opt/itcmdb/certs
cp certificates/ca_cert.pem /opt/itcmdb/certs/
cp certificates/server_cert.pem /opt/itcmdb/certs/
cp certificates/server_key.pem /opt/itcmdb/certs/

# 2. 设置权限
chmod 644 /opt/itcmdb/certs/*.pem
chmod 600 /opt/itcmdb/certs/*_key.pem

# 3. 修改docker-compose.yml
```

编辑 `docker-compose.yml`:

```yaml
services:
  cmdb-service:
    environment:
      # 启用mTLS
      - CMDB_GRPC_MTLS_ENABLED=true
      - CMDB_GRPC_MTLS_SERVER_CERT=/certs/server_cert.pem
      - CMDB_GRPC_MTLS_SERVER_KEY=/certs/server_key.pem
      - CMDB_GRPC_MTLS_CA_CERT=/certs/ca_cert.pem
    volumes:
      # 挂载证书目录
      - /opt/itcmdb/certs:/certs:ro
```

#### Agent客户端

```bash
# 1. 创建证书目录
mkdir -p /etc/hardware_agent/certs

# 2. 复制客户端证书
cp certificates/client_cert.pem /etc/hardware_agent/certs/
cp certificates/client_key.pem /etc/hardware_agent/certs/
cp certificates/ca_cert.pem /etc/hardware_agent/certs/

# 3. 设置权限（只有root可读）
chmod 600 /etc/hardware_agent/certs/*
chown root:root /etc/hardware_agent/certs/*

# 4. 更新Agent配置文件
```

编辑 `/etc/hardware_agent/config.yaml`:

```yaml
report:
  enabled: true
  type: grpc
  grpc:
    address: cmdb-service:50002
    server_name: cmdb-service  # 重要！必须与服务端证书的CN或SAN匹配

    mtls:
      client_cert: /etc/hardware_agent/certs/client_cert.pem
      client_key: /etc/hardware_agent/certs/client_key.pem
      ca_cert: /etc/hardware_agent/certs/ca_cert.pem
```

### 3. 重启服务

```bash
# 重启CMDB服务
docker-compose down
docker-compose build cmdb-service
docker-compose up -d

# 查看启动日志
docker-compose logs -f cmdb-service | grep -i "mtls\|tls"

# 期望看到：
# mTLS enabled server_cert=/certs/server_cert.pem ca_cert=/certs/ca_cert.pem
# CMDB gRPC service starting with mTLS port=50002 auth=mtls
```

```bash
# 重启Agent
systemctl restart hardware-agent

# 查看Agent日志
journalctl -u hardware-agent -f
```

## 验证配置

### 1. 测试服务端启动

```bash
# 查看CMDB服务日志
docker-compose logs cmdb-service | grep "gRPC service"

# 应该看到：
# ✅ CMDB gRPC service starting with mTLS
# ❌ 如果看到 "without mTLS" 或 "insecure"，说明配置未生效
```

### 2. 测试Agent连接

```bash
# 手动运行Agent测试
hardware-agent --once --config /etc/hardware_agent/config.yaml

# 成功输出：
# ✅ 使用 gRPC 上报方式（已配置mTLS双向认证）
# ✅ 连接成功
# ✅ 数据上报成功

# 失败输出（证书错误）：
# ❌ 连接 gRPC 服务器失败: x509: certificate signed by unknown authority
# ❌ 连接 gRPC 服务器失败: x509: certificate is valid for localhost, not cmdb-service
```

### 3. 使用grpcurl测试

```bash
# 安装grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 测试1: 不带证书（应该失败）
grpcurl -plaintext localhost:50002 cmdb.CMDBService/GetCITypes
# ❌ Error: connection closed before server preface received

# 测试2: 带客户端证书（应该成功）
grpcurl \
  -cert /etc/hardware_agent/certs/client_cert.pem \
  -key /etc/hardware_agent/certs/client_key.pem \
  -cacert /etc/hardware_agent/certs/ca_cert.pem \
  localhost:50002 \
  cmdb.CMDBService/GetCITypes
# ✅ 返回CI类型列表
```

## 配置详解

### 服务端配置项

| 环境变量 | 说明 | 示例值 | 必填 |
|---------|------|--------|------|
| `CMDB_GRPC_MTLS_ENABLED` | 是否启用mTLS | `true` | 是 |
| `CMDB_GRPC_MTLS_SERVER_CERT` | 服务端证书路径 | `/certs/server_cert.pem` | 是 |
| `CMDB_GRPC_MTLS_SERVER_KEY` | 服务端私钥路径 | `/certs/server_key.pem` | 是 |
| `CMDB_GRPC_MTLS_CA_CERT` | CA证书路径 | `/certs/ca_cert.pem` | 是 |

### 客户端配置项

| 配置项 | 说明 | 示例值 | 必填 |
|-------|------|--------|------|
| `grpc.address` | 服务端地址 | `cmdb-service:50002` | 是 |
| `grpc.server_name` | 服务器名称（证书验证） | `cmdb-service` | 是 |
| `grpc.mtls.client_cert` | 客户端证书路径 | `/etc/hardware_agent/certs/client_cert.pem` | 是 |
| `grpc.mtls.client_key` | 客户端私钥路径 | `/etc/hardware_agent/certs/client_key.pem` | 是 |
| `grpc.mtls.ca_cert` | CA证书路径 | `/etc/hardware_agent/certs/ca_cert.pem` | 是 |

## 常见问题

### Q1: Agent报错 "certificate signed by unknown authority"

**原因**: Agent使用的CA证书与服务端不一致

**解决**:
```bash
# 检查Agent的CA证书
sha256sum /etc/hardware_agent/certs/ca_cert.pem

# 检查服务端的CA证书
docker-compose exec cmdb-service sha256sum /certs/ca_cert.pem

# 两者应该一致！如果不一致，重新复制
```

### Q2: Agent报错 "certificate is valid for localhost, not cmdb-service"

**原因**: 服务器名称配置错误

**解决**:
```bash
# 查看服务端证书的实际CN或SAN
openssl x509 -in certificates/server_cert.pem -noout -text | grep -A 1 "Subject:"

# 或者查看SAN
openssl x509 -in certificates/server_cert.pem -noout -text | grep "Subject Alternative Name"

# 确保配置文件中的 server_name 与证书一致
# 如果证书是 localhost，配置改为：
# server_name: localhost

# 如果需要重新生成证书（指定正确的SAN）
# 修改 generate-certificates.sh 中的 subjectAltName
```

### Q3: 服务端报错 "private key does not match public key"

**原因**: 证书和私钥不匹配

**解决**:
```bash
# 重新生成证书
./scripts/generate-certificates.sh

# 重新部署
cp certificates/server_* /opt/itcmdb/certs/
docker-compose restart cmdb-service
```

### Q4: 如何生成多个Agent证书？

**方法1: 使用相同证书（简单但不推荐）**
```bash
# 所有Agent使用同一个client_cert.pem和client_key.pem
# 适合内网环境，Agent数量少
```

**方法2: 为每个Agent生成独立证书（推荐）**
```bash
# 为Agent-1生成证书
openssl genrsa -out agent1_key.pem 4096
openssl req -new -key agent1_key.pem -out agent1.csr -subj "/CN=agent1"
openssl x509 -req -in agent1.csr -CA ca_cert.pem -CAkey ca_key.pem \
    -CAcreateserial -out agent1_cert.pem -days 365 -sha256

# 为Agent-2生成证书
openssl genrsa -out agent2_key.pem 4096
openssl req -new -key agent2_key.pem -out agent2.csr -subj "/CN=agent2"
openssl x509 -req -in agent2.csr -CA ca_cert.pem -CAkey ca_key.pem \
    -CAcreateserial -out agent2_cert.pem -days 365 -sha256
```

### Q5: 证书过期怎么办？

**查看证书有效期**:
```bash
openssl x509 -in certificates/server_cert.pem -noout -dates
```

**重新生成证书**:
```bash
# 修改脚本中的 DAYS 变量
DAYS=1095 ./scripts/generate-certificates.sh  # 3年有效期

# 重新部署
cp certificates/*.pem /opt/itcmdb/certs/
docker-compose restart cmdb-service
systemctl restart hardware-agent
```

### Q6: 生产环境安全建议

1. **保护CA私钥**:
   ```bash
   # CA私钥离线保存，不要放在生产服务器
   chmod 400 ca_key.pem
   ```

2. **使用强加密算法**:
   ```bash
   # 脚本已使用：
   # - RSA 4096位
   # - SHA256签名
   # - TLS 1.2+
   ```

3. **定期轮换证书**:
   ```bash
   # 建议：每1-2年轮换一次服务端证书
   # 每6个月-1年轮换一次客户端证书
   ```

4. **监控证书过期**:
   ```bash
   # 创建定时任务检查证书有效期
   0 0 * * * /usr/local/bin/check-cert-expiry.sh
   ```

5. **限制证书撤销**:
   ```bash
   # 生产环境可以考虑使用CRL或OCSP
   # 但对于小型内网部署，直接重新生成证书更简单
   ```

## 开发环境配置

如果开发环境不想配置mTLS，可以禁用：

```yaml
# docker-compose.yml
services:
  cmdb-service:
    environment:
      - CMDB_GRPC_MTLS_ENABLED=false  # 禁用mTLS
```

**警告**: 禁用mTLS后，任何人都可以连接到服务端！仅用于开发测试。

## 参考资料

- [gRPC Go官方文档 - mTLS](https://github.com/grpc/grpc-go/blob/master/examples/features/encryption/README.md)
- [OpenSSL证书生成](https://www.openssl.org/docs/)
- [TLS最佳实践](https://wiki.mozilla.org/Security/Server_Side_TLS)

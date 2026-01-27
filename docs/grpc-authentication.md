# CMDB gRPC 认证配置说明

## 概述

CMDB服务的gRPC接口现已启用认证机制，Agent需要提供有效的Token才能上报硬件信息。

## 架构说明

### gRPC服务器端（CMDB Service）

- **端口**: 50002（可通过环境变量配置）
- **认证方式**: Bearer Token
- **拦截器链**:
  1. `LoggingInterceptor`: 记录所有gRPC调用日志
  2. `UnaryAgentAuthInterceptor`: 验证Agent的Bearer Token

### Agent客户端（Hardware Agent）

- **上报方式**: gRPC
- **认证方式**: Per-RPC Credentials（在每个请求中添加Authorization头）

## 配置步骤

### 1. 配置CMDB服务端Token

#### 方式一：环境变量（推荐）

在`docker-compose.yml`中添加环境变量：

```yaml
services:
  cmdb-service:
    environment:
      - CMDB_GRPC_AGENT_TOKEN=your-secret-agent-token-here
```

#### 方式二：配置文件

在`config.yaml`中添加：

```yaml
grpc:
  port: "50002"
  agent_token: "your-secret-agent-token-here"
```

### 2. 配置Agent端Token

编辑Agent配置文件`/etc/hardware_agent/config.yaml`：

```yaml
report:
  enabled: true
  type: grpc

  grpc:
    address: cmdb-service:50002  # 或实际IP:端口
    token: your-secret-agent-token-here  # 必须与服务端配置一致
```

### 3. 重启服务

```bash
# 重启CMDB服务
docker-compose restart cmdb-service

# 重启Agent
systemctl restart hardware-agent
```

## 认证流程

```
┌─────────────┐                           ┌──────────────┐
│    Agent    │                           │ CMDB Service │
└──────┬──────┘                           └──────┬───────┘
       │                                         │
       │ 1. ReportHardwareInfo (with Bearer Token)│
       │----------------------------------------->│
       │                                         │
       │                              2. 拦截器验证Token
       │                                         │
       │                    ┌────────────────────┴────────────┐
       │                    │                                  │
       │              Token有效?                         Token无效?
       │                    │                                  │
       │                    ▼                                  ▼
       │            3. 处理请求                   返回 Unauthenticated
       │                    │                          Error
       │ 4. 返回成功响应    │                                  │
       │<-----------------------------------------┘
       │                                         │
```

## 安全建议

### 生产环境

1. **使用强随机Token**:
   ```bash
   # 生成32字节随机token
   openssl rand -base64 32
   ```

2. **启用TLS加密**:
   - 修改`grpc_reporter.go`: `RequireTransportSecurity()` 返回 `true`
   - 配置证书路径

3. **定期轮换Token**:
   - 每90天更换一次Token
   - 使用配置管理工具（如Vault）管理Token

4. **使用环境变量或密钥管理系统**:
   - 不要将Token硬编码在配置文件中
   - 使用Kubernetes Secrets或Docker Secrets

### 开发环境

默认Token: `hardware-agent-token-default`

**警告**: 仅用于开发和测试，生产环境必须修改！

## 验证配置

### 1. 检查服务端日志

```bash
docker-compose logs -f cmdb-service | grep grpc
```

正常启动应该看到：
```
CMDB gRPC service starting with authentication port=50002 auth=agent-token
```

### 2. 测试Agent连接

```bash
# 手动运行Agent测试
hardware-agent --once --config /etc/hardware_agent/config.yaml
```

成功输出：
```
使用 gRPC 上报方式（已配置认证Token）
开始收集硬件信息...
数据上报成功
收集完成
```

失败输出（Token错误）：
```
rpc error: code = Unauthenticated desc = invalid agent token
```

### 3. 使用grpcurl测试

```bash
# 安装grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 测试连接（带Token）
grpcurl -plaintext \
  -authorization "Bearer your-token-here" \
  localhost:50002 \
  cmdb.CMDBService/GetCITypes
```

## 故障排查

### 错误1: `rpc error: code = Unauthenticated desc = missing authorization token`

**原因**: Agent未配置Token

**解决**:
1. 检查Agent配置文件中是否设置`report.grpc.token`
2. 确认配置文件路径正确

### 错误2: `rpc error: code = Unauthenticated desc = invalid agent token`

**原因**: Token与服务端不匹配

**解决**:
1. 检查CMDB服务端日志，确认配置的Token
2. 确保Agent和CMDB使用相同的Token
3. 检查Token前后是否有空格

### 错误3: `connection refused`

**原因**:
- CMDB服务未启动
- 端口配置错误
- 网络不通

**解决**:
1. 检查CMDB服务是否运行: `docker-compose ps`
2. 检查端口: `docker-compose exec cmdb-service env | grep GRPC`
3. 测试网络: `telnet cmdb-service 50002`

## 环境变量参考

| 环境变量 | 说明 | 默认值 |
|---------|------|--------|
| `CMDB_GRPC_PORT` | gRPC监听端口 | `50002` |
| `CMDB_GRPC_AGENT_TOKEN` | Agent认证Token | `hardware-agent-token-default` |
| `CMDB_AGENT_TOKEN` | Agent认证Token（别名） | `hardware-agent-token-default` |

## 相关文件

### 服务端

- `/root/itcmdb/services/cmdb-service/cmd/main.go` - gRPC服务器启动
- `/root/itcmdb/services/shared/pkg/middleware/grpc_interceptor.go` - 拦截器实现
- `/root/itcmdb/services/cmdb-service/internal/grpc/server.go` - gRPC Handler

### Agent端

- `/mnt/c/.../hardware_agent/cmd/agent/main.go` - Agent主程序
- `/mnt/c/.../hardware_agent/internal/reporter/grpc_reporter.go` - gRPC上报器
- `/mnt/c/.../hardware_agent/internal/config/config.go` - 配置解析

## 参考资料

- [gRPC Go官方文档 - 拦截器](https://github.com/grpc/grpc-go/blob/master/examples/features/interceptor/README.md)
- [gRPC认证最佳实践](https://grpc.io/docs/guides/auth/)
- [Per-RPC Credentials](https://pkg.go.dev/google.golang.org/grpc/credentials#PerRPCCredentials)

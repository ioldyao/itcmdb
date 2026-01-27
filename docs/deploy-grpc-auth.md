# gRPC认证部署和验证指南

## 问题诊断

如果Agent能够成功上报但Token验证似乎没有生效，**最可能的原因是CMDB服务没有重新构建**，Docker容器中运行的还是旧代码。

## 部署步骤

### 1. 重新构建CMDB服务镜像

```bash
# 在itcmdb根目录执行
cd /root/itcmdb

# 停止现有服务
docker-compose down

# 重新构建cmdb-service镜像
docker-compose build cmdb-service

# 启动服务
docker-compose up -d
```

### 2. 验证服务启动

```bash
# 查看服务状态
docker-compose ps

# 查看启动日志，应该看到认证相关信息
docker-compose logs -f cmdb-service | grep -i "grpc\|auth\|interceptor"
```

**期望看到的日志**:
```
CMDB gRPC service starting with authentication port=50002 auth=agent-token
```

### 3. 验证Token配置

检查环境变量是否正确设置：

```bash
# 检查实际配置的Token
docker-compose exec cmdb-service env | grep -i token

# 或者查看完整环境变量
docker-compose exec cmdb-service env | grep CMDB
```

## 验证认证是否生效

### 方法1: 使用测试脚本（推荐）

```bash
# 给脚本执行权限
chmod +x /root/itcmdb/scripts/test-grpc-auth.sh

# 测试默认token
./scripts/test-grpc-auth.sh hardware-agent-token-default

# 测试自定义token（如果配置了）
./scripts/test-grpc-auth.sh your-custom-token
```

### 方法2: 使用grpcurl手动测试

```bash
# 安装grpcurl（如果没有）
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 测试1: 不带token（应该失败）
grpcurl -plaintext localhost:50002 cmdb.CMDBService/GetCITypes
# 预期: Error: 代码 = Unauthenticated

# 测试2: 带错误token（应该失败）
grpcurl -plaintext \
    -authorization "Bearer wrong-token" \
    localhost:50002 \
    cmdb.CMDBService/GetCITypes
# 预期: Error: 代码 = Unauthenticated desc = invalid agent token

# 测试3: 带正确token（应该成功）
grpcurl -plaintext \
    -authorization "Bearer hardware-agent-token-default" \
    localhost:50002 \
    cmdb.CMDBService/GetCITypes
# 预期: 返回CI类型列表
```

### 方法3: 查看Agent上报日志

```bash
# 如果Agent上报成功，查看CMDB日志
docker-compose logs -f cmdb-service | grep "gRPC call"

# 应该看到类似这样的日志：
# DEBUG gRPC call method=/cmdb.HardwareService/ReportHardwareInfo
```

## 常见问题排查

### Q1: Agent还是能上报成功，Token验证没生效？

**检查清单**:
1. ✅ 确认重新构建了镜像: `docker images | grep cmdb-service`
2. ✅ 确认重启了容器: `docker-compose ps`
3. ✅ 查看启动日志，确认有"authentication"字样
4. ✅ 运行测试脚本验证认证是否生效

**诊断命令**:
```bash
# 查看容器创建时间（应该是刚构建的）
docker-compose ps cmdb-service

# 查看镜像构建时间
docker inspect itcmdb-cmdb-service | grep Created

# 进入容器查看代码
docker-compose exec cmdb-service cat /app/cmd/main.go | grep ChainUnaryInterceptor
```

### Q2: Agent上报失败，提示"Unauthenticated"

**原因**: Token配置不匹配

**解决**:
```bash
# 1. 查看CMDB期望的Token
docker-compose exec -e CMDB_GRPC_AGENT_TOKEN cmdb-service env | grep AGENT_TOKEN

# 2. 查看Agent配置的Token
cat /etc/hardware_agent/config.yaml | grep token

# 3. 确保两者一致，然后重启Agent
systemctl restart hardware-agent
```

### Q3: 重新构建后还是不生效？

**可能原因**: Docker使用了缓存

**强制重新构建**:
```bash
# 不使用缓存重新构建
docker-compose build --no-cache cmdb-service

# 删除旧容器和镜像
docker-compose down
docker rmi itcmdb-cmdb-service

# 重新构建和启动
docker-compose build cmdb-service
docker-compose up -d
```

## 验证检查表

- [ ] 代码已修改并提交（commit ac8fee4）
- [ ] CMDB服务已重新构建: `docker-compose build cmdb-service`
- [ ] CMDB服务已重启: `docker-compose up -d cmdb-service`
- [ ] 启动日志包含"authentication"字样
- [ ] 运行测试脚本验证认证生效
- [ ] Agent配置的Token与CMDB一致
- [ ] Agent能够成功上报数据

## 快速重新部署命令

```bash
cd /root/itcmdb
docker-compose down
docker-compose build --no-cache cmdb-service
docker-compose up -d
docker-compose logs -f cmdb-service | grep -i "grpc.*auth"
```

## 下一步

认证验证通过后，可以：

1. **修改默认Token**（生产环境必须）
   ```yaml
   # docker-compose.yml
   environment:
     - CMDB_GRPC_AGENT_TOKEN=<生成强随机token>
   ```

2. **配置Agent使用新Token**
   ```yaml
   # /etc/hardware_agent/config.yaml
   report:
     grpc:
       token: <与CMDB相同的新token>
   ```

3. **启用TLS**（生产环境推荐）
   - 修改Agent的`RequireTransportSecurity()`返回`true`
   - 配置证书路径

## 参考资料

- [gRPC认证部署和验证指南](/root/itcmdb/docs/grpc-authentication.md)
- [测试脚本](/root/itcmdb/scripts/test-grpc-auth.sh)
- [gRPC官方文档 - 拦截器](https://github.com/grpc/grpc-go/blob/master/examples/features/interceptor/README.md)

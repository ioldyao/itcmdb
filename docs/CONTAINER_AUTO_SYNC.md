# 容器自动同步功能说明

## 概述

本系统实现了从 VictoriaMetrics 自动发现和同步容器信息到 CMDB 的功能。系统会定期查询 VictoriaMetrics 中的容器指标，自动创建、更新容器 CI 实例，并跟踪容器的生命周期。

## 功能特性

### 1. 自动容器发现
- 从 VictoriaMetrics 自动发现所有容器
- 支持 HTTPS 连接并跳过证书验证
- 使用基本认证（用户名/密码）

### 2. 容器生命周期管理
- **新容器自动创建**：发现新容器时自动创建 CI 实例
- **容器状态同步**：实时更新容器在线/离线状态
- **容器重建检测**：检测容器名称相同但 ID 变化的情况（容器重建）
- **历史记录保留**：记录容器的所有 ID 历史

### 3. 智能同步策略（方案 C）
- 以 `container_name` 作为主键
- 记录 `container_id` 历史，追踪容器重建
- 在线状态判断：最近 5 分钟内有指标数据
- 离线判断：连续 2 个同步周期无数据

## 配置说明

### 配置文件 (config.yaml)

```yaml
victoriametrics:
  endpoint: https://10.120.43.230:8109
  username: admin
  password: prometheus@123
  sync_interval: 5m  # 同步间隔，默认 5 分钟
```

### 环境变量 (compose.yml)

```yaml
environment:
  CMDB_VICTORIAMETRICS_ENDPOINT: https://10.120.43.230:8109
  CMDB_VICTORIAMETRICS_USERNAME: admin
  CMDB_VICTORIAMETRICS_PASSWORD: prometheus@123
  CMDB_VICTORIAMETRICS_SYNC_INTERVAL: 5m  # 可选
```

## 使用前提

### 1. 创建容器 CI 类型

在使用自动同步功能前，需要在 CMDB 中手动创建名为 **"容器"** 的 CI 类型：

1. 登录 CMDB 前端
2. 进入 CI 类型管理
3. 创建新的 CI 类型：
   - 名称：`容器`（必须是这个名称）
   - 描述：`Docker 容器实例`
   - 图标：`container`（可选）

### 2. VictoriaMetrics 指标要求

系统需要以下 Prometheus 指标：

- `container_last_seen{name="容器名"}` - 容器最后发现时间
- `container_cpu_usage_seconds_total{name="容器名"}` - CPU 使用
- `container_memory_working_set_bytes{name="容器名"}` - 内存使用
- `container_spec_memory_limit_bytes{name="容器名"}` - 内存限制
- `container_network_receive_bytes_total{name="容器名"}` - 网络接收
- `container_network_transmit_bytes_total{name="容器名"}` - 网络发送
- `container_fs_usage_bytes{name="容器名"}` - 磁盘使用
- `container_start_time_seconds{name="容器名"}` - 启动时间

## 容器 CI 实例属性

自动创建的容器 CI 实例包含以下属性：

```json
{
  "container_name": "容器名称",
  "container_id": "当前容器ID",
  "container_image": "镜像名称",
  "is_online": true/false,
  "last_seen": "2026-01-25T10:30:00Z",
  "container_id_history": ["id1", "id2", "id3"],
  "sync_source": "victoriametrics",
  "auto_discovered": true,
  "last_online_at": "2026-01-25T10:30:00Z",
  "last_offline_at": "2026-01-25T09:00:00Z",
  "rebuild_detected_at": "2026-01-25T10:00:00Z"
}
```

## 部署步骤

### 1. 重新构建服务

```bash
docker compose up -d --build cmdb-service
```

### 2. 查看日志

```bash
docker compose logs -f cmdb-service
```

你应该看到类似的日志：

```
INFO    VictoriaMetrics client initialized    endpoint=https://10.120.43.230:8109
INFO    Container auto-sync service started   interval=5m0s
INFO    Starting container synchronization
INFO    Discovered containers from VictoriaMetrics    total=15
INFO    Container synchronization completed   created=10 updated=3 marked_offline=2 marked_online=0 rebuilt=0
```

### 3. 验证功能

#### 检查 VictoriaMetrics 连接

```bash
curl -X GET "http://localhost:8000/api/v1/monitoring/victoriametrics/health" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

#### 查看同步的容器

登录 CMDB 前端，进入 CI 实例列表，筛选 CI 类型为"容器"，你应该能看到自动同步的容器实例。

## API 接口

### 1. 获取容器监控数据

```
GET /api/v1/monitoring/containers/{ci_id}/stats
```

返回容器的实时监控数据（CPU、内存、网络、磁盘等）。

### 2. VictoriaMetrics 健康检查

```
GET /api/v1/monitoring/victoriametrics/health
```

检查 VictoriaMetrics 服务是否可用。

## 同步逻辑详解

### 同步流程

1. **发现阶段**
   - 查询 VictoriaMetrics 获取所有容器列表
   - 检查每个容器的在线状态

2. **对比阶段**
   - 获取 CMDB 中现有的容器 CI 实例
   - 创建容器名称到实例的映射

3. **同步阶段**
   - **新容器**：创建新的 CI 实例
   - **现有容器**：
     - 检查容器 ID 是否变化（重建检测）
     - 更新在线状态
     - 更新镜像信息
     - 更新最后发现时间
   - **消失的容器**：标记为离线

### 容器重建检测

当检测到容器名称相同但 ID 变化时：

1. 更新 `container_id` 为新 ID
2. 将新 ID 添加到 `container_id_history` 数组
3. 记录 `rebuild_detected_at` 时间戳
4. 记录日志：`Container rebuild detected`

### 在线/离线判断

- **在线**：最近 5 分钟内有 `container_cpu_usage_seconds_total` 指标数据
- **离线**：连续 2 个同步周期（默认 10 分钟）查询不到指标数据

## 故障排除

### 1. 容器未自动同步

**检查项：**
- VictoriaMetrics 连接是否正常
- 是否创建了"容器" CI 类型
- 查看 cmdb-service 日志是否有错误

**解决方法：**
```bash
# 检查日志
docker compose logs cmdb-service | grep -i "container sync"

# 测试 VictoriaMetrics 连接
curl -k -u admin:prometheus@123 \
  "https://10.120.43.230:8109/api/v1/query?query=up"
```

### 2. 证书验证失败

确保配置中使用了 `https://` 协议，系统会自动跳过证书验证。

### 3. 同步间隔太长/太短

修改配置文件中的 `sync_interval` 或环境变量 `CMDB_VICTORIAMETRICS_SYNC_INTERVAL`：

```yaml
# 更快的同步（2 分钟）
sync_interval: 2m

# 更慢的同步（10 分钟）
sync_interval: 10m
```

### 4. 容器类型不存在错误

如果看到错误：`容器 CI 类型不存在，请先在 CMDB 中创建名为'容器'的 CI 类型`

需要手动在 CMDB 前端创建名为"容器"的 CI 类型。

## 性能考虑

- **同步间隔**：默认 5 分钟，可根据需要调整
- **批量查询**：每次同步最多查询 10000 个 CI 实例
- **缓存**：监控数据有 30 秒 Redis 缓存
- **并发**：同步服务在后台独立运行，不影响 API 响应

## 安全注意事项

1. **凭证管理**：VictoriaMetrics 密码存储在配置文件中，建议使用环境变量
2. **证书验证**：当前跳过 TLS 证书验证，生产环境建议配置正确的证书
3. **访问控制**：确保 VictoriaMetrics 只允许授权访问

## 未来改进

- [ ] 支持容器过滤规则（按名称前缀、标签等）
- [ ] 支持手动触发同步
- [ ] 添加同步统计和监控指标
- [ ] 支持多个 VictoriaMetrics 数据源
- [ ] 容器关系自动发现（容器 → 宿主机）

## 更新日期

最后更新：2026-01-25

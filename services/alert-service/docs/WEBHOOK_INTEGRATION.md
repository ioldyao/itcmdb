# Webhook 集成功能文档

## 概述

ITCMDB的Webhook集成功能允许您：

1. **接收外部告警** - 外部监控系统（如Alertmanager、Prometheus、VictoriaMetrics）可以向ITCMDB推送告警
2. **ITCMDB推送** - 将ITCMDB中的告警推送到外部系统，支持多种推送目标

## 架构设计

### 前端组件

- **位置**: `frontend/src/pages/Admin/AlertIntegration/Webhook.tsx`
- **技术栈**: React + TypeScript + Ant Design
- **特性**:
  - 使用 Tabs 组件分离两个功能区域
  - 支持CRUD操作
  - 实时统计数据展示
  - 一键复制Webhook URL

### 后端服务

- **位置**: `services/alert-service/`
- **技术栈**: Go + Gin + GORM
- **核心组件**:
  - `internal/models/webhook.go` - 数据模型定义
  - `internal/handlers/webhook.go` - HTTP处理器
  - `internal/services/webhook_service.go` - 业务逻辑

## API 端点

### 接收外部告警 (Inbound Webhooks)

#### 创建接收地址
```http
POST /api/v1/webhooks/inbound
Content-Type: application/json
Authorization: Bearer <jwt_token>

{
  "name": "Alertmanager主集群",
  "source_type": "alertmanager",
  "description": "接收主集群Alertmanager推送的告警"
}
```

响应:
```json
{
  "id": 1,
  "name": "Alertmanager主集群",
  "webhook_url": "https://itcmdb.example.com/api/v1/webhooks/inbound/abc123xyz",
  "source_type": "alertmanager",
  "enabled": true,
  "created_at": "2024-01-29T10:30:00Z"
}
```

#### 查询接收地址列表
```http
GET /api/v1/webhooks/inbound?page=1&page_size=20&source_type=alertmanager&enabled=true
```

#### 更新接收地址
```http
PUT /api/v1/webhooks/inbound/{id}
Content-Type: application/json

{
  "name": "更新后的名称",
  "enabled": true,
  "description": "更新后的描述"
}
```

#### 删除接收地址
```http
DELETE /api/v1/webhooks/inbound/{id}
```

#### 接收外部告警 (公开端点)
```http
POST /api/v1/webhooks/inbound/{token}
Content-Type: application/json

# Alertmanager格式
{
  "version": "4",
  "groupKey": "...",
  "status": "firing",
  "receiver": "email",
  "groupLabels": {...},
  "commonLabels": {...},
  "commonAnnotations": {...},
  "alerts": [...]
}
```

### ITCMDB推送 (Outbound Webhooks)

#### 创建推送目标
```http
POST /api/v1/webhooks/outbound
Content-Type: application/json
Authorization: Bearer <jwt_token>

{
  "name": "钉钉告警通知",
  "target_type": "dingtalk",
  "endpoint_url": "https://oapi.dingtalk.com/robot/send?access_token=xxx",
  "description": "推送告警到钉钉群",
  "secret": "SEC..."
}
```

支持的 target_type:
- `alertmanager` - Alertmanager
- `dingtalk` - 钉钉
- `wechat` - 企业微信
- `feishu` - 飞书
- `email` - 邮件
- `webhook` - 自定义Webhook

#### 测试推送目标
```http
POST /api/v1/webhooks/outbound/{id}/test
```

## 支持的告警格式

### Alertmanager
```json
{
  "version": "4",
  "groupKey": "...",
  "status": "firing",
  "alerts": [
    {
      "status": "firing",
      "labels": {
        "alertname": "HighMemory",
        "instance": "server1"
      },
      "annotations": {
        "summary": "High memory usage",
        "description": "Memory usage is above 90%"
      },
      "startsAt": "2024-01-29T10:00:00Z",
      "endsAt": "0001-01-01T00:00:00Z",
      "fingerprint": "..."
    }
  ]
}
```

### Prometheus
```json
{
  "version": "4",
  "alerts": [
    {
      "labels": {...},
      "annotations": {...},
      "startsAt": "2024-01-29T10:00:00Z",
      "endsAt": "0001-01-01T00:00:00Z"
    }
  ]
}
```

### VictoriaMetrics
```json
{
  "receiverName": "email",
  "status": "firing",
  "alerts": [
    {
      "id": "...",
      "status": "firing",
      "labels": {...},
      "annotations": {...}
    }
  ]
}
```

### 自定义格式
```json
{
  "alert_id": "custom-001",
  "title": "自定义告警",
  "content": "这是一条自定义告警",
  "severity": "critical",
  "status": "firing",
  "metadata": {...}
}
```

## 数据库表结构

### inbound_webhooks
存储接收外部告警的Webhook配置。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | SERIAL | 主键 |
| name | VARCHAR(255) | 名称 |
| webhook_url | VARCHAR(500) | 自动生成的唯一接收地址 |
| source_type | VARCHAR(50) | 来源类型 |
| enabled | BOOLEAN | 是否启用 |
| description | TEXT | 描述 |
| last_received | TIMESTAMP | 最后接收时间 |

### outbound_webhooks
存储推送到外部系统的Webhook配置。

| 字段 | 类型 | 说明 |
|------|------|------|
| id | SERIAL | 主键 |
| name | VARCHAR(255) | 名称 |
| target_type | VARCHAR(50) | 目标类型 |
| endpoint_url | TEXT | 推送端点URL |
| enabled | BOOLEAN | 是否启用 |
| description | TEXT | 描述 |
| secret | VARCHAR(255) | 签名密钥 |
| last_sent | TIMESTAMP | 最后推送时间 |

### inbound_webhook_logs
记录Webhook接收日志。

### outbound_webhook_logs
记录Webhook推送日志。

## 部署步骤

### 1. 执行数据库迁移
```bash
psql -h localhost -U postgres -d itcmdb -f services/alert-service/migrations/002_webhook_integration.sql
```

### 2. 重新构建并启动服务
```bash
cd services/alert-service
go build -o alert-service ./cmd
./alert-service
```

### 3. 使用Docker Compose
```bash
docker-compose restart alert-service
```

## 配置示例

### Alertmanager 推送到 ITCMDB

在Alertmanager配置中添加webhook配置：
```yaml
receivers:
  - name: 'itcmdb'
    webhook_configs:
      - url: 'https://itcmdb.example.com/api/v1/webhooks/inbound/abc123xyz'
        send_resolved: true

route:
  receiver: 'itcmdb'
  group_by: ['alertname', 'cluster']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
```

### ITCMDB 推送到钉钉

1. 在钉钉群中添加自定义机器人，获取Webhook地址
2. 在ITCMDB中创建推送目标：
```json
{
  "name": "钉钉告警通知",
  "target_type": "dingtalk",
  "endpoint_url": "https://oapi.dingtalk.com/robot/send?access_token=xxx",
  "secret": "SEC..." // 可选，用于签名验证
}
```

## 使用场景

### 场景1: 集中管理告警

- **问题**: 多个监控系统产生的告警分散在不同平台
- **解决**: 配置多个Inbound Webhook，统一接收告警到ITCMDB

### 场景2: 告警多渠道分发

- **问题**: 告警需要同时发送到钉钉、企业微信、邮件等多个渠道
- **解决**: 配置多个Outbound Webhook，一次告警自动推送到所有渠道

### 场景3: 告警聚合和转发

- **问题**: Alertmanager告警需要转发到企业微信
- **解决**: 配置Inbound Webhook接收Alertmanager告警，再配置Outbound Webhook推送到企业微信

## 常见问题

### Q: Webhook URL是如何生成的？
A: 系统使用加密随机数生成器生成32位十六进制token，确保唯一性和安全性。

### Q: 如何确保Webhook调用的安全性？
A:
- 使用HTTPS加密传输
- Token随机生成，难以猜测
- 记录所有调用日志（IP、时间、内容）
- 可配置是否启用

### Q: 推送失败会重试吗？
A: 当前版本记录失败日志但不自动重试。可以通过查询日志手动处理。

### Q: 支持哪些告警格式？
A: 原生支持Alertmanager、Prometheus、VictoriaMetrics格式，同时支持自定义JSON格式。

## 后续优化

1. **重试机制** - 添加自动重试和指数退避
2. **告警聚合** - 相似告警聚合成一条推送
3. **限流控制** - 防止告警风暴
4. **模板定制** - 自定义不同渠道的告警消息模板
5. **邮件发送** - 完善SMTP邮件发送功能
6. **签名验证** - 增强Webhook调用的安全性

## 相关文档

- [Alertmanager Webhook配置](https://prometheus.io/docs/alerting/latest/configuration/#webhook_config)
- [钉钉机器人API](https://open.dingtalk.com/document/robots/custom-robot-access)
- [企业微信机器人API](https://developer.work.weixin.qq.com/document/path/91770)
- [飞书机器人API](https://open.feishu.cn/document/ukTMukTMukTM/uUTNz4SN1MjL1UzM)

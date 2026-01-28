# 工作流引擎部署指南

## 一键部署（已自动配置）

所有配置文件已自动更新，直接运行：

```bash
cd /root/itcmdb

# 停止现有服务
docker-compose down

# 重新构建（包含新的依赖）
docker-compose build

# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f workflow-engine
docker-compose logs -f alert-service
docker-compose logs -f frontend
```

## 验证部署

### 1. 检查服务状态
```bash
# 检查所有容器
docker-compose ps

# 应该看到 itcmdb-workflow-engine 和 itcmdb-alert-service 都在运行
```

### 2. 初始化数据库
```bash
# 执行迁移脚本
docker exec -i itcmdb-postgres psql -U postgres -d itcmdb < services/alert-service/migrations/002_create_workflow_tables.sql
```

### 3. 测试 Workflow Engine
```bash
# 健康检查
curl http://localhost:8000/

# 应该返回：
# {"service":"ITCMDB Workflow Engine","version":"1.0.0","status":"running",...}
```

### 4. 测试 Alert Service 集成
```bash
# 健康检查
curl http://localhost:5004/health

# 测试 Webhook 端点（需要 token）
# curl -X POST http://localhost:5004/api/v1/webhooks/test_token \
#   -H "Content-Type: application/json" \
#   -d '{"alert_data": {"status": "firing"}}'
```

### 5. 访问前端
```
http://localhost:8082
导航到：系统管理 > 告警集成 > Webhook
```

## 功能测试

### 测试 Inbound Webhook
1. 在前端创建一个 Inbound Webhook
2. 复制自动生成的 URL
3. 使用 curl 发送测试告警：
```bash
curl -X POST http://localhost:5004/api/v1/webhooks/生成的token \
  -H "Content-Type: application/json" \
  -d '[
    {
      "status": "firing",
      "labels": {
        "alertname": "HighCPU",
        "instance": "server-001"
      },
      "annotations": {
        "description": "CPU usage is above 90%"
      }
    }
  ]'
```

### 测试工作流可视化
1. 访问前端工作流编辑器
2. 拖拽节点创建流程
3. 保存后查看生成的 Pipeline JSON

## 故障排查

### Workflow Engine 无法启动
```bash
# 查看日志
docker-compose logs workflow-engine

# 检查 Python 模块
docker exec -it itcmdb-workflow-engine python3 -c "import workflow_engine; print('OK')"
```

### Alert Service 连接失败
```bash
# 检查环境变量
docker exec -it itcmdb-alert-service env | grep WORKFLOW

# 测试网络连通性
docker exec -it itcmdb-alert-service wget -O- http://workflow-engine:8000/
```

### 前端 React Flow 报错
```bash
# 检查依赖是否安装
docker exec -it itcmdb-frontend npm list @xyflow/react

# 如果没有，重新构建 frontend
docker-compose build frontend
docker-compose up -d frontend
```

## 配置说明

### 环境变量
```yaml
# Alert Service
ALERT_WORKFLOW_ENGINE_URL: http://workflow-engine:8000  # Python Engine 地址
ALERT_WORKFLOW_BASE_URL: http://localhost:5004         # 生成 Webhook URL 的基础地址

# Workflow Engine
WORKFLOW_ENGINE_HOST: 0.0.0.0
WORKFLOW_ENGINE_PORT: 8000
PYTHONPATH: /usr/local/lib/python3.11/site-packages
```

### 数据库表
- `workflows` - 工作流配置
- `webhooks` - Webhook 配置
- `workflow_executions` - 执行记录

### API 端点
```
# Workflow Engine (Python:8000)
GET  /                                    # 健康检查
POST /api/v1/workflow/execute             # 执行工作流
POST /api/v1/webhooks/{token}             # 接收告警
GET  /api/v1/workflows                    # 工作流列表

# Alert Service (Go:5004)
POST /api/v1/webhooks/:token              # Webhook 接收（代理到 Python）
GET  /api/v1/workflows                    # 工作流管理
GET  /api/v1/webhooks                     # Webhook 管理
```

## 下一步

1. ✅ 部署完成
2. 创建第一个工作流（接收 → 过滤 → 推送）
3. 配置实际的钉钉/企业微信 Webhook
4. 从 Alertmanager 发送测试告警
5. 查看执行日志和监控

需要帮助请查看日志文件！

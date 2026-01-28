# ITCMDB Workflow Engine

基于 Bamboo-Engine 的告警集成工作流引擎。

## 功能特性

- ✅ 可视化工作流配置
- ✅ 接收外部告警（Alertmanager/Prometheus/VictoriaMetrics）
- ✅ 推送到外部系统（钉钉/企业微信/飞书/Alertmanager）
- ✅ 告警过滤和转换
- ✅ 并行和条件分支
- ✅ 工作流暂停/恢复

## 安装依赖

```bash
cd /root/itcmdb/services/alert-service/python
pip install -r requirements.txt
```

## 启动服务

```bash
# 开发模式
python workflow_engine/main.py

# 或使用 uvicorn
uvicorn workflow_engine.api:app --host 0.0.0.0 --port 8000 --reload
```

## API 文档

启动后访问：http://localhost:8000/docs

### 核心端点

- `POST /api/v1/workflow/execute` - 执行工作流
- `POST /api/v1/workflow` - 创建工作流
- `GET /api/v1/workflows` - 获取工作流列表
- `GET /api/v1/workflow/{pipeline_id}/status` - 查询工作流状态
- `POST /api/v1/workflow/{pipeline_id}/pause` - 暂停工作流
- `POST /api/v1/workflow/{pipeline_id}/resume` - 恢复工作流
- `POST /api/v1/webhooks/{token}` - 接收 Webhook 告警
- `GET /api/v1/webhooks` - 获取 Webhook 列表

## 组件列表

### 接收组件
- `alert_receiver` - 接收告警

### 处理组件
- `alert_filter` - 过滤告警
- `alert_transform` - 转换告警格式

### 发送组件
- `sender_alertmanager` - 推送到 Alertmanager
- `sender_dingtalk` - 推送到钉钉
- `sender_wechat` - 推送到企业微信
- `sender_feishu` - 推送到飞书

## 工作流示例

### 简单工作流：接收告警 → 推送到钉钉

```python
{
  "id": "wf_001",
  "start_event": {...},
  "activities": {
    "act_1": {
      "type": "ServiceActivity",
      "component": {"code": "alert_receiver"},
      "inputs": {"source_type": "alertmanager"}
    },
    "act_2": {
      "type": "ServiceActivity",
      "component": {"code": "sender_dingtalk"},
      "inputs": {
        "webhook_url": "https://oapi.dingtalk.com/robot/send?access_token=xxx"
      }
    }
  },
  "flows": {...}
}
```

### 并行推送：推送到多个渠道

```python
{
  "activities": {
    "send_am": {"component": {"code": "sender_alertmanager"}},
    "send_dt": {"component": {"code": "sender_dingtalk"}},
    "send_wx": {"component": {"code": "sender_wechat"}}
  },
  "gateways": {
    "parallel": {"type": "ParallelGateway"},
    "converge": {"type": "ConvergeGateway"}
  }
}
```

## 开发说明

### 添加自定义组件

1. 在 `components/` 目录创建新文件
2. 继承 `Component` 类
3. 实现 `execute()` 和 `outputs()` 方法
4. 在 `components/__init__.py` 中导出

### 测试

```bash
# 运行测试（需要添加测试文件）
pytest
```

## 配置

环境变量：
- `WORKFLOW_ENGINE_HOST` - 服务地址（默认：0.0.0.0）
- `WORKFLOW_ENGINE_PORT` - 服务端口（默认：8000）
- `DATABASE_URL` - 数据库连接（可选）
- `REDIS_URL` - Redis 连接（可选）

## 注意事项

1. 当前使用内存存储，重启后数据丢失
2. 生产环境需要配置持久化存储
3. 需要配合 Django Runtime 使用 Bamboo-Engine
4. 组件执行需要处理超时和错误重试

## 后续开发

- [ ] 集成数据库持久化
- [ ] 添加工作流版本管理
- [ ] 实现工作流调试模式
- [ ] 添加执行日志和审计
- [ ] 性能优化和缓存

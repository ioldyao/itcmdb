-- 工作流相关表
-- 执行: psql -h localhost -U postgres -d itcmdb -f services/alert-service/migrations/002_create_workflow_tables.sql

-- 工作流表
CREATE TABLE IF NOT EXISTS workflows (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    direction VARCHAR(20) NOT NULL CHECK (direction IN ('inbound', 'outbound')),
    type VARCHAR(50) NOT NULL CHECK (type IN ('alertmanager', 'prometheus', 'victoriametrics', 'workflow')),
    pipeline TEXT NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE workflows IS '工作流配置表';
COMMENT ON COLUMN workflows.direction IS 'inbound: 接收外部告警, outbound: 推送到外部系统';
COMMENT ON COLUMN workflows.pipeline IS 'Bamboo-Engine Pipeline JSON 配置';

-- Webhook 配置表
CREATE TABLE IF NOT EXISTS webhooks (
    id SERIAL PRIMARY KEY,
    workflow_id INTEGER REFERENCES workflows(id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    direction VARCHAR(20) NOT NULL CHECK (direction IN ('inbound', 'outbound')),
    webhook_token VARCHAR(100) UNIQUE,
    webhook_url TEXT,
    type VARCHAR(50) NOT NULL CHECK (type IN ('alertmanager', 'prometheus', 'victoriametrics', 'workflow')),
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE webhooks IS 'Webhook 配置表';
COMMENT ON COLUMN webhooks.webhook_token IS 'Inbound Webhook Token（自动生成）';
COMMENT ON COLUMN webhooks.webhook_url IS 'Outbound 目标 URL';

-- 工作流执行记录表
CREATE TABLE IF NOT EXISTS workflow_executions (
    id SERIAL PRIMARY KEY,
    workflow_id INTEGER REFERENCES workflows(id) ON DELETE CASCADE,
    webhook_id INTEGER REFERENCES webhooks(id) ON DELETE SET NULL,
    pipeline_id VARCHAR(255) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('running', 'completed', 'failed', 'paused')),
    input_data TEXT,
    output_data TEXT,
    error_msg TEXT,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    finished_at TIMESTAMP
);

COMMENT ON TABLE workflow_executions IS '工作流执行记录表';
COMMENT ON COLUMN workflow_executions.pipeline_id IS 'Bamboo-Engine Pipeline 执行 ID';
COMMENT ON COLUMN workflow_executions.input_data IS '输入数据（JSON）';
COMMENT ON COLUMN workflow_executions.output_data IS '输出数据（JSON）';

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_webhooks_token ON webhooks(webhook_token);
CREATE INDEX IF NOT EXISTS idx_webhooks_direction ON webhooks(direction);
CREATE INDEX IF NOT EXISTS idx_webhooks_type ON webhooks(type);
CREATE INDEX IF NOT EXISTS idx_workflows_direction ON workflows(direction);
CREATE INDEX IF NOT EXISTS idx_workflows_type ON workflows(type);
CREATE INDEX IF NOT EXISTS idx_executions_pipeline_id ON workflow_executions(pipeline_id);
CREATE INDEX IF NOT EXISTS idx_executions_workflow_id ON workflow_executions(workflow_id);
CREATE INDEX IF NOT EXISTS idx_executions_status ON workflow_executions(status);
CREATE INDEX IF NOT EXISTS idx_executions_started_at ON workflow_executions(started_at DESC);

-- 插入示例工作流（可选）
INSERT INTO workflows (name, description, direction, type, pipeline, enabled) VALUES
(
    '示例：接收告警推送到钉钉',
    '从 Alertmanager 接收告警，经过过滤后推送到钉钉',
    'inbound',
    'workflow',
    '{
        "id": "example_pipeline",
        "start_event": {"id": "start_1", "type": "EmptyStartEvent", "name": "开始", "incoming": "", "outgoing": ""},
        "end_event": {"id": "end_1", "type": "EmptyEndEvent", "name": "结束", "incoming": [], "outgoing": ""},
        "activities": {
            "receive": {
                "id": "receive",
                "type": "ServiceActivity",
                "name": "接收告警",
                "component": {"code": "alert_receiver", "inputs": {}},
                "incoming": [],
                "outgoing": "",
                "error_ignore": false,
                "optional": false,
                "retryable": true,
                "skippable": true,
                "timeout": 0
            },
            "filter": {
                "id": "filter",
                "type": "ServiceActivity",
                "name": "过滤告警",
                "component": {"code": "alert_filter", "inputs": {}},
                "incoming": [],
                "outgoing": "",
                "error_ignore": false,
                "optional": false,
                "retryable": true,
                "skippable": true,
                "timeout": 0
            },
            "send": {
                "id": "send",
                "type": "ServiceActivity",
                "name": "推送到钉钉",
                "component": {"code": "sender_dingtalk", "inputs": {}},
                "incoming": [],
                "outgoing": "",
                "error_ignore": false,
                "optional": false,
                "retryable": true,
                "skippable": true,
                "timeout": 0
            }
        },
        "gateways": {},
        "flows": {},
        "data": {"inputs": {}, "outputs": []}
    }',
    false
) ON CONFLICT DO NOTHING;

-- Webhook集成表结构
-- 执行命令: psql -h localhost -U postgres -d itcmdb -f services/alert-service/migrations/002_webhook_integration.sql

-- 接收外部告警的Webhook配置表
CREATE TABLE IF NOT EXISTS inbound_webhooks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    webhook_url VARCHAR(500) NOT NULL UNIQUE,
    source_type VARCHAR(50) NOT NULL CHECK (source_type IN ('alertmanager', 'prometheus', 'victoriametrics', 'custom')),
    enabled BOOLEAN DEFAULT true,
    description TEXT,
    last_received TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_inbound_webhooks_source_type ON inbound_webhooks(source_type);
CREATE INDEX IF NOT EXISTS idx_inbound_webhooks_enabled ON inbound_webhooks(enabled);
CREATE INDEX IF NOT EXISTS idx_inbound_webhooks_created_at ON inbound_webhooks(created_at);

-- 推送到外部系统的Webhook配置表
CREATE TABLE IF NOT EXISTS outbound_webhooks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    target_type VARCHAR(50) NOT NULL CHECK (target_type IN ('alertmanager', 'receiver')),
    receiver_id INTEGER REFERENCES alert_receivers(id) ON DELETE SET NULL,
    endpoint_url TEXT,
    enabled BOOLEAN DEFAULT true,
    description TEXT,
    last_sent TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_outbound_webhooks_target_type ON outbound_webhooks(target_type);
CREATE INDEX IF NOT EXISTS idx_outbound_webhooks_receiver_id ON outbound_webhooks(receiver_id);
CREATE INDEX IF NOT EXISTS idx_outbound_webhooks_enabled ON outbound_webhooks(enabled);
CREATE INDEX IF NOT EXISTS idx_outbound_webhooks_created_at ON outbound_webhooks(created_at);

-- Webhook接收日志表
CREATE TABLE IF NOT EXISTS inbound_webhook_logs (
    id SERIAL PRIMARY KEY,
    webhook_id INTEGER NOT NULL REFERENCES inbound_webhooks(id) ON DELETE CASCADE,
    source_ip VARCHAR(50),
    user_agent TEXT,
    status_code INTEGER,
    request_data JSONB,
    response_data TEXT,
    error_message TEXT,
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_inbound_webhook_logs_webhook_id ON inbound_webhook_logs(webhook_id);
CREATE INDEX IF NOT EXISTS idx_inbound_webhook_logs_created_at ON inbound_webhook_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_inbound_webhook_logs_status_code ON inbound_webhook_logs(status_code);

-- Webhook推送日志表
CREATE TABLE IF NOT EXISTS outbound_webhook_logs (
    id SERIAL PRIMARY KEY,
    webhook_id INTEGER NOT NULL REFERENCES outbound_webhooks(id) ON DELETE CASCADE,
    alert_id VARCHAR(255),
    target_url TEXT,
    status_code INTEGER,
    request_data JSONB,
    response_data TEXT,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    sent_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_outbound_webhook_logs_webhook_id ON outbound_webhook_logs(webhook_id);
CREATE INDEX IF NOT EXISTS idx_outbound_webhook_logs_alert_id ON outbound_webhook_logs(alert_id);
CREATE INDEX IF NOT EXISTS idx_outbound_webhook_logs_created_at ON outbound_webhook_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_outbound_webhook_logs_status_code ON outbound_webhook_logs(status_code);

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_inbound_webhooks_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_outbound_webhooks_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建触发器
DROP TRIGGER IF EXISTS trigger_update_inbound_webhooks_updated_at ON inbound_webhooks;
CREATE TRIGGER trigger_update_inbound_webhooks_updated_at
    BEFORE UPDATE ON inbound_webhooks
    FOR EACH ROW
    EXECUTE FUNCTION update_inbound_webhooks_updated_at();

DROP TRIGGER IF EXISTS trigger_update_outbound_webhooks_updated_at ON outbound_webhooks;
CREATE TRIGGER trigger_update_outbound_webhooks_updated_at
    BEFORE UPDATE ON outbound_webhooks
    FOR EACH ROW
    EXECUTE FUNCTION update_outbound_webhooks_updated_at();

-- 插入示例数据
INSERT INTO inbound_webhooks (name, webhook_url, source_type, description)
VALUES
    ('Alertmanager主集群', 'https://itcmdb.example.com/api/v1/webhooks/inbound/abc123xyz', 'alertmanager', '接收主集群Alertmanager推送的告警'),
    ('VictoriaMetrics环境', 'https://itcmdb.example.com/api/v1/webhooks/inbound/def456uvw', 'victoriametrics', '接收VictoriaMetrics推送的告警')
ON CONFLICT (webhook_url) DO NOTHING;

INSERT INTO outbound_webhooks (name, target_type, endpoint_url, description, enabled)
VALUES
    ('推送至主Alertmanager', 'alertmanager', 'http://alertmanager:9093/api/v1/alerts', '将ITCMDB告警推送到主集群Alertmanager', true)
ON CONFLICT DO NOTHING;

-- 注意：receiver 类型的推送目标需要在配置告警接收人后手动创建

-- 添加注释
COMMENT ON TABLE inbound_webhooks IS '接收外部告警的Webhook配置表';
COMMENT ON TABLE outbound_webhooks IS '推送到外部系统的Webhook配置表';
COMMENT ON TABLE inbound_webhook_logs IS 'Webhook接收日志表';
COMMENT ON TABLE outbound_webhook_logs IS 'Webhook推送日志表';

COMMENT ON COLUMN inbound_webhooks.webhook_url IS '自动生成的唯一接收地址';
COMMENT ON COLUMN inbound_webhooks.source_type IS '告警来源类型';
COMMENT ON COLUMN inbound_webhooks.last_received IS '最后接收时间';

COMMENT ON COLUMN outbound_webhooks.endpoint_url IS '推送目标URL';
COMMENT ON COLUMN outbound_webhooks.target_type IS '推送目标类型';
COMMENT ON COLUMN outbound_webhooks.secret IS '签名密钥（可选）';
COMMENT ON COLUMN outbound_webhooks.last_sent IS '最后推送时间';

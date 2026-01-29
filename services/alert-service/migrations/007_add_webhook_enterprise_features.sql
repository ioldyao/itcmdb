-- 添加死信队列表
CREATE TABLE IF NOT EXISTS dead_letter_queues (
    id SERIAL PRIMARY KEY,
    webhook_id INTEGER NOT NULL,
    webhook_type VARCHAR(20) NOT NULL,
    alert_data JSONB,
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    last_retry_at TIMESTAMP,
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_dlq_webhook_id ON dead_letter_queues(webhook_id);
CREATE INDEX idx_dlq_status ON dead_letter_queues(status);

-- 添加Webhook指标表
CREATE TABLE IF NOT EXISTS webhook_metrics (
    id SERIAL PRIMARY KEY,
    webhook_id INTEGER NOT NULL UNIQUE,
    webhook_type VARCHAR(20) NOT NULL,
    total_requests BIGINT DEFAULT 0,
    success_requests BIGINT DEFAULT 0,
    failed_requests BIGINT DEFAULT 0,
    avg_response_time DOUBLE PRECISION DEFAULT 0,
    last_request_at TIMESTAMP,
    circuit_state VARCHAR(20) DEFAULT 'closed',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_webhook_metrics_webhook ON webhook_metrics(webhook_id, webhook_type);

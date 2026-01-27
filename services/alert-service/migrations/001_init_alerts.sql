-- 告警平台数据库Schema初始化脚本
-- 创建时间: 2026-01-28

-- ============================================
-- 1. 告警规则表 (alert_rules)
-- ============================================
CREATE TABLE IF NOT EXISTS alert_rules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,

    -- 规则配置
    metric_query TEXT NOT NULL,                    -- PromQL查询或简单指标名
    threshold_operator VARCHAR(10) NOT NULL,        -- >, <, >=, <=, ==, !=
    threshold_value FLOAT NOT NULL,                -- 阈值
    duration INTEGER DEFAULT 300,                  -- 持续时间(秒)，默认5分钟

    -- 告警属性
    severity VARCHAR(20) NOT NULL,                 -- critical, high, medium, low
    enabled BOOLEAN DEFAULT true,

    -- 关联配置
    ci_type_id INTEGER REFERENCES ci_types(id),    -- 关联CI类型
    notification_channels JSONB,                    -- 通知渠道配置

    -- 静默配置
    silenced_until TIMESTAMP,                       -- 静默截止时间

    -- 审计字段
    created_by INTEGER REFERENCES users(id),
    updated_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- 约束
    CONSTRAINT alert_rules_severity_check
        CHECK (severity IN ('critical', 'high', 'medium', 'low')),
    CONSTRAINT alert_rules_operator_check
        CHECK (threshold_operator IN ('>', '<', '>=', '<=', '==', '!='))
);

-- 索引
CREATE INDEX idx_alert_rules_enabled ON alert_rules(enabled) WHERE deleted_at IS NULL;
CREATE INDEX idx_alert_rules_severity ON alert_rules(severity);
CREATE INDEX idx_alert_rules_ci_type ON alert_rules(ci_type_id);

-- ============================================
-- 2. 告警实例表 (alert_instances)
-- ============================================
CREATE TABLE IF NOT EXISTS alert_instances (
    id SERIAL PRIMARY KEY,
    alert_id VARCHAR(64) NOT NULL UNIQUE,          -- 告警唯一标识（UUID）
    rule_id INTEGER REFERENCES alert_rules(id),

    -- 告警基本信息
    title VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'firing',   -- firing, acknowledged, resolved, closed

    -- 分类信息
    category VARCHAR(100),                         -- 用户体验/业务应用/CPU使用率等
    tags JSONB,                                    -- 标签数组
    object_type VARCHAR(100),                      -- 对象类型

    -- 目标信息
    target_info JSONB,                             -- 目标信息(host_ip, host_id, host_group等)
    affected_ci_id INTEGER REFERENCES ci_instances(id),

    -- 触发条件
    trigger_conditions JSONB,                      -- 触发条件详情
    metrics JSONB,                                 -- 指标数据(current_value, threshold_value, deviation)

    -- 去重指纹
    fingerprint VARCHAR(64) NOT NULL,              -- 用于告警去重和聚合

    -- 时间信息
    first_triggered TIMESTAMP NOT NULL,            -- 首次触发时间
    last_triggered TIMESTAMP NOT NULL,             -- 最后触发时间
    recovered_at TIMESTAMP,                        -- 恢复时间
    closed_at TIMESTAMP,                           -- 关闭时间

    -- 计数
    count INTEGER DEFAULT 1,                       -- 触发次数

    -- 处理信息
    handler INTEGER REFERENCES users(id),          -- 处理人
    handling_status VARCHAR(20),                   -- 未处理/处理中/已完成
    handling_notes TEXT,                           -- 处理备注
    acknowledged_at TIMESTAMP,                     -- 确认时间

    -- 通知信息
    notification_sent BOOLEAN DEFAULT false,       -- 是否已发送通知
    notification_channels JSONB,                   -- 通知渠道

    -- 审计字段
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- 约束
    CONSTRAINT alert_instances_severity_check
        CHECK (severity IN ('critical', 'high', 'medium', 'low')),
    CONSTRAINT alert_instances_status_check
        CHECK (status IN ('firing', 'acknowledged', 'resolved', 'closed'))
);

-- 索引
CREATE INDEX idx_alert_instances_status ON alert_instances(status);
CREATE INDEX idx_alert_instances_fingerprint ON alert_instances(fingerprint);
CREATE INDEX idx_alert_instances_rule_id ON alert_instances(rule_id);
CREATE INDEX idx_alert_instances_severity ON alert_instances(severity);
CREATE INDEX idx_alert_instances_ci_id ON alert_instances(affected_ci_id);
CREATE INDEX idx_alert_instances_time_range ON alert_instances(first_triggered, last_triggered);
CREATE INDEX idx_alert_instances_category ON alert_instances(category);

-- 复合索引（用于常见查询）
CREATE INDEX idx_alert_instances_composite ON alert_instances(status, severity, last_triggered DESC);

-- ============================================
-- 3. 告警历史表 (alert_history)
-- ============================================
CREATE TABLE IF NOT EXISTS alert_history (
    id SERIAL PRIMARY KEY,
    alert_id INTEGER REFERENCES alert_instances(id) ON DELETE CASCADE,

    -- 事件信息
    event_type VARCHAR(50) NOT NULL,               -- triggered, updated, acknowledged, resolved, closed
    old_status VARCHAR(20),
    new_status VARCHAR(20),

    -- 操作信息
    operated_by INTEGER REFERENCES users(id),
    operated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- 详细信息
    message TEXT,
    details JSONB,                                 -- 额外的详细信息

    -- 约束
    CONSTRAINT alert_history_event_type_check
        CHECK (event_type IN ('triggered', 'updated', 'acknowledged', 'resolved', 'closed'))
);

-- 索引
CREATE INDEX idx_alert_history_alert_id ON alert_history(alert_id);
CREATE INDEX idx_alert_history_operated_at ON alert_history(operated_at DESC);
CREATE INDEX idx_alert_history_event_type ON alert_history(event_type);

-- ============================================
-- 4. 告警静默表 (alert_silences)
-- ============================================
CREATE TABLE IF NOT EXISTS alert_silences (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    comment TEXT,

    -- 匹配规则
    matchers JSONB NOT NULL,                       -- 匹配规则 JSON数组

    -- 时间范围
    starts_at TIMESTAMP NOT NULL,
    ends_at TIMESTAMP NOT NULL,

    -- 创建信息
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- 状态
    active BOOLEAN DEFAULT true
);

-- 索引
CREATE INDEX idx_alert_silences_time ON alert_silences(starts_at, ends_at);
CREATE INDEX idx_alert_silences_active ON alert_silences(active) WHERE active = true;

-- ============================================
-- 5. 告警聚合表 (alert_aggregations)
-- ============================================
CREATE TABLE IF NOT EXISTS alert_aggregations (
    id SERIAL PRIMARY KEY,
    aggregation_key VARCHAR(255) NOT NULL UNIQUE,  -- 聚合键

    -- 聚合信息
    base_alert_id INTEGER REFERENCES alert_instances(id),
    alert_count INTEGER DEFAULT 1,                 -- 聚合的告警数量
    related_alert_ids JSONB,                       -- 相关告警ID列表

    -- 时间信息
    first_triggered TIMESTAMP NOT NULL,
    last_triggered TIMESTAMP NOT NULL,

    -- 更新时间
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_alert_aggregations_key ON alert_aggregations(aggregation_key);
CREATE INDEX idx_alert_aggregations_time ON alert_aggregations(first_triggered, last_triggered);

-- ============================================
-- 6. 触发器函数：自动更新 updated_at
-- ============================================
CREATE OR REPLACE FUNCTION update_alert_rules_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_alert_rules_updated_at
    BEFORE UPDATE ON alert_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_alert_rules_updated_at();

CREATE OR REPLACE FUNCTION update_alert_instances_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_alert_instances_updated_at
    BEFORE UPDATE ON alert_instances
    FOR EACH ROW
    EXECUTE FUNCTION update_alert_instances_updated_at();

CREATE OR REPLACE FUNCTION update_alert_silences_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_alert_silences_updated_at
    BEFORE UPDATE ON alert_silences
    FOR EACH ROW
    EXECUTE FUNCTION update_alert_silences_updated_at();

-- ============================================
-- 7. 触发器函数：自动记录告警历史
-- ============================================
CREATE OR REPLACE FUNCTION log_alert_changes()
RETURNS TRIGGER AS $$
BEGIN
    -- 状态变更时记录历史
    IF TG_OP = 'UPDATE' AND OLD.status IS DISTINCT FROM NEW.status THEN
        INSERT INTO alert_history (
            alert_id, event_type, old_status, new_status,
            operated_by, operated_at, message
        ) VALUES (
            NEW.id,
            CASE NEW.status
                WHEN 'acknowledged' THEN 'acknowledged'
                WHEN 'resolved' THEN 'resolved'
                WHEN 'closed' THEN 'closed'
                ELSE 'updated'
            END,
            OLD.status,
            NEW.status,
            NEW.handler,
            CURRENT_TIMESTAMP,
            '状态自动变更'
        );
    ELSIF TG_OP = 'INSERT' THEN
        -- 新建告警时记录触发事件
        INSERT INTO alert_history (
            alert_id, event_type, new_status, operated_at, message
        ) VALUES (
            NEW.id,
            'triggered',
            NEW.status,
            CURRENT_TIMESTAMP,
            '告警首次触发'
        );
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_log_alert_changes
    AFTER INSERT OR UPDATE ON alert_instances
    FOR EACH ROW
    EXECUTE FUNCTION log_alert_changes();

-- ============================================
-- 8. 注释
-- ============================================

-- alert_rules 注释
COMMENT ON TABLE alert_rules IS '告警规则表：定义监控告警的规则和条件';
COMMENT ON COLUMN alert_rules.metric_query IS 'PromQL查询语句或简单指标名';
COMMENT ON COLUMN alert_rules.threshold_operator IS '比较运算符：>, <, >=, <=, ==, !=';
COMMENT ON COLUMN alert_rules.duration IS '持续时间（秒），指标超过阈值的持续时间';
COMMENT ON COLUMN alert_rules.ci_type_id IS '关联的CI类型，用于限定规则适用范围';

-- alert_instances 注释
COMMENT ON TABLE alert_instances IS '告警实例表：存储实际触发的告警';
COMMENT ON COLUMN alert_instances.alert_id IS '告警唯一标识（UUID格式）';
COMMENT ON COLUMN alert_instances.fingerprint IS '告警指纹，用于去重和聚合';
COMMENT ON COLUMN alert_instances.target_info IS '目标信息JSON：包含IP、主机ID、主机组等';
COMMENT ON COLUMN alert_instances.metrics IS '指标数据JSON：包含当前值、阈值、偏差等';
COMMENT ON COLUMN alert_instances.status IS '告警状态：firing(触发中), acknowledged(已确认), resolved(已恢复), closed(已关闭)';

-- alert_history 注释
COMMENT ON TABLE alert_history IS '告警历史表：记录告警的所有变更历史';
COMMENT ON COLUMN alert_history.event_type IS '事件类型：triggered, updated, acknowledged, resolved, closed';

-- alert_silences 注释
COMMENT ON TABLE alert_silences IS '告警静默表：定义告警抑制规则';
COMMENT ON COLUMN alert_silences.matchers IS '匹配规则JSON数组，定义哪些告警需要被静默';

-- alert_aggregations 注释
COMMENT ON TABLE alert_aggregations IS '告警聚合表：存储聚合后的告警信息';
COMMENT ON COLUMN alert_aggregations.aggregation_key IS '聚合键，由规则ID+目标等生成';

-- ============================================
-- 9. 初始化数据（可选）
-- ============================================

-- 插入一些示例告警规则（仅用于演示）
INSERT INTO alert_rules (name, description, metric_query, threshold_operator, threshold_value, duration, severity, enabled, created_by, updated_by) VALUES
('CPU使用率告警', 'CPU使用率超过80%持续5分钟', '100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)', '>', 80.0, 300, 'high', true, 1, 1),
('内存使用率告警', '内存使用率超过90%持续5分钟', '(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100', '>', 90.0, 300, 'critical', true, 1, 1),
('磁盘使用率告警', '磁盘使用率超过85%', '(1 - (node_filesystem_avail_bytes / node_filesystem_size_bytes)) * 100', '>', 85.0, 300, 'high', true, 1, 1),
('容器CPU使用率告警', '容器CPU使用率超过80%', 'rate(container_cpu_usage_seconds_total{name!=""}[5m]) * 100', '>', 80.0, 300, 'medium', true, 1, 1),
('Ping不可达告警', '主机Ping不可达', 'icmplatency', '>', 9999, 60, 'critical', true, 1, 1)
ON CONFLICT (name) DO NOTHING;

-- ============================================
-- 完成提示
-- ============================================

DO $$
BEGIN
    RAISE NOTICE '============================================';
    RAISE NOTICE '告警平台数据库Schema初始化完成！';
    RAISE NOTICE '创建的表：';
    RAISE NOTICE '  - alert_rules (告警规则表)';
    RAISE NOTICE '  - alert_instances (告警实例表)';
    RAISE NOTICE '  - alert_history (告警历史表)';
    RAISE NOTICE '  - alert_silences (告警静默表)';
    RAISE NOTICE '  - alert_aggregations (告警聚合表)';
    RAISE NOTICE '============================================';
END $$;

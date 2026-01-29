-- 告警路由规则和模板系统
-- 执行命令: psql -h localhost -U postgres -d itcmdb -f services/alert-service/migrations/003_alert_routing_and_templates.sql

-- ============================================
-- 告警路由规则表 (类似Alertmanager的route配置)
-- ============================================
CREATE TABLE IF NOT EXISTS alert_routing_rules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- 匹配条件 (JSON格式存储label匹配规则)
    -- 例如: {"severity": "critical", "cluster": "prod"}
    matchers JSONB NOT NULL DEFAULT '{}',

    -- 匹配模式: match (完全匹配) 或 match_re (正则匹配)
    match_type VARCHAR(20) NOT NULL DEFAULT 'match' CHECK (match_type IN ('match', 'match_re')),

    -- 接收人组ID (匹配成功后发送给的接收人组)
    receiver_group_id INTEGER REFERENCES alert_receiver_groups(id) ON DELETE SET NULL,

    -- 是否继续匹配子路由
    continue BOOLEAN DEFAULT false,

    -- 优先级 (数字越小优先级越高)
    priority INTEGER NOT NULL DEFAULT 0,

    -- 是否启用
    enabled BOOLEAN DEFAULT true,

    -- 审计字段
    created_by INTEGER,
    updated_by INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_alert_routing_rules_enabled ON alert_routing_rules(enabled);
CREATE INDEX IF NOT EXISTS idx_alert_routing_rules_priority ON alert_routing_rules(priority);
CREATE INDEX IF NOT EXISTS idx_alert_routing_rules_receiver_group_id ON alert_routing_rules(receiver_group_id);

-- ============================================
-- 告警通知模板表
-- ============================================
CREATE TABLE IF NOT EXISTS alert_notification_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,

    -- 模板类型: dingtalk, feishu, wechat, email
    template_type VARCHAR(50) NOT NULL CHECK (template_type IN ('dingtalk', 'feishu', 'wechat', 'email')),

    -- 模板内容 (使用Go template语法)
    -- 可用变量: .AlertID, .Title, .Content, .Severity, .Status, .Instance, .Labels, .Annotations, .Timestamp
    template_content TEXT NOT NULL,

    -- 是否为默认模板
    is_default BOOLEAN DEFAULT false,

    -- 审计字段
    created_by INTEGER,
    updated_by INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_alert_notification_templates_type ON alert_notification_templates(template_type);
CREATE INDEX IF NOT EXISTS idx_alert_notification_templates_default ON alert_notification_templates(is_default);

-- ============================================
-- 修改alert_instances表，添加路由规则关联
-- ============================================
ALTER TABLE alert_instances
ADD COLUMN IF NOT EXISTS routing_rule_id INTEGER REFERENCES alert_routing_rules(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_alert_instances_routing_rule_id ON alert_instances(routing_rule_id);

-- ============================================
-- 修改inbound_webhooks表，添加默认接收人组
-- ============================================
ALTER TABLE inbound_webhooks
ADD COLUMN IF NOT EXISTS default_receiver_group_id INTEGER REFERENCES alert_receiver_groups(id) ON DELETE SET NULL;

-- ============================================
-- 创建更新时间触发器函数
-- ============================================
CREATE OR REPLACE FUNCTION update_alert_routing_rules_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_alert_notification_templates_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- 创建触发器
-- ============================================
DROP TRIGGER IF EXISTS trigger_update_alert_routing_rules_updated_at ON alert_routing_rules;
CREATE TRIGGER trigger_update_alert_routing_rules_updated_at
    BEFORE UPDATE ON alert_routing_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_alert_routing_rules_updated_at();

DROP TRIGGER IF EXISTS trigger_update_alert_notification_templates_updated_at ON alert_notification_templates;
CREATE TRIGGER trigger_update_alert_notification_templates_updated_at
    BEFORE UPDATE ON alert_notification_templates
    FOR EACH ROW
    EXECUTE FUNCTION update_alert_notification_templates_updated_at();

-- ============================================
-- 插入默认模板
-- ============================================

-- 钉钉默认模板 (Markdown格式)
INSERT INTO alert_notification_templates (name, description, template_type, template_content, is_default)
VALUES (
    '钉钉默认模板',
    '钉钉Markdown格式告警模板',
    'dingtalk',
    '#### {{ .Title }}

> **告警ID**: {{ .AlertID }}
> **严重级别**: {{ .Severity }}
> **状态**: {{ .Status }}
> **空间/对象**: {{ .Instance }}
> **时间**: {{ .Timestamp }}

{{ if .Content }}**详情**: {{ .Content }}{{ end }}

{{ if .Labels }}
**标签**:
{{ range $key, $value := .Labels }}
- {{ $key }}: {{ $value }}
{{ end }}
{{ end }}',
    true
) ON CONFLICT (name) DO NOTHING;

-- 飞书默认模板 (富文本格式)
INSERT INTO alert_notification_templates (name, description, template_type, template_content, is_default)
VALUES (
    '飞书默认模板',
    '飞书富文本格式告警模板',
    'feishu',
    '{{ .Title }}

告警ID: {{ .AlertID }}
严重级别: {{ .Severity }}
状态: {{ .Status }}
空间/对象: {{ .Instance }}
时间: {{ .Timestamp }}

{{ if .Content }}详情: {{ .Content }}{{ end }}',
    true
) ON CONFLICT (name) DO NOTHING;

-- 企业微信默认模板 (Markdown格式)
INSERT INTO alert_notification_templates (name, description, template_type, template_content, is_default)
VALUES (
    '企业微信默认模板',
    '企业微信Markdown格式告警模板',
    'wechat',
    '### {{ .Title }}

**告警ID**: {{ .AlertID }}
**严重级别**: {{ .Severity }}
**状态**: {{ .Status }}
**空间/对象**: {{ .Instance }}
**时间**: {{ .Timestamp }}

{{ if .Content }}**详情**: {{ .Content }}{{ end }}

{{ if .Labels }}
**标签**:
{{ range $key, $value := .Labels }}
- {{ $key }}: {{ $value }}
{{ end }}
{{ end }}',
    true
) ON CONFLICT (name) DO NOTHING;

-- ============================================
-- 插入示例路由规则
-- ============================================
-- 注意：这些规则需要在配置接收人组后才能生效
-- 这里只提供示例，实际使用时需要根据alert_receiver_groups表的数据调整

-- 严重程度为critical的告路由发送给紧急响应组
-- INSERT INTO alert_routing_rules (name, description, matchers, match_type, priority)
-- VALUES (
--     '严重告警路由',
--     '将严重级别为critical的告警路由到紧急响应组',
--     '{"severity": "critical"}'::jsonb,
--     'match',
--     1
-- );

-- 生产环境告路由发送给运维组
-- INSERT INTO alert_routing_rules (name, description, matchers, match_type, priority)
-- VALUES (
--     '生产环境告警路由',
--     '将生产环境的告警路由到运维组',
--     '{"env": "production"}'::jsonb,
--     'match',
--     2
-- );

-- ============================================
-- 添加注释
-- ============================================
COMMENT ON TABLE alert_routing_rules IS '告警路由规则表';
COMMENT ON TABLE alert_notification_templates IS '告警通知模板表';

COMMENT ON COLUMN alert_routing_rules.matchers IS '标签匹配条件 (JSON格式)';
COMMENT ON COLUMN alert_routing_rules.match_type IS '匹配类型: match(完全匹配) 或 match_re(正则匹配)';
COMMENT ON COLUMN alert_routing_rules.receiver_group_id IS '匹配成功后发送给的接收人组';
COMMENT ON COLUMN alert_routing_rules.continue IS '是否继续匹配子路由';
COMMENT ON COLUMN alert_routing_rules.priority IS '优先级 (数字越小优先级越高)';

COMMENT ON COLUMN alert_notification_templates.template_type IS '模板类型: dingtalk, feishu, wechat, email';
COMMENT ON COLUMN alert_notification_templates.template_content IS '模板内容 (Go template语法)';
COMMENT ON COLUMN alert_notification_templates.is_default IS '是否为默认模板';

COMMENT ON COLUMN alert_instances.routing_rule_id IS '匹配的路由规则ID';
COMMENT ON COLUMN inbound_webhooks.default_receiver_group_id IS '默认接收人组ID (未匹配到路由规则时使用)';

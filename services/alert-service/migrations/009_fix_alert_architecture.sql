-- ============================================
-- 告警中心架构修复
-- ============================================
-- 修复日期: 2026-01-30
-- 目的: 修复告警路由和通知模板的关联关系

-- 1. 为路由规则添加模板ID字段
ALTER TABLE alert_routing_rules
ADD COLUMN IF NOT EXISTS template_id INTEGER REFERENCES alert_notification_templates(id) ON DELETE SET NULL;

COMMENT ON COLUMN alert_routing_rules.template_id IS '路由规则指定的通知模板ID（优先级最高）';

CREATE INDEX IF NOT EXISTS idx_alert_routing_rules_template_id ON alert_routing_rules(template_id);

-- 2. 为接收人添加默认模板ID字段
ALTER TABLE alert_receivers
ADD COLUMN IF NOT EXISTS default_template_id INTEGER REFERENCES alert_notification_templates(id) ON DELETE SET NULL;

COMMENT ON COLUMN alert_receivers.default_template_id IS '接收人的默认通知模板ID（如果路由规则未指定模板则使用此模板）';

CREATE INDEX IF NOT EXISTS idx_alert_receivers_default_template_id ON alert_receivers(default_template_id);

-- 3. 删除 outbound_webhooks 相关表（功能与 alert_receivers 重复）
DROP TABLE IF EXISTS outbound_webhook_logs CASCADE;
DROP TABLE IF EXISTS outbound_webhooks CASCADE;

-- 4. 删除 webhook_metrics 表（简化架构）
DROP TABLE IF EXISTS webhook_metrics CASCADE;

-- 5. 删除 dead_letter_queues 表（使用 notification_logs 的重试机制替代）
DROP TABLE IF EXISTS dead_letter_queues CASCADE;

-- 6. 为 notification_logs 添加缺失的索引
CREATE INDEX IF NOT EXISTS idx_notification_logs_routing_rule_id ON notification_logs(routing_rule_id);

-- 7. 更新注释说明新的架构
COMMENT ON TABLE alert_routing_rules IS '告警路由规则表：定义告警如何路由到接收组，以及使用哪个通知模板';
COMMENT ON TABLE alert_receivers IS '告警接收人表：定义通知发送的目标（钉钉/飞书/企业微信等），每个接收人可指定默认模板';
COMMENT ON TABLE alert_notification_templates IS '通知模板表：定义不同类型接收人的消息格式模板';
COMMENT ON TABLE notification_logs IS '统一通知日志表：记录所有通知的发送状态、重试次数和结果';

-- 8. 添加模板选择优先级说明
COMMENT ON DATABASE itcmdb IS '
告警通知模板选择优先级：
1. 路由规则指定的模板 (alert_routing_rules.template_id)
2. 接收人的默认模板 (alert_receivers.default_template_id)
3. 系统默认模板 (alert_notification_templates.is_default = true)
';

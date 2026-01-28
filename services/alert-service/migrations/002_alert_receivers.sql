-- 告警接收组表
CREATE TABLE IF NOT EXISTS alert_receiver_groups (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 告警接收人表
CREATE TABLE IF NOT EXISTS alert_receivers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'wechat', 'dingtalk', 'feishu', 'email', 'sms'
    webhook_url TEXT, -- webhook地址
    at_mobiles TEXT[], -- @手机号列表
    at_user_ids TEXT[], -- @用户ID列表
    secret VARCHAR(255), -- 签名密钥（钉钉、企业微信）
    config JSONB, -- 其他配置（扩展字段）
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 接收组-接收人关联表
CREATE TABLE IF NOT EXISTS alert_receiver_group_members (
    id SERIAL PRIMARY KEY,
    group_id INTEGER NOT NULL REFERENCES alert_receiver_groups(id) ON DELETE CASCADE,
    receiver_id INTEGER NOT NULL REFERENCES alert_receivers(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(group_id, receiver_id)
);

-- 告警规则-接收组关联表
CREATE TABLE IF NOT EXISTS alert_rule_receiver_groups (
    id SERIAL PRIMARY KEY,
    rule_id INTEGER NOT NULL REFERENCES alert_rules(id) ON DELETE CASCADE,
    group_id INTEGER NOT NULL REFERENCES alert_receiver_groups(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(rule_id, group_id)
);

-- 触发器：自动更新updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_alert_receiver_groups_updated_at
    BEFORE UPDATE ON alert_receiver_groups
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_alert_receivers_updated_at
    BEFORE UPDATE ON alert_receivers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 插入示例数据
INSERT INTO alert_receiver_groups (name, description, enabled) VALUES
('运维组', '负责处理系统告警', true),
('开发组', '负责处理应用告警', true),
('测试组', '负责处理测试环境告警', true)
ON CONFLICT (name) DO NOTHING;

-- 插入示例接收人（webhook_url需要用户配置）
INSERT INTO alert_receivers (name, type, webhook_url, enabled) VALUES
('钉钉-运维组', 'dingtalk', '', true),
('企业微信-运维组', 'wechat', '', true),
('飞书-运维组', 'feishu', '', true)
ON CONFLICT DO NOTHING;

-- 关联组和接收人
INSERT INTO alert_receiver_group_members (group_id, receiver_id)
SELECT g.id, r.id
FROM alert_receiver_groups g
CROSS JOIN alert_receivers r
WHERE g.name = '运维组'
ON CONFLICT DO NOTHING;

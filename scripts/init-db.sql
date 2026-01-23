-- ITCMDB 数据库初始化脚本
-- 创建数据库结构

-- ============================================
-- 用户与权限模块
-- ============================================

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(100),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 角色表
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 权限表
CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT,
    UNIQUE(resource, action)
);

-- 角色权限关联表
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
    permission_id INTEGER REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- 用户角色关联表
CREATE TABLE IF NOT EXISTS user_roles (
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    role_id INTEGER REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- ============================================
-- CMDB模块
-- ============================================

-- CI类型表
CREATE TABLE IF NOT EXISTS ci_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    icon VARCHAR(50),
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- CI属性定义表
CREATE TABLE IF NOT EXISTS ci_attributes (
    id SERIAL PRIMARY KEY,
    ci_type_id INTEGER REFERENCES ci_types(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    type VARCHAR(20) NOT NULL, -- string, number, boolean, date, select
    options JSONB, -- 用于select类型的选项
    is_required BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- CI实例表
CREATE TABLE IF NOT EXISTS ci_instances (
    id SERIAL PRIMARY KEY,
    ci_type_id INTEGER REFERENCES ci_types(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    attributes JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- CI关系表
CREATE TABLE IF NOT EXISTS ci_relations (
    id SERIAL PRIMARY KEY,
    parent_id INTEGER REFERENCES ci_instances(id) ON DELETE CASCADE,
    child_id INTEGER REFERENCES ci_instances(id) ON DELETE CASCADE,
    relation_type VARCHAR(50) NOT NULL, -- depends, contains, connects
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(parent_id, child_id, relation_type)
);

-- CI变更历史表
CREATE TABLE IF NOT EXISTS ci_history (
    id SERIAL PRIMARY KEY,
    ci_id INTEGER REFERENCES ci_instances(id) ON DELETE CASCADE,
    changed_by INTEGER REFERENCES users(id),
    field_name VARCHAR(50),
    old_value JSONB,
    new_value JSONB,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 工单模块
-- ============================================

-- 工作流定义表
CREATE TABLE IF NOT EXISTS ticket_workflows (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    states JSONB NOT NULL, -- ["open", "in_progress", "resolved", "closed"]
    transitions JSONB NOT NULL, -- 状态转换规则
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 工单模板表
CREATE TABLE IF NOT EXISTS ticket_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    workflow_id INTEGER REFERENCES ticket_workflows(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 工单主表
CREATE TABLE IF NOT EXISTS tickets (
    id SERIAL PRIMARY KEY,
    ticket_number VARCHAR(20) UNIQUE NOT NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    template_id INTEGER REFERENCES ticket_templates(id),
    workflow_id INTEGER REFERENCES ticket_workflows(id),
    status VARCHAR(20) DEFAULT 'open',
    priority VARCHAR(20) DEFAULT 'medium', -- low, medium, high, critical
    assignee_id INTEGER REFERENCES users(id),
    requester_id INTEGER REFERENCES users(id),
    sla_deadline TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 工单评论表
CREATE TABLE IF NOT EXISTS ticket_comments (
    id SERIAL PRIMARY KEY,
    ticket_id INTEGER REFERENCES tickets(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id),
    content TEXT NOT NULL,
    is_internal BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 工单附件表
CREATE TABLE IF NOT EXISTS ticket_attachments (
    id SERIAL PRIMARY KEY,
    ticket_id INTEGER REFERENCES tickets(id) ON DELETE CASCADE,
    file_name VARCHAR(255) NOT NULL,
    file_url VARCHAR(500) NOT NULL,
    file_size BIGINT,
    uploaded_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 工单历史表
CREATE TABLE IF NOT EXISTS ticket_history (
    id SERIAL PRIMARY KEY,
    ticket_id INTEGER REFERENCES tickets(id) ON DELETE CASCADE,
    field_name VARCHAR(50),
    old_value TEXT,
    new_value TEXT,
    changed_by INTEGER REFERENCES users(id),
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 告警模块
-- ============================================

-- 告警规则表
CREATE TABLE IF NOT EXISTS alert_rules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    condition JSONB NOT NULL, -- 告警条件
    severity VARCHAR(20) NOT NULL, -- critical, high, medium, low
    notification_channels JSONB, -- ["email", "wechat", "sms"]
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 告警阈值表
CREATE TABLE IF NOT EXISTS alert_thresholds (
    id SERIAL PRIMARY KEY,
    rule_id INTEGER REFERENCES alert_rules(id) ON DELETE CASCADE,
    metric VARCHAR(100) NOT NULL,
    operator VARCHAR(20) NOT NULL, -- >, <, =, >=, <=
    threshold DECIMAL(10,2) NOT NULL,
    duration INTEGER, -- 持续时间(秒)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 告警实例表
CREATE TABLE IF NOT EXISTS alert_instances (
    id SERIAL PRIMARY KEY,
    rule_id INTEGER REFERENCES alert_rules(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    severity VARCHAR(20) NOT NULL,
    status VARCHAR(20) DEFAULT 'active', -- active, acknowledged, closed
    affected_ci_id INTEGER REFERENCES ci_instances(id),
    triggered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    acknowledged_at TIMESTAMP,
    acknowledged_by INTEGER REFERENCES users(id),
    closed_at TIMESTAMP
);

-- 告警历史表
CREATE TABLE IF NOT EXISTS alert_history (
    id SERIAL PRIMARY KEY,
    alert_id INTEGER REFERENCES alert_instances(id) ON DELETE CASCADE,
    event_type VARCHAR(50) NOT NULL, -- triggered, acknowledged, closed, escalated
    event_data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 通知模块
-- ============================================

-- 通知模板表
CREATE TABLE IF NOT EXISTS notification_templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL, -- email, wechat, sms
    subject VARCHAR(200),
    content TEXT NOT NULL,
    variables JSONB, -- 可用变量列表
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 通知历史表
CREATE TABLE IF NOT EXISTS notification_history (
    id SERIAL PRIMARY KEY,
    template_id INTEGER REFERENCES notification_templates(id),
    recipient VARCHAR(200) NOT NULL,
    channel VARCHAR(20) NOT NULL, -- email, wechat, sms
    status VARCHAR(20) DEFAULT 'pending', -- pending, sent, failed
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    sent_at TIMESTAMP
);

-- ============================================
-- 审计日志模块
-- ============================================

-- 审计日志表
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    resource_id INTEGER,
    details JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 报表模块
-- ============================================

-- 报表配置表
CREATE TABLE IF NOT EXISTS report_configs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    type VARCHAR(50) NOT NULL, -- cmdb_assets, ticket_stats, alert_trends
    config JSONB NOT NULL,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- 索引创建
-- ============================================

-- 用户表索引
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);

-- CI实例表索引
CREATE INDEX idx_ci_instances_type ON ci_instances(ci_type_id);
CREATE INDEX idx_ci_instances_status ON ci_instances(status);
CREATE INDEX idx_ci_instances_name ON ci_instances(name);

-- CI关系表索引
CREATE INDEX idx_ci_relations_parent ON ci_relations(parent_id);
CREATE INDEX idx_ci_relations_child ON ci_relations(child_id);

-- 工单表索引
CREATE INDEX idx_tickets_status ON tickets(status);
CREATE INDEX idx_tickets_priority ON tickets(priority);
CREATE INDEX idx_tickets_assignee ON tickets(assignee_id);
CREATE INDEX idx_tickets_requester ON tickets(requester_id);
CREATE INDEX idx_tickets_created_at ON tickets(created_at);

-- 告警表索引
CREATE INDEX idx_alerts_status ON alert_instances(status);
CREATE INDEX idx_alerts_severity ON alert_instances(severity);
CREATE INDEX idx_alerts_triggered_at ON alert_instances(triggered_at);

-- 审计日志表索引
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);

-- ============================================
-- 初始化数据
-- ============================================

-- 插入默认管理员用户 (密码: admin123)
INSERT INTO users (username, email, password_hash, full_name) VALUES
('admin', 'admin@itcmdb.com', '$2a$10$YourHashedPasswordHere', '系统管理员')
ON CONFLICT (username) DO NOTHING;

-- 插入默认角色
INSERT INTO roles (name, description) VALUES
('admin', '系统管理员，拥有所有权限'),
('operator', '运维工程师，可操作CMDB和工单'),
('developer', '开发工程师，只读权限和创建工单'),
('viewer', '只读用户，仅查看权限')
ON CONFLICT (name) DO NOTHING;

-- 插入管理员角色关联
INSERT INTO user_roles (user_id, role_id)
SELECT u.id, r.id FROM users u, roles r
WHERE u.username = 'admin' AND r.name = 'admin'
ON CONFLICT DO NOTHING;

-- 插入默认权限
INSERT INTO permissions (resource, action, description) VALUES
-- CMDB权限
('cmdb:server', 'create', '创建服务器'),
('cmdb:server', 'read', '查看服务器'),
('cmdb:server', 'update', '更新服务器'),
('cmdb:server', 'delete', '删除服务器'),
('cmdb:network', 'create', '创建网络设备'),
('cmdb:network', 'read', '查看网络设备'),
('cmdb:network', 'update', '更新网络设备'),
('cmdb:network', 'delete', '删除网络设备'),
('cmdb:application', 'create', '创建应用服务'),
('cmdb:application', 'read', '查看应用服务'),
('cmdb:application', 'update', '更新应用服务'),
('cmdb:application', 'delete', '删除应用服务'),
-- 工单权限
('ticket:incident', 'create', '创建故障工单'),
('ticket:incident', 'read', '查看故障工单'),
('ticket:incident', 'update', '更新故障工单'),
('ticket:incident', 'delete', '删除故障工单'),
('ticket:incident', 'approve', '审批故障工单'),
-- 告警权限
('alert:rule', 'create', '创建告警规则'),
('alert:rule', 'read', '查看告警规则'),
('alert:rule', 'update', '更新告警规则'),
('alert:rule', 'delete', '删除告警规则'),
('alert:instance', 'ack', '确认告警'),
('alert:instance', 'close', '关闭告警'),
-- 用户管理权限
('user', 'create', '创建用户'),
('user', 'read', '查看用户'),
('user', 'update', '更新用户'),
('user', 'delete', '删除用户'),
('role', 'create', '创建角色'),
('role', 'read', '查看角色'),
('role', 'update', '更新角色'),
('role', 'delete', '删除角色'),
-- 审计权限
('audit', 'read', '查看审计日志')
ON CONFLICT (resource, action) DO NOTHING;

-- 插入管理员所有权限
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin'
ON CONFLICT DO NOTHING;

-- 插入默认CI类型
INSERT INTO ci_types (name, icon, description) VALUES
('server', 'ServerOutlined', '服务器/主机'),
('network', 'CloudServerOutlined', '网络设备'),
('application', 'AppstoreOutlined', '应用服务'),
('container', 'ContainerOutlined', '容器/K8s集群')
ON CONFLICT (name) DO NOTHING;

-- 插入默认工作流
INSERT INTO ticket_workflows (name, states, transitions) VALUES
('默认工单流程', '["open", "in_progress", "resolved", "closed"]', '{"open": ["in_progress", "closed"], "in_progress": ["resolved", "closed"], "resolved": ["closed"]}'::jsonb),
('故障工单流程', '["open", "assigned", "in_progress", "resolved", "closed"]', '{"open": ["assigned", "closed"], "assigned": ["in_progress", "closed"], "in_progress": ["resolved", "closed"], "resolved": ["closed"]}'::jsonb)
ON CONFLICT (name) DO NOTHING;

-- 插入默认工单模板
INSERT INTO ticket_templates (name, description, workflow_id) VALUES
('故障工单', '系统故障类工单', (SELECT id FROM ticket_workflows WHERE name = '故障工单流程' LIMIT 1)),
('服务请求', '用户服务请求类工单', (SELECT id FROM ticket_workflows WHERE name = '默认工单流程' LIMIT 1)),
('变更申请', '系统变更申请类工单', (SELECT id FROM ticket_workflows WHERE name = '默认工单流程' LIMIT 1))
ON CONFLICT DO NOTHING;

-- ============================================
-- 函数和触发器
-- ============================================

-- 更新时间戳函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 为需要的表添加更新时间戳触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ci_types_updated_at BEFORE UPDATE ON ci_types
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ci_instances_updated_at BEFORE UPDATE ON ci_instances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tickets_updated_at BEFORE UPDATE ON tickets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_alert_rules_updated_at BEFORE UPDATE ON alert_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_notification_templates_updated_at BEFORE UPDATE ON notification_templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_report_configs_updated_at BEFORE UPDATE ON report_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 工单编号生成函数
CREATE OR REPLACE FUNCTION generate_ticket_number()
RETURNS TRIGGER AS $$
DECLARE
    ticket_num VARCHAR(20);
    date_part VARCHAR(8);
    seq_num INTEGER;
BEGIN
    date_part := TO_CHAR(CURRENT_DATE, 'YYYYMMDD');

    SELECT COALESCE(MAX(CAST(SUBSTRING(ticket_number FROM 10) AS INTEGER)), 0) + 1
    INTO seq_num
    FROM tickets
    WHERE ticket_number LIKE 'TKT-' || date_part || '-%';

    ticket_num := 'TKT-' || date_part || '-' || LPAD(seq_num::TEXT, 4, '0');
    NEW.ticket_number := ticket_num;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER generate_ticket_number_trigger
    BEFORE INSERT ON tickets
    FOR EACH ROW
    WHEN (NEW.ticket_number IS NULL)
    EXECUTE FUNCTION generate_ticket_number();

-- ============================================
-- 完成
-- ============================================

-- 创建完成后输出信息
DO $$
BEGIN
    RAISE NOTICE 'ITCMDB 数据库初始化完成！';
    RAISE NOTICE '默认管理员账号: admin';
    RAISE NOTICE '请在生产环境中修改默认密码！';
END $$;

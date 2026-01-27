-- ITCMDB 数据库初始化脚本
-- 创建完整的数据库结构
-- 版本: v2.0
-- 更新日期: 2026-01-26
--
-- 此脚本整合了所有数据库迁移，可直接在新设备上部署
-- 执行方式: docker exec -i itcmdb-postgres psql -U itcmdb -d itcmdb < scripts/init-db.sql

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


-- 用户直接权限关联表
CREATE TABLE IF NOT EXISTS user_permissions (
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    permission_id INTEGER REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, permission_id)
);

-- ============================================
-- CMDB模块
-- ============================================

-- CI类型表
CREATE TABLE IF NOT EXISTS ci_types (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
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
    display_name VARCHAR(100) NOT NULL,
    type VARCHAR(20) NOT NULL, -- string, number, boolean, date, select
    options JSONB, -- 用于select类型的选项
    is_required BOOLEAN DEFAULT false,
    is_unique BOOLEAN DEFAULT false,
    default_value VARCHAR(255),
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- CI实例表
CREATE TABLE IF NOT EXISTS ci_instances (
    id SERIAL PRIMARY KEY,
    ci_type_id INTEGER REFERENCES ci_types(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    attributes JSONB,
    tags JSONB,
    created_by INTEGER REFERENCES users(id),
    updated_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- CI关系表
CREATE TABLE IF NOT EXISTS ci_relations (
    id SERIAL PRIMARY KEY,
    parent_id INTEGER REFERENCES ci_instances(id) ON DELETE CASCADE,
    child_id INTEGER REFERENCES ci_instances(id) ON DELETE CASCADE,
    relation_type VARCHAR(50) NOT NULL, -- depends_on, runs_on, connects_to, contains
    description VARCHAR(255),
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(parent_id, child_id, relation_type)
);

-- CI变更历史表
CREATE TABLE IF NOT EXISTS ci_history (
    id SERIAL PRIMARY KEY,
    ci_id INTEGER REFERENCES ci_instances(id) ON DELETE CASCADE,
    changed_by INTEGER REFERENCES users(id),
    action VARCHAR(20) NOT NULL, -- create, update, delete
    field_name VARCHAR(50),
    old_value JSONB,
    new_value JSONB,
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- CI角色表 (技术角色)
CREATE TABLE IF NOT EXISTS ci_roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(200) NOT NULL,
    description TEXT,
    color VARCHAR(20),
    icon VARCHAR(50),
    priority INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 负责人角色表
CREATE TABLE IF NOT EXISTS owner_roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(200) NOT NULL,
    description TEXT,
    level INTEGER DEFAULT 0,
    responsibilities JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- CI实例角色关联表
CREATE TABLE IF NOT EXISTS ci_instance_roles (
    id SERIAL PRIMARY KEY,
    ci_id INTEGER NOT NULL REFERENCES ci_instances(id) ON DELETE CASCADE,
    role_type VARCHAR(20) NOT NULL, -- 'ci_role' 或 'owner_role'
    role_id INTEGER NOT NULL,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    assigned_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT check_role_type CHECK (role_type IN ('ci_role', 'owner_role')),
    UNIQUE(ci_id, role_type, role_id, COALESCE(user_id, 0))
);

-- 标签分类表
CREATE TABLE IF NOT EXISTS tag_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    display_name VARCHAR(200) NOT NULL,
    description TEXT,
    color VARCHAR(20),
    icon VARCHAR(50),
    sort_order INTEGER DEFAULT 0,
    is_system BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 标签定义表
CREATE TABLE IF NOT EXISTS tags (
    id SERIAL PRIMARY KEY,
    category_id INTEGER REFERENCES tag_categories(id) ON DELETE SET NULL,
    name VARCHAR(100) NOT NULL,
    display_name VARCHAR(200) NOT NULL,
    color VARCHAR(20),
    description TEXT,
    usage_count INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(category_id, name)
);

-- CI实例标签关联表
CREATE TABLE IF NOT EXISTS ci_tags (
    id SERIAL PRIMARY KEY,
    ci_id INTEGER NOT NULL REFERENCES ci_instances(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    tagged_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    tagged_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ci_id, tag_id)
);

-- 标签使用历史表
CREATE TABLE IF NOT EXISTS tag_history (
    id SERIAL PRIMARY KEY,
    ci_id INTEGER REFERENCES ci_instances(id) ON DELETE CASCADE,
    tag_id INTEGER REFERENCES tags(id) ON DELETE CASCADE,
    action VARCHAR(20) NOT NULL,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT check_tag_action CHECK (action IN ('added', 'removed'))
);

-- 系统配置表
CREATE TABLE IF NOT EXISTS system_configs (
    id SERIAL PRIMARY KEY,
    category VARCHAR(50),
    key VARCHAR(100) NOT NULL,
    value TEXT,
    description TEXT,
    is_encrypted BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_by INTEGER REFERENCES users(id),
    updated_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(category, key)
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
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);


-- 用户权限表索引
CREATE INDEX IF NOT EXISTS idx_user_permissions_user_id ON user_permissions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_permissions_permission_id ON user_permissions(permission_id);

-- CI实例表索引
CREATE INDEX IF NOT EXISTS idx_ci_instances_type ON ci_instances(ci_type_id);
CREATE INDEX IF NOT EXISTS idx_ci_instances_status ON ci_instances(status);
CREATE INDEX IF NOT EXISTS idx_ci_instances_name ON ci_instances(name);
CREATE INDEX IF NOT EXISTS idx_ci_instances_created_by ON ci_instances(created_by);
CREATE INDEX IF NOT EXISTS idx_ci_instances_updated_by ON ci_instances(updated_by);
CREATE INDEX IF NOT EXISTS idx_ci_instances_deleted_at ON ci_instances(deleted_at);

-- CI属性表索引
CREATE INDEX IF NOT EXISTS idx_ci_attributes_deleted_at ON ci_attributes(deleted_at);

-- CI关系表索引
CREATE INDEX IF NOT EXISTS idx_ci_relations_parent ON ci_relations(parent_id);
CREATE INDEX IF NOT EXISTS idx_ci_relations_child ON ci_relations(child_id);
CREATE INDEX IF NOT EXISTS idx_ci_relations_created_by ON ci_relations(created_by);
CREATE INDEX IF NOT EXISTS idx_ci_relations_deleted_at ON ci_relations(deleted_at);

-- CI角色关联表索引
CREATE INDEX IF NOT EXISTS idx_ci_instance_roles_ci_id ON ci_instance_roles(ci_id);
CREATE INDEX IF NOT EXISTS idx_ci_instance_roles_type_id ON ci_instance_roles(role_type, role_id);

-- 标签表索引
CREATE INDEX IF NOT EXISTS idx_ci_tags_ci_id ON ci_tags(ci_id);
CREATE INDEX IF NOT EXISTS idx_ci_tags_tag_id ON ci_tags(tag_id);
CREATE INDEX IF NOT EXISTS idx_tag_history_ci_id ON tag_history(ci_id);
CREATE INDEX IF NOT EXISTS idx_tag_history_tag_id ON tag_history(tag_id);

-- 工单表索引
CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status);
CREATE INDEX IF NOT EXISTS idx_tickets_priority ON tickets(priority);
CREATE INDEX IF NOT EXISTS idx_tickets_assignee ON tickets(assignee_id);
CREATE INDEX IF NOT EXISTS idx_tickets_requester ON tickets(requester_id);
CREATE INDEX IF NOT EXISTS idx_tickets_created_at ON tickets(created_at);

-- 告警表索引
CREATE INDEX IF NOT EXISTS idx_alerts_status ON alert_instances(status);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alert_instances(severity);
CREATE INDEX IF NOT EXISTS idx_alerts_triggered_at ON alert_instances(triggered_at);

-- 审计日志表索引
CREATE INDEX IF NOT EXISTS idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

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

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ci_types_updated_at BEFORE UPDATE ON ci_types
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ci_attributes_updated_at BEFORE UPDATE ON ci_attributes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ci_instances_updated_at BEFORE UPDATE ON ci_instances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ci_relations_updated_at BEFORE UPDATE ON ci_relations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ci_roles_updated_at BEFORE UPDATE ON ci_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_owner_roles_updated_at BEFORE UPDATE ON owner_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tag_categories_updated_at BEFORE UPDATE ON tag_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tags_updated_at BEFORE UPDATE ON tags
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tickets_updated_at BEFORE UPDATE ON tickets
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
-- 初始化数据
-- ============================================

-- 插入默认管理员用户 (密码: admin123)
-- 这是bcrypt哈希后的密码
INSERT INTO users (username, email, password_hash, full_name) VALUES
('admin', 'admin@itcmdb.com', '$2b$10$ZvxirUvXL48w01pBuGX3vetlzJ3imte3x/FO83ub23lLCiJJFNlIy', '系统管理员')
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

-- 插入系统权限（使用新的简化格式）
INSERT INTO permissions (resource, action, description) VALUES
-- 用户管理权限
('user', 'create', '创建用户'),
('user', 'update', '更新用户'),
('user', 'delete', '删除用户'),
('user', 'view', '查看用户'),
('user', 'manage', '管理用户'),
-- 角色管理权限
('role', 'create', '创建角色'),
('role', 'update', '更新角色'),
('role', 'delete', '删除角色'),
('role', 'view', '查看角色'),
('role', 'manage', '管理角色'),
-- 权限管理权限
('permission', 'create', '创建权限'),
('permission', 'delete', '删除权限'),
('permission', 'view', '查看权限'),
-- 配置管理权限
('config', 'create', '创建配置'),
('config', 'update', '更新配置'),
('config', 'delete', '删除配置'),
('config', 'view', '查看配置'),
-- CI管理权限
('ci', 'create', '创建配置项'),
('ci', 'update', '更新配置项'),
('ci', 'delete', '删除配置项'),
('ci', 'view', '查看配置项'),
-- 标签管理权限
('tag', 'create', '创建标签'),
('tag', 'update', '更新标签'),
('tag', 'delete', '删除标签'),
('tag', 'view', '查看标签'),
-- 工单管理权限
('ticket', 'create', '创建工单'),
('ticket', 'update', '更新工单'),
('ticket', 'delete', '删除工单'),
('ticket', 'view', '查看工单'),
-- 告警管理权限
('alert', 'create', '创建告警'),
('alert', 'update', '更新告警'),
('alert', 'delete', '删除告警'),
('alert', 'view', '查看告警'),
-- 审计日志权限
('audit', 'view', '查看审计日志'),
-- 超级管理员权限
('*', '*', '超级管理员全部权限')
ON CONFLICT (resource, action) DO NOTHING;

-- 为admin角色分配超级管理员权限
DO $$
DECLARE
    admin_role_id INT;
    super_admin_perm_id INT;
BEGIN
    SELECT id INTO admin_role_id FROM roles WHERE name = 'admin';
    SELECT id INTO super_admin_perm_id FROM permissions WHERE resource = '*' AND action = '*';

    IF admin_role_id IS NOT NULL AND super_admin_perm_id IS NOT NULL THEN
        INSERT INTO role_permissions (role_id, permission_id)
        VALUES (admin_role_id, super_admin_perm_id)
        ON CONFLICT (role_id, permission_id) DO NOTHING;
    END IF;
END $$;

-- 插入默认CI类型
INSERT INTO ci_types (name, display_name, icon, description) VALUES
('server', '服务器', 'ServerOutlined', '服务器/主机'),
('network', '网络设备', 'CloudServerOutlined', '网络设备'),
('application', '应用服务', 'AppstoreOutlined', '应用服务'),
('container', '容器', 'ContainerOutlined', '容器/K8s集群')
ON CONFLICT (name) DO NOTHING;

-- 插入预设标签分类
INSERT INTO tag_categories (name, display_name, description, color, is_system, sort_order) VALUES
('environment', '环境', '环境标签，标识CI所处的环境阶段', 'blue', true, 1),
('business', '业务', '业务标签，标识CI的业务重要性', 'green', true, 2),
('location', '位置', '位置标签，标识CI的物理位置', 'orange', true, 3),
('custom', '自定义', '用户自定义标签', 'gray', false, 4)
ON CONFLICT (name) DO NOTHING;

-- 插入预设标签
INSERT INTO tags (category_id, name, display_name, color, description) VALUES
-- 环境标签
((SELECT id FROM tag_categories WHERE name = 'environment'), 'prod', '生产环境', '#f50', '生产环境'),
((SELECT id FROM tag_categories WHERE name = 'environment'), 'staging', '预发环境', '#2db7f5', '预发布/准生产环境'),
((SELECT id FROM tag_categories WHERE name = 'environment'), 'test', '测试环境', '#87d068', '测试/集成环境'),
((SELECT id FROM tag_categories WHERE name = 'environment'), 'dev', '开发环境', '#108ee9', '开发/调试环境'),
((SELECT id FROM tag_categories WHERE name = 'environment'), 'dr', '灾备环境', '#722ed1', '灾难恢复环境'),
-- 业务标签
((SELECT id FROM tag_categories WHERE name = 'business'), 'critical', '核心业务', '#f50', '核心关键业务系统'),
((SELECT id FROM tag_categories WHERE name = 'business'), 'important', '重要业务', '#faad14', '重要业务系统'),
((SELECT id FROM tag_categories WHERE name = 'business'), 'normal', '普通业务', '#52c41a', '一般业务系统'),
((SELECT id FROM tag_categories WHERE name = 'business'), 'internal', '内部工具', '#8c8c8c', '内部支撑工具'),
-- 位置标签
((SELECT id FROM tag_categories WHERE name = 'location'), 'idc_1', 'IDC机房1', '#1890ff', '一号数据中心'),
((SELECT id FROM tag_categories WHERE name = 'location'), 'idc_2', 'IDC机房2', '#096dd9', '二号数据中心'),
((SELECT id FROM tag_categories WHERE name = 'location'), 'cloud_aliyun', '阿里云', '#ffec3d', '阿里云平台'),
((SELECT id FROM tag_categories WHERE name = 'location'), 'cloud_tencent', '腾讯云', '#ffbb96', '腾讯云平台')
ON CONFLICT (category_id, name) DO NOTHING;

-- 插入预设CI角色
INSERT INTO ci_roles (name, display_name, description, color, icon, priority) VALUES
('primary_db', '主数据库', '主要数据库节点', '#f50', 'database', 1),
('standby_db', '备数据库', '备用/只读数据库节点', '#87d068', 'database', 2),
('web_server', 'Web服务器', 'Web应用服务器', '#108ee9', 'server', 3),
('app_server', '应用服务器', '后端应用服务器', '#2db7f5', 'server', 4),
('lb_server', '负载均衡', '负载均衡器', '#722ed1', 'workflow', 5),
('cache_node', '缓存节点', 'Redis/Memcached缓存节点', '#faad14', 'cpu', 6),
('mq_node', '消息队列', '消息队列节点', '#eb2f96', 'workflow', 7),
('storage', '存储节点', '文件/对象存储节点', '#13c2c2', 'hard-drive', 8)
ON CONFLICT (name) DO NOTHING;

-- 插入预设负责人角色
INSERT INTO owner_roles (name, display_name, description, level, responsibilities) VALUES
('admin', '系统管理员', '负责系统的整体管理和维护', 0, '{"permissions": ["all"], "duties": ["系统配置", "用户管理", "安全策略"]}'::jsonb),
('ops_owner', '运维负责人', '负责运维相关的操作和维护', 1, '{"permissions": ["deploy", "monitor", "backup"], "duties": ["部署发布", "监控告警", "备份恢复"]}'::jsonb),
('biz_owner', '业务负责人', '负责业务需求和业务逻辑', 2, '{"permissions": ["requirement", "acceptance"], "duties": ["需求提报", "验收测试"]}'::jsonb),
('sec_owner', '安全负责人', '负责安全审计和合规性检查', 3, '{"permissions": ["audit", "security"], "duties": ["安全审计", "漏洞扫描"]}'::jsonb),
('dev_owner', '开发负责人', '负责开发任务和技术实现', 4, '{"permissions": ["develop", "review"], "duties": ["代码开发", "代码评审"]}'::jsonb)
ON CONFLICT (name) DO NOTHING;

-- 插入默认工作流
INSERT INTO ticket_workflows (name, states, transitions) VALUES
('默认工单流程', '["open", "in_progress", "resolved", "closed"]'::jsonb, '{"open": ["in_progress", "closed"], "in_progress": ["resolved", "closed"], "resolved": ["closed"]}'::jsonb),
('故障工单流程', '["open", "assigned", "in_progress", "resolved", "closed"]'::jsonb, '{"open": ["assigned", "closed"], "assigned": ["in_progress", "closed"], "in_progress": ["resolved", "closed"], "resolved": ["closed"]}'::jsonb)
ON CONFLICT (name) DO NOTHING;

-- 插入默认工单模板
INSERT INTO ticket_templates (name, description, workflow_id) VALUES
('故障工单', '系统故障类工单', (SELECT id FROM ticket_workflows WHERE name = '故障工单流程' LIMIT 1)),
('服务请求', '用户服务请求类工单', (SELECT id FROM ticket_workflows WHERE name = '默认工单流程' LIMIT 1)),
('变更申请', '系统变更申请类工单', (SELECT id FROM ticket_workflows WHERE name = '默认工单流程' LIMIT 1))
ON CONFLICT DO NOTHING;

-- ============================================
-- 添加表注释
-- ============================================

COMMENT ON TABLE ci_roles IS 'CI技术角色表，定义配置项在系统中的技术职能';
COMMENT ON TABLE owner_roles IS '负责人角色表，定义资产负责人类型和职责';
COMMENT ON TABLE ci_instance_roles IS 'CI实例角色关联表，关联CI与角色';
COMMENT ON TABLE tag_categories IS '标签分类表，用于组织标签';
COMMENT ON TABLE tags IS '标签定义表';
COMMENT ON TABLE ci_tags IS 'CI实例标签关联表';
COMMENT ON TABLE tag_history IS '标签使用历史表';
COMMENT ON TABLE system_configs IS '系统配置表';

-- ============================================
-- 完成
-- ============================================

-- 创建完成后输出信息
DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'ITCMDB 数据库初始化完成！';
    RAISE NOTICE '========================================';
    RAISE NOTICE '默认管理员账号: admin';
    RAISE NOTICE '默认管理员密码: admin123';
    RAISE NOTICE '';
    RAISE NOTICE '重要提示：请在生产环境中立即修改默认密码！';
    RAISE NOTICE '========================================';
END $$;

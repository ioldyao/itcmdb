-- =====================================================
-- 角色和标签模块数据库迁移脚本
-- 版本: v1.0
-- 日期: 2026-01-24
-- =====================================================

-- =====================================================
-- 1. 角色模块表结构
-- =====================================================

-- CI角色表 (技术角色)
CREATE TABLE IF NOT EXISTS ci_roles (
  id                  SERIAL PRIMARY KEY,
  name                VARCHAR(100) NOT NULL UNIQUE,
  display_name        VARCHAR(200) NOT NULL,
  description         TEXT,
  color               VARCHAR(20),
  icon                VARCHAR(50),
  priority            INTEGER DEFAULT 0,
  is_active           BOOLEAN DEFAULT TRUE,
  created_at          TIMESTAMP DEFAULT NOW(),
  updated_at          TIMESTAMP DEFAULT NOW()
);

-- 负责人角色表
CREATE TABLE IF NOT EXISTS owner_roles (
  id                  SERIAL PRIMARY KEY,
  name                VARCHAR(100) NOT NULL UNIQUE,
  display_name        VARCHAR(200) NOT NULL,
  description         TEXT,
  level               INTEGER DEFAULT 0,
  responsibilities    JSONB DEFAULT '{}',
  is_active           BOOLEAN DEFAULT TRUE,
  created_at          TIMESTAMP DEFAULT NOW(),
  updated_at          TIMESTAMP DEFAULT NOW()
);

-- 角色权限定义表
CREATE TABLE IF NOT EXISTS role_permissions (
  id                  SERIAL PRIMARY KEY,
  role_name           VARCHAR(100) NOT NULL UNIQUE,
  permissions         JSONB NOT NULL DEFAULT '[]',
  description         TEXT,
  created_at          TIMESTAMP DEFAULT NOW(),
  updated_at          TIMESTAMP DEFAULT NOW()
);

-- CI实例角色关联表
CREATE TABLE IF NOT EXISTS ci_instance_roles (
  id                  SERIAL PRIMARY KEY,
  ci_id               INTEGER NOT NULL REFERENCES ci_instances(id) ON DELETE CASCADE,
  role_type           VARCHAR(20) NOT NULL,  -- 'ci_role' 或 'owner_role'
  role_id             INTEGER NOT NULL,
  user_id             INTEGER REFERENCES users(id) ON DELETE SET NULL,
  assigned_at         TIMESTAMP DEFAULT NOW(),
  assigned_by         INTEGER REFERENCES users(id) ON DELETE SET NULL,
  CONSTRAINT check_role_type CHECK (role_type IN ('ci_role', 'owner_role'))
);

CREATE INDEX idx_ci_instance_roles_ci_id ON ci_instance_roles(ci_id);
CREATE INDEX idx_ci_instance_roles_type_id ON ci_instance_roles(role_type, role_id);
CREATE UNIQUE INDEX idx_ci_instance_roles_unique ON ci_instance_roles(ci_id, role_type, role_id, COALESCE(user_id, 0));

-- =====================================================
-- 2. 标签模块表结构
-- =====================================================

-- 标签分类表
CREATE TABLE IF NOT EXISTS tag_categories (
  id                  SERIAL PRIMARY KEY,
  name                VARCHAR(100) NOT NULL UNIQUE,
  display_name        VARCHAR(200) NOT NULL,
  description         TEXT,
  color               VARCHAR(20),
  icon                VARCHAR(50),
  sort_order          INTEGER DEFAULT 0,
  is_system           BOOLEAN DEFAULT FALSE,
  created_at          TIMESTAMP DEFAULT NOW(),
  updated_at          TIMESTAMP DEFAULT NOW()
);

-- 标签定义表
CREATE TABLE IF NOT EXISTS tags (
  id                  SERIAL PRIMARY KEY,
  category_id         INTEGER REFERENCES tag_categories(id) ON DELETE SET NULL,
  name                VARCHAR(100) NOT NULL,
  display_name        VARCHAR(200) NOT NULL,
  color               VARCHAR(20),
  description         TEXT,
  usage_count         INTEGER DEFAULT 0,
  is_active           BOOLEAN DEFAULT TRUE,
  created_at          TIMESTAMP DEFAULT NOW(),
  updated_at          TIMESTAMP DEFAULT NOW(),
  UNIQUE(category_id, name)
);

-- CI实例标签关联表
CREATE TABLE IF NOT EXISTS ci_tags (
  id                  SERIAL PRIMARY KEY,
  ci_id               INTEGER NOT NULL REFERENCES ci_instances(id) ON DELETE CASCADE,
  tag_id              INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
  tagged_by           INTEGER REFERENCES users(id) ON DELETE SET NULL,
  tagged_at           TIMESTAMP DEFAULT NOW(),
  UNIQUE(ci_id, tag_id)
);

-- 标签使用历史表
CREATE TABLE IF NOT EXISTS tag_history (
  id                  SERIAL PRIMARY KEY,
  ci_id               INTEGER REFERENCES ci_instances(id) ON DELETE CASCADE,
  tag_id              INTEGER REFERENCES tags(id) ON DELETE CASCADE,
  action              VARCHAR(20) NOT NULL,
  user_id             INTEGER REFERENCES users(id) ON DELETE SET NULL,
  created_at          TIMESTAMP DEFAULT NOW(),
  CONSTRAINT check_action CHECK (action IN ('added', 'removed'))
);

CREATE INDEX idx_ci_tags_ci_id ON ci_tags(ci_id);
CREATE INDEX idx_ci_tags_tag_id ON ci_tags(tag_id);
CREATE INDEX idx_tag_history_ci_id ON tag_history(ci_id);
CREATE INDEX idx_tag_history_tag_id ON tag_history(tag_id);

-- 更新时间戳触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_ci_roles_updated_at BEFORE UPDATE ON ci_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_owner_roles_updated_at BEFORE UPDATE ON owner_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_role_permissions_updated_at BEFORE UPDATE ON role_permissions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tag_categories_updated_at BEFORE UPDATE ON tag_categories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_tags_updated_at BEFORE UPDATE ON tags
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- 3. 预设数据
-- =====================================================

-- 插入预设标签分类
INSERT INTO tag_categories (name, display_name, description, color, is_system, sort_order) VALUES
('environment', '环境', '环境标签，标识CI所处的环境阶段', 'blue', TRUE, 1),
('business', '业务', '业务标签，标识CI的业务重要性', 'green', TRUE, 2),
('location', '位置', '位置标签，标识CI的物理位置', 'orange', TRUE, 3),
('custom', '自定义', '用户自定义标签', 'gray', FALSE, 4)
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
('admin', '系统管理员', '负责系统的整体管理和维护', 0, '{"permissions": ["all"], "duties": ["系统配置", "用户管理", "安全策略"]}'),
('ops_owner', '运维负责人', '负责运维相关的操作和维护', 1, '{"permissions": ["deploy", "monitor", "backup"], "duties": ["部署发布", "监控告警", "备份恢复"]}'),
('biz_owner', '业务负责人', '负责业务需求和业务逻辑', 2, '{"permissions": ["requirement", "acceptance"], "duties": ["需求提报", "验收测试"]}'),
('sec_owner', '安全负责人', '负责安全审计和合规性检查', 3, '{"permissions": ["audit", "security"], "duties": ["安全审计", "漏洞扫描"]}'),
('dev_owner', '开发负责人', '负责开发任务和技术实现', 4, '{"permissions": ["develop", "review"], "duties": ["代码开发", "代码评审"]}')
ON CONFLICT (name) DO NOTHING;

-- 插入角色权限定义
INSERT INTO role_permissions (role_name, permissions, description) VALUES
('admin', '["*:*"]', '管理员全部权限'),
('developer', '["cmdb:read", "cmdb:create", "ticket:read", "ticket:create", "alert:read"]', '开发者权限'),
('ops', '["cmdb:*", "ticket:*", "alert:*", "report:read"]', '运维工程师权限'),
('reader', '["cmdb:read", "ticket:read", "alert:read"]', '只读用户权限'),
('auditor', '["*:*:read"]', '审计员权限')
ON CONFLICT (role_name) DO NOTHING;

COMMENT ON TABLE ci_roles IS 'CI技术角色表，定义配置项在系统中的技术职能';
COMMENT ON TABLE owner_roles IS '负责人角色表，定义资产负责人类型和职责';
COMMENT ON TABLE role_permissions IS '角色权限表，定义用户角色的权限集合';
COMMENT ON TABLE ci_instance_roles IS 'CI实例角色关联表，关联CI与角色';
COMMENT ON TABLE tag_categories IS '标签分类表，用于组织标签';
COMMENT ON TABLE tags IS '标签定义表';
COMMENT ON TABLE ci_tags IS 'CI实例标签关联表';
COMMENT ON TABLE tag_history IS '标签使用历史表';

-- 008: 告警空间与路由规则
-- 告警空间
CREATE TABLE IF NOT EXISTS alert_spaces (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    description VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 空间-角色关联（对应 auth-service 的 roles 表）
CREATE TABLE IF NOT EXISTS alert_space_roles (
    space_id INTEGER NOT NULL REFERENCES alert_spaces(id) ON DELETE CASCADE,
    role_id INTEGER NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (space_id, role_id)
);

-- 空间路由规则
CREATE TABLE IF NOT EXISTS alert_space_routes (
    id SERIAL PRIMARY KEY,
    field_name VARCHAR(100) NOT NULL,
    field_value VARCHAR(255) NOT NULL,
    space_id INTEGER NOT NULL REFERENCES alert_spaces(id) ON DELETE CASCADE,
    priority INTEGER DEFAULT 0,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_alert_space_routes_enabled ON alert_space_routes(enabled);
CREATE INDEX IF NOT EXISTS idx_alert_space_routes_space_id ON alert_space_routes(space_id);

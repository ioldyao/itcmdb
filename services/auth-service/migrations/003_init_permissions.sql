-- 初始化系统权限和角色权限关联
-- 这个脚本应该在系统首次部署时运行，或者在权限系统修复时运行

-- 1. 创建系统默认权限
INSERT INTO permissions (resource, action) VALUES
-- 用户管理权限
('user', 'create'),
('user', 'update'),
('user', 'delete'),
('user', 'view'),
('user', 'manage'),

-- 角色管理权限
('role', 'create'),
('role', 'update'),
('role', 'delete'),
('role', 'view'),
('role', 'manage'),

-- 权限管理权限
('permission', 'create'),
('permission', 'delete'),
('permission', 'view'),

-- 配置管理权限
('config', 'create'),
('config', 'update'),
('config', 'delete'),
('config', 'view'),

-- CI管理权限
('ci', 'create'),
('ci', 'update'),
('ci', 'delete'),
('ci', 'view'),

-- 标签管理权限
('tag', 'create'),
('tag', 'update'),
('tag', 'delete'),
('tag', 'view'),

-- 工单管理权限
('ticket', 'create'),
('ticket', 'update'),
('ticket', 'delete'),
('ticket', 'view'),

-- 告警管理权限
('alert', 'create'),
('alert', 'update'),
('alert', 'delete'),
('alert', 'view'),

-- 审计日志权限
('audit', 'view')

ON CONFLICT (resource, action) DO NOTHING;

-- 2. 为 admin 角色分配所有权限
-- 首先确保 admin 角色存在
INSERT INTO roles (name, description, created_at, updated_at) VALUES
('admin', '系统管理员，拥有所有权限', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- 获取 admin 角色 ID 和超级管理员权限 ID
DO $$
DECLARE
    admin_role_id INT;
    super_admin_perm_id INT;
    perm_record RECORD;
BEGIN
    -- 获取 admin 角色 ID
    SELECT id INTO admin_role_id FROM roles WHERE name = 'admin';

    -- 获取超级管理员权限 ID (*:*)
    SELECT id INTO super_admin_perm_id FROM permissions WHERE resource = '*' AND action = '*';

    -- 如果超级管理员权限不存在，创建它
    IF super_admin_perm_id IS NULL THEN
        INSERT INTO permissions (resource, action)
        VALUES ('*', '*')
        RETURNING id INTO super_admin_perm_id;
    END IF;

    -- 为 admin 角色分配超级管理员权限
    INSERT INTO role_permissions (role_id, permission_id)
    VALUES (admin_role_id, super_admin_perm_id)
    ON CONFLICT (role_id, permission_id) DO NOTHING;

    -- 为 admin 角色分配所有其他权限（作为备份）
    FOR perm_record IN SELECT id FROM permissions WHERE id != super_admin_perm_id LOOP
        INSERT INTO role_permissions (role_id, permission_id)
        VALUES (admin_role_id, perm_record.id)
        ON CONFLICT (role_id, permission_id) DO NOTHING;
    END LOOP;

    RAISE NOTICE 'Admin role permissions initialized successfully';
END $$;

-- 3. 确保 admin 用户拥有 admin 角色
DO $$
DECLARE
    admin_user_id INT;
    admin_role_id INT;
BEGIN
    -- 获取 admin 用户 ID
    SELECT id INTO admin_user_id FROM users WHERE username = 'admin';

    -- 获取 admin 角色 ID
    SELECT id INTO admin_role_id FROM roles WHERE name = 'admin';

    -- 如果 admin 用户存在，确保其拥有 admin 角色
    IF admin_user_id IS NOT NULL AND admin_role_id IS NOT NULL THEN
        INSERT INTO user_roles (user_id, role_id)
        VALUES (admin_user_id, admin_role_id)
        ON CONFLICT (user_id, role_id) DO NOTHING;

        RAISE NOTICE 'Admin user role assigned successfully';
    ELSE
        RAISE NOTICE 'Admin user or admin role not found';
    END IF;
END $$;

-- 4. 创建普通用户角色（可选）
INSERT INTO roles (name, description, created_at, updated_at) VALUES
('user', '普通用户，只能查看和创建自己的资源', NOW(), NOW()),
('operator', '运维人员，可以管理CI和查看配置', NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

-- 为普通用户角色分配基本权限
DO $$
DECLARE
    user_role_id INT;
    operator_role_id INT;
BEGIN
    SELECT id INTO user_role_id FROM roles WHERE name = 'user';
    SELECT id INTO operator_role_id FROM roles WHERE name = 'operator';

    -- 普通用户权限：查看和创建工单、查看CI
    INSERT INTO role_permissions (role_id, permission_id)
    SELECT user_role_id, id FROM permissions
    WHERE (resource = 'ticket' AND action IN ('create', 'view'))
       OR (resource = 'ci' AND action = 'view')
       OR (resource = 'alert' AND action = 'view')
    ON CONFLICT (role_id, permission_id) DO NOTHING;

    -- 运维人员权限：管理CI、查看配置、管理标签
    INSERT INTO role_permissions (role_id, permission_id)
    SELECT operator_role_id, id FROM permissions
    WHERE (resource = 'ci' AND action IN ('create', 'update', 'delete', 'view'))
       OR (resource = 'config' AND action = 'view')
       OR (resource = 'tag' AND action IN ('create', 'update', 'delete', 'view'))
       OR (resource = 'ticket' AND action IN ('create', 'update', 'view'))
       OR (resource = 'alert' AND action IN ('view', 'update'))
    ON CONFLICT (role_id, permission_id) DO NOTHING;

    RAISE NOTICE 'User and operator role permissions initialized successfully';
END $$;

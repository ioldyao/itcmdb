-- 修复CI类型的display_name字段
-- 日期: 2026-01-26
-- 问题: CI types的display_name字段没有正确设置中文名称

UPDATE ci_types SET display_name = '服务器' WHERE name = 'server';
UPDATE ci_types SET display_name = '网络设备' WHERE name = 'network';
UPDATE ci_types SET display_name = '应用服务' WHERE name = 'application';
UPDATE ci_types SET display_name = '容器' WHERE name = 'container';

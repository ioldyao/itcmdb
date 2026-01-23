#!/bin/bash
# 数据库迁移脚本

set -e

echo "执行数据库迁移..."

# 检查 PostgreSQL 是否可用
until docker exec -it itcmdb-postgres pg_isready -U postgres; do
    echo "等待 PostgreSQL 启动..."
    sleep 2
done

echo "PostgreSQL 已就绪，执行初始化脚本..."

# 执行初始化脚本
docker exec -i itcmdb-postgres psql -U postgres -d itcmdb < scripts/init-db.sql

echo "数据库迁移完成！"

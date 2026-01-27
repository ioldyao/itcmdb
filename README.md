# ITCMDB - 一体化运维管理平台

## 项目简介

ITCMDB 是一个功能完整的运维管理平台，集成了 **CMDB**、**工单系统**、**告警系统** 和 **审计系统**，采用微服务架构设计。

平台核心特色：
- 🎯 **CMDB资产管理** - 服务器、网络设备、应用服务、容器的全生命周期管理
- 🏷️ **角色与标签系统** - 灵活的技术角色、负责人角色和业务标签管理
- 🔄 **容器自动发现** - 自动从 VictoriaMetrics 同步容器信息
- 📊 **变更历史追踪** - 完整的配置项变更记录和时间线展示
- 🔐 **基于权限的访问控制** - 细粒度的 RBAC 权限管理
- 🎫 **工单系统** - 支持工作流引擎和工单模板
- 🚨 **告警管理** - 告警规则引擎和实时告警处理
- 📝 **审计日志** - 完整的操作审计追踪

## 技术栈

### 前端
- **React 18** + TypeScript
- **Ant Design 5** - UI组件库
- **Zustand** - 轻量级状态管理
- **React Router v6** - 路由管理
- **Vite** - 快速构建工具

### 后端
- **Go 1.21** - 高性能后端语言
- **Gin** - HTTP框架
- **GORM** - ORM框架
- **gRPC** - 服务间通信
- **JWT** - 认证授权

### 基础设施
- **PostgreSQL 15** - 主数据库
- **Redis 7** - 缓存和会话存储
- **Kafka 4.1.0** (KRaft模式) - 消息队列
- **VictoriaMetrics** - 监控数据存储
- **Docker & Docker Compose** - 容器化部署

## 核心功能

### 1. 用户认证与权限 (auth-service:5001)

**已实现功能：**
- ✅ JWT令牌签发与验证
- ✅ 用户登录/登出/注册
- ✅ 密码加密存储 (bcrypt)
- ✅ RBAC权限管理
- ✅ 用户角色分配
- ✅ 权限缓存 (Redis)
- ✅ gRPC服务间认证
- ✅ 审计日志集成

**预定义角色：**
- `admin` - 系统管理员，拥有所有权限
- `operator` - 运维工程师，可操作CMDB和工单
- `developer` - 开发工程师，只读权限和创建工单
- `viewer` - 只读用户，仅查看权限

### 2. CMDB服务 (cmdb-service:5002)

**已实现功能：**
- ✅ 配置项(CI)全生命周期管理
- ✅ 四种CI类型：服务器、网络设备、应用服务、容器
- ✅ 动态属性定义 (JSONB)
- ✅ CI关系管理 (depends_on, runs_on, connects_to, contains)
- ✅ CI变更历史追踪
- ✅ **容器自动发现** - 从VictoriaMetrics自动同步容器
- ✅ **角色管理** - 技术角色和负责人角色
- ✅ **标签系统** - 环境标签、业务标签、位置标签
- ✅ **硬件信息采集** - cAdvisor集成
- ✅ CI导入/导出功能

**CI类型详情：**

| 类型 | 说明 | 主要属性 |
|------|------|----------|
| **服务器** | 物理机、虚拟机 | 主机名、IP、CPU、内存、操作系统等 |
| **网络设备** | 交换机、路由器 | 设备型号、管理IP、端口数等 |
| **应用服务** | 中间件、数据库 | 服务名称、版本、端口、配置等 |
| **容器** | Docker容器 | 容器名、镜像、状态、资源使用等 |

**容器自动发现特性：**
- 每5分钟自动从VictoriaMetrics同步容器列表
- 自动检测容器上下线状态
- 容器重建检测（Container ID变更）
- 容器生命周期历史记录
- 实时资源监控数据

### 3. 工单服务 (ticket-service:5003)

**已实现功能：**
- ✅ 工单基础数据模型
- ✅ 工作流定义
- ✅ 工单模板
- ✅ 自动工单编号生成 (TKT-YYYYMMDD-序号)

**开发中功能：**
- 🚧 工单业务逻辑处理
- 🚧 SLA监控
- 🚧 工单状态机

### 4. 告警服务 (alert-service:5004)

**已实现功能：**
- ✅ 告警基础数据模型
- ✅ 告警规则定义
- ✅ 告警阈值配置

**开发中功能：**
- 🚧 告警规则引擎
- 🚧 实时告警处理
- 🚧 告警通知集成

### 5. 通知服务 (notification-service:5005)

**计划支持渠道：**
- 企业微信
- 钉钉
- 邮件
- 短信

**开发中功能：**
- 🚧 通知发送逻辑
- 🚧 通知模板管理
- 🚧 发送失败重试

### 6. 报表服务 (report-service:5006)

**计划功能：**
- CMDB资产报表
- 工单统计分析
- 告警趋势分析
- 报表导出

### 7. 审计服务 (audit-service:5007)

**已实现功能：**
- ✅ 审计日志采集 (Kafka消费)
- ✅ 审计日志存储
- ✅ 批量处理机制
- ✅ 优雅关闭处理

## 快速开始

### 前置要求

- **Docker** 20.10+
- **Docker Compose** 2.0+
- **Go** 1.21+ (本地开发)
- **Node.js** 20+ (本地开发)

### 一键部署

```bash
# 克隆项目
git clone https://codeberg.org/iEZELL/itcmdb.git
cd itcmdb

# 启动所有服务（包括数据库初始化）
./scripts/start.sh

# 等待服务启动完成（约30-60秒）
docker compose ps

# 查看服务日志
docker compose logs -f
```

### 服务访问地址

| 服务 | 地址 | 说明 |
|------|------|------|
| **前端应用** | http://localhost | React前端 |
| **API网关** | http://localhost:8000 | Nginx反向代理 |
| **认证服务** | http://localhost:5001 | auth-service |
| **CMDB服务** | http://localhost:5002 | cmdb-service |
| **工单服务** | http://localhost:5003 | ticket-service |
| **告警服务** | http://localhost:5004 | alert-service |
| **通知服务** | http://localhost:5005 | notification-service |
| **报表服务** | http://localhost:5006 | report-service |
| **审计服务** | http://localhost:5007 | audit-service |
| **Kafka UI** | http://localhost:8081 | Kafka管理界面 |

### 默认账号

```
用户名: admin
密码: admin123
```

⚠️ **重要提示：请在生产环境中立即修改默认密码！**

### 数据库初始化

数据库会在首次启动时自动通过 `scripts/init-db.sql` 初始化，包括：
- 创建所有数据表
- 创建索引和触发器
- 插入默认管理员账号
- 插入预定义角色和权限
- 插入CI类型、标签分类、预设标签等

如需手动重新初始化数据库：
```bash
docker exec -i itcmdb-postgres psql -U itcmdb -d itcmdb < scripts/init-db.sql
```

## 停止服务

```bash
# 停止所有服务
./scripts/stop.sh

# 或使用 docker compose
docker compose down

# 删除数据卷 (⚠️ 警告: 会删除所有数据)
docker compose down -v
```

## 项目结构

```
itcmdb/
├── frontend/                    # React前端应用
│   ├── src/
│   │   ├── components/         # 可复用组件
│   │   │   ├── ProtectedRoute.tsx
│   │   │   ├── PermissionGuard.tsx
│   │   │   ├── CIHistoryTimeline.tsx
│   │   │   ├── CIRelationGraph.tsx
│   │   │   └── ...
│   │   ├── pages/              # 页面组件
│   │   │   ├── Auth/           # 认证页面
│   │   │   ├── Dashboard/      # 仪表盘
│   │   │   ├── CMDB/           # CMDB模块
│   │   │   ├── Tickets/        # 工单模块
│   │   │   ├── Alerts/         # 告警模块
│   │   │   └── Admin/          # 管理模块
│   │   ├── services/           # API服务层
│   │   ├── stores/             # Zustand状态管理
│   │   ├── types/              # TypeScript类型定义
│   │   └── router.tsx          # 路由配置
│   ├── package.json
│   └── Dockerfile
│
├── services/                    # Go微服务
│   ├── auth-service/           # 认证服务 (5001/50001)
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── config/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   └── model/
│   │   └── migrations/         # 数据库迁移
│   │
│   ├── cmdb-service/           # CMDB服务 (5002/50002)
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   │   ├── service/
│   │   │   │   └── container_sync_service.go  # 容器自动同步
│   │   │   ├── handler/
│   │   │   └── model/
│   │
│   ├── ticket-service/         # 工单服务 (5003)
│   ├── alert-service/          # 告警服务 (5004)
│   ├── notification-service/   # 通知服务 (5005)
│   ├── report-service/         # 报表服务 (5006)
│   ├── audit-service/          # 审计服务 (5007)
│   │
│   └── shared/                 # 共享代码库
│       └── pkg/
│           ├── auth/           # JWT认证工具
│           ├── cache/          # Redis客户端
│           ├── database/       # PostgreSQL连接
│           ├── kafka/          # Kafka生产者/消费者
│           ├── logger/         # Zap日志
│           ├── rbac/           # 权限控制
│           ├── grpc/           # gRPC客户端
│           └── proto/          # Protocol Buffers定义
│
├── scripts/                    # 部署和维护脚本
│   ├── init-db.sql            # 数据库初始化脚本 (v2.0)
│   ├── start.sh               # 一键启动脚本
│   ├── stop.sh                # 停止脚本
│   ├── build.sh               # 构建所有服务
│   ├── docker-build.sh        # Docker镜像构建
│   └── migrations/            # 数据库迁移文件
│
├── deploy/                     # 部署配置
│   ├── nginx/                 # Nginx配置
│   └── kubernetes/            # Kubernetes部署文件
│
├── compose.yml                 # Docker Compose配置
├── CLAUDE.md                  # Claude开发指南
└── README.md                  # 本文档
```

## 本地开发

### 前端开发

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 构建生产版本
npm run build

# 类型检查
npm run type-check
```

### 后端开发

```bash
# 启动基础设施 (PostgreSQL, Redis, Kafka)
docker compose up -d postgres redis kafka

# 启动单个服务 (例如: auth-service)
cd services/auth-service
go run cmd/main.go

# 运行测试
go test ./...

# 构建
go build -o bin/auth-service cmd/main.go
```

## 权限系统

### 权限格式

权限定义格式：`<resource>:<action>`

**资源类型：**
- `user` - 用户管理
- `role` - 角色管理
- `permission` - 权限管理
- `config` - 配置管理
- `ci` - 配置项管理
- `tag` - 标签管理
- `ticket` - 工单管理
- `alert` - 告警管理
- `audit` - 审计日志
- `*` - 所有资源（超级管理员）

**操作类型：**
- `create` - 创建
- `update` - 更新
- `delete` - 删除
- `view` - 查看
- `manage` - 管理
- `*` - 所有操作（超级管理员）

### 权限示例

```javascript
// 前端权限检查
{hasPermission('ci', 'create') && (
  <Button>创建CI</Button>
)}

// 后端权限检查
if !hasPermission(c, "ci", "delete") {
    return
}
```

## 容器自动发现

### 配置说明

ITCMDB支持通过**Web界面**配置多个VictoriaMetrics数据源，实现从多个监控源自动同步容器信息。

#### 通过Web界面配置（推荐）

1. 登录ITCMDB管理界面
2. 进入 **系统配置** -> **新增配置**
3. 配置参数：
   - **分类**: `monitoring`
   - **配置键**: `victoriametrics_datasources`
   - **配置值**: JSON格式的数据源数组

配置示例（JSON格式）：
```json
[
  {
    "name": "主数据中心",
    "id": "primary-dc",
    "endpoint": "https://victoriametrics-primary.example.com:8429",
    "username": "admin",
    "password": "your_password",
    "enabled": true,
    "container_prefix": ["prod-", "staging-"],
    "labels": {
      "location": "主数据中心",
      "environment": "production"
    }
  },
  {
    "name": "备数据中心",
    "id": "secondary-dc",
    "endpoint": "https://victoriametrics-backup.example.com:8429",
    "username": "admin",
    "password": "backup_password",
    "enabled": true,
    "labels": {
      "location": "备数据中心",
      "environment": "dr"
    }
  },
  {
    "name": "开发环境",
    "id": "dev-env",
    "endpoint": "https://victoriametrics-dev.example.com:8429",
    "username": "dev_user",
    "password": "dev_password",
    "enabled": false,
    "labels": {
      "location": "开发环境",
      "environment": "development"
    }
  }
]
```

4. 保存配置后，容器同步服务会自动从数据库读取配置并启动

#### 配置字段说明

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `name` | string | 是 | 数据源显示名称 |
| `id` | string | 是 | 数据源唯一标识符 |
| `endpoint` | string | 是 | VictoriaMetrics API地址 |
| `username` | string | 否 | 认证用户名 |
| `password` | string | 否 | 认证密码 |
| `enabled` | boolean | 否 | 是否启用（默认true） |
| `container_prefix` | []string | 否 | 容器名前缀过滤，只同步匹配前缀的容器 |
| `labels` | object | 否 | 自动添加到容器的标签 |

### 多数据源特性

- **并发同步** - 从多个数据源并发发现和同步容器
- **数据源隔离** - 每个数据源独立配置认证信息和标签
- **容器过滤** - 通过 `container_prefix` 指定数据源负责的容器
- **自动标签** - 数据源的 `labels` 会自动添加到同步的容器上
- **健康检查** - 定期检查所有数据源的健康状态
- **容错机制** - 单个数据源故障不影响其他数据源的同步
- **动态配置** - 通过Web界面实时修改，无需重启服务

### 容器属性

从多数据源同步的容器包含以下额外属性：

```json
{
  "datasource_id": "primary-dc",           // 数据源ID
  "datasource_name": "主数据中心",         // 数据源名称
  "datasource_labels": {                  // 数据源标签
    "location": "主数据中心",
    "environment": "production"
  },
  "container_name": "prod-web-server-1",
  "container_id": "abc123...",
  "is_online": true,
  "last_seen": "2026-01-26T12:00:00Z"
}
```

### 同步特性

- **自动发现** - 自动发现新创建的容器
- **状态检测** - 检测容器在线/离线状态
- **重建检测** - 识别容器重建（Container ID变化）
- **历史记录** - 记录容器ID变更历史
- **资源监控** - 同步CPU、内存、网络等指标
- **数据源切换** - 检测容器在不同数据源间的迁移

### 支持的指标

- `container_last_seen` - 最后活跃时间
- `container_cpu_usage_seconds_total` - CPU使用量
- `container_memory_working_set_bytes` - 内存使用量
- `container_network_receive_bytes_total` - 网络接收字节数
- `container_network_transmit_bytes_total` - 网络发送字节数
- `container_fs_usage_bytes` - 文件系统使用量

## 数据库表结构

### 核心表

**用户与权限模块：**
- `users` - 用户表
- `roles` - 角色表
- `permissions` - 权限表
- `user_roles` - 用户角色关联表
- `role_permissions` - 角色权限关联表

**CMDB模块：**
- `ci_types` - CI类型表
- `ci_attributes` - CI属性定义表
- `ci_instances` - CI实例表
- `ci_relations` - CI关系表
- `ci_history` - CI变更历史表
- `ci_roles` - CI角色表（技术角色）
- `owner_roles` - 负责人角色表
- `ci_instance_roles` - CI实例角色关联表
- `tag_categories` - 标签分类表
- `tags` - 标签定义表
- `ci_tags` - CI标签关联表
- `tag_history` - 标签历史表
- `system_config` - 系统配置表

**工单模块：**
- `ticket_workflows` - 工作流定义表
- `ticket_templates` - 工单模板表
- `tickets` - 工单主表
- `ticket_comments` - 工单评论表
- `ticket_attachments` - 工单附件表
- `ticket_history` - 工单历史表

**告警模块：**
- `alert_rules` - 告警规则表
- `alert_thresholds` - 告警阈值表
- `alert_instances` - 告警实例表
- `alert_history` - 告警历史表

**通知模块：**
- `notification_templates` - 通知模板表
- `notification_history` - 通知历史表

**审计模块：**
- `audit_logs` - 审计日志表

**报表模块：**
- `report_configs` - 报表配置表

### 查看数据库

```bash
# 连接到PostgreSQL
docker exec -it itcmdb-postgres psql -U itcmdb -d itcmdb

# 查看所有表
\dt

# 查看表结构
\d ci_instances

# 退出
\q
```

## Kafka主题

系统使用以下Kafka主题进行事件驱动：

- `audit_logs` - 审计日志事件
- `ticket.created` - 工单创建事件
- `ticket.status.changed` - 工单状态变更事件
- `ticket.assigned` - 工单分配事件
- `ticket.sla.breached` - SLA违约事件
- `alert.triggered` - 告警触发事件
- `alert.acknowledged` - 告警确认事件
- `alert.closed` - 告警关闭事件
- `alert.escalated` - 告警升级事件
- `ci.changed` - CI变更事件
- `ci.deleted` - CI删除事件
- `ci.relationship.changed` - CI关系变更事件
- `notification.send` - 通知发送事件

## 配置说明

### 环境变量

各服务的配置文件位于 `services/<service-name>/internal/config/config.yaml`

支持通过环境变量覆盖配置：
- `AUTH_DATABASE_HOST` - 认证服务数据库主机
- `CMDB_DATABASE_HOST` - CMDB服务数据库主机
- `REDIS_HOST` - Redis主机地址
- `KAFKA_BROKERS` - Kafka集群地址

### Nginx API网关

Nginx配置位于 `deploy/nginx/nginx.conf`，提供：
- 反向代理到各个后端服务
- CORS配置
- 请求路由

## 故障排查

### 服务无法启动

```bash
# 查看服务日志
docker compose logs -f <service-name>

# 检查服务状态
docker compose ps

# 重启单个服务
docker compose restart <service-name>
```

### 数据库连接失败

```bash
# 检查PostgreSQL状态
docker compose logs postgres

# 验证数据库连接
docker exec -it itcmdb-postgres psql -U itcmdb -d itcmdb
```

### 前端无法访问后端

```bash
# 检查API网关
docker compose logs nginx

# 检查后端服务状态
docker compose ps
```

## 文档

- [容器自动同步详细说明](docs/CONTAINER_AUTO_SYNC.md)
- [镜像优化指南](docs/MIRROR_OPTIMIZATION.md)
- [微服务架构设计](docs/plans/2026-01-24-itcmdb-microservices-design.md)

## 开发路线图

### 已完成 ✅

- [x] 用户认证与权限管理
- [x] CMDB核心功能
- [x] 容器自动发现
- [x] 角色与标签系统
- [x] CI变更历史
- [x] 审计日志系统
- [x] 前端24个页面
- [x] 权限UI控制

### 开发中 🚧

- [ ] 工单业务逻辑
- [ ] 告警规则引擎
- [ ] 通知服务实现
- [ ] 报表服务实现

### 计划中 📋

- [ ] 工作流引擎
- [ ] SLA监控
- [ ] 高级报表功能
- [ ] 移动端适配
- [ ] 多租户支持

## 贡献指南

欢迎贡献代码、报告问题或提出建议！

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

MIT License

## 联系方式

如有问题或建议，请提交 Issue 或 Pull Request。

---

**ITCMDB** - 让运维管理更简单、更高效！

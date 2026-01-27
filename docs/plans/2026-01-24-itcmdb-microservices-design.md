# ITCMDB 微服务系统设计文档

**日期:** 2026-01-24
**版本:** v1.0

---

## 1. 项目概述

### 1.1 项目目标
构建一个 CMDB + 工单系统 + 告警系统的一体化微服务平台，服务于内部运维团队。

### 1.2 技术栈

| 层级 | 技术选型 |
|------|----------|
| 前端 | React 18 + TypeScript + Ant Design 5 + Zustand + Vite |
| 后端 | Go + gRPC + Gin |
| 数据库 | PostgreSQL 15 |
| 缓存 | Redis 7 |
| 消息队列 | Kafka |
| 容器化 | Docker + Docker Compose |
| API网关 | Kong/Traefik |

---

## 2. 整体架构

### 2.1 微服务架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                         API Gateway (Kong/Traefik)               │
│                     认证 + 限流 + 路由                           │
└─────────────────────────────────────────────────────────────────┘
                                    │
        ┌───────────────────────────┼───────────────────────────┐
        │                           │                           │
        ▼                           ▼                           ▼
┌──────────────┐          ┌──────────────┐          ┌──────────────┐
│   前端应用    │          │  用户认证服务 │          │   公共API    │
│  (React)     │◄────────►│   (Auth)     │◄────────►│  (Swagger)   │
└──────────────┘          └──────────────┘          └──────────────┘
                                       │
        ┌──────────────────────────────┼──────────────────────────────┐
        │                              │                              │
        ▼                              ▼                              ▼
┌──────────────┐          ┌──────────────┐          ┌──────────────┐
│  CMDB服务    │          │  工单服务    │          │  告警服务    │
│  (Asset)     │          │  (Ticket)    │          │  (Alert)     │
└──────────────┘          └──────────────┘          └──────────────┘
        │                              │                              │
        └──────────────────────────────┼──────────────────────────────┘
                                       │
        ┌──────────────────────────────┼──────────────────────────────┐
        │                              │                              │
        ▼                              ▼                              ▼
┌──────────────┐          ┌──────────────┐          ┌──────────────┐
│  通知服务    │          │  报表服务    │          │  审计日志    │
│ (Notification│          │  (Report)    │          │   (Audit)    │
└──────────────┘          └──────────────┘          └──────────────┘
```

### 2.2 服务间通信模式

| 场景 | 通信方式 | 说明 |
|------|----------|------|
| 前端→后端 | REST/gRPC | API Gateway 转发 |
| 同步调用 | gRPC | 服务间高性能调用 |
| 异步事件 | Kafka | 工单状态变更、告警触发等 |
| 缓存 | Redis | 热点数据、会话 |
| 持久化 | PostgreSQL | 关系型数据存储 |

---

## 3. 微服务职责划分

### 3.1 用户认证服务 (auth-service:5001)

**职责：**
- JWT令牌签发与验证
- 用户登录/登出
- 密码加密存储
- 权限验证中间件

**核心API：**
```
POST   /api/v1/auth/login
POST   /api/v1/auth/logout
POST   /api/v1/auth/refresh
GET    /api/v1/users/me
PUT    /api/v1/users/me
GET    /api/v1/users/:id/permissions
```

### 3.2 CMDB服务 (cmdb-service:5002)

**职责：**
- 配置项(CI)CRUD管理
- 动态属性模型支持
- CI关系图谱查询
- 自动发现数据接入
- 变更历史审计

**CMDB资源类型：**
- 服务器/主机
- 网络设备
- 应用服务
- 容器/K8s

**核心API：**
```
GET    /api/v1/ci/types
POST   /api/v1/ci/instances
GET    /api/v1/ci/instances/:id
PUT    /api/v1/ci/instances/:id
DELETE /api/v1/ci/instances/:id
GET    /api/v1/ci/relations
POST   /api/v1/ci/sync
```

### 3.3 工单服务 (ticket-service:5003)

**职责：**
- 工单全生命周期管理
- 工作流状态机引擎
- 工单分派与升级规则
- SLA计时与监控

**工作流模式：** 预设流程 + 可扩展设计

**核心API：**
```
GET    /api/v1/tickets
POST   /api/v1/tickets
GET    /api/v1/tickets/:id
PUT    /api/v1/tickets/:id/status
POST   /api/v1/tickets/:id/comments
GET    /api/v1/workflows
GET    /api/v1/tickets/:id/sla
```

### 3.4 告警服务 (alert-service:5004)

**职责：**
- 告警规则引擎
- 实时告警处理与去重
- 告警升级与聚合
- CMDB事件驱动告警

**触发源：**
- 监控系统集成 (Prometheus/Zabbix)
- CMDB事件驱动
- 规则引擎
- 工单联动

**核心API：**
```
GET    /api/v1/alerts
POST   /api/v1/alerts/:id/ack
POST   /api/v1/alerts/:id/close
GET    /api/v1/rules
POST   /api/v1/rules
POST   /api/v1/alerts/ingest
```

### 3.5 通知服务 (notification-service:5005)

**职责：**
- 统一通知发送
- 通知模板管理
- 发送失败重试
- 通知历史记录

**通知渠道：**
- 企业微信/钉钉
- 邮件
- 短信/语音

### 3.6 报表服务 (report-service:5006)

**职责：**
- CMDB资产报表
- 工单统计分析
- 告警趋势分析
- 报表导出

---

## 4. 数据模型设计

### 4.1 用户与权限模块

```sql
-- 用户表
users (
  id, username, email, password_hash, full_name, status, created_at, updated_at
)

-- 角色表
roles (
  id, name, description, created_at, updated_at
)

-- 权限表
permissions (
  id, resource, action, description
)

-- 角色权限关联表
role_permissions (
  role_id, permission_id
)

-- 用户角色关联表
user_roles (
  user_id, role_id
)
```

### 4.2 CMDB模块

```sql
-- CI类型表
ci_types (
  id, name, icon, description, is_active
)

-- CI属性定义表
ci_attributes (
  id, ci_type_id, name, type, options, is_required
)

-- CI实例表
ci_instances (
  id, ci_type_id, name, status, attributes, created_at, updated_at
)

-- CI关系表
ci_relations (
  id, parent_id, child_id, relation_type, created_at
)

-- CI变更历史表
ci_history (
  id, ci_id, changed_by, old_value, new_value, changed_at
)
```

### 4.3 工单模块

```sql
-- 工单模板表
ticket_templates (
  id, name, description, workflow_id, created_at
)

-- 工作流定义表
ticket_workflows (
  id, name, states, transitions, created_at
)

-- 工单主表
tickets (
  id, title, description, template_id, workflow_id, status, priority,
  assignee_id, requester_id, sla_deadline, created_at, updated_at
)

-- 工单评论表
ticket_comments (
  id, ticket_id, user_id, content, is_internal, created_at
)

-- 工单附件表
ticket_attachments (
  id, ticket_id, file_name, file_url, uploaded_by, created_at
)

-- 工单历史表
ticket_history (
  id, ticket_id, field_name, old_value, new_value, changed_by, changed_at
)
```

### 4.4 告警模块

```sql
-- 告警规则表
alert_rules (
  id, name, condition, severity, notification_channels, is_active
)

-- 告警实例表
alert_instances (
  id, rule_id, title, description, severity, status, affected_ci_id,
  triggered_at, acknowledged_at, acknowledged_by, closed_at
)

-- 告警历史表
alert_history (
  id, alert_id, event_type, event_data, created_at
)

-- 告警阈值表
alert_thresholds (
  id, rule_id, metric, operator, threshold, duration
)
```

---

## 5. RBAC权限模型

### 5.1 权限设计

采用 **资源:操作** 粒度的权限控制：

**资源类型：**
- `cmdb:server`, `cmdb:network`, `cmdb:application`, `cmdb:container`
- `ticket:incident`, `ticket:request`, `ticket:change`
- `alert:rule`, `alert:instance`
- `user`, `role`, `audit`

**操作类型：**
- `create`, `read`, `update`, `delete`, `approve`

**权限示例：**
- `cmdb:server:update` - 更新服务器信息
- `ticket:incident:approve` - 审批故障工单
- `alert:rule:delete` - 删除告警规则

### 5.2 预设角色

| 角色 | 权限范围 |
|------|----------|
| 管理员 | 全部权限 |
| 运维工程师 | CMDB读写、工单处理、告警确认 |
| 开发工程师 | CMDB只读、工单创建 |
| 只读用户 | 仅查看权限 |

---

## 6. Kafka事件设计

### 6.1 主题列表

```
# 工单事件
ticket.created          - 工单创建
ticket.status.changed   - 状态变更
ticket.assigned         - 工单分派
ticket.sla.breached     - SLA违约

# 告警事件
alert.triggered         - 告警触发
alert.acknowledged      - 告警确认
alert.closed            - 告警关闭
alert.escalated         - 告警升级

# CMDB事件
ci.changed              - CI变更
ci.deleted              - CI删除
ci.relationship.changed - 关系变更

# 通知事件
notification.send       - 发送通知
```

### 6.2 事件示例

**工单创建事件：**
```json
{
  "event_type": "ticket.created",
  "timestamp": "2026-01-24T10:00:00Z",
  "data": {
    "ticket_id": "TKT-001",
    "title": "服务器CPU异常",
    "priority": "high",
    "requester_id": 123
  }
}
```

---

## 7. 前端架构

### 7.1 技术栈

```
React 18 + TypeScript
Ant Design 5
Zustand (状态管理)
React Router v6
React Query
Vite
```

### 7.2 路由结构

```
/login                    - 登录页
/dashboard                - 首页仪表板
/cmdb                     - CMDB模块
  /cmdb/servers           - 服务器管理
  /cmdb/networks          - 网络设备
  /cmdb/applications      - 应用服务
  /cmdb/containers        - 容器/K8s
/tickets                  - 工单模块
  /tickets/list           - 工单列表
  /tickets/create         - 创建工单
  /tickets/:id            - 工单详情
/alerts                   - 告警模块
  /alerts/list            - 告警列表
  /alerts/rules           - 告警规则
  /alerts/history         - 告警历史
/admin                    - 管理模块
  /admin/users            - 用户管理
  /admin/roles            - 角色权限
  /admin/audit            - 审计日志
/reports                  - 报表模块
```

### 7.3 Zustand Store设计

```typescript
// stores/authStore.ts
interface AuthState {
  user: User | null
  token: string
  permissions: string[]
  login: (credentials) => Promise<void>
  logout: () => void
  hasPermission: (resource: string, action: string) => boolean
}

// stores/cmdbStore.ts
interface CMDBState {
  ciInstances: CIInstance[]
  ciTypes: CIType[]
  selectedCI: CIInstance | null
  fetchInstances: () => Promise<void>
  createInstance: (data) => Promise<void>
  updateInstance: (id, data) => Promise<void>
}

// stores/ticketStore.ts
interface TicketState {
  tickets: Ticket[]
  workflows: Workflow[]
  filters: TicketFilters
  fetchTickets: () => Promise<void>
  createTicket: (data) => Promise<void>
  updateStatus: (id, status) => Promise<void>
}

// stores/alertStore.ts
interface AlertState {
  alerts: Alert[]
  alertStats: AlertStats
  fetchAlerts: () => Promise<void>
  acknowledgeAlert: (id) => Promise<void>
}
```

---

## 8. 部署架构

### 8.1 Docker服务清单

```yaml
services:
  # 基础设施
  postgres:        # PostgreSQL数据库 :5433
  redis:           # Redis缓存 :6379
  zookeeper:       # Kafka依赖 :2181
  kafka:           # Kafka消息队列 :9092
  kafka-ui:        # Kafka管理界面 :8080

  # 后端服务
  auth-service:        # 认证服务 :5001
  cmdb-service:        # CMDB服务 :5002
  ticket-service:      # 工单服务 :5003
  alert-service:       # 告警服务 :5004
  notification-service:# 通知服务 :5005
  report-service:      # 报表服务 :5006

  # 前端与网关
  frontend:       # React前端 :80
  api-gateway:    # Kong/Traefik :8000
```

### 8.2 网络隔离

```
frontend-network (对外可访问)
  ├─ frontend (80)
  └─ api-gateway (8000)

backend-network (服务间通信)
  ├─ auth-service (5001)
  ├─ cmdb-service (5002)
  ├─ ticket-service (5003)
  ├─ alert-service (5004)
  ├─ notification-service (5005)
  └─ report-service (5006)

data-network (数据层)
  ├─ postgres (5433)
  ├─ redis (6379)
  └─ kafka (9092)
```

---

## 9. 项目目录结构

```
itcmdb/
├── frontend/                    # React前端
│   ├── src/
│   │   ├── components/         # 通用组件
│   │   ├── pages/              # 页面组件
│   │   ├── services/           # API调用
│   │   ├── stores/             # Zustand状态管理
│   │   ├── hooks/              # 自定义Hooks
│   │   ├── utils/              # 工具函数
│   │   ├── types/              # TypeScript类型
│   │   └── main.tsx            # 入口文件
│   ├── public/
│   ├── package.json
│   ├── vite.config.ts
│   └── Dockerfile
│
├── services/                    # Go微服务
│   ├── auth-service/           # 认证服务
│   │   ├── cmd/
│   │   ├── internal/
│   │   └── go.mod
│   ├── cmdb-service/           # CMDB服务
│   ├── ticket-service/         # 工单服务
│   ├── alert-service/          # 告警服务
│   ├── notification-service/   # 通知服务
│   ├── report-service/         # 报表服务
│   └── shared/                 # 共享代码
│       ├── pkg/
│       │   ├── auth/           # JWT中间件
│       │   ├── database/       # 数据库封装
│       │   ├── cache/          # Redis封装
│       │   ├── kafka/          # Kafka封装
│       │   ├── rbac/           # RBAC权限检查
│       │   ├── response/       # 统一响应格式
│       │   └── logger/         # 日志工具
│       └── proto/              # gRPC Protobuf定义
│
├── deploy/                     # 部署配置
│   ├── docker-compose.yml
│   ├── kubernetes/            # K8s配置（可选）
│   └── nginx/                 # API Gateway配置
│
├── scripts/                   # 脚本工具
│   ├── init-db.sql            # 数据库初始化
│   ├── migrate.sh             # 迁移脚本
│   └── build.sh               # 构建脚本
│
└── docs/                      # 文档
    ├── api/                   # API文档
    └── architecture.md        # 架构文档
```

---

## 10. 开发计划

### Phase 1: 基础设施搭建
- [ ] 项目目录结构创建
- [ ] Docker环境配置
- [ ] 数据库表结构设计与初始化
- [ ] 共享代码库开发（JWT、数据库、缓存等）

### Phase 2: 核心服务开发
- [ ] 认证服务
- [ ] CMDB服务
- [ ] 工单服务
- [ ] 告警服务

### Phase 3: 扩展服务开发
- [ ] 通知服务
- [ ] 报表服务

### Phase 4: 前端开发
- [ ] 基础框架搭建
- [ ] 认证页面
- [ ] CMDB模块
- [ ] 工单模块
- [ ] 告警模块
- [ ] 管理模块

### Phase 5: 集成测试与优化
- [ ] API集成测试
- [ ] 性能优化
- [ ] 安全加固

---

**文档状态:** ✅ 已完成
**下一步:** 创建项目骨架

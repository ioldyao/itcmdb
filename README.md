# ITCMDB - 运维管理平台

## 项目简介

ITCMDB 是一个一体化的运维管理平台，集成了 CMDB、工单系统和告警系统，采用微服务架构设计。

## 技术栈

### 前端
- React 18 + TypeScript
- Ant Design 5
- Zustand (状态管理)
- React Router v6
- React Query
- Vite

### 后端
- Go 1.21
- Gin (HTTP框架)
- GORM (ORM)
- JWT (认证)

### 基础设施
- PostgreSQL 15 (数据库)
- Redis 7 (缓存)
- Kafka (消息队列)
- Docker & Docker Compose (容器化部署)

## 功能模块

### 1. 用户认证服务 (auth-service:5001)
- JWT令牌签发与验证
- 用户登录/登出
- 权限管理
- RBAC权限控制

### 2. CMDB服务 (cmdb-service:5002)
- 配置项(CI)管理
- 资源类型:
  - 服务器/主机
  - 网络设备
  - 应用服务
  - 容器/K8s
- CI关系图谱
- 变更历史

### 3. 工单服务 (ticket-service:5003)
- 工单全生命周期管理
- 工作流引擎
- 工单分派
- SLA监控
- 工单评论与附件

### 4. 告警服务 (alert-service:5004)
- 告警规则引擎
- 实时告警处理
- 告警确认与关闭
- 告警升级
- 外部系统集成

### 5. 通知服务 (notification-service:5005)
- 统一通知发送
- 通知渠道:
  - 企业微信/钉钉
  - 邮件
  - 短信/语音
- 通知模板管理

### 6. 报表服务 (report-service:5006)
- CMDB资产报表
- 工单统计分析
- 告警趋势分析
- 报表导出

## 快速开始

### 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- Go 1.21+ (本地开发)
- Node.js 20+ (本地开发)

### 使用 Docker Compose 启动

```bash
# 克隆项目
git clone <repository-url>
cd itcmdb

# 启动所有服务
./scripts/start.sh

# 查看服务状态
docker compose ps

# 查看日志
docker compose logs -f
```

### 服务访问地址

| 服务 | 地址 |
|------|------|
| 前端应用 | http://localhost |
| API网关 | http://localhost:8000 |
| 认证服务 | http://localhost:5001 |
| CMDB服务 | http://localhost:5002 |
| 工单服务 | http://localhost:5003 |
| 告警服务 | http://localhost:5004 |
| 通知服务 | http://localhost:5005 |
| 报表服务 | http://localhost:5006 |
| Kafka UI | http://localhost:8080 |

### 默认账号

```
用户名: admin
密码: admin123
```

**重要**: 请在生产环境中修改默认密码！

## 项目结构

```
itcmdb/
├── frontend/                    # React前端
│   ├── src/
│   │   ├── components/         # 组件
│   │   ├── pages/              # 页面
│   │   ├── services/           # API服务
│   │   ├── stores/             # Zustand状态管理
│   │   ├── hooks/              # 自定义Hooks
│   │   ├── types/              # TypeScript类型
│   │   └── utils/              # 工具函数
│   ├── package.json
│   └── Dockerfile
│
├── services/                    # Go微服务
│   ├── auth-service/           # 认证服务
│   ├── cmdb-service/           # CMDB服务
│   ├── ticket-service/         # 工单服务
│   ├── alert-service/          # 告警服务
│   ├── notification-service/   # 通知服务
│   ├── report-service/         # 报表服务
│   └── shared/                 # 共享代码库
│       └── pkg/
│           ├── auth/           # JWT认证
│           ├── cache/          # Redis缓存
│           ├── database/       # 数据库封装
│           ├── kafka/          # Kafka封装
│           ├── logger/         # 日志工具
│           ├── rbac/           # 权限控制
│           └── response/       # 统一响应
│
├── deploy/                     # 部署配置
│   ├── docker-compose.yml
│   └── nginx/                 # Nginx配置
│
├── scripts/                    # 脚本
│   ├── start.sh               # 启动脚本
│   ├── stop.sh                # 停止脚本
│   ├── build.sh               # 构建脚本
│   └── init-db.sql            # 数据库初始化
│
└── docs/                       # 文档
    └── plans/                 # 设计文档
```

## 开发指南

### 本地开发

#### 前端开发

```bash
cd frontend
npm install
npm run dev
```

#### 后端开发

```bash
# 启动基础设施 (PostgreSQL, Redis, Kafka)
docker compose up -d postgres redis kafka zookeeper

# 启动单个服务 (例如: auth-service)
cd services/auth-service
go run cmd/main.go
```

### API文档

请参考 `docs/api/` 目录下的API文档。

### 数据库

数据库表结构定义在 `scripts/init-db.sql` 中。

## 停止服务

```bash
./scripts/stop.sh

# 或使用 docker compose
docker compose down

# 删除数据卷 (警告: 会删除所有数据)
docker compose down -v
```

## 配置说明

### 环境变量

各服务的配置文件位于 `services/<service-name>/internal/config/config.yaml`

环境变量格式: `<SERVICE>_CONFIG_KEY`

例如:
- `AUTH_DATABASE_HOST`
- `CMDB_DATABASE_HOST`
- `TICKET_JWT_SECRET`

### 权限系统

权限格式: `<resource>:<action>`

示例:
- `cmdb:server:update` - 更新服务器信息
- `ticket:incident:approve` - 审批故障工单
- `alert:rule:delete` - 删除告警规则

## 贡献指南

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

[MIT License](LICENSE)

## 联系方式

如有问题或建议，请提交 Issue。

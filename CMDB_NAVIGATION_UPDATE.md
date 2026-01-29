# CMDB 二级导航实现完成

## 完成的任务

### 1. 删除 /admin 系统的告警集成 ✅
- 已从 `/admin` 路由中删除 `alert-integration/webhook` 路径
- 告警集成已完全移至 `/alerts/integration/webhook`

### 2. 删除 /admin 系统的 VictoriaMetrics 配置 ✅
- 已从 `/admin` 路由中删除 `victoriametrics` 路径
- VictoriaMetrics 配置已移至 `/cmdb/victoriametrics`

### 3. 实现 CMDB 二级导航 ✅

**主导航**（默认显示）：
- 仪表板
- CMDB（点击后切换到子导航）
- 工单
- 告警（点击后切换到子导航）
- 报表
- 系统

**CMDB 子导航**（点击 CMDB 后显示）：
- ← 返回按钮
- CMDB（/cmdb）
- VictoriaMetrics配置（/cmdb/victoriametrics）

**告警子导航**（点击告警后显示）：
- ← 返回按钮
- 工单（/alerts）
- 配置（/alerts/rules）
- Webhook（/alerts/integration/webhook）

## 实现细节

### 1. 路由配置更新

**文件**: `frontend/src/router.tsx`

**CMDB 路由**：
```typescript
{
  path: 'cmdb',
  element: <CMDBLayout />,
  children: [
    { index: true, element: <CMDBDefaultPage /> },
    { path: 'servers', element: <CMDBServers /> },
    { path: 'networks', element: <CMDBNetworks /> },
    { path: 'applications', element: <CMDBApplications /> },
    { path: 'containers', element: <CMDBContainers /> },
    { path: 'roles', element: <CIRoles /> },
    { path: 'tags', element: <Tags /> },
    { path: 'instances/:id', element: <CIDetail /> },
    { path: 'victoriametrics', element: <VictoriaMetrics /> }, // 新增
  ],
}
```

**Admin 路由**（已清理）：
```typescript
{
  path: 'admin',
  element: <AdminLayout />,
  children: [
    { index: true, element: <AdminDefaultPage /> },
    { path: 'users', element: <AdminUsers /> },
    { path: 'roles', element: <AdminRoles /> },
    { path: 'audit', element: <AdminAudit /> },
    { path: 'alert-receivers', element: <AdminAlertReceivers /> },
    { path: 'alert-receiver-groups', element: <AdminAlertReceiverGroups /> },
    // 已删除: victoriametrics
    // 已删除: alert-integration/webhook
  ],
}
```

### 2. 导航栏组件更新

**文件**: `frontend/src/components/Layout/MainLayout.tsx`

**新增状态管理**：
```typescript
const [showSubNav, setShowSubNav] = useState(false)
const [currentSubNav, setCurrentSubNav] = useState<'cmdb' | 'alerts' | null>(null)
```

**新增菜单配置**：
```typescript
const cmdbSubMenuItems = [
  { key: '/cmdb', label: 'CMDB', icon: Server },
  { key: '/cmdb/victoriametrics', label: 'VictoriaMetrics配置', icon: Monitor },
]
```

**新增处理函数**：
```typescript
const handleCMDBClick = () => {
  setShowSubNav(true)
  setCurrentSubNav('cmdb')
  navigate('/cmdb')
}
```

**自动检测当前模块**：
```typescript
useEffect(() => {
  const isInAlerts = location.pathname.startsWith('/alerts')
  const isInCMDB = location.pathname.startsWith('/cmdb')

  if (isInAlerts) {
    setShowSubNav(true)
    setCurrentSubNav('alerts')
  } else if (isInCMDB) {
    setShowSubNav(true)
    setCurrentSubNav('cmdb')
  } else {
    setShowSubNav(false)
    setCurrentSubNav(null)
  }
}, [location.pathname])
```

### 3. 动画效果

- **主导航 → 子导航**：主导航向左淡出，子导航从右淡入
- **子导航 → 主导航**：子导航向右淡出，主导航从左淡入
- **子导航切换**：CMDB 和告警子导航之间切换时也有流畅动画
- 动画时长：300ms，使用 easeInOut 缓动函数

## 路径变更对照表

| 功能 | 旧路径 | 新路径 |
|------|--------|--------|
| VictoriaMetrics 配置 | `/admin/victoriametrics` | `/cmdb/victoriametrics` |
| 告警 Webhook | `/admin/alert-integration/webhook` | `/alerts/integration/webhook` |

## 部署状态

- ✅ 路由配置已更新
- ✅ 导航栏组件已更新
- ✅ 前端镜像已重新构建
- ✅ 前端服务已重启

## 访问路径

### CMDB 模块
- CMDB 主页：http://your-host:8082/cmdb
- VictoriaMetrics 配置：http://your-host:8082/cmdb/victoriametrics

### 告警模块
- 告警列表：http://your-host:8082/alerts
- 告警配置：http://your-host:8082/alerts/rules
- 告警 Webhook：http://your-host:8082/alerts/integration/webhook

### 系统管理
- 用户管理：http://your-host:8082/admin/users
- 角色管理：http://your-host:8082/admin/roles
- 审计日志：http://your-host:8082/admin/audit
- 告警接收人：http://your-host:8082/admin/alert-receivers
- 告警接收组：http://your-host:8082/admin/alert-receiver-groups

## 用户体验

1. **点击 CMDB**：自动切换到 CMDB 子导航，显示 CMDB 和 VictoriaMetrics 配置
2. **点击告警**：自动切换到告警子导航，显示工单、配置、Webhook
3. **点击返回按钮**：返回主导航并跳转到仪表板
4. **在模块内切换页面**：子导航保持显示
5. **离开模块**：自动切换回主导航
6. **所有切换都有流畅的淡入淡出动画**

## 技术实现

- **React Router** - 路由管理
- **Framer Motion** - 动画效果
- **AnimatePresence** - 组件进出场动画
- **Lucide React** - 图标库
- **TypeScript** - 类型安全
- **Tailwind CSS** - 样式

## 注意事项

1. VictoriaMetrics 配置已从系统管理移到 CMDB 模块
2. 告警集成已从系统管理移到告警模块
3. 原路径已失效，请使用新路径
4. 子导航会根据当前路由自动显示/隐藏
5. 支持多个模块的二级导航（CMDB 和告警）

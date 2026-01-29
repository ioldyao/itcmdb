# 多层级导航栏实现完成

## 功能说明

实现了告警模块的二级导航栏，点击"告警"后会切换到子导航，并且有淡入淡出动画效果。

## 实现内容

### 1. 导航栏结构

**主导航**（默认显示）：
- 仪表板
- CMDB
- 工单
- 告警（点击后切换到子导航）
- 报表
- 系统

**子导航**（点击告警后显示）：
- ← 返回按钮
- 工单（/alerts）
- 配置（/alerts/rules）
- Webhook（/alerts/integration/webhook）

### 2. 动画效果

- **主导航 → 子导航**：主导航向左淡出（opacity + translateX），子导航从右淡入
- **子导航 → 主导航**：子导航向右淡出，主导航从左淡入
- 动画时长：300ms，使用 easeInOut 缓动函数

### 3. 路由调整

**移动告警集成**：
- 从：`/admin/alert-integration/webhook`
- 到：`/alerts/integration/webhook`

**新增路由**：
```typescript
{
  path: 'alerts',
  children: [
    { index: true, element: <AlertList /> },
    { path: ':id', element: <AlertDetail /> },
    { path: 'rules', element: <AlertRules /> },
    { path: 'history', element: <AlertHistory /> },
    { path: 'integration/webhook', element: <AlertIntegrationWebhook /> },
  ],
}
```

### 4. 代码修改

**文件**：`frontend/src/components/Layout/MainLayout.tsx`

**主要改动**：
1. 添加 `useState` 管理子导航显示状态
2. 添加 `useEffect` 监听路由变化自动切换导航
3. 使用 `AnimatePresence` 和 `motion.div` 实现动画
4. 添加返回按钮处理函数
5. 新增子导航菜单项配置

**新增图标**：
- `ArrowLeft` - 返回按钮
- `Webhook` - Webhook 菜单项
- `SlidersHorizontal` - 配置菜单项

### 5. 用户体验

- 点击"告警"菜单项自动切换到子导航并跳转到告警列表页
- 点击返回按钮返回主导航并跳转到仪表板
- 在告警模块内切换页面时，子导航保持显示
- 离开告警模块时，自动切换回主导航
- 所有切换都有流畅的淡入淡出动画

## 部署状态

- ✅ 前端代码已修改
- ✅ 路由配置已更新
- ✅ 前端镜像已重新构建
- ✅ 前端服务已重启

## 访问路径

- 告警列表：http://your-host:8082/alerts
- 告警配置：http://your-host:8082/alerts/rules
- 告警 Webhook：http://your-host:8082/alerts/integration/webhook

## 技术实现

使用了以下技术：
- **React Router** - 路由管理
- **Framer Motion** - 动画效果
- **AnimatePresence** - 组件进出场动画
- **Lucide React** - 图标库
- **Tailwind CSS** - 样式

## 注意事项

1. 告警集成已从系统管理模块移到告警模块
2. 原路径 `/admin/alert-integration/webhook` 已失效
3. 新路径为 `/alerts/integration/webhook`
4. 子导航会根据当前路由自动显示/隐藏

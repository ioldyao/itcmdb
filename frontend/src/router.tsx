import { createBrowserRouter, Navigate } from 'react-router-dom'
import MainLayout from '@/components/Layout/MainLayout'
import ProtectedRoute from '@/components/ProtectedRoute'
import PermissionGuard from '@/components/PermissionGuard'

// Direct imports instead of lazy loading
import Login from '@/pages/Login'
import Dashboard from '@/pages/Dashboard'
import Profile from '@/pages/Profile'

// CMDB Pages
import CMDBLayout from '@/pages/CMDB'
import CMDBDefaultPage from '@/pages/CMDB/DefaultPage'
import CMDBServers from '@/pages/CMDB/Servers'
import CMDBNetworks from '@/pages/CMDB/Networks'
import CMDBApplications from '@/pages/CMDB/Applications'
import CMDBContainers from '@/pages/CMDB/Containers'
import CIRoles from '@/pages/CMDB/Roles'
import Tags from '@/pages/CMDB/Tags'
import CIDetail from '@/pages/CMDB/CIDetail'

// Ticket Pages
import TicketList from '@/pages/Ticket/List'
import TicketCreate from '@/pages/Ticket/Create'
import TicketDetail from '@/pages/Ticket/Detail'

// Alert Pages
import AlertList from '@/pages/Alert/List'
import AlertDetail from '@/pages/Alert/Detail'
import AlertRules from '@/pages/Alert/Rules'
import AlertHistory from '@/pages/Alert/History'
import AlertReceivers from '@/pages/Alerts/AlertReceivers'

// Admin Pages
import AdminLayout from '@/pages/Admin'
import AdminUsers from '@/pages/Admin/Users'
import AdminRoles from '@/pages/Admin/Roles'
import AdminAudit from '@/pages/Admin/Audit'
import AdminDefaultPage from '@/pages/Admin/DefaultPage'
import AlertIntegrationWebhook from '@/pages/Admin/AlertIntegration/Webhook'
import VictoriaMetrics from '@/pages/Monitoring/VictoriaMetrics'

const router = createBrowserRouter([
  {
    path: '/login',
    element: <Login />,
  },
  {
    path: '/',
    element: (
      <ProtectedRoute>
        <MainLayout />
      </ProtectedRoute>
    ),
    children: [
      { index: true, element: <Navigate to="/dashboard" replace /> },
      { path: 'dashboard', element: <Dashboard /> },
      { path: 'profile', element: <Profile /> },
      {
        path: 'cmdb',
        element: <CMDBLayout />,
        children: [
          { index: true, element: <PermissionGuard resource="ci" action="view"><CMDBDefaultPage /></PermissionGuard> },
          { path: 'servers', element: <PermissionGuard resource="ci" action="view"><CMDBServers /></PermissionGuard> },
          { path: 'networks', element: <PermissionGuard resource="ci" action="view"><CMDBNetworks /></PermissionGuard> },
          { path: 'applications', element: <PermissionGuard resource="ci" action="view"><CMDBApplications /></PermissionGuard> },
          { path: 'containers', element: <PermissionGuard resource="ci" action="view"><CMDBContainers /></PermissionGuard> },
          { path: 'roles', element: <PermissionGuard resource="ci" action="manage"><CIRoles /></PermissionGuard> },
          { path: 'tags', element: <PermissionGuard resource="ci" action="view"><Tags /></PermissionGuard> },
          { path: 'instances/:id', element: <PermissionGuard resource="ci" action="view"><CIDetail /></PermissionGuard> },
          { path: 'victoriametrics', element: <PermissionGuard resource="monitoring" action="view"><VictoriaMetrics /></PermissionGuard> },
        ],
      },
      {
        path: 'tickets',
        children: [
          { index: true, element: <PermissionGuard resource="ticket" action="view"><TicketList /></PermissionGuard> },
          { path: 'create', element: <PermissionGuard resource="ticket" action="create"><TicketCreate /></PermissionGuard> },
          { path: ':id', element: <PermissionGuard resource="ticket" action="view"><TicketDetail /></PermissionGuard> },
        ],
      },
      {
        path: 'alerts',
        children: [
          { index: true, element: <PermissionGuard resource="alert" action="view"><AlertList /></PermissionGuard> },
          { path: ':id', element: <PermissionGuard resource="alert" action="view"><AlertDetail /></PermissionGuard> },
          { path: 'rules', element: <PermissionGuard resource="alert_rule" action="view"><AlertRules /></PermissionGuard> },
          { path: 'history', element: <PermissionGuard resource="alert" action="view"><AlertHistory /></PermissionGuard> },
          { path: 'integration/webhook', element: <PermissionGuard resource="webhook" action="view"><AlertIntegrationWebhook /></PermissionGuard> },
          { path: 'receivers', element: <PermissionGuard resource="alert_receiver" action="view"><AlertReceivers /></PermissionGuard> },
        ],
      },
      {
        path: 'admin',
        element: <AdminLayout />,
        children: [
          { index: true, element: <AdminDefaultPage /> },
          {
            path: 'users',
            element: (
              <PermissionGuard resource="user" action="view">
                <AdminUsers />
              </PermissionGuard>
            )
          },
          {
            path: 'roles',
            element: (
              <PermissionGuard resource="role" action="view">
                <AdminRoles />
              </PermissionGuard>
            )
          },
          { path: 'audit', element: <PermissionGuard resource="audit" action="view"><AdminAudit /></PermissionGuard> },
        ],
      },
    ],
  },
])

export default router

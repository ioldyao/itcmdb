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
          { index: true, element: <CMDBDefaultPage /> },
          { path: 'servers', element: <CMDBServers /> },
          { path: 'networks', element: <CMDBNetworks /> },
          { path: 'applications', element: <CMDBApplications /> },
          { path: 'containers', element: <CMDBContainers /> },
          { path: 'roles', element: <CIRoles /> },
          { path: 'tags', element: <Tags /> },
          { path: 'instances/:id', element: <CIDetail /> },
          { path: 'victoriametrics', element: <VictoriaMetrics /> },
        ],
      },
      {
        path: 'tickets',
        children: [
          { index: true, element: <TicketList /> },
          { path: 'create', element: <TicketCreate /> },
          { path: ':id', element: <TicketDetail /> },
        ],
      },
      {
        path: 'alerts',
        children: [
          { index: true, element: <AlertList /> },
          { path: ':id', element: <AlertDetail /> },
          { path: 'rules', element: <AlertRules /> },
          { path: 'history', element: <AlertHistory /> },
          { path: 'integration/webhook', element: <AlertIntegrationWebhook /> },
          { path: 'receivers', element: <AlertReceivers /> },
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
          { path: 'audit', element: <AdminAudit /> },
        ],
      },
    ],
  },
])

export default router

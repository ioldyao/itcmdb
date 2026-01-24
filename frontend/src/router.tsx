import { createBrowserRouter, Navigate } from 'react-router-dom'
import MainLayout from '@/components/Layout/MainLayout'
import ProtectedRoute from '@/components/ProtectedRoute'

// Direct imports instead of lazy loading
import Login from '@/pages/Login'
import Dashboard from '@/pages/Dashboard'
import Profile from '@/pages/Profile'

// CMDB Pages
import CMDBLayout from '@/pages/CMDB'
import CMDBServers from '@/pages/CMDB/Servers'
import CMDBNetworks from '@/pages/CMDB/Networks'
import CMDBApplications from '@/pages/CMDB/Applications'
import CMDBContainers from '@/pages/CMDB/Containers'
import CIRoles from '@/pages/CMDB/Roles'
import Tags from '@/pages/CMDB/Tags'

// Ticket Pages
import TicketList from '@/pages/Ticket/List'
import TicketCreate from '@/pages/Ticket/Create'
import TicketDetail from '@/pages/Ticket/Detail'

// Alert Pages
import AlertList from '@/pages/Alert/List'
import AlertRules from '@/pages/Alert/Rules'
import AlertHistory from '@/pages/Alert/History'

// Admin Pages
import AdminUsers from '@/pages/Admin/Users'
import AdminRoles from '@/pages/Admin/Roles'
import AdminAudit from '@/pages/Admin/Audit'

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
          { index: true, element: <Navigate to="/cmdb/servers" replace /> },
          { path: 'servers', element: <CMDBServers /> },
          { path: 'networks', element: <CMDBNetworks /> },
          { path: 'applications', element: <CMDBApplications /> },
          { path: 'containers', element: <CMDBContainers /> },
          { path: 'roles', element: <CIRoles /> },
          { path: 'tags', element: <Tags /> },
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
          { path: 'rules', element: <AlertRules /> },
          { path: 'history', element: <AlertHistory /> },
        ],
      },
      {
        path: 'admin',
        children: [
          { index: true, element: <Navigate to="/admin/users" replace /> },
          { path: 'users', element: <AdminUsers /> },
          { path: 'roles', element: <AdminRoles /> },
          { path: 'audit', element: <AdminAudit /> },
        ],
      },
    ],
  },
])

export default router

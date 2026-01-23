import { createBrowserRouter, Navigate } from 'react-router-dom'
import { lazy } from 'react'
import MainLayout from '@/components/Layout/MainLayout'
import ProtectedRoute from '@/components/ProtectedRoute'

// Lazy load pages
const Login = lazy(() => import('@/pages/Login'))
const Dashboard = lazy(() => import('@/pages/Dashboard'))
const Profile = lazy(() => import('@/pages/Profile'))

// CMDB Pages
const CMDBServers = lazy(() => import('@/pages/CMDB/Servers'))
const CMDBNetworks = lazy(() => import('@/pages/CMDB/Networks'))
const CMDBApplications = lazy(() => import('@/pages/CMDB/Applications'))
const CMDBContainers = lazy(() => import('@/pages/CMDB/Containers'))

// Ticket Pages
const TicketList = lazy(() => import('@/pages/Ticket/List'))
const TicketCreate = lazy(() => import('@/pages/Ticket/Create'))
const TicketDetail = lazy(() => import('@/pages/Ticket/Detail'))

// Alert Pages
const AlertList = lazy(() => import('@/pages/Alert/List'))
const AlertRules = lazy(() => import('@/pages/Alert/Rules'))
const AlertHistory = lazy(() => import('@/pages/Alert/History'))

// Admin Pages
const AdminUsers = lazy(() => import('@/pages/Admin/Users'))
const AdminRoles = lazy(() => import('@/pages/Admin/Roles'))
const AdminAudit = lazy(() => import('@/pages/Admin/Audit'))

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
        children: [
          { index: true, element: <Navigate to="/cmdb/servers" replace /> },
          { path: 'servers', element: <CMDBServers /> },
          { path: 'networks', element: <CMDBNetworks /> },
          { path: 'applications', element: <CMDBApplications /> },
          { path: 'containers', element: <CMDBContainers /> },
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

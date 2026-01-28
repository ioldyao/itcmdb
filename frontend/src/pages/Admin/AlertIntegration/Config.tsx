import { Routes, Route, Navigate } from 'react-router-dom'

export default function AlertIntegrationConfig() {
  return (
    <Routes>
      <Route index element={<Navigate to="receivers" replace />} />
      <Route path="receivers" element={<AlertReceivers />} />
      <Route path="groups" element={<AlertReceiverGroups />} />
    </Routes>
  )
}

// 从父目录导入
import AlertReceivers from '../AlertReceivers'
import AlertReceiverGroups from '../AlertReceiverGroups'

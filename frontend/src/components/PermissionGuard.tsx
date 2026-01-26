import { ReactNode } from 'react'
import { Result, Button } from 'antd'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'

interface PermissionGuardProps {
  children: ReactNode
  resource: string
  action: string
}

export default function PermissionGuard({ children, resource, action }: PermissionGuardProps) {
  const navigate = useNavigate()
  const hasPermission = useAuthStore((state) => state.hasPermission)

  if (!hasPermission(resource, action)) {
    return (
      <Result
        status="403"
        title="403"
        subTitle="抱歉，您没有权限访问此页面。"
        extra={
          <Button type="primary" onClick={() => navigate('/admin')}>
            返回系统管理
          </Button>
        }
      />
    )
  }

  return <>{children}</>
}

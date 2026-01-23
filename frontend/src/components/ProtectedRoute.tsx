import { Navigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'
import { Result, Button } from 'antd'

interface ProtectedRouteProps {
  children: React.ReactNode
  requiredPermission?: string
}

export default function ProtectedRoute({ children, requiredPermission }: ProtectedRouteProps) {
  const { isAuthenticated, token, hasPermission } = useAuthStore()

  // 检查是否已登录
  if (!isAuthenticated || !token) {
    return <Navigate to="/login" replace />
  }

  // 检查权限（如果需要）
  if (requiredPermission) {
    const [resource, action] = requiredPermission.split(':')
    if (!hasPermission(resource, action)) {
      return (
        <Result
          status="403"
          title="403"
          subTitle="抱歉，您没有权限访问此页面。"
          extra={
            <Button type="primary" onClick={() => window.history.back()}>
              返回上一页
            </Button>
          }
        />
      )
    }
  }

  return <>{children}</>
}

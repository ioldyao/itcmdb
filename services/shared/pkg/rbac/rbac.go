package rbac

import (
	"github.com/gin-gonic/gin"
	"github.com/itcmdb/shared/pkg/auth"
)

// Permission 权限检查函数类型
type PermissionFunc func(c *gin.Context) bool

// RequirePermission 权限中间件
func RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := auth.GetRoles(c)
		if !exists {
			c.JSON(403, gin.H{"code": 403, "message": "forbidden: no roles found"})
			c.Abort()
			return
		}

		// TODO: 从数据库或缓存中查询用户权限
		// 这里简化处理，实际应该查询用户的所有权限
		for _, role := range roles {
			if role == "admin" {
				c.Next()
				return
			}
		}

		c.JSON(403, gin.H{"code": 403, "message": "forbidden: insufficient permissions"})
		c.Abort()
	}
}

// RequireRole 角色中间件
func RequireRole(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, exists := auth.GetRoles(c)
		if !exists {
			c.JSON(403, gin.H{"code": 403, "message": "forbidden: no roles found"})
			c.Abort()
			return
		}

		for _, required := range requiredRoles {
			for _, role := range roles {
				if role == required {
					c.Next()
					return
				}
			}
		}

		c.JSON(403, gin.H{"code": 403, "message": "forbidden: insufficient role"})
		c.Abort()
	}
}

// HasPermission 检查是否有权限
func HasPermission(c *gin.Context, resource, action string) bool {
	roles, exists := auth.GetRoles(c)
	if !exists {
		return false
	}

	for _, role := range roles {
		if role == "admin" {
			return true
		}
	}

	return false
}

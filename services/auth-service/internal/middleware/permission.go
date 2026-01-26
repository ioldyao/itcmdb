package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/itcmdb/auth-service/internal/service"
	"github.com/itcmdb/shared/pkg/auth"
	"github.com/itcmdb/shared/pkg/logger"
	"github.com/itcmdb/shared/pkg/response"
	"go.uber.org/zap"
)

// RequirePermission 创建一个权限检查中间件
// resource: 资源类型 (如 "user", "role", "permission")
// action: 操作类型 (如 "create", "update", "delete", "view")
func RequirePermission(userService service.UserService, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户ID（由 AuthMiddleware 设置）
		userID, exists := auth.GetUserID(c)
		if !exists {
			logger.Warn("User ID not found in context")
			c.JSON(401, response.Error("Unauthorized", "authentication required"))
			c.Abort()
			return
		}

		// 检查用户权限
		allowed, err := userService.CheckPermission(uint(userID), resource, action)
		if err != nil {
			logger.Error("Failed to check permission",
				zap.Error(err),
				zap.Int64("user_id", userID),
				zap.String("resource", resource),
				zap.String("action", action))
			c.JSON(500, response.Error("Internal Error", "failed to check permission"))
			c.Abort()
			return
		}

		if !allowed {
			logger.Warn("Permission denied",
				zap.Int64("user_id", userID),
				zap.String("resource", resource),
				zap.String("action", action))
			c.JSON(403, response.Error("Forbidden", "insufficient permissions"))
			c.Abort()
			return
		}

		// 权限检查通过，继续处理请求
		c.Next()
	}
}

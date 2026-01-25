package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	grpcclient "github.com/itcmdb/shared/pkg/grpc"
	"github.com/itcmdb/shared/pkg/logger"
	"github.com/itcmdb/shared/pkg/response"
	"go.uber.org/zap"
)

// GRPCAuthMiddleware 基于gRPC的认证中间件
func GRPCAuthMiddleware(authClient *grpcclient.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, response.Error("未提供认证token", ""))
			c.Abort()
			return
		}

		// 解析Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(401, response.Error("token格式错误", ""))
			c.Abort()
			return
		}

		token := parts[1]

		// 通过gRPC调用Auth服务验证token
		resp, err := authClient.ValidateToken(c.Request.Context(), token)
		if err != nil {
			logger.Error("Failed to validate token via gRPC", zap.Error(err))
			c.JSON(401, response.Error("token验证失败", err.Error()))
			c.Abort()
			return
		}

		if !resp.Valid {
			c.JSON(401, response.Error("token无效或已过期", resp.Error))
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", resp.UserId)
		c.Set("username", resp.Username)

		c.Next()
	}
}

// GRPCPermissionMiddleware 基于gRPC的权限检查中间件
func GRPCPermissionMiddleware(authClient *grpcclient.AuthClient, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(403, response.Error("未找到用户信息", ""))
			c.Abort()
			return
		}

		// 通过gRPC调用Auth服务检查权限
		resp, err := authClient.CheckPermission(c.Request.Context(), userID.(uint64), resource, action)
		if err != nil {
			logger.Error("Failed to check permission via gRPC", zap.Error(err))
			c.JSON(403, response.Error("权限检查失败", err.Error()))
			c.Abort()
			return
		}

		if !resp.Allowed {
			c.JSON(403, response.Error("没有权限执行此操作", ""))
			c.Abort()
			return
		}

		c.Next()
	}
}

// GRPCAdminOnlyMiddleware 管理员权限检查中间件
func GRPCAdminOnlyMiddleware(authClient *grpcclient.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(403, response.Error("未找到用户信息", ""))
			c.Abort()
			return
		}

		// 通过gRPC调用Auth服务检查是否为管理员
		resp, err := authClient.CheckPermission(c.Request.Context(), userID.(uint64), "system", "admin")
		if err != nil {
			logger.Error("Failed to check admin permission via gRPC", zap.Error(err))
			c.JSON(403, response.Error("权限检查失败", err.Error()))
			c.Abort()
			return
		}

		if !resp.Allowed {
			c.JSON(403, response.Error("需要管理员权限", ""))
			c.Abort()
			return
		}

		c.Next()
	}
}

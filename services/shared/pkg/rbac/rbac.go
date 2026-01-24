package rbac

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/itcmdb/shared/pkg/auth"
	"github.com/itcmdb/shared/pkg/cache"
	"github.com/itcmdb/shared/pkg/database"
	"github.com/itcmdb/shared/pkg/logger"
	"go.uber.org/zap"
)

// Permission 权限检查函数类型
type PermissionFunc func(c *gin.Context) bool

const (
	// PermissionCacheTTL 权限缓存过期时间
	PermissionCacheTTL = 5 * time.Minute
)

// GetUserPermissionsWithCache 从缓存或数据库获取用户权限
func GetUserPermissionsWithCache(ctx context.Context, userID int64) ([]string, error) {
	redisClient := cache.Get()
	cacheKey := fmt.Sprintf("user:permissions:%d", userID)

	// 尝试从Redis获取缓存
	if redisClient != nil {
		cachedData, err := redisClient.Get(ctx, cacheKey).Result()
		if err == nil && cachedData != "" {
			var permissions []string
			if err := json.Unmarshal([]byte(cachedData), &permissions); err == nil {
				logger.Debug("Permissions loaded from cache", zap.Int64("user_id", userID))
				return permissions, nil
			}
		}
	}

	// 缓存未命中，从数据库查询
	db := database.Get()
	if db == nil {
		return nil, fmt.Errorf("database not available")
	}

	// 通过角色获取的权限
	var rolePermissions []string
	err := db.Table("permissions").
		Joins("JOIN role_permissions ON permissions.id = role_permissions.permission_id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Pluck("CONCAT(resource, ':', action)", &rolePermissions).
		Error

	if err != nil {
		logger.Error("Failed to query role permissions", zap.Error(err), zap.Int64("user_id", userID))
		return nil, err
	}

	// 直接分配给用户的权限
	var userPermissions []string
	err = db.Table("permissions").
		Joins("JOIN user_permissions ON permissions.id = user_permissions.permission_id").
		Where("user_permissions.user_id = ?", userID).
		Pluck("CONCAT(resource, ':', action)", &userPermissions).
		Error

	if err != nil {
		logger.Error("Failed to query user permissions", zap.Error(err), zap.Int64("user_id", userID))
		return nil, err
	}

	// 合并权限并去重
	permissionMap := make(map[string]bool)
	for _, p := range rolePermissions {
		permissionMap[p] = true
	}
	for _, p := range userPermissions {
		permissionMap[p] = true
	}

	// 转换为切片
	permissions := make([]string, 0, len(permissionMap))
	for p := range permissionMap {
		permissions = append(permissions, p)
	}

	// 存入Redis缓存
	if redisClient != nil {
		permissionsJSON, err := json.Marshal(permissions)
		if err == nil {
			if err := redisClient.Set(ctx, cacheKey, permissionsJSON, PermissionCacheTTL).Err(); err != nil {
				logger.Warn("Failed to cache permissions", zap.Error(err), zap.Int64("user_id", userID))
			} else {
				logger.Debug("Permissions cached", zap.Int64("user_id", userID), zap.Int("count", len(permissions)))
			}
		}
	}

	return permissions, nil
}

// ClearUserPermissionsCache 清除用户权限缓存
func ClearUserPermissionsCache(ctx context.Context, userID int64) error {
	redisClient := cache.Get()
	if redisClient == nil {
		logger.Warn("Redis client not available, cannot clear permission cache")
		return nil
	}

	cacheKey := fmt.Sprintf("user:permissions:%d", userID)
	err := redisClient.Del(ctx, cacheKey).Err()
	if err != nil {
		logger.Error("Failed to clear permission cache", zap.Error(err), zap.Int64("user_id", userID))
		return err
	}

	logger.Info("Permission cache cleared", zap.Int64("user_id", userID))
	return nil
}

// RequirePermission 权限中间件
func RequirePermission(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := auth.GetUserID(c)
		if !exists {
			c.JSON(403, gin.H{"code": 403, "message": "forbidden: no user context"})
			c.Abort()
			return
		}

		// 获取用户权限（从缓存或数据库）
		permissions, err := GetUserPermissionsWithCache(c.Request.Context(), userID)
		if err != nil {
			logger.Error("Failed to get user permissions", zap.Error(err), zap.Int64("user_id", userID))
			c.JSON(500, gin.H{"code": 500, "message": "internal server error"})
			c.Abort()
			return
		}

		// 检查是否有所需权限
		requiredPerm := resource + ":" + action
		hasPermission := false

		for _, perm := range permissions {
			// 精确匹配
			if perm == requiredPerm {
				hasPermission = true
				break
			}
			// 资源通配符匹配 (例如 "cmdb:*")
			if perm == resource+":*" {
				hasPermission = true
				break
			}
			// 超级管理员通配符 (例如 "*:*")
			if perm == "*:*" {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			logger.Warn("Permission denied",
				zap.Int64("user_id", userID),
				zap.String("required_permission", requiredPerm))
			c.JSON(403, gin.H{
				"code":    403,
				"message": fmt.Sprintf("forbidden: missing permission %s", requiredPerm),
			})
			c.Abort()
			return
		}

		c.Next()
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
	userID, exists := auth.GetUserID(c)
	if !exists {
		return false
	}

	permissions, err := GetUserPermissionsWithCache(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get user permissions", zap.Error(err), zap.Int64("user_id", userID))
		return false
	}

	requiredPerm := resource + ":" + action
	for _, perm := range permissions {
		if perm == requiredPerm || perm == resource+":*" || perm == "*:*" {
			return true
		}
	}

	return false
}

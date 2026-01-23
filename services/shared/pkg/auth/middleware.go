package auth

import (
	"strings"
	"github.com/gin-gonic/gin"
)

const (
	AuthorizationHeader = "Authorization"
	BearerPrefix        = "Bearer "
	UserIDKey           = "user_id"
	UsernameKey         = "username"
	RolesKey            = "roles"
)

// AuthMiddleware JWT认证中间件
func (m *JWTManager) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(AuthorizationHeader)
		if authHeader == "" {
			c.JSON(401, gin.H{"code": 401, "message": "missing authorization header"})
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, BearerPrefix) {
			c.JSON(401, gin.H{"code": 401, "message": "invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, BearerPrefix)
		claims, err := m.Verify(tokenString)
		if err != nil {
			c.JSON(401, gin.H{"code": 401, "message": err.Error()})
			c.Abort()
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(UsernameKey, claims.Username)
		c.Set(RolesKey, claims.Roles)
		c.Next()
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return 0, false
	}
	return userID.(int64), true
}

// GetUsername 从上下文获取用户名
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get(UsernameKey)
	if !exists {
		return "", false
	}
	return username.(string), true
}

// GetRoles 从上下文获取角色
func GetRoles(c *gin.Context) ([]string, bool) {
	roles, exists := c.Get(RolesKey)
	if !exists {
		return nil, false
	}
	return roles.([]string), true
}

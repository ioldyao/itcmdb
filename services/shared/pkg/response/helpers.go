package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetUserID 从 gin context 安全提取用户 ID (uint64)
// 适配 gRPC auth middleware 和本地 JWT middleware 两种来源
func GetUserID(c *gin.Context) (uint64, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, Error("Unauthorized", "missing user context"))
		return 0, false
	}

	switch v := userID.(type) {
	case uint64:
		return v, true
	case int64:
		return uint64(v), true
	case uint:
		return uint64(v), true
	case int:
		return uint64(v), true
	default:
		c.JSON(http.StatusUnauthorized, Error("Unauthorized", "invalid user context"))
		return 0, false
	}
}

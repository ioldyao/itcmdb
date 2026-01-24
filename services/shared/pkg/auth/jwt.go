package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/itcmdb/shared/pkg/cache"
	"github.com/itcmdb/shared/pkg/logger"
	"go.uber.org/zap"
)

var (
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenInvalid     = errors.New("token invalid")
	ErrTokenMalformed   = errors.New("token malformed")
	ErrTokenBlacklisted = errors.New("token has been revoked")
)

type Claims struct {
	UserID   int64    `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secretKey     []byte
	tokenDuration time.Duration
}

func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
	}
}

// Generate 生成JWT token
func (m *JWTManager) Generate(userID int64, username string, roles []string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Roles:    roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

// Verify 验证JWT token
func (m *JWTManager) Verify(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrTokenInvalid
		}
		return m.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		if errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, ErrTokenMalformed
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// Refresh 刷新token
func (m *JWTManager) Refresh(tokenString string) (string, error) {
	claims, err := m.Verify(tokenString)
	if err != nil {
		return "", err
	}

	// 检查是否在刷新窗口期内（例如，剩余时间小于1小时）
	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return "", errors.New("token not ready for refresh")
	}

	return m.Generate(claims.UserID, claims.Username, claims.Roles)
}

// hashToken 生成token的SHA256哈希值
func (m *JWTManager) hashToken(tokenString string) string {
	hash := sha256.Sum256([]byte(tokenString))
	return hex.EncodeToString(hash[:])
}

// AddToBlacklist 将token添加到黑名单
func (m *JWTManager) AddToBlacklist(ctx context.Context, tokenString string, claims *Claims) error {
	redisClient := cache.Get()
	if redisClient == nil {
		logger.Warn("Redis client not available, cannot blacklist token")
		return errors.New("redis not available")
	}

	// 使用token哈希作为key，提高安全性
	tokenHash := m.hashToken(tokenString)
	key := "token:blacklist:" + tokenHash

	// 计算TTL：token剩余有效时间
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		// Token已过期，无需加入黑名单
		return nil
	}

	// 将token哈希存入Redis，值为用户ID（用于审计）
	err := redisClient.Set(ctx, key, claims.UserID, ttl).Err()
	if err != nil {
		logger.Error("Failed to add token to blacklist", zap.Error(err), zap.String("key", key))
		return err
	}

	logger.Info("Token added to blacklist",
		zap.Int64("user_id", claims.UserID),
		zap.String("username", claims.Username),
		zap.Duration("ttl", ttl))
	return nil
}

// IsBlacklisted 检查token是否在黑名单中
func (m *JWTManager) IsBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	redisClient := cache.Get()
	if redisClient == nil {
		// Redis不可用时，采用fail-safe策略：允许访问但记录警告
		logger.Warn("Redis client not available, cannot check token blacklist - allowing access")
		return false, nil
	}

	tokenHash := m.hashToken(tokenString)
	key := "token:blacklist:" + tokenHash

	// 检查key是否存在
	exists, err := redisClient.Exists(ctx, key).Result()
	if err != nil {
		// Redis查询失败，采用fail-safe策略：允许访问但记录错误
		logger.Error("Failed to check token blacklist", zap.Error(err), zap.String("key", key))
		return false, nil
	}

	return exists > 0, nil
}

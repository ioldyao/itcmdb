package service

import (
	"errors"
	"time"

	"github.com/itcmdb/auth-service/internal/models"
	"github.com/itcmdb/shared/pkg/auth"
)

type AuthService interface {
	ValidateToken(token string) (*auth.Claims, error)
	GenerateToken(user *models.User) (string, error)
}

type authService struct {
	jwtManager *auth.JWTManager
}

func NewAuthService(jwtManager *auth.JWTManager) AuthService {
	return &authService{
		jwtManager: jwtManager,
	}
}

// ValidateToken 验证JWT令牌
func (s *authService) ValidateToken(token string) (*auth.Claims, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// GenerateToken 生成JWT令牌
func (s *authService) GenerateToken(user *models.User) (string, error) {
	claims := &auth.Claims{
		UserID:   user.ID,
		Username: user.Username,
		Email:    user.Email,
	}

	token, err := s.jwtManager.GenerateToken(claims)
	if err != nil {
		return "", err
	}

	return token, nil
}

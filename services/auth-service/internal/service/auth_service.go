package service

import (
	"errors"

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
	claims, err := s.jwtManager.Verify(token)
	if err != nil {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

// GenerateToken 生成JWT令牌
func (s *authService) GenerateToken(user *models.User) (string, error) {
	token, err := s.jwtManager.Generate(int64(user.ID), user.Username, []string{})
	if err != nil {
		return "", err
	}

	return token, nil
}

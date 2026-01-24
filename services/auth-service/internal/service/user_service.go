package service

import (
	"errors"

	"github.com/itcmdb/auth-service/internal/models"
	"github.com/itcmdb/auth-service/internal/repository"
	"github.com/itcmdb/shared/pkg/database"
	"github.com/itcmdb/shared/pkg/logger"
	passwordpkg "github.com/itcmdb/shared/pkg/password"
	"go.uber.org/zap"
)

type UserService interface {
	Register(username, email, password, fullName string) (*models.User, error)
	ValidateUser(username, password string) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	GetAllUsers() ([]models.User, error)
	UpdateUser(id uint, updates map[string]interface{}) error
	DeleteUser(id uint) error
	GetUserPermissions(userID uint) ([]string, error)
	CheckPermission(userID uint, resource, action string) (bool, error)
}

type userService struct {
	repo repository.UserRepository
}

// NewUserService 创建用户服务实例
func NewUserService() UserService {
	return &userService{
		repo: repository.NewUserRepository(),
	}
}

// Register 用户注册
func (s *userService) Register(username, email, password, fullName string) (*models.User, error) {
	// 验证密码强度
	if err := passwordpkg.ValidatePassword(password); err != nil {
		return nil, err
	}

	// 检查用户名是否已存在
	if _, err := s.repo.FindByUsername(username); err == nil {
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	db := database.Get()
	var existingUser models.User
	if err := db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil, errors.New("email already exists")
	}

	// 哈希密码
	hashedPassword, err := passwordpkg.HashPassword(password)
	if err != nil {
		logger.Error("Failed to hash password", zap.Error(err))
		return nil, err
	}

	// 创建用户
	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		FullName:     fullName,
		Status:       "active",
	}

	// 分配默认角色（viewer）
	if err := s.repo.Create(user); err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		return nil, err
	}

	logger.Info("User registered successfully", zap.String("username", username))

	return user, nil
}

// ValidateUser 验证用户登录
func (s *userService) ValidateUser(username, password string) (*models.User, error) {
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// 验证密码
	logger.Info("Validating password", zap.String("username", username), zap.Int("pwd_len", len(password)), zap.Int("hash_len", len(user.PasswordHash)))
	if err := passwordpkg.CheckPassword(password, user.PasswordHash); err != nil {
		logger.Warn("Password validation failed", zap.String("username", username), zap.Error(err))
		return nil, errors.New("invalid password")
	}

	if user.Status != "active" {
		return nil, errors.New("user account is disabled")
	}

	logger.Info("User validated successfully", zap.String("username", username))

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *userService) GetUserByID(id uint) (*models.User, error) {
	return s.repo.FindByID(id)
}

// UpdateUser 更新用户信息
func (s *userService) UpdateUser(id uint, updates map[string]interface{}) error {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	// 不允许修改用户名和邮箱
	if _, ok := updates["username"]; ok {
		return errors.New("cannot update username")
	}
	if _, ok := updates["email"]; ok {
		return errors.New("cannot update email")
	}

	// 如果要修改密码
	if newPassword, ok := updates["password"].(string); ok {
		if err := passwordpkg.ValidatePassword(newPassword); err != nil {
			return err
		}
		hashedPassword, err := passwordpkg.HashPassword(newPassword)
		if err != nil {
			return err
		}
		user.PasswordHash = hashedPassword
	}

	// 应用其他更新
	if fullName, ok := updates["full_name"].(string); ok {
		user.FullName = fullName
	}

	return s.repo.Update(user)
}

// GetAllUsers 获取所有用户列表
func (s *userService) GetAllUsers() ([]models.User, error) {
	return s.repo.GetAllUsers()
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(id uint) error {
	return s.repo.Delete(id)
}

// GetUserPermissions 获取用户权限
func (s *userService) GetUserPermissions(userID uint) ([]string, error) {
	return s.repo.GetUserPermissions(userID)
}

// CheckPermission 检查用户权限
func (s *userService) CheckPermission(userID uint, resource, action string) (bool, error) {
	permissions, err := s.GetUserPermissions(userID)
	if err != nil {
		return false, err
	}

	// 检查是否有匹配的权限
	requiredPermission := resource + ":" + action
	for _, perm := range permissions {
		if perm == requiredPermission || perm == resource+":*" || perm == "*:*" {
			return true, nil
		}
	}

	return false, nil
}

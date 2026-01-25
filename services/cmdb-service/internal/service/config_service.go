package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"github.com/itcmdb/cmdb-service/internal/models"
	"github.com/itcmdb/cmdb-service/internal/repository"
)

type ConfigService interface {
	GetAllConfigs() ([]models.SystemConfig, error)
	GetConfigsByCategory(category string) ([]models.SystemConfig, error)
	GetConfig(category, key string) (*models.SystemConfig, error)
	GetConfigValue(category, key string) (string, error)
	CreateConfig(req *CreateConfigRequest, userID uint) (*models.SystemConfig, error)
	UpdateConfig(id uint, req *UpdateConfigRequest, userID uint) (*models.SystemConfig, error)
	DeleteConfig(id uint) error
	BatchUpdateConfigs(configs []BatchConfigRequest, userID uint) error
}

type configService struct {
	repo          repository.ConfigRepository
	encryptionKey []byte // 32 bytes for AES-256
}

type CreateConfigRequest struct {
	Category    string `json:"category" binding:"required"`
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value" binding:"required"`
	Description string `json:"description"`
	IsEncrypted bool   `json:"is_encrypted"`
}

type UpdateConfigRequest struct {
	Value       string `json:"value"`
	Description string `json:"description"`
	IsEncrypted bool   `json:"is_encrypted"`
	IsActive    *bool  `json:"is_active"`
}

type BatchConfigRequest struct {
	Category    string `json:"category" binding:"required"`
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value" binding:"required"`
	Description string `json:"description"`
	IsEncrypted bool   `json:"is_encrypted"`
}

func NewConfigService(repo repository.ConfigRepository, encryptionKey string) ConfigService {
	// 如果没有提供加密密钥，使用默认密钥（生产环境应该从环境变量读取）
	if encryptionKey == "" {
		encryptionKey = "itcmdb-default-encryption-key-32" // 32 bytes
	}

	// 确保密钥长度为 32 字节
	key := []byte(encryptionKey)
	if len(key) < 32 {
		// 填充到 32 字节
		padding := make([]byte, 32-len(key))
		key = append(key, padding...)
	} else if len(key) > 32 {
		// 截断到 32 字节
		key = key[:32]
	}

	return &configService{
		repo:          repo,
		encryptionKey: key,
	}
}

func (s *configService) GetAllConfigs() ([]models.SystemConfig, error) {
	configs, err := s.repo.GetAllConfigs()
	if err != nil {
		return nil, err
	}

	// 解密加密的配置值
	for i := range configs {
		if configs[i].IsEncrypted {
			decrypted, err := s.decrypt(configs[i].Value)
			if err == nil {
				configs[i].Value = decrypted
			}
		}
	}

	return configs, nil
}

func (s *configService) GetConfigsByCategory(category string) ([]models.SystemConfig, error) {
	configs, err := s.repo.GetConfigsByCategory(category)
	if err != nil {
		return nil, err
	}

	// 解密加密的配置值
	for i := range configs {
		if configs[i].IsEncrypted {
			decrypted, err := s.decrypt(configs[i].Value)
			if err == nil {
				configs[i].Value = decrypted
			}
		}
	}

	return configs, nil
}

func (s *configService) GetConfig(category, key string) (*models.SystemConfig, error) {
	config, err := s.repo.GetConfig(category, key)
	if err != nil {
		return nil, err
	}

	// 解密加密的配置值
	if config.IsEncrypted {
		decrypted, err := s.decrypt(config.Value)
		if err == nil {
			config.Value = decrypted
		}
	}

	return config, nil
}

func (s *configService) GetConfigValue(category, key string) (string, error) {
	config, err := s.GetConfig(category, key)
	if err != nil {
		return "", err
	}
	return config.Value, nil
}

func (s *configService) CreateConfig(req *CreateConfigRequest, userID uint) (*models.SystemConfig, error) {
	value := req.Value

	// 如果需要加密，加密值
	if req.IsEncrypted {
		encrypted, err := s.encrypt(value)
		if err != nil {
			return nil, err
		}
		value = encrypted
	}

	config := &models.SystemConfig{
		Category:    req.Category,
		Key:         req.Key,
		Value:       value,
		Description: req.Description,
		IsEncrypted: req.IsEncrypted,
		IsActive:    true,
		UpdatedBy:   userID,
	}

	if err := s.repo.CreateConfig(config); err != nil {
		return nil, err
	}

	// 返回时解密
	if config.IsEncrypted {
		config.Value = req.Value
	}

	return config, nil
}

func (s *configService) UpdateConfig(id uint, req *UpdateConfigRequest, userID uint) (*models.SystemConfig, error) {
	config, err := s.repo.GetConfig("", "")
	if err != nil {
		return nil, errors.New("config not found")
	}

	value := req.Value

	// 如果需要加密，加密值
	if req.IsEncrypted {
		encrypted, err := s.encrypt(value)
		if err != nil {
			return nil, err
		}
		value = encrypted
	}

	config.Value = value
	config.Description = req.Description
	config.IsEncrypted = req.IsEncrypted
	if req.IsActive != nil {
		config.IsActive = *req.IsActive
	}
	config.UpdatedBy = userID

	if err := s.repo.UpdateConfig(config); err != nil {
		return nil, err
	}

	// 返回时解密
	if config.IsEncrypted {
		config.Value = req.Value
	}

	return config, nil
}

func (s *configService) DeleteConfig(id uint) error {
	return s.repo.DeleteConfig(id)
}

func (s *configService) BatchUpdateConfigs(configs []BatchConfigRequest, userID uint) error {
	for _, req := range configs {
		value := req.Value

		// 如果需要加密，加密值
		if req.IsEncrypted {
			encrypted, err := s.encrypt(value)
			if err != nil {
				return err
			}
			value = encrypted
		}

		config := &models.SystemConfig{
			Category:    req.Category,
			Key:         req.Key,
			Value:       value,
			Description: req.Description,
			IsEncrypted: req.IsEncrypted,
			IsActive:    true,
			UpdatedBy:   userID,
		}

		if err := s.repo.UpsertConfig(config); err != nil {
			return err
		}
	}

	return nil
}

// encrypt 加密字符串
func (s *configService) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt 解密字符串
func (s *configService) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/StellarisJAY/beepbot/internal/crypto"
	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/types"
	"gorm.io/datatypes"
)

// ProviderService 供应商服务
type ProviderService struct {
	repo      repository.ProviderRepository
	encryptor *crypto.Encryptor
}

func NewProviderService(repo repository.ProviderRepository, encryptor *crypto.Encryptor) *ProviderService {
	return &ProviderService{
		repo:      repo,
		encryptor: encryptor,
	}
}

// CreateProviderRequest 创建供应商请求
type CreateProviderRequest struct {
	Name         string             `json:"name" binding:"required"`
	ProviderType types.ProviderType `json:"provider_type" binding:"required"`
	APIKey       string             `json:"api_key" binding:"required"`
	BaseURL      string             `json:"base_url"`
	ExtraConfig  map[string]any     `json:"extra_config"`
	IsDefault    bool               `json:"is_default"`
}

// UpdateProviderRequest 更新供应商请求
type UpdateProviderRequest struct {
	Name        string         `json:"name"`
	APIKey      string         `json:"api_key"`
	BaseURL     string         `json:"base_url"`
	ExtraConfig map[string]any `json:"extra_config"`
	IsDefault   bool           `json:"is_default"`
}

// ProviderResponse 供应商响应
type ProviderResponse struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	ProviderType types.ProviderType `json:"provider_type"`
	APIKeyMasked string             `json:"api_key_masked"`
	BaseURL      string             `json:"base_url"`
	ExtraConfig  map[string]any     `json:"extra_config"`
	IsDefault    bool               `json:"is_default"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

// CreateProvider 创建供应商
func (s *ProviderService) CreateProvider(req *CreateProviderRequest) (*types.Provider, error) {
	// 检查名称是否已存在
	if _, err := s.repo.GetByName(req.Name); err == nil {
		return nil, errors.New("provider name already exists")
	}

	// 加密 API Key
	encryptedKey, err := s.encryptor.Encrypt(req.APIKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt api key failed: %w", err)
	}

	// 处理 ExtraConfig
	var extraConfig datatypes.JSON
	if req.ExtraConfig != nil {
		data, _ := json.Marshal(req.ExtraConfig)
		extraConfig = data
	}

	provider := &types.Provider{
		Name:         req.Name,
		ProviderType: req.ProviderType,
		APIKey:       encryptedKey,
		BaseURL:      req.BaseURL,
		ExtraConfig:  extraConfig,
		IsDefault:    req.IsDefault,
	}

	// 设置 ID 和时间
	provider.ID = types.GenerateUUIDv7()
	now := time.Now()
	provider.CreatedAt = now
	provider.UpdatedAt = now

	if err := s.repo.Create(provider); err != nil {
		return nil, err
	}

	// 如果设置为默认，更新其他同类型供应商
	if req.IsDefault {
		_ = s.repo.SetDefault(provider.ID)
	}

	return provider, nil
}

// UpdateProvider 更新供应商
func (s *ProviderService) UpdateProvider(id string, req *UpdateProviderRequest) (*types.Provider, error) {
	provider, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// 更新字段
	if req.Name != "" {
		// 检查名称是否被其他供应商使用
		if existing, err := s.repo.GetByName(req.Name); err == nil && existing.ID != id {
			return nil, errors.New("provider name already exists")
		}
		provider.Name = req.Name
	}

	if req.APIKey != "" {
		encryptedKey, err := s.encryptor.Encrypt(req.APIKey)
		if err != nil {
			return nil, fmt.Errorf("encrypt api key failed: %w", err)
		}
		provider.APIKey = encryptedKey
	}

	if req.BaseURL != "" {
		provider.BaseURL = req.BaseURL
	}

	if req.ExtraConfig != nil {
		data, _ := json.Marshal(req.ExtraConfig)
		provider.ExtraConfig = data
	}

	provider.IsDefault = req.IsDefault
	provider.UpdatedAt = time.Now()

	if err := s.repo.Update(provider); err != nil {
		return nil, err
	}

	// 如果设置为默认，更新其他同类型供应商
	if req.IsDefault {
		_ = s.repo.SetDefault(provider.ID)
	}

	return provider, nil
}

// DeleteProvider 删除供应商
func (s *ProviderService) DeleteProvider(id string) error {
	// TODO: 检查是否有智能体正在使用此供应商
	return s.repo.Delete(id)
}

// GetProvider 获取供应商
func (s *ProviderService) GetProvider(id string) (*ProviderResponse, error) {
	provider, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	return s.toResponse(provider), nil
}

// GetProviderByName 根据名称获取供应商
func (s *ProviderService) GetProviderByName(name string) (*ProviderResponse, error) {
	provider, err := s.repo.GetByName(name)
	if err != nil {
		return nil, err
	}

	return s.toResponse(provider), nil
}

// ListProviders 列出供应商
func (s *ProviderService) ListProviders(page, pageSize int) ([]ProviderResponse, int64, error) {
	providers, total, err := s.repo.List(page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]ProviderResponse, len(providers))
	for i, p := range providers {
		responses[i] = *s.toResponse(&p)
	}

	return responses, total, nil
}

// GetProvidersByType 根据类型获取供应商列表
func (s *ProviderService) GetProvidersByType(providerType types.ProviderType) ([]ProviderResponse, error) {
	providers, err := s.repo.GetByType(providerType)
	if err != nil {
		return nil, err
	}

	responses := make([]ProviderResponse, len(providers))
	for i, p := range providers {
		responses[i] = *s.toResponse(&p)
	}

	return responses, nil
}

// SetDefaultProvider 设置默认供应商
func (s *ProviderService) SetDefaultProvider(id string) error {
	return s.repo.SetDefault(id)
}

// GetDecryptedAPIKey 获取解密后的 API Key（内部使用）
func (s *ProviderService) GetDecryptedAPIKey(providerID string) (string, error) {
	provider, err := s.repo.GetByID(providerID)
	if err != nil {
		return "", err
	}
	return s.encryptor.Decrypt(provider.APIKey)
}

// GetProviderEntity 获取供应商实体（包含加密的 API Key）
func (s *ProviderService) GetProviderEntity(id string) (*types.Provider, error) {
	return s.repo.GetByID(id)
}

// toResponse 转换为响应格式
func (s *ProviderService) toResponse(provider *types.Provider) *ProviderResponse {
	// 解密 API Key 用于脱敏显示
	apiKey, _ := s.encryptor.Decrypt(provider.APIKey)

	var extraConfig map[string]any
	if provider.ExtraConfig != nil {
		_ = json.Unmarshal(provider.ExtraConfig, &extraConfig)
	}

	return &ProviderResponse{
		ID:           provider.ID,
		Name:         provider.Name,
		ProviderType: provider.ProviderType,
		APIKeyMasked: crypto.MaskAPIKey(apiKey),
		BaseURL:      provider.BaseURL,
		ExtraConfig:  extraConfig,
		IsDefault:    provider.IsDefault,
		CreatedAt:    provider.CreatedAt,
		UpdatedAt:    provider.UpdatedAt,
	}
}

package api

import (
	"strconv"

	"github.com/StellarisJAY/beepbot/internal/service"
	"github.com/StellarisJAY/beepbot/internal/types"
	"github.com/gin-gonic/gin"
)

// ProviderHandler 供应商处理器
type ProviderHandler struct {
	service *service.ProviderService
}

func NewProviderHandler(service *service.ProviderService) *ProviderHandler {
	return &ProviderHandler{service: service}
}

// ListProviders 获取供应商列表
// GET /api/v1/providers
func (h *ProviderHandler) ListProviders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	providers, total, err := h.service.ListProviders(page, pageSize)
	if err != nil {
		InternalError(c, "failed to list providers: "+err.Error())
		return
	}

	SuccessWithPage(c, providers, total, page, pageSize)
}

// GetProvider 获取单个供应商
// GET /api/v1/providers/:id
func (h *ProviderHandler) GetProvider(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "provider id is required")
		return
	}

	provider, err := h.service.GetProvider(id)
	if err != nil {
		NotFound(c, "provider not found")
		return
	}

	Success(c, provider)
}

// CreateProvider 创建供应商
// POST /api/v1/providers
func (h *ProviderHandler) CreateProvider(c *gin.Context) {
	var req service.CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// 验证供应商类型
	if req.ProviderType != types.ProviderTypeOpenAI && req.ProviderType != types.ProviderTypeDashScope {
		BadRequest(c, "invalid provider type, must be 'openai' or 'dashscope'")
		return
	}

	provider, err := h.service.CreateProvider(&req)
	if err != nil {
		Error(c, 500, "failed to create provider: "+err.Error())
		return
	}

	Success(c, provider)
}

// UpdateProvider 更新供应商
// PUT /api/v1/providers/:id
func (h *ProviderHandler) UpdateProvider(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "provider id is required")
		return
	}

	var req service.UpdateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	provider, err := h.service.UpdateProvider(id, &req)
	if err != nil {
		Error(c, 500, "failed to update provider: "+err.Error())
		return
	}

	Success(c, provider)
}

// DeleteProvider 删除供应商
// DELETE /api/v1/providers/:id
func (h *ProviderHandler) DeleteProvider(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "provider id is required")
		return
	}

	if err := h.service.DeleteProvider(id); err != nil {
		Error(c, 500, "failed to delete provider: "+err.Error())
		return
	}

	Success(c, nil)
}

// SetDefaultProvider 设置默认供应商
// PUT /api/v1/providers/:id/default
func (h *ProviderHandler) SetDefaultProvider(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "provider id is required")
		return
	}

	if err := h.service.SetDefaultProvider(id); err != nil {
		Error(c, 500, "failed to set default provider: "+err.Error())
		return
	}

	Success(c, nil)
}

// GetProvidersByType 根据类型获取供应商列表
// GET /api/v1/providers/type/:type
func (h *ProviderHandler) GetProvidersByType(c *gin.Context) {
	providerType := c.Param("type")
	if providerType == "" {
		BadRequest(c, "provider type is required")
		return
	}

	providers, err := h.service.GetProvidersByType(types.ProviderType(providerType))
	if err != nil {
		Error(c, 500, "failed to get providers by type: "+err.Error())
		return
	}

	Success(c, providers)
}

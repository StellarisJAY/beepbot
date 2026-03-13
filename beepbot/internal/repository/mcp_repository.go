package repository

import (
	"github.com/StellarisJAY/beepbot/internal/types"
	"gorm.io/gorm"
)

// MCPServerRepository MCP 服务器仓储接口
type MCPServerRepository interface {
	Repository[types.MCPServer]

	// ListWithQuery 带筛选条件的分页查询
	ListWithQuery(page, pageSize int, query *types.MCPServerQuery) ([]types.MCPServer, int64, error)

	// GetByName 根据名称获取 MCP 服务器
	GetByName(name string) (*types.MCPServer, error)

	// GetByStatus 根据状态获取 MCP 服务器列表
	GetByStatus(status types.MCPServerStatus) ([]types.MCPServer, error)

	// GetActiveServers 获取所有活跃的 MCP 服务器
	GetActiveServers() ([]types.MCPServer, error)

	// UpdateStatus 更新服务器状态
	UpdateStatus(id string, status types.MCPServerStatus) error
}

type mcpServerRepository struct {
	*BaseRepository[types.MCPServer]
	db *gorm.DB
}

// NewMCPServerRepository 创建 MCP 服务器仓储
func NewMCPServerRepository(db *gorm.DB) MCPServerRepository {
	return &mcpServerRepository{
		BaseRepository: NewBaseRepository[types.MCPServer](db),
		db:             db,
	}
}

// ListWithQuery 带筛选条件的分页查询
func (r *mcpServerRepository) ListWithQuery(page, pageSize int, query *types.MCPServerQuery) ([]types.MCPServer, int64, error) {
	var servers []types.MCPServer
	var total int64

	db := r.db.Model(&types.MCPServer{})

	// 动态拼接筛选条件
	if query != nil {
		if query.Name != "" {
			db = db.Where("name LIKE ?", "%"+query.Name+"%")
		}
		if query.Status != "" {
			db = db.Where("status = ?", query.Status)
		}
	}

	// 统计总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&servers).Error; err != nil {
		return nil, 0, err
	}

	return servers, total, nil
}

// GetByName 根据名称获取 MCP 服务器
func (r *mcpServerRepository) GetByName(name string) (*types.MCPServer, error) {
	var server types.MCPServer
	err := r.db.Where("name = ?", name).First(&server).Error
	if err != nil {
		return nil, err
	}
	return &server, nil
}

// GetByStatus 根据状态获取 MCP 服务器列表
func (r *mcpServerRepository) GetByStatus(status types.MCPServerStatus) ([]types.MCPServer, error) {
	var servers []types.MCPServer
	err := r.db.Where("status = ?", status).Find(&servers).Error
	return servers, err
}

// GetActiveServers 获取所有活跃的 MCP 服务器
func (r *mcpServerRepository) GetActiveServers() ([]types.MCPServer, error) {
	return r.GetByStatus(types.MCPServerStatusActive)
}

// UpdateStatus 更新服务器状态
func (r *mcpServerRepository) UpdateStatus(id string, status types.MCPServerStatus) error {
	return r.db.Model(&types.MCPServer{}).Where("id = ?", id).Update("status", status).Error
}
package api

import (
	"strconv"

	"github.com/StellarisJAY/beepbot/internal/service"
	"github.com/StellarisJAY/beepbot/internal/types"
	"github.com/gin-gonic/gin"
)

// SkillHandler 技能处理器
type SkillHandler struct {
	service *service.SkillService
}

// NewSkillHandler 创建技能处理器
func NewSkillHandler(service *service.SkillService) *SkillHandler {
	return &SkillHandler{service: service}
}

// ListSkills 获取技能列表
// GET /api/v1/skills
func (h *SkillHandler) ListSkills(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	status := types.SkillStatus(c.Query("status"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	skills, total, err := h.service.ListSkills(page, size, status)
	if err != nil {
		InternalError(c, "failed to list skills: "+err.Error())
		return
	}

	SuccessWithPage(c, skills, total, page, size)
}

// GetSkill 获取技能详情
// GET /api/v1/skills/:id
func (h *SkillHandler) GetSkill(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "skill id is required")
		return
	}

	skill, err := h.service.GetSkill(id)
	if err != nil {
		NotFound(c, "skill not found")
		return
	}

	Success(c, skill)
}

// GetSkillWithFiles 获取技能详情（包含文件列表）
// GET /api/v1/skills/:id/detail
func (h *SkillHandler) GetSkillWithFiles(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "skill id is required")
		return
	}

	result, err := h.service.GetSkillWithFiles(id)
	if err != nil {
		NotFound(c, "skill not found")
		return
	}

	Success(c, result)
}

// GetSkillFiles 获取技能文件列表
// GET /api/v1/skills/:id/files
func (h *SkillHandler) GetSkillFiles(c *gin.Context) {
	skillID := c.Param("id")
	if skillID == "" {
		BadRequest(c, "skill id is required")
		return
	}

	files, err := h.service.GetSkillFiles(skillID)
	if err != nil {
		InternalError(c, "failed to get skill files: "+err.Error())
		return
	}

	Success(c, files)
}

// GetSkillFileContent 获取文件内容
// GET /api/v1/skills/:id/files/:fileId
func (h *SkillHandler) GetSkillFileContent(c *gin.Context) {
	skillID := c.Param("id")
	fileID := c.Param("fileId")

	if skillID == "" {
		BadRequest(c, "skill id is required")
		return
	}
	if fileID == "" {
		BadRequest(c, "file id is required")
		return
	}

	result, err := h.service.GetFileContent(skillID, fileID)
	if err != nil {
		NotFound(c, "file not found")
		return
	}

	Success(c, result)
}

// UploadSkill 上传安装技能
// POST /api/v1/skills/upload
func (h *SkillHandler) UploadSkill(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		BadRequest(c, "file is required")
		return
	}
	defer file.Close()

	result, err := h.service.UploadSkill(file, header)
	if err != nil {
		Error(c, 400, err.Error())
		return
	}

	Success(c, result)
}

// DeleteSkill 删除技能
// DELETE /api/v1/skills/:id
func (h *SkillHandler) DeleteSkill(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "skill id is required")
		return
	}

	if err := h.service.DeleteSkill(id); err != nil {
		Error(c, 500, "failed to delete skill: "+err.Error())
		return
	}

	Success(c, nil)
}

// UpdateSkillStatus 更新技能状态
// PUT /api/v1/skills/:id/status
func (h *SkillHandler) UpdateSkillStatus(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		BadRequest(c, "skill id is required")
		return
	}

	var req struct {
		Status types.SkillStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "invalid request: "+err.Error())
		return
	}

	// 验证状态值
	if req.Status != types.SkillStatusActive && req.Status != types.SkillStatusInactive {
		BadRequest(c, "invalid status value")
		return
	}

	if err := h.service.UpdateSkillStatus(id, req.Status); err != nil {
		Error(c, 500, "failed to update skill status: "+err.Error())
		return
	}

	Success(c, nil)
}

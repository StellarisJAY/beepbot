package api

import (
	"net/http"

	"github.com/StellarisJAY/beepbot/internal/service"
	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login 登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "用户名和密码不能为空")
		return
	}

	resp, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			Unauthorized(c, "用户名或密码错误")
			return
		}
		InternalServerError(c, err.Error())
		return
	}

	Success(c, resp)
}

// GetMe 获取当前用户信息
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		Unauthorized(c, "未登录")
		return
	}

	userInfo, err := h.authService.GetUserInfo(userID.(string))
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, userInfo)
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		Unauthorized(c, "未登录")
		return
	}

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "参数错误")
		return
	}

	err := h.authService.ChangePassword(userID.(string), req.OldPassword, req.NewPassword)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			Unauthorized(c, "旧密码错误")
		case service.ErrSamePassword:
			BadRequest(c, "新密码不能与旧密码相同")
		default:
			InternalServerError(c, err.Error())
		}
		return
	}

	Success(c, gin.H{"message": "密码修改成功，请重新登录"})
}

// ChangeUsername 修改用户名
func (h *AuthHandler) ChangeUsername(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		Unauthorized(c, "未登录")
		return
	}

	var req struct {
		NewUsername string `json:"new_username" binding:"required,min=3,max=64"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "用户名长度需要在 3-64 个字符之间")
		return
	}

	err := h.authService.ChangeUsername(userID.(string), req.NewUsername)
	if err != nil {
		switch err {
		case service.ErrUsernameAlreadyExists:
			BadRequest(c, "用户名已存在")
		default:
			InternalServerError(c, err.Error())
		}
		return
	}

	Success(c, gin.H{"message": "用户名修改成功"})
}

// AuthMiddleware JWT 认证中间件
func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 Authorization Header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未登录",
			})
			return
		}

		// 解析 Bearer Token
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的认证信息",
			})
			return
		}

		tokenString := authHeader[7:]

		// 验证 Token
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "登录已过期，请重新登录",
			})
			return
		}

		// 将用户信息存入 context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)

		c.Next()
	}
}
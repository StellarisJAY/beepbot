package service

import (
	"errors"
	"time"

	"github.com/StellarisJAY/beepbot/internal/repository"
	"github.com/StellarisJAY/beepbot/internal/types"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// 默认管理员凭据
const (
	DefaultAdminUsername = "admin"
	DefaultAdminPassword = "admin"
)

// 认证相关错误
var (
	ErrInvalidCredentials     = errors.New("invalid username or password")
	ErrUserNotFound           = errors.New("user not found")
	ErrSamePassword           = errors.New("new password cannot be same as old password")
	ErrUsernameAlreadyExists  = errors.New("username already exists")
)

// JWTClaims JWT 声明
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// AuthService 认证服务
type AuthService struct {
	repo      repository.AdminRepository
	jwtSecret []byte
	tokenTTL  time.Duration
}

// NewAuthService 创建认证服务
func NewAuthService(repo repository.AdminRepository, jwtSecret string) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: []byte(jwtSecret),
		tokenTTL:  24 * time.Hour,
	}
}

// Login 登录验证，返回 JWT Token
func (s *AuthService) Login(username, password string) (*types.LoginResponse, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 生成 JWT Token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &types.LoginResponse{
		Token:               token,
		RequirePasswordChange: user.RequirePasswordChange,
	}, nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(userID, oldPassword, newPassword string) error {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}

	// 检查新密码是否与旧密码相同
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(newPassword)); err == nil {
		return ErrSamePassword
	}

	// 加密新密码
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(newHash)
	user.RequirePasswordChange = false

	return s.repo.Update(user)
}

// ChangeUsername 修改用户名
func (s *AuthService) ChangeUsername(userID, newUsername string) error {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return ErrUserNotFound
	}

	// 检查用户名是否已被使用
	existing, err := s.repo.GetByUsername(newUsername)
	if err == nil && existing.ID != userID {
		return ErrUsernameAlreadyExists
	}

	user.Username = newUsername
	return s.repo.Update(user)
}

// ValidateToken 验证 JWT Token
func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetUserInfo 获取用户信息
func (s *AuthService) GetUserInfo(userID string) (*types.AdminUserInfo, error) {
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return &types.AdminUserInfo{
		ID:                   user.ID,
		Username:             user.Username,
		RequirePasswordChange: user.RequirePasswordChange,
	}, nil
}

// CreateDefaultAdmin 创建默认管理员
func (s *AuthService) CreateDefaultAdmin() error {
	// 检查是否已存在管理员
	exists, err := s.repo.Exists()
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	// 加密默认密码
	hash, err := bcrypt.GenerateFromPassword([]byte(DefaultAdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &types.AdminUser{
		Username:             DefaultAdminUsername,
		PasswordHash:         string(hash),
		RequirePasswordChange: true,
	}

	return s.repo.Create(user)
}

// generateToken 生成 JWT Token
func (s *AuthService) generateToken(user *types.AdminUser) (string, error) {
	claims := &JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
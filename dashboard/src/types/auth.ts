// 认证相关类型定义

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  require_password_change: boolean
}

export interface AdminUser {
  id: string
  username: string
  require_password_change: boolean
}

export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

export interface ChangeUsernameRequest {
  new_username: string
}
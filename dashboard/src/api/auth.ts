import { http } from '@/utils/http'
import type { LoginRequest, LoginResponse, AdminUser, ChangePasswordRequest, ChangeUsernameRequest } from '@/types/auth'

// 登录
export function login(data: LoginRequest): Promise<LoginResponse> {
  return http.post<LoginResponse>('/auth/login', data).then((res) => res.data)
}

// 获取当前用户信息
export function getMe(): Promise<AdminUser> {
  return http.get<AdminUser>('/auth/me').then((res) => res.data)
}

// 修改密码
export function changePassword(data: ChangePasswordRequest): Promise<void> {
  return http.put('/auth/password', data).then(() => {})
}

// 修改用户名
export function changeUsername(data: ChangeUsernameRequest): Promise<void> {
  return http.put('/auth/username', data).then(() => {})
}
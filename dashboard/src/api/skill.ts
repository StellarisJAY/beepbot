import { http, type ApiResponse, type PageResponse } from '@/utils/http'
import type { Skill, SkillFile, SkillWithFiles, SkillFileContent, UploadSkillResult } from '@/types/skill'

/**
 * 技能 API 服务
 */
export const skillApi = {
  /**
   * 获取技能列表（分页）
   */
  list(page: number = 1, size: number = 10, status?: string): Promise<PageResponse<Skill>> {
    const params: Record<string, string | number> = { page, size }
    if (status) {
      params.status = status
    }
    return http.get<Skill[]>('/skills', { params }) as Promise<PageResponse<Skill>>
  },

  /**
   * 获取技能详情
   */
  get(id: string): Promise<ApiResponse<Skill>> {
    return http.get<Skill>(`/skills/${id}`)
  },

  /**
   * 获取技能详情（包含文件列表）
   */
  getWithFiles(id: string): Promise<ApiResponse<SkillWithFiles>> {
    return http.get<SkillWithFiles>(`/skills/${id}/detail`)
  },

  /**
   * 获取技能文件列表
   */
  getFiles(id: string): Promise<ApiResponse<SkillFile[]>> {
    return http.get<SkillFile[]>(`/skills/${id}/files`)
  },

  /**
   * 获取文件内容
   */
  getFileContent(skillId: string, fileId: string): Promise<ApiResponse<SkillFileContent>> {
    return http.get<SkillFileContent>(`/skills/${skillId}/files/${fileId}`)
  },

  /**
   * 上传安装技能
   */
  upload(file: File): Promise<ApiResponse<UploadSkillResult>> {
    const formData = new FormData()
    formData.append('file', file)
    return http.post<UploadSkillResult>('/skills/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },

  /**
   * 删除技能
   */
  delete(id: string): Promise<ApiResponse<void>> {
    return http.delete<void>(`/skills/${id}`)
  },

  /**
   * 更新技能状态
   */
  updateStatus(id: string, status: string): Promise<ApiResponse<void>> {
    return http.put<void>(`/skills/${id}/status`, { status })
  },
}

export default skillApi
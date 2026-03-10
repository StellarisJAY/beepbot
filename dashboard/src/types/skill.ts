// 技能状态
export type SkillStatus = 'active' | 'inactive'

// 技能
export interface Skill {
  id: string
  name: string
  description: string
  version: string
  author: string
  path: string
  status: SkillStatus
  installed_at: string
  updated_at: string
}

// 技能文件
export interface SkillFile {
  id: string
  skill_id: string
  file_name: string
  file_path: string
  file_type: string
  file_size: number
  created_at: string
}

// 技能详情（包含文件列表）
export interface SkillWithFiles extends Skill {
  files: SkillFile[]
}

// 文件内容
export interface SkillFileContent {
  id: string
  file_name: string
  file_type: string
  content: string
  file_path: string
  file_size: number
}

// 上传技能结果
export interface UploadSkillResult {
  id: string
  name: string
  description: string
  version: string
  author: string
}
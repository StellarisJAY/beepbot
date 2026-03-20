/**
 * 团队状态枚举
 */
export enum TeamStatus {
  Active = 'active',
  Inactive = 'inactive',
}

/**
 * 团队状态显示名称映射
 */
export const TeamStatusLabels: Record<TeamStatus, string> = {
  [TeamStatus.Active]: '活跃',
  [TeamStatus.Inactive]: '停用',
}

/**
 * 团队状态选项（用于下拉选择）
 */
export const TeamStatusOptions = [
  { value: TeamStatus.Active, label: TeamStatusLabels[TeamStatus.Active] },
  { value: TeamStatus.Inactive, label: TeamStatusLabels[TeamStatus.Inactive] },
]

/**
 * 成员角色枚举
 */
export enum MemberRole {
  Leader = 'leader',
  Member = 'member',
}

/**
 * 团队成员简要信息
 */
export interface TeamMemberBrief {
  agent_id: string
  agent_name: string
  member_name: string
  role: string
  description: string
}

/**
 * 团队配置
 */
export interface Team {
  id: string
  name: string
  description: string
  leader_id: string
  leader?: TeamMemberBrief
  members?: TeamMemberBrief[]
  status: TeamStatus
  created_at: string
  updated_at: string
}

/**
 * 创建团队请求
 */
export interface CreateTeamRequest {
  name: string
  description?: string
  leader_id: string
  members?: MemberRequest[]
}

/**
 * 成员请求
 */
export interface MemberRequest {
  agent_id: string
  member_name: string
  description?: string
}

/**
 * 更新团队请求
 */
export interface UpdateTeamRequest {
  name?: string
  description?: string
  leader_id?: string
  members?: MemberRequest[]
  status?: TeamStatus
}
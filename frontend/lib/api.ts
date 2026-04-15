export const API_BASE = process.env.NEXT_PUBLIC_API_URL || ''

export interface DashboardMetrics {
  active_companies: number
  messages_today: number
  broadcasts_today: number
  success_rate: number
  last_update: string
  sessions_active: number
  messages_sent: number
  messages_failed: number
  broadcasts_created: number
  alerts: Alert[]
}

export interface Alert {
  type: string
  level: string
  message: string
}

export async function getMetrics(): Promise<DashboardMetrics> {
  const res = await fetch(`${API_BASE}/metrics`)
  if (!res.ok) {
    throw new Error('Failed to fetch metrics')
  }
  return res.json()
}

export interface Company {
  account_id: string
  status: string
  last_message?: string
  updated_at: string
}

export interface CompaniesResponse {
  companies: Company[]
}

export async function getCompanies(): Promise<CompaniesResponse> {
  const res = await fetch(`${API_BASE}/companies`)
  if (!res.ok) {
    throw new Error('Failed to fetch companies')
  }
  return res.json()
}

export interface AdminMessage {
  id: number
  account_id: string
  to: string
  content: string
  status: string
  created_at: string
}

export interface MessagesResponse {
  messages: AdminMessage[]
  total: number
}

export async function getAdminMessages(filters?: {
  account_id?: string
  status?: string
  limit?: number
}): Promise<MessagesResponse> {
  const params = new URLSearchParams()
  if (filters?.account_id) params.set('account_id', filters.account_id)
  if (filters?.status) params.set('status', filters.status)
  if (filters?.limit) params.set('limit', String(filters.limit))

  const res = await fetch(`${API_BASE}/admin/messages?${params}`)
  if (!res.ok) {
    throw new Error('Failed to fetch messages')
  }
  return res.json()
}

export interface SessionInfo {
  account_id: string
  status: string
  qr_string?: string
  updated_at: string
}

export interface SessionsResponse {
  sessions: SessionInfo[]
}

export async function getAdminSessions(): Promise<SessionsResponse> {
  const res = await fetch(`${API_BASE}/admin/sessions`)
  if (!res.ok) {
    throw new Error('Failed to fetch sessions')
  }
  return res.json()
}

export async function postAdminSession(action: string, accountId: string): Promise<{status: string}> {
  const res = await fetch(`${API_BASE}/admin/sessions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ account_id: accountId, action }),
  })
  if (!res.ok) {
    throw new Error('Failed to perform action')
  }
  return res.json()
}

export interface BroadcastInfo {
  reference_id: string
  ruc_empresa: string
  total: number
  status: string
  success: number
  failed: number
  created_at: string
}

export interface BroadcastsResponse {
  broadcasts: BroadcastInfo[]
}

export async function getAdminBroadcasts(accountId?: string): Promise<BroadcastsResponse> {
  const params = accountId ? `?account_id=${accountId}` : ''
  const res = await fetch(`${API_BASE}/admin/broadcasts${params}`)
  if (!res.ok) {
    throw new Error('Failed to fetch broadcasts')
  }
  return res.json()
}
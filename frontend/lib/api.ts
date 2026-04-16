export const API_BASE = process.env.NEXT_PUBLIC_API_URL || "";

export interface DashboardMetrics {
  active_companies: number;
  messages_today: number;
  broadcasts_today: number;
  success_rate: number;
  last_update: string;
  sessions_active: number;
  messages_sent: number;
  messages_failed: number;
  broadcasts_created: number;
  alerts: Alert[];
}

export interface Alert {
  type: string;
  level: string;
  message: string;
}

export async function getMetrics(): Promise<DashboardMetrics> {
  const res = await fetch(`${API_BASE}/api/dashboard/metricas`, {
    headers: authHeaders(),
  });
  if (!res.ok) {
    throw new Error("Failed to fetch metrics");
  }
  return res.json();
}

export interface Company {
  account_id: string;
  status: string;
  last_message?: string;
  updated_at: string;
}

export interface CompaniesResponse {
  companies: Company[];
}

export async function getCompanies(): Promise<CompaniesResponse> {
  const res = await fetch(`${API_BASE}/companies`);
  if (!res.ok) {
    throw new Error("Failed to fetch companies");
  }
  return res.json();
}

// ---- Empresa (admin API) ----

export interface Empresa {
  id: number;
  ruc: string;
  nombre: string;
  nombre_comercial?: string;
  telefono?: string;
  direccion?: string;
  activo: boolean;
  created_at: string;
  updated_at: string;
}

export interface EmpresasListResponse {
  ok: boolean;
  empresas: Empresa[];
  total: number;
  page: number;
  limit: number;
}

export interface EmpresaResponse {
  ok: boolean;
  empresa: Empresa;
  error?: string;
}

export interface EmpresaCreateRequest {
  ruc: string;
  nombre: string;
  nombre_comercial?: string;
  telefono?: string;
  direccion?: string;
}

function authHeaders(): HeadersInit {
  const token =
    typeof window !== "undefined" ? localStorage.getItem("admin_token") : null;
  return {
    "Content-Type": "application/json",
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
}

export async function getEmpresas(params?: {
  page?: number;
  limit?: number;
  busqueda?: string;
  estado?: string;
}): Promise<EmpresasListResponse> {
  const q = new URLSearchParams();
  if (params?.page) q.set("page", String(params.page));
  if (params?.limit) q.set("limit", String(params.limit));
  if (params?.busqueda) q.set("busqueda", params.busqueda);
  if (params?.estado) q.set("estado", params.estado);
  const res = await fetch(`${API_BASE}/api/companies?${q}`, {
    headers: authHeaders(),
  });
  if (!res.ok) throw new Error("Error al obtener empresas");
  return res.json();
}

export async function createEmpresa(
  data: EmpresaCreateRequest,
): Promise<EmpresaResponse> {
  const res = await fetch(`${API_BASE}/api/companies`, {
    method: "POST",
    headers: authHeaders(),
    body: JSON.stringify(data),
  });
  const json = await res.json();
  if (!res.ok) throw new Error(json.error || "Error al crear empresa");
  return json;
}

export async function updateEmpresa(
  id: number,
  data: Partial<EmpresaCreateRequest>,
): Promise<EmpresaResponse> {
  const res = await fetch(`${API_BASE}/api/companies/?id=${id}`, {
    method: "PUT",
    headers: authHeaders(),
    body: JSON.stringify(data),
  });
  const json = await res.json();
  if (!res.ok) throw new Error(json.error || "Error al actualizar empresa");
  return json;
}

export async function deleteEmpresa(id: number): Promise<void> {
  const res = await fetch(`${API_BASE}/api/companies/?id=${id}`, {
    method: "DELETE",
    headers: authHeaders(),
  });
  if (!res.ok) {
    const json = await res.json().catch(() => ({}));
    throw new Error(json.error || "Error al eliminar empresa");
  }
}

export interface AdminMessage {
  id: number;
  account_id: string;
  to: string;
  content: string;
  status: string;
  created_at: string;
}

export interface MessagesResponse {
  messages: AdminMessage[];
  total: number;
}

export async function getAdminMessages(filters?: {
  account_id?: string;
  status?: string;
  limit?: number;
}): Promise<MessagesResponse> {
  const params = new URLSearchParams();
  if (filters?.account_id) params.set("account_id", filters.account_id);
  if (filters?.status) params.set("status", filters.status);
  if (filters?.limit) params.set("limit", String(filters.limit));

  const res = await fetch(`${API_BASE}/admin/messages?${params}`);
  if (!res.ok) {
    throw new Error("Failed to fetch messages");
  }
  return res.json();
}

export interface SessionInfo {
  account_id: string;
  status: string;
  qr_string?: string;
  updated_at: string;
}

export interface SessionsResponse {
  sessions: SessionInfo[];
}

export async function getAdminSessions(): Promise<SessionsResponse> {
  const res = await fetch(`${API_BASE}/admin/sessions`);
  if (!res.ok) {
    throw new Error("Failed to fetch sessions");
  }
  return res.json();
}

export async function postAdminSession(
  action: string,
  accountId: string,
): Promise<{ status: string }> {
  const res = await fetch(`${API_BASE}/admin/sessions`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ account_id: accountId, action }),
  });
  if (!res.ok) {
    throw new Error("Failed to perform action");
  }
  return res.json();
}

export interface BroadcastInfo {
  reference_id: string;
  ruc_empresa: string;
  total: number;
  status: string;
  success: number;
  failed: number;
  created_at: string;
}

export interface BroadcastsResponse {
  broadcasts: BroadcastInfo[];
}

export async function getAdminBroadcasts(
  accountId?: string,
): Promise<BroadcastsResponse> {
  const params = accountId ? `?account_id=${accountId}` : "";
  const res = await fetch(`${API_BASE}/admin/broadcasts${params}`);
  if (!res.ok) {
    throw new Error("Failed to fetch broadcasts");
  }
  return res.json();
}

// ---- Users Management ----

export interface UserAdminRol {
  id: number;
  username: string;
  email: string;
  rol: string;
  role_id?: number;
  is_root: boolean;
  activo: boolean;
  empresa_id?: number;
  created_at: string;
  updated_at: string;
}

export interface Role {
  id: number;
  name: string;
  description: string;
  is_root: boolean;
}

export interface Module {
  id: number;
  name: string;
  slug: string;
  description: string;
}

export interface UsersResponse {
  users: UserAdminRol[];
  total: number;
}

export interface RolesResponse {
  roles: Role[];
}

export interface ModulesResponse {
  modules: Module[];
}

async function fetchWithAuth(url: string, options?: RequestInit) {
  const token = localStorage.getItem("admin_token");
  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...options?.headers,
  };
  const res = await fetch(url, { ...options, headers });
  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: "Request failed" }));
    throw new Error(error.error || "Request failed");
  }
  return res.json();
}

export async function getUsers(page = 1, limit = 20): Promise<UsersResponse> {
  return fetchWithAuth(
    `${API_BASE}/api/admin/users?page=${page}&limit=${limit}`,
  );
}

export async function getUser(id: number): Promise<UserAdminRol> {
  return fetchWithAuth(`${API_BASE}/api/admin/users/${id}`);
}

export interface CreateUserRequest {
  username: string;
  password: string;
  email: string;
  role_id?: number;
  empresa_id?: number;
}

export async function createUser(
  data: CreateUserRequest,
): Promise<UserAdminRol> {
  return fetchWithAuth(`${API_BASE}/api/admin/users`, {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export interface UpdateUserRequest {
  email?: string;
  role_id?: number;
  is_active?: boolean;
  empresa_id?: number;
}

export async function updateUser(
  id: number,
  data: UpdateUserRequest,
): Promise<UserAdminRol> {
  return fetchWithAuth(`${API_BASE}/api/admin/users/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export async function deleteUser(id: number): Promise<void> {
  await fetchWithAuth(`${API_BASE}/api/admin/users/${id}`, {
    method: "DELETE",
  });
}

export async function promoteToRoot(userId: number): Promise<UserAdminRol> {
  return fetchWithAuth(`${API_BASE}/api/admin/users/${userId}/promote`, {
    method: "POST",
  });
}

export async function assignUserModules(
  userId: number,
  moduleIds: number[],
): Promise<void> {
  await fetchWithAuth(`${API_BASE}/api/admin/users/${userId}/modules`, {
    method: "PUT",
    body: JSON.stringify({ module_ids: moduleIds }),
  });
}

export async function getRoles(): Promise<RolesResponse> {
  return fetchWithAuth(`${API_BASE}/api/admin/roles`);
}

export async function getModules(): Promise<ModulesResponse> {
  return fetchWithAuth(`${API_BASE}/api/admin/modules`);
}

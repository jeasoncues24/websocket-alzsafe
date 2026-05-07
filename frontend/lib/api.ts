export const API_BASE =
  process.env.NEXT_PUBLIC_API_URL?.replace(/\/$/, "") ?? "";

if (!API_BASE) {
  throw new Error("NEXT_PUBLIC_API_URL is required in frontend/.env.local");
}

export function buildAdminWsUrl(path: string, token?: string) {
  const url = new URL(API_BASE);
  url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
  url.pathname = path;
  url.search = "";
  if (token) {
    url.searchParams.set("token", token);
  }
  return url.toString();
}

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

type ApiEnvelope = {
  ok?: boolean;
  error?: string;
  message?: string;
  [key: string]: any;
};

async function parseApiBody(res: Response): Promise<ApiEnvelope> {
  const contentType = res.headers.get("content-type") || "";
  if (contentType.includes("application/json")) {
    return res.json();
  }
  const text = await res.text();
  return { ok: res.ok, error: text || undefined, message: text || undefined };
}

async function requestJSON<T = ApiEnvelope>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(url, options);
  const payload: ApiEnvelope = await parseApiBody(res).catch(() => ({ error: "Request failed" }));
  if (!res.ok || payload.ok === false) {
    throw new Error(payload.message || payload.error || "Request failed");
  }
  return payload as T;
}

export async function getMetrics(): Promise<DashboardMetrics> {
  return requestJSON(`${API_BASE}/api/admin/metricas`, {
    headers: authHeaders(),
  }) as Promise<DashboardMetrics>;
}

// ---- Empresa (admin API) ----

export interface Empresa {
  id: number;
  ruc: string;
  nombre: string;
  nombre_comercial?: string;
  telefono_contacto?: string;
  direccion?: string;
  token_version: number;
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

export interface AdminTelefono {
  id: number;
  empresa_id: number;
  codigo_pais: string;
  numero: string;
  numero_completo: string;
  status: string;
  qr_string?: string;
  last_connected?: string | null;
  created_at: string;
  updated_at: string;
  runtime_connected?: boolean;
  mismatch?: boolean;
  mismatch_reason?: string;
}

export interface TelefonosResponse {
  ok: boolean;
  telefonos: AdminTelefono[];
  total: number;
  error?: string;
}

export interface ApiKey {
  id: number;
  empresa_id: number;
  telefono_id: number;
  nombre: string;
  key_prefix: string;
  scopes?: string[];
  activo: boolean;
  created_by_user_id?: number | null;
  created_at: string;
  updated_at: string;
  last_used_at?: string | null;
  expires_at?: string | null;
  revoked_at?: string | null;
  rotated_from_id?: number | null;
}

export interface ApiKeyListResponse {
  ok: boolean;
  api_keys: ApiKey[];
  error?: string;
}

export interface ApiKeyResponse {
  ok: boolean;
  api_key?: ApiKey;
  error?: string;
}

export interface ApiKeyCreateResponse {
  ok: boolean;
  api_key?: ApiKey;
  secret?: string;
  message?: string;
  error?: string;
}

export interface ApiKeyUsageDaily {
  day: string;
  api_key_id: number;
  empresa_id: number;
  telefono_id: number;
  request_count: number;
  success_count: number;
  error_count: number;
  latency_avg_ms: number;
  messages_sent: number;
  broadcasts_sent: number;
  bytes_in: number;
  bytes_out: number;
}

export interface ApiKeyAuditEvent {
  id: number;
  api_key_id: number;
  empresa_id: number;
  telefono_id: number;
  action: string;
  actor_user_id?: number | null;
  metadata?: unknown;
  created_at: string;
}

export interface EmpresaCreateRequest {
  ruc: string;
  nombre: string;
  nombre_comercial?: string;
  telefono_contacto?: string;
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
  return requestJSON(`${API_BASE}/api/admin/empresas?${q}`, {
    headers: authHeaders(),
  }) as Promise<EmpresasListResponse>;
}

export async function getEmpresa(id: number): Promise<EmpresaResponse> {
  return fetchWithAuth(`${API_BASE}/api/admin/empresas/${id}`);
}

export async function getAdminEmpresaTelefonos(
  id: number,
): Promise<TelefonosResponse> {
  return fetchWithAuth(`${API_BASE}/api/admin/empresas/${id}/telefonos`);
}

export interface AdminTelefonoRequest {
  codigo_pais: string;
  numero: string;
  status?: string;
}

export async function createAdminTelefono(
  empresaId: number,
  data: AdminTelefonoRequest,
): Promise<{ ok: boolean; telefono: AdminTelefono; error?: string }> {
  return fetchWithAuth(
    `${API_BASE}/api/admin/empresas/${empresaId}/telefonos`,
    {
      method: "POST",
      body: JSON.stringify(data),
    },
  );
}

export async function updateAdminTelefono(
  telefonoId: number,
  data: Partial<AdminTelefonoRequest>,
): Promise<{ ok: boolean; telefono: AdminTelefono; error?: string }> {
  return fetchWithAuth(`${API_BASE}/api/admin/telefonos/${telefonoId}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export async function deleteAdminTelefono(
  telefonoId: number,
): Promise<{ ok: boolean }> {
  return fetchWithAuth(`${API_BASE}/api/admin/telefonos/${telefonoId}`, {
    method: "DELETE",
  });
}

export async function createEmpresa(
  data: EmpresaCreateRequest,
): Promise<EmpresaResponse> {
  return requestJSON(`${API_BASE}/api/admin/empresas`, {
    method: "POST",
    headers: authHeaders(),
    body: JSON.stringify(data),
  }) as Promise<EmpresaResponse>;
}

export async function updateEmpresa(
  id: number,
  data: Partial<EmpresaCreateRequest>,
): Promise<EmpresaResponse> {
  return requestJSON(`${API_BASE}/api/admin/empresas/${id}`, {
    method: "PUT",
    headers: authHeaders(),
    body: JSON.stringify(data),
  }) as Promise<EmpresaResponse>;
}

export async function deleteEmpresa(id: number): Promise<void> {
  await requestJSON(`${API_BASE}/api/admin/empresas/${id}`, {
    method: "DELETE",
    headers: authHeaders(),
  });
}

export async function restoreEmpresa(
  id: number,
): Promise<{ ok: boolean; empresa: Empresa }> {
  return fetchWithAuth(`${API_BASE}/api/admin/empresas/${id}/restore`, {
    method: "POST",
  });
}

export async function getAdminTelefonoApiKeys(
  telefonoId: number,
): Promise<ApiKeyListResponse> {
  return fetchWithAuth(
    `${API_BASE}/api/admin/telefonos/${telefonoId}/api-keys`,
  );
}

export async function createAdminTelefonoApiKey(
  telefonoId: number,
  data: { nombre: string; scopes: string[]; expires_at?: string },
): Promise<ApiKeyCreateResponse> {
  return fetchWithAuth(
    `${API_BASE}/api/admin/telefonos/${telefonoId}/api-keys`,
    {
      method: "POST",
      body: JSON.stringify(data),
    },
  );
}

export async function getAdminApiKey(id: number): Promise<ApiKeyResponse> {
  return fetchWithAuth(`${API_BASE}/api/admin/api-keys/${id}`);
}

export async function rotateAdminApiKey(
  id: number,
): Promise<ApiKeyCreateResponse> {
  return fetchWithAuth(`${API_BASE}/api/admin/api-keys/${id}/rotate`, {
    method: "POST",
  });
}

export async function revokeAdminApiKey(id: number): Promise<ApiKeyResponse> {
  return fetchWithAuth(`${API_BASE}/api/admin/api-keys/${id}/revoke`, {
    method: "POST",
  });
}

export async function getAdminApiKeyUsage(
  id: number,
): Promise<{ ok: boolean; usage: ApiKeyUsageDaily[] }> {
  return fetchWithAuth(`${API_BASE}/api/admin/api-keys/${id}/usage`);
}

export async function getAdminApiKeyAudit(
  id: number,
): Promise<{ ok: boolean; audit: ApiKeyAuditEvent[] }> {
  return fetchWithAuth(`${API_BASE}/api/admin/api-keys/${id}/audit`);
}

export interface AdminMessage {
  id: number;
  reference_id?: string;
  account_id: string;
  to: string;
  content: string;
  status: string;
  error_reason?: string;
  retry_count?: number;
  adjuntos?: AttachmentInfo[];
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

  return fetchWithAuth(`${API_BASE}/api/admin/mensajes?${params}`);
}

export interface SessionEvent {
  timestamp: string;
  type: string;
  details?: string;
}

export interface SessionSummary {
  total: number;
  active: number;
  disconnected: number;
  mismatch: number;
  qr_pending: number;
  initializing: number;
}

export interface SessionInfo {
  account_id: string;
  status: string;
  qr_string?: string;
  updated_at: string;
  telefono_id?: number;
  empresa_id?: number;
  empresa_nombre?: string;
  runtime_connected?: boolean;
  mismatch?: boolean;
  last_connected?: string;
  events?: SessionEvent[];
}

export interface SessionsResponse {
  sessions: SessionInfo[];
  summary?: SessionSummary;
}

export async function getAdminSessions(): Promise<SessionsResponse> {
  return fetchWithAuth(`${API_BASE}/api/admin/sesiones`);
}

export interface ReconnectSessionResponse {
  ok: boolean;
  status?: string;
  qr_string?: string;
  error?: string;
}

export async function reconnectAdminSession(
  telefonoId: number,
): Promise<ReconnectSessionResponse> {
  return fetchWithAuth(
    `${API_BASE}/api/admin/telefonos/${telefonoId}/connect`,
    { method: "POST" },
  ) as Promise<ReconnectSessionResponse>;
}

export interface MessageRetryResponse {
  ok: boolean;
  reference_id: string;
  estado?: string;
  error?: string;
}

// export async function retryMessage(referenceId: string): Promise<MessageRetryResponse> {
//   const res = await fetchWithAuth(`${API_BASE}/api/mensajes/${referenceId}/reintentar`, {
//     method: "POST",
//   });
//   if (!res.ok) {
//     const data = await res.json();
//     throw new Error(data.error || "Error al reintentar mensaje");
//   }
//   return res.json();
// }

export async function retryMessageAdmin(
  referenceId: string,
): Promise<MessageRetryResponse> {
  return fetchWithAuth(
    `${API_BASE}/api/admin/mensajes/${referenceId}`,
    {
      method: "POST",
    },
  ) as Promise<MessageRetryResponse>;
}

export async function updateMessage(
  referenceId: string,
  data: { contenido?: string; destino?: string },
): Promise<{ ok: boolean; reference_id: string }> {
  return fetchWithAuth(`${API_BASE}/api/mensajes/${referenceId}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  }) as Promise<{ ok: boolean; reference_id: string }>;
}

export interface EmpresaTelefonoSessionData {
  telefono_id: number;
  numeroCompleto: string;
  status: string;
  lastConnected?: string | null;
  qr_string?: string;
  expires_in?: number;
}

export interface EmpresaTelefonoResponse {
  ok: boolean;
  data?: EmpresaTelefonoSessionData;
  message?: string;
  error?: string;
}

async function fetchWithEmpresaAuth<T = ApiEnvelope>(url: string, options?: RequestInit): Promise<T> {
  return requestJSON<T>(url, {
    ...options,
    headers: {
      ...authHeaders(),
      ...options?.headers,
    },
  });
}

export async function getEmpresaTelefono(
  telefonoId: number,
): Promise<EmpresaTelefonoResponse> {
  const json = await fetchWithEmpresaAuth(
    `${API_BASE}/api/admin/telefonos/${telefonoId}`,
  );
  return {
    ok: !!json.ok,
    data: json.telefono
      ? {
          telefono_id: json.telefono.id,
          numeroCompleto: json.telefono.numero_completo,
          status: json.telefono.status,
          lastConnected: json.telefono.last_connected ?? null,
          qr_string: json.telefono.qr_string,
        }
      : undefined,
    message: json.message,
    error: json.error,
  };
}

export async function connectEmpresaTelefono(
  telefonoId: number,
): Promise<EmpresaTelefonoResponse> {
  const json = await fetchWithEmpresaAuth(
    `${API_BASE}/api/admin/telefonos/${telefonoId}/connect`,
    {
      method: "POST",
    },
  );
  return {
    ok: !!json.ok,
    data: json.ok
      ? {
          telefono_id: json.telefono_id,
          numeroCompleto: json.numeroCompleto,
          status: json.status,
          lastConnected: json.lastConnected ?? null,
          qr_string: json.qr_string,
          expires_in: json.expires_in,
        }
      : undefined,
    message: json.message,
    error: json.error,
  };
}

export async function postAdminSession(
  action: string,
  accountId: string,
): Promise<{ status: string }> {
  return fetchWithAuth(`${API_BASE}/api/admin/sesiones`, {
    method: "POST",
    body: JSON.stringify({ account_id: accountId, action }),
  });
}

export interface BroadcastInfo {
  reference_id: string;
  ruc_empresa: string;
  total: number;
  status: string;
  success: number;
  failed: number;
  adjuntos?: AttachmentInfo[];
  created_at: string;
}

export interface AttachmentInfo {
  nombre: string;
  sha256_hash: string;
  tamano_bytes: number;
}

export interface BroadcastsResponse {
  broadcasts: BroadcastInfo[];
}

export async function getAdminBroadcasts(
  accountId?: string,
): Promise<BroadcastsResponse> {
  const params = accountId ? `?account_id=${accountId}` : "";
  return fetchWithAuth(`${API_BASE}/api/admin/difusiones${params}`);
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
  created_at: string;
  updated_at: string;
}

export interface Role {
  id: number;
  name: string;
  description?: string;
  is_root: boolean;
  permissions: string[];
  usage_count?: number;
}

export interface Module {
  id: number;
  name: string;
  slug: string;
  description: string;
}

export interface UsersResponse {
  ok?: boolean;
  users: UserAdminRol[];
  total: number;
}

export interface RolesResponse {
  roles: Role[];
}

export interface ModulesResponse {
  modules: Module[];
}

async function fetchWithAuth<T = ApiEnvelope>(url: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem("admin_token");
  return requestJSON<T>(url, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...options?.headers,
    },
  });
}

function normalizeUserPayload(payload: any): UserAdminRol {
  return payload?.user ?? payload;
}

function normalizeUsersResponse(payload: any): UsersResponse {
  return {
    ok: payload?.ok,
    users: payload?.users ?? [],
    total: payload?.total ?? 0,
  };
}

export async function getUsuarioAdmins(page = 1, limit = 20): Promise<UsersResponse> {
  const payload = await fetchWithAuth(
    `${API_BASE}/api/admin/usuario_admin?page=${page}&limit=${limit}`,
  );
  return normalizeUsersResponse(payload);
}

export async function getUsuarioAdmin(id: number): Promise<UserAdminRol> {
  return normalizeUserPayload(
    await fetchWithAuth(`${API_BASE}/api/admin/usuario_admin/${id}`),
  );
}

export interface CreateUserRequest {
  username: string;
  password: string;
  email: string;
  role_id?: number;
}

export async function createUsuarioAdmin(
  data: CreateUserRequest,
): Promise<UserAdminRol> {
  return normalizeUserPayload(
    await fetchWithAuth(`${API_BASE}/api/admin/usuario_admin`, {
      method: "POST",
      body: JSON.stringify(data),
    }),
  );
}

export interface UpdateUserRequest {
  email?: string;
  role_id?: number;
  is_active?: boolean;
}

export async function updateUsuarioAdmin(
  id: number,
  data: UpdateUserRequest,
): Promise<UserAdminRol> {
  return normalizeUserPayload(
    await fetchWithAuth(`${API_BASE}/api/admin/usuario_admin/${id}`, {
      method: "PUT",
      body: JSON.stringify(data),
    }),
  );
}

export async function deleteUsuarioAdmin(id: number): Promise<{ status: string }> {
  return fetchWithAuth(`${API_BASE}/api/admin/usuario_admin/${id}`, {
    method: "DELETE",
  });
}

export async function promoteUsuarioAdmin(userId: number): Promise<UserAdminRol> {
  return normalizeUserPayload(
    await fetchWithAuth(`${API_BASE}/api/admin/usuario_admin/${userId}/promote`, {
      method: "POST",
    }),
  );
}

export async function getUsuarioAdminModules(userId: number): Promise<{ module_ids: number[] }> {
  return fetchWithAuth(`${API_BASE}/api/admin/usuario_admin/${userId}/modulos`);
}

export async function assignUsuarioAdminModules(
  userId: number,
  moduleIds: number[],
): Promise<void> {
  await fetchWithAuth(`${API_BASE}/api/admin/usuario_admin/${userId}/modulos`, {
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

export interface RoleRequest {
  name: string;
  description?: string;
  is_root?: boolean;
  permissions: string[];
}

export async function createRole(data: RoleRequest): Promise<{ role: Role }> {
  return fetchWithAuth(`${API_BASE}/api/admin/roles`, {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function updateRole(
  id: number,
  data: RoleRequest,
): Promise<{ role: Role }> {
  return fetchWithAuth(`${API_BASE}/api/admin/roles/${id}`, {
    method: "PUT",
    body: JSON.stringify(data),
  });
}

export async function deleteRole(id: number): Promise<{ status: string }> {
  return fetchWithAuth(`${API_BASE}/api/admin/roles/${id}`, {
    method: "DELETE",
  });
}

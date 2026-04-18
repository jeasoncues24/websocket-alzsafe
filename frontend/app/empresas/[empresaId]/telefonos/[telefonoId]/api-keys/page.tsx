"use client";

import { useEffect, useMemo, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import {
  ArrowLeft,
  Building2,
  Check,
  Copy,
  KeyRound,
  Loader2,
  RefreshCw,
  ShieldAlert,
  ShieldCheck,
  Smartphone,
  Activity,
  FileText,
} from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Textarea } from "@/components/ui/textarea";
import {
  createAdminTelefonoApiKey,
  getAdminApiKeyAudit,
  getAdminApiKeyUsage,
  getAdminEmpresaTelefonos,
  getAdminTelefonoApiKeys,
  getEmpresa,
  revokeAdminApiKey,
  rotateAdminApiKey,
  type AdminTelefono,
  type ApiKey,
  type ApiKeyAuditEvent,
  type ApiKeyUsageDaily,
  type Empresa,
} from "@/lib/api";

type ApiKeyAction = "rotate" | "revoke" | null;

function formatDate(value?: string | null) {
  if (!value) return "—";
  return new Date(value).toLocaleString("es-PE", {
    dateStyle: "medium",
    timeStyle: "short",
  });
}

function formatStatus(status: string) {
  switch (status) {
    case "active":
      return "Activo";
    case "qr_pending":
      return "QR pendiente";
    case "disconnected":
      return "Desconectado";
    default:
      return status;
  }
}

export default function PhoneApiKeysPage() {
  const router = useRouter();
  const params = useParams<{ empresaId: string; telefonoId: string }>();
  const empresaId = Number(params?.empresaId);
  const telefonoId = Number(params?.telefonoId);

  const [empresa, setEmpresa] = useState<Empresa | null>(null);
  const [telefono, setTelefono] = useState<AdminTelefono | null>(null);
  const [apiKeys, setApiKeys] = useState<ApiKey[]>([]);
  const [selectedKeyId, setSelectedKeyId] = useState<number | null>(null);
  const [usage, setUsage] = useState<ApiKeyUsageDaily[]>([]);
  const [audit, setAudit] = useState<ApiKeyAuditEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [keysLoading, setKeysLoading] = useState(false);
  const [usageLoading, setUsageLoading] = useState(false);
  const [auditLoading, setAuditLoading] = useState(false);
  const [error, setError] = useState("");
  const [actionError, setActionError] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [actionType, setActionType] = useState<ApiKeyAction>(null);
  const [actionTarget, setActionTarget] = useState<ApiKey | null>(null);
  const [secretOpen, setSecretOpen] = useState(false);
  const [generatedSecret, setGeneratedSecret] = useState("");
  const [generatedKey, setGeneratedKey] = useState<ApiKey | null>(null);
  const [copied, setCopied] = useState(false);
  const [creating, setCreating] = useState(false);
  const [acting, setActing] = useState(false);
  const [createName, setCreateName] = useState("");
  const [createScopes, setCreateScopes] = useState("messages:read\nmessages:write\nbroadcasts:read\nbroadcasts:write");
  const [createExpiresAt, setCreateExpiresAt] = useState("");

  const selectedKey = useMemo(
    () => apiKeys.find((item) => item.id === selectedKeyId) ?? apiKeys[0] ?? null,
    [apiKeys, selectedKeyId],
  );

  useEffect(() => {
    const token = localStorage.getItem("admin_token");
    if (!token) {
      router.push("/login");
      return;
    }

    if (!Number.isFinite(empresaId) || !Number.isFinite(telefonoId)) return;

    let cancelled = false;

    async function load() {
      setLoading(true);
      setError("");

      try {
        const [empresaResp, telefonosResp, keysResp] = await Promise.all([
          getEmpresa(empresaId),
          getAdminEmpresaTelefonos(empresaId),
          getAdminTelefonoApiKeys(telefonoId),
        ]);

        if (cancelled) return;

        const selectedPhone = telefonosResp.telefonos.find((item) => item.id === telefonoId) ?? null;

        setEmpresa(empresaResp.empresa ?? null);
        setTelefono(selectedPhone);
        setApiKeys(keysResp.api_keys ?? []);

        const firstKey = (keysResp.api_keys ?? [])[0] ?? null;
        setSelectedKeyId(firstKey ? firstKey.id : null);
        setGeneratedSecret("");
        setGeneratedKey(null);
        setSecretOpen(false);
      } catch (err: unknown) {
        if (!cancelled) {
          setEmpresa(null);
          setTelefono(null);
          setApiKeys([]);
          setError(err instanceof Error ? err.message : "Error al cargar la vista");
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();

    return () => {
      cancelled = true;
    };
  }, [empresaId, router, telefonoId]);

  useEffect(() => {
    if (!selectedKey?.id) {
      setUsage([]);
      setAudit([]);
      return;
    }

    let cancelled = false;

    async function loadKeyDetails() {
      setUsageLoading(true);
      setAuditLoading(true);

      try {
        const [usageResp, auditResp] = await Promise.all([
          getAdminApiKeyUsage(selectedKey.id),
          getAdminApiKeyAudit(selectedKey.id),
        ]);

        if (cancelled) return;

        setUsage(usageResp.usage ?? []);
        setAudit(auditResp.audit ?? []);
      } catch {
        if (!cancelled) {
          setUsage([]);
          setAudit([]);
        }
      } finally {
        if (!cancelled) {
          setUsageLoading(false);
          setAuditLoading(false);
        }
      }
    }

    loadKeyDetails();

    return () => {
      cancelled = true;
    };
  }, [selectedKey?.id]);

  async function refreshKeys(nextSelectedKeyId?: number | null) {
    setKeysLoading(true);
    try {
      const resp = await getAdminTelefonoApiKeys(telefonoId);
      const keys = resp.api_keys ?? [];
      setApiKeys(keys);

      const targetId = nextSelectedKeyId ?? selectedKeyId;
      const found = keys.find((item) => item.id === targetId) ?? keys[0] ?? null;
      setSelectedKeyId(found ? found.id : null);
    } catch (err: unknown) {
      setActionError(err instanceof Error ? err.message : "Error al recargar keys");
    } finally {
      setKeysLoading(false);
    }
  }

  function openCreate() {
    setActionError("");
    setCreateName(`API key ${telefono?.numero_completo ?? telefonoId}`);
    setCreateScopes("messages:read\nmessages:write\nbroadcasts:read\nbroadcasts:write");
    setCreateExpiresAt("");
    setCreateOpen(true);
  }

  async function handleCreate() {
    setCreating(true);
    setActionError("");
    try {
      const scopes = createScopes
        .split("\n")
        .map((item) => item.trim())
        .filter(Boolean);

      const payload: { nombre: string; scopes: string[]; expires_at?: string } = {
        nombre: createName.trim() || `API key ${telefono?.numero_completo ?? telefonoId}`,
        scopes,
      };

      if (createExpiresAt) {
        payload.expires_at = new Date(createExpiresAt).toISOString();
      }

      const resp = await createAdminTelefonoApiKey(telefonoId, payload);
      if (resp.api_key) {
        setGeneratedKey(resp.api_key);
        setGeneratedSecret(resp.secret ?? "");
        setSecretOpen(true);
      }
      setCreateOpen(false);
      await refreshKeys(resp.api_key?.id);
    } catch (err: unknown) {
      setActionError(err instanceof Error ? err.message : "No se pudo crear la API key");
    } finally {
      setCreating(false);
    }
  }

  function askAction(type: ApiKeyAction, key: ApiKey) {
    setActionType(type);
    setActionTarget(key);
    setConfirmOpen(true);
    setActionError("");
  }

  async function handleAction() {
    if (!actionType || !actionTarget) return;
    setActing(true);
    setActionError("");

    try {
      if (actionType === "rotate") {
        const resp = await rotateAdminApiKey(actionTarget.id);
        if (resp.api_key) {
          setGeneratedKey(resp.api_key);
          setGeneratedSecret(resp.secret ?? "");
          setSecretOpen(true);
        }
        await refreshKeys(resp.api_key?.id);
      }

      if (actionType === "revoke") {
        await revokeAdminApiKey(actionTarget.id);
        await refreshKeys(selectedKeyId === actionTarget.id ? null : selectedKeyId);
      }

      setConfirmOpen(false);
    } catch (err: unknown) {
      setActionError(err instanceof Error ? err.message : "No se pudo completar la acción");
    } finally {
      setActing(false);
    }
  }

  async function copySecret() {
    if (!generatedSecret) return;
    await navigator.clipboard.writeText(generatedSecret);
    setCopied(true);
  }

  const selectedPhoneTitle = telefono ? telefono.numero_completo : `${telefonoId}`;

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div className="space-y-2">
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Building2 className="h-4 w-4" />
            <span>{empresa?.nombre ?? "Empresa"}</span>
            <span>•</span>
            <Smartphone className="h-4 w-4" />
            <span>{selectedPhoneTitle}</span>
          </div>
          <h1 className="text-3xl font-bold tracking-tight">API Keys del teléfono</h1>
          <p className="text-muted-foreground">
            Administra creación, rotación y revocación con foco en el número WhatsApp.
          </p>
        </div>

        <div className="flex flex-wrap gap-2">
          <Button variant="outline" onClick={() => router.push(`/empresas/${empresaId}/telefonos`)}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            Volver a teléfonos
          </Button>
          <Button variant="outline" onClick={() => refreshKeys(selectedKeyId)} disabled={keysLoading || loading}>
            {keysLoading || loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <RefreshCw className="mr-2 h-4 w-4" />}
            Recargar
          </Button>
          <Button onClick={openCreate} disabled={!telefono}>
            <KeyRound className="mr-2 h-4 w-4" />
            Nueva API key
          </Button>
        </div>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}
      {actionError && <p className="text-sm text-destructive">{actionError}</p>}

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader>
            <CardDescription>Empresa</CardDescription>
            <CardTitle>{empresa?.nombre ?? (loading ? "Cargando..." : "—")}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader>
            <CardDescription>Teléfono</CardDescription>
            <CardTitle>{telefono?.numero_completo ?? selectedPhoneTitle}</CardTitle>
          </CardHeader>
        </Card>
        <Card>
          <CardHeader>
            <CardDescription>Estado</CardDescription>
            <CardTitle>
              <Badge variant={telefono?.status === "active" ? "default" : "secondary"}>
                {telefono ? formatStatus(telefono.status) : "—"}
              </Badge>
            </CardTitle>
          </CardHeader>
        </Card>
      </div>

      <Tabs defaultValue="keys" className="space-y-4">
        <TabsList className="grid w-full grid-cols-3">
          <TabsTrigger value="keys">Claves</TabsTrigger>
          <TabsTrigger value="usage">Uso</TabsTrigger>
          <TabsTrigger value="audit">Auditoría</TabsTrigger>
        </TabsList>

        <TabsContent value="keys" className="space-y-4">
          <div className="grid gap-4 xl:grid-cols-[1.7fr_1fr]">
            <Card>
              <CardHeader>
                <CardTitle>Keys del teléfono</CardTitle>
                <CardDescription>
                  Cada key se muestra una sola vez al crearla o rotarla.
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-3">
                {loading ? (
                  <div className="space-y-3">
                    <Skeleton className="h-20 w-full" />
                    <Skeleton className="h-20 w-full" />
                    <Skeleton className="h-20 w-full" />
                  </div>
                ) : apiKeys.length === 0 ? (
                  <div className="rounded-xl border border-dashed p-8 text-center text-sm text-muted-foreground">
                    No hay API keys para este teléfono.
                  </div>
                ) : (
                  apiKeys.map((key) => (
                    <div
                      key={key.id}
                      onClick={() => setSelectedKeyId(key.id)}
                      role="button"
                      tabIndex={0}
                      className={`w-full rounded-xl border p-4 text-left transition ${
                        selectedKey?.id === key.id ? "border-primary bg-primary/5" : "hover:bg-muted/40"
                      }`}
                    >
                      <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                        <div className="space-y-2">
                          <div className="flex flex-wrap items-center gap-2">
                            <span className="font-medium">{key.nombre}</span>
                            <Badge variant={key.activo ? "default" : "secondary"}>
                              {key.activo ? "Activa" : "Revocada"}
                            </Badge>
                            <Badge variant="outline" className="font-mono">
                              {key.key_prefix}
                            </Badge>
                          </div>
                          <div className="grid gap-2 text-sm text-muted-foreground md:grid-cols-2">
                            <span>Creada: {formatDate(key.created_at)}</span>
                            <span>Último uso: {formatDate(key.last_used_at)}</span>
                            <span>Expira: {formatDate(key.expires_at)}</span>
                            <span>Rotada desde: {key.rotated_from_id ?? "—"}</span>
                          </div>
                        </div>

                        <div className="flex flex-wrap gap-2 md:justify-end">
                          <Button variant="outline" size="sm" onClick={() => askAction("rotate", key)}>
                            <RefreshCw className="mr-2 h-4 w-4" />
                            Rotar
                          </Button>
                          <Button
                            variant="destructive"
                            size="sm"
                            onClick={() => askAction("revoke", key)}
                            disabled={!key.activo}
                          >
                            <ShieldAlert className="mr-2 h-4 w-4" />
                            Revocar
                          </Button>
                        </div>
                      </div>
                    </div>
                  ))
                )}
              </CardContent>
            </Card>

            <Card>
              <CardHeader>
                <CardTitle>Detalle de la key</CardTitle>
                <CardDescription>
                  Estado actual y metadatos de la API key seleccionada.
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4 text-sm">
                {selectedKey ? (
                  <>
                    <div className="rounded-lg border bg-muted/20 p-3">
                      <div className="text-xs text-muted-foreground">Nombre</div>
                      <div className="mt-1 font-medium">{selectedKey.nombre}</div>
                    </div>
                    <div className="grid grid-cols-2 gap-3">
                      <div className="rounded-lg border bg-muted/20 p-3">
                        <div className="text-xs text-muted-foreground">Prefijo</div>
                        <div className="mt-1 font-mono text-xs">{selectedKey.key_prefix}</div>
                      </div>
                      <div className="rounded-lg border bg-muted/20 p-3">
                        <div className="text-xs text-muted-foreground">Estado</div>
                        <div className="mt-1">
                          <Badge variant={selectedKey.activo ? "default" : "secondary"}>
                            {selectedKey.activo ? "Activa" : "Revocada"}
                          </Badge>
                        </div>
                      </div>
                      <div className="rounded-lg border bg-muted/20 p-3">
                        <div className="text-xs text-muted-foreground">Creada</div>
                        <div className="mt-1">{formatDate(selectedKey.created_at)}</div>
                      </div>
                      <div className="rounded-lg border bg-muted/20 p-3">
                        <div className="text-xs text-muted-foreground">Último uso</div>
                        <div className="mt-1">{formatDate(selectedKey.last_used_at)}</div>
                      </div>
                    </div>
                    <div className="rounded-lg border bg-muted/20 p-3">
                      <div className="text-xs text-muted-foreground">Scopes</div>
                      <div className="mt-2 flex flex-wrap gap-2">
                        {(selectedKey.scopes ?? []).length > 0 ? (
                          selectedKey.scopes?.map((scope) => (
                            <Badge key={scope} variant="outline">
                              {scope}
                            </Badge>
                          ))
                        ) : (
                          <span className="text-muted-foreground">Sin scopes definidos</span>
                        )}
                      </div>
                    </div>
                  </>
                ) : (
                  <div className="rounded-xl border border-dashed p-6 text-center text-muted-foreground">
                    Selecciona una key para ver sus detalles.
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
        </TabsContent>

        <TabsContent value="usage">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Activity className="h-4 w-4" />
                Uso diario
              </CardTitle>
              <CardDescription>Métricas agregadas para la key seleccionada.</CardDescription>
            </CardHeader>
            <CardContent>
              {usageLoading ? (
                <div className="space-y-3">
                  <Skeleton className="h-16 w-full" />
                  <Skeleton className="h-16 w-full" />
                </div>
              ) : usage.length === 0 ? (
                <div className="rounded-xl border border-dashed p-8 text-center text-sm text-muted-foreground">
                  Sin datos de uso todavía.
                </div>
              ) : (
                <div className="space-y-3">
                  {usage.map((row) => (
                    <div key={`${row.day}-${row.api_key_id}`} className="rounded-xl border p-4 text-sm">
                      <div className="flex flex-wrap items-center justify-between gap-2">
                        <div className="font-medium">{row.day}</div>
                        <Badge variant="outline">{row.request_count} requests</Badge>
                      </div>
                      <div className="mt-3 grid gap-2 text-muted-foreground md:grid-cols-4">
                        <span>Éxitos: {row.success_count}</span>
                        <span>Errores: {row.error_count}</span>
                        <span>Latencia: {row.latency_avg_ms} ms</span>
                        <span>Mensajes: {row.messages_sent}</span>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="audit">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <FileText className="h-4 w-4" />
                Auditoría
              </CardTitle>
              <CardDescription>Eventos de creación, rotación y revocación.</CardDescription>
            </CardHeader>
            <CardContent>
              {auditLoading ? (
                <div className="space-y-3">
                  <Skeleton className="h-16 w-full" />
                  <Skeleton className="h-16 w-full" />
                </div>
              ) : audit.length === 0 ? (
                <div className="rounded-xl border border-dashed p-8 text-center text-sm text-muted-foreground">
                  Sin eventos de auditoría.
                </div>
              ) : (
                <div className="space-y-3">
                  {audit.map((item) => (
                    <div key={item.id} className="rounded-xl border p-4 text-sm">
                      <div className="flex flex-wrap items-center justify-between gap-2">
                        <div className="flex items-center gap-2">
                          {item.action === "created" ? (
                            <ShieldCheck className="h-4 w-4 text-green-600" />
                          ) : item.action === "rotated" ? (
                            <RefreshCw className="h-4 w-4 text-blue-600" />
                          ) : (
                            <ShieldAlert className="h-4 w-4 text-red-600" />
                          )}
                          <span className="font-medium">{item.action}</span>
                        </div>
                        <span className="text-muted-foreground">{formatDate(item.created_at)}</span>
                      </div>
                      <p className="mt-2 text-muted-foreground">
                        Actor: {item.actor_user_id ?? "sistema"}
                      </p>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>

      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>Nueva API key</DialogTitle>
            <DialogDescription>
              Se generará un secreto visible una sola vez para este teléfono.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium">Nombre</label>
              <Input value={createName} onChange={(e) => setCreateName(e.target.value)} placeholder="Integración principal" />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Scopes</label>
              <Textarea
                value={createScopes}
                onChange={(e) => setCreateScopes(e.target.value)}
                className="min-h-32 font-mono text-xs"
                placeholder="messages:read\nmessages:write"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium">Expiración opcional</label>
              <Input
                type="datetime-local"
                value={createExpiresAt}
                onChange={(e) => setCreateExpiresAt(e.target.value)}
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateOpen(false)}>
              Cancelar
            </Button>
            <Button onClick={handleCreate} disabled={creating}>
              {creating ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
              Crear key
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={confirmOpen} onOpenChange={setConfirmOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>
              {actionType === "rotate" ? "Rotar API key" : "Revocar API key"}
            </DialogTitle>
            <DialogDescription>
              {actionType === "rotate"
                ? "La key anterior quedará inválida y se mostrará una nueva una sola vez."
                : "Esta acción corta el acceso de la integración asociada inmediatamente."}
            </DialogDescription>
          </DialogHeader>

          <DialogFooter>
            <Button variant="outline" onClick={() => setConfirmOpen(false)}>
              Cancelar
            </Button>
            <Button variant="destructive" onClick={handleAction} disabled={acting}>
              {acting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : null}
              Confirmar
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={secretOpen} onOpenChange={setSecretOpen}>
        <DialogContent className="sm:max-w-xl">
          <DialogHeader>
            <DialogTitle>Secreto generado</DialogTitle>
            <DialogDescription>
              Guarda este secreto ahora. No volverá a mostrarse.
            </DialogDescription>
          </DialogHeader>

          <Alert>
            <ShieldAlert className="h-4 w-4" />
            <AlertDescription>
              {generatedKey?.nombre} · {generatedKey?.key_prefix}
            </AlertDescription>
          </Alert>

          <div className="space-y-2">
            <label className="text-sm font-medium">API key</label>
            <Input readOnly value={generatedSecret} className="font-mono text-xs" />
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={copySecret} disabled={!generatedSecret}>
              {copied ? <Check className="mr-2 h-4 w-4" /> : <Copy className="mr-2 h-4 w-4" />}
              {copied ? "Copiada" : "Copiar"}
            </Button>
            <Button onClick={() => setSecretOpen(false)}>Cerrar</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}

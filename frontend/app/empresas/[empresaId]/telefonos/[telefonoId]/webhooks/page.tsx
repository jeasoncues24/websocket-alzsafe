"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, ChevronDown, ChevronRight, Copy, Loader2, Phone, RefreshCw, Webhook } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { DataEmptyState } from "@/components/feedback/data-empty-state";
import {
  getAdminEmpresaTelefonos,
  getAdminTelefonoWebhooks,
  type AdminTelefono,
  type AdminWebhook,
} from "@/lib/api";

const EVENTO_LABELS: Record<string, string> = {
  "message.received": "Mensaje recibido",
  "message.status_update": "Estado de mensaje",
  "session.connected": "Sesión conectada",
  "session.disconnected": "Sesión desconectada",
};

function etiquetaEvento(ev: string): string {
  return EVENTO_LABELS[ev] ?? ev;
}

function etiquetaEstadoWebhook(wh: AdminWebhook): { label: string; variant: "default" | "secondary" | "destructive" } {
  if (!wh.activo) return { label: "Inactivo", variant: "secondary" };
  if (wh.failure_count > 0) return { label: "Con fallos", variant: "destructive" };
  return { label: "Activo", variant: "default" };
}

function tiempoRelativo(fecha?: string | null): string {
  if (!fecha) return "Nunca";
  const diff = Date.now() - new Date(fecha).getTime();
  const min = Math.floor(diff / 60000);
  if (min < 1) return "hace un momento";
  if (min < 60) return `hace ${min} min`;
  const hrs = Math.floor(min / 60);
  if (hrs < 24) return `hace ${hrs}h`;
  return `hace ${Math.floor(hrs / 24)} días`;
}

function truncarUrl(url: string, max = 40): string {
  return url.length > max ? url.slice(0, max) + "..." : url;
}

function formatDate(value?: string | null) {
  if (!value) return "—";
  return new Date(value).toLocaleDateString("es-PE", { dateStyle: "medium" });
}

export default function TelefonoWebhooksPage() {
  const router = useRouter();
  const params = useParams<{ empresaId: string; telefonoId: string }>();
  const empresaId = Number(params?.empresaId);
  const telefonoId = Number(params?.telefonoId);

  const [telefono, setTelefono] = useState<AdminTelefono | null>(null);
  const [webhooks, setWebhooks] = useState<AdminWebhook[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [expanded, setExpanded] = useState<Set<number>>(new Set());
  const [copied, setCopied] = useState<number | null>(null);

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
        const [telefonosResp, webhooksResp] = await Promise.all([
          getAdminEmpresaTelefonos(empresaId),
          getAdminTelefonoWebhooks(telefonoId),
        ]);
        if (cancelled) return;
        const found = telefonosResp.telefonos?.find((t) => t.id === telefonoId) ?? null;
        setTelefono(found);
        setWebhooks(webhooksResp.webhooks ?? []);
      } catch (err: unknown) {
        if (!cancelled) setError(err instanceof Error ? err.message : "Error cargando webhooks");
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();
    return () => { cancelled = true; };
  }, [empresaId, telefonoId, router]);

  function toggle(id: number) {
    setExpanded((prev) => {
      const next = new Set(prev);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });
  }

  async function copyUrl(id: number, url: string) {
    await navigator.clipboard.writeText(url);
    setCopied(id);
    setTimeout(() => setCopied(null), 1500);
  }

  const activos = webhooks.filter((w) => w.activo && w.failure_count === 0).length;
  const inactivos = webhooks.filter((w) => !w.activo).length;
  const conFallos = webhooks.filter((w) => w.activo && w.failure_count > 0).length;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between gap-3">
        <div>
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Phone className="h-4 w-4" />
            <span>{telefono?.numero_completo ?? "Cargando..."}</span>
          </div>
          <h1 className="text-3xl font-bold tracking-tight">Webhooks del teléfono</h1>
          <p className="text-sm text-muted-foreground">
            Solo lectura — los webhooks son creados y gestionados por los integradores a través de la API.
          </p>
        </div>

        <div className="flex flex-wrap gap-2">
          <Button
            variant="outline"
            onClick={() => router.push(`/empresas/${empresaId}/telefonos`)}
          >
            <ArrowLeft className="mr-2 h-4 w-4" />
            Volver
          </Button>
          <Button variant="outline" onClick={() => router.refresh()} disabled={loading}>
            {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <RefreshCw className="mr-2 h-4 w-4" />}
            Recargar
          </Button>
        </div>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      <Alert>
        <AlertDescription>
          Solo lectura — los webhooks son creados y gestionados por los integradores a través de la API.
        </AlertDescription>
      </Alert>

      {/* Fila bento resumen */}
      {loading ? (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          {Array(3).fill(0).map((_, i) => (
            <Card key={i}>
              <CardHeader className="pb-2"><Skeleton className="h-4 w-24" /></CardHeader>
              <CardContent><Skeleton className="h-8 w-12" /></CardContent>
            </Card>
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Activos</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-3xl font-bold">{activos}</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Inactivos</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-3xl font-bold text-muted-foreground">{inactivos}</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">Con fallos</CardTitle>
            </CardHeader>
            <CardContent className="flex items-center gap-2">
              <p className="text-3xl font-bold">{conFallos}</p>
              {conFallos > 0 && <Badge variant="destructive">Atención</Badge>}
            </CardContent>
          </Card>
        </div>
      )}

      {/* Tabla */}
      {loading ? (
        <Card>
          <CardContent className="pt-6 space-y-3">
            {Array(3).fill(0).map((_, i) => <Skeleton key={i} className="h-12 w-full" />)}
          </CardContent>
        </Card>
      ) : webhooks.length === 0 ? (
        <DataEmptyState
          icon={Webhook}
          title="Sin webhooks"
          description="Este teléfono no tiene webhooks registrados. Los webhooks son creados por los integradores a través de la API."
        />
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-8" />
                  <TableHead>URL</TableHead>
                  <TableHead>Eventos</TableHead>
                  <TableHead>Estado</TableHead>
                  <TableHead>Fallos</TableHead>
                  <TableHead>Último éxito</TableHead>
                  <TableHead>Registrado</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {webhooks.map((wh) => {
                  const estado = etiquetaEstadoWebhook(wh);
                  const isExpanded = expanded.has(wh.id);
                  return (
                    <>
                      <TableRow
                        key={wh.id}
                        className="cursor-pointer hover:bg-muted/50"
                        onClick={() => toggle(wh.id)}
                      >
                        <TableCell>
                          {isExpanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
                        </TableCell>
                        <TableCell className="font-mono text-xs" title={wh.url}>
                          {truncarUrl(wh.url)}
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-wrap gap-1">
                            {wh.eventos.map((ev) => (
                              <Badge key={ev} variant="secondary" className="text-xs">
                                {etiquetaEvento(ev)}
                              </Badge>
                            ))}
                          </div>
                        </TableCell>
                        <TableCell>
                          <Badge variant={estado.variant}>{estado.label}</Badge>
                        </TableCell>
                        <TableCell className={wh.failure_count > 0 ? "font-semibold text-destructive" : ""}>
                          {wh.failure_count}
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {tiempoRelativo(wh.last_success_at)}
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {formatDate(wh.created_at)}
                        </TableCell>
                      </TableRow>
                      {isExpanded && (
                        <TableRow key={`${wh.id}-detail`}>
                          <TableCell colSpan={7} className="bg-muted/30 p-4">
                            <div className="space-y-3 text-sm">
                              <div>
                                <p className="mb-1 font-medium text-muted-foreground">URL completa</p>
                                <div className="flex items-center gap-2">
                                  <code className="break-all rounded bg-muted px-2 py-1 text-xs">{wh.url}</code>
                                  <Button
                                    size="sm"
                                    variant="ghost"
                                    className="shrink-0"
                                    onClick={(e) => { e.stopPropagation(); copyUrl(wh.id, wh.url); }}
                                  >
                                    <Copy className="h-3.5 w-3.5" />
                                    {copied === wh.id ? " Copiado" : ""}
                                  </Button>
                                </div>
                              </div>
                              <div>
                                <p className="mb-1 font-medium text-muted-foreground">API key asociada</p>
                                <span className="text-xs">API key #{wh.api_key_id}</span>
                              </div>
                              <div>
                                <p className="mb-1 font-medium text-muted-foreground">Eventos suscritos</p>
                                <div className="flex flex-wrap gap-1">
                                  {wh.eventos.map((ev) => (
                                    <Badge key={ev} variant="secondary" className="text-xs">
                                      {etiquetaEvento(ev)}
                                    </Badge>
                                  ))}
                                </div>
                              </div>
                              {wh.last_error && (
                                <div>
                                  <p className="mb-1 font-medium text-muted-foreground">Último error</p>
                                  <code className="break-all rounded bg-destructive/10 px-2 py-1 text-xs text-destructive">
                                    {wh.last_error}
                                  </code>
                                </div>
                              )}
                            </div>
                          </TableCell>
                        </TableRow>
                      )}
                    </>
                  );
                })}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </div>
  );
}

"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, Building2, KeyRound, Link, Loader2, Pencil, Phone, Plus, QrCode, RefreshCw, Trash2, Webhook } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { cn } from "@/lib/utils";
import { DataEmptyState } from "@/components/feedback/data-empty-state";
import {
  createAdminTelefono,
  deleteAdminTelefono,
  getAdminEmpresaTelefonos,
  getEmpresa,
  updateAdminTelefono,
  type AdminTelefono,
  type Empresa,
} from "@/lib/api";
import { TelefonoFormModal, type TelefonoFormData } from "@/components/companies/telefono-form-modal";

type BadgeVariant = "default" | "secondary" | "destructive" | "outline";

function estadoTelefono(t: AdminTelefono): { label: string; variant: BadgeVariant; className?: string } {
  if (t.mismatch) return { label: "Desajuste", variant: "outline", className: "text-amber-600 border-amber-400" };
  if (t.status === "active" && t.runtime_connected) return { label: "Conectado", variant: "default" };
  if (t.status === "disconnected" || !t.runtime_connected) return { label: "Desconectado", variant: "destructive" };
  return { label: "En espera", variant: "secondary" };
}

export default function CompanyPhonesPage() {
  const router = useRouter();
  const params = useParams<{ empresaId: string }>();
  const empresaId = Number(params?.empresaId);

  const [empresa, setEmpresa] = useState<Empresa | null>(null);
  const [telefonos, setTelefonos] = useState<AdminTelefono[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [mutationError, setMutationError] = useState("");
  const [formOpen, setFormOpen] = useState(false);
  const [editingTelefono, setEditingTelefono] = useState<AdminTelefono | null>(null);

  useEffect(() => {
    const token = localStorage.getItem("admin_token");
    if (!token) {
      router.push("/login");
      return;
    }

    if (!Number.isFinite(empresaId) || empresaId <= 0) return;

    let cancelled = false;

    async function load() {
      setLoading(true);
      setError("");

      try {
        const [empresaResp, telefonosResp] = await Promise.all([
          getEmpresa(empresaId),
          getAdminEmpresaTelefonos(empresaId),
        ]);

        if (cancelled) return;

        setEmpresa(empresaResp.empresa ?? null);
        setTelefonos(telefonosResp.telefonos ?? []);
      } catch (err: unknown) {
        if (!cancelled) {
          setEmpresa(null);
          setTelefonos([]);
          setError(err instanceof Error ? err.message : "Error cargando teléfonos");
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();

    return () => {
      cancelled = true;
    };
  }, [empresaId, router]);

  async function reload() {
    const [empresaResp, telefonosResp] = await Promise.all([
      getEmpresa(empresaId),
      getAdminEmpresaTelefonos(empresaId),
    ]);
    setEmpresa(empresaResp.empresa ?? null);
    setTelefonos(telefonosResp.telefonos ?? []);
  }

  function openCreate() {
    setMutationError("");
    setEditingTelefono(null);
    setFormOpen(true);
  }

  function openEdit(telefono: AdminTelefono) {
    setMutationError("");
    setEditingTelefono(telefono);
    setFormOpen(true);
  }

  async function saveTelefono(data: TelefonoFormData) {
    setMutationError("");
    try {
      if (editingTelefono) {
        await updateAdminTelefono(editingTelefono.id, data);
      } else {
        await createAdminTelefono(empresaId, data);
      }
      await reload();
    } catch (err: unknown) {
      setMutationError(err instanceof Error ? err.message : "Error guardando teléfono");
      throw err;
    }
  }

  async function handleDelete(telefono: AdminTelefono) {
    if (!confirm(`¿Eliminar el teléfono ${telefono.numero_completo}?`)) return;
    setMutationError("");
    try {
      await deleteAdminTelefono(telefono.id);
      await reload();
    } catch (err: unknown) {
      setMutationError(err instanceof Error ? err.message : "Error eliminando teléfono");
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between gap-3">
        <div>
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <Building2 className="h-4 w-4" />
            <span>Teléfonos por empresa</span>
          </div>
          <h1 className="text-3xl font-bold tracking-tight">
            {empresa ? empresa.nombre : "Cargando empresa..."}
          </h1>
          <p className="text-muted-foreground">
            Selecciona un teléfono para gestionar sus API keys y consumo.
          </p>
        </div>

        <div className="flex flex-wrap gap-2">
          <Button onClick={openCreate}>
            <Plus className="mr-2 h-4 w-4" />
            Nuevo teléfono
          </Button>
          <Button variant="outline" onClick={() => router.push("/empresas")}>
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
      {mutationError && <p className="text-sm text-destructive">{mutationError}</p>}

      {loading ? (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          {Array(3).fill(0).map((_, i) => (
            <Card key={i}>
              <CardHeader className="pb-3">
                <Skeleton className="h-5 w-40" />
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="flex gap-4">
                  <Skeleton className="h-4 w-24" />
                  <Skeleton className="h-4 w-24" />
                </div>
                <div className="flex gap-2">
                  <Skeleton className="h-8 w-20" />
                  <Skeleton className="h-8 w-20" />
                  <Skeleton className="h-8 w-24" />
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      ) : telefonos.length === 0 ? (
        <DataEmptyState
          icon={Phone}
          title="Sin teléfonos"
          description="No hay teléfonos registrados para esta empresa."
        />
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          {telefonos.map((telefono) => {
            const estado = estadoTelefono(telefono);
            const isLarge = telefono.status === "active" && telefono.runtime_connected && (telefono.webhook_count ?? 0) > 0;

            return (
              <Card
                key={telefono.id}
                className={cn(
                  "transition-shadow hover:shadow-md",
                  isLarge && "md:col-span-2"
                )}
              >
                <CardHeader className="pb-3">
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex items-center gap-2 min-w-0">
                      <Phone className="h-4 w-4 shrink-0 text-muted-foreground" />
                      <CardTitle className="truncate text-base font-semibold">
                        {telefono.numero_completo}
                      </CardTitle>
                    </div>
                    <Badge variant={estado.variant} className={estado.className}>
                      {estado.label}
                    </Badge>
                  </div>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="flex gap-4 text-sm text-muted-foreground">
                    <span className="flex items-center gap-1">
                      <KeyRound className="h-3.5 w-3.5" />
                      {telefono.api_key_count ?? 0} claves activas
                    </span>
                    <span className="flex items-center gap-1">
                      <Webhook className="h-3.5 w-3.5" />
                      {telefono.webhook_count ?? 0} webhooks
                    </span>
                  </div>

                  <div className="flex flex-wrap gap-2">
                    <Button size="sm" variant="outline" onClick={() => openEdit(telefono)}>
                      <Pencil className="mr-1.5 h-3.5 w-3.5" />
                      Editar
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => router.push(`/empresas/${empresaId}/telefonos/${telefono.id}/api-keys`)}
                    >
                      <KeyRound className="mr-1.5 h-3.5 w-3.5" />
                      API Keys
                    </Button>
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => router.push(`/empresas/${empresaId}/telefonos/${telefono.id}/webhooks`)}
                    >
                      <Link className="mr-1.5 h-3.5 w-3.5" />
                      Webhooks
                    </Button>
                    {telefono.status !== "active" && (
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => router.push(`/empresas/${empresaId}/telefonos/${telefono.id}/connect`)}
                      >
                        <QrCode className="mr-1.5 h-3.5 w-3.5" />
                        {telefono.status === "disconnected" ? "Conectar" : "Ver QR"}
                      </Button>
                    )}
                    <Button size="sm" variant="destructive" onClick={() => handleDelete(telefono)}>
                      <Trash2 className="mr-1.5 h-3.5 w-3.5" />
                      Eliminar
                    </Button>
                  </div>
                </CardContent>
              </Card>
            );
          })}
        </div>
      )}

      <TelefonoFormModal
        open={formOpen}
        onClose={() => setFormOpen(false)}
        onSave={saveTelefono}
        telefono={editingTelefono}
      />
    </div>
  );
}

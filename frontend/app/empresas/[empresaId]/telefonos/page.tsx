"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, Building2, KeyRound, Loader2, Phone, Pencil, Plus, RefreshCw, Trash2 } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
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

      <Card>
        <CardHeader>
          <CardTitle>Lista de teléfonos</CardTitle>
          <CardDescription>
            Cada número puede tener una o varias API keys asociadas.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {loading ? (
            <div className="space-y-3">
              <Skeleton className="h-16 w-full" />
              <Skeleton className="h-16 w-full" />
              <Skeleton className="h-16 w-full" />
            </div>
          ) : telefonos.length === 0 ? (
            <div className="rounded-lg border border-dashed p-8 text-center text-sm text-muted-foreground">
              No hay teléfonos registrados para esta empresa.
            </div>
          ) : (
            telefonos.map((telefono) => (
              <div
                key={telefono.id}
                className="flex flex-col gap-4 rounded-xl border bg-card p-4 md:flex-row md:items-center md:justify-between"
              >
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <Phone className="h-4 w-4 text-muted-foreground" />
                    <span className="font-medium">{telefono.numero_completo}</span>
                    <Badge variant={telefono.status === "active" ? "default" : "secondary"}>
                      {formatStatus(telefono.status)}
                    </Badge>
                  </div>
                  <p className="text-sm text-muted-foreground">
                    {telefono.codigo_pais} {telefono.numero}
                  </p>
                </div>

                <div className="flex flex-wrap gap-2">
                  <Button variant="outline" onClick={() => openEdit(telefono)}>
                    <Pencil className="mr-2 h-4 w-4" />
                    Editar
                  </Button>
                  <Button variant="destructive" onClick={() => handleDelete(telefono)}>
                    <Trash2 className="mr-2 h-4 w-4" />
                    Eliminar
                  </Button>
                  <Button
                    variant="outline"
                    onClick={() => router.push(`/empresas/${empresaId}/telefonos/${telefono.id}/api-keys`)}
                  >
                    <KeyRound className="mr-2 h-4 w-4" />
                    Gestionar API Keys
                  </Button>
                </div>
              </div>
            ))
          )}
        </CardContent>
      </Card>

      <TelefonoFormModal
        open={formOpen}
        onClose={() => setFormOpen(false)}
        onSave={saveTelefono}
        telefono={editingTelefono}
      />
    </div>
  );
}

"use client";

import { useState, useEffect } from "react";
import { Loader2, Sparkles } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { buscarClientePorDocumento, type Empresa, type EmpresaCreateRequest } from "@/lib/api";

interface Props {
  open: boolean;
  onClose: () => void;
  onSave: (data: EmpresaCreateRequest) => Promise<void>;
  empresa?: Empresa | null;
}

const EMPTY: EmpresaCreateRequest = {
  ruc: "",
  nombre: "",
  nombre_comercial: "",
  telefono_contacto: "",
  direccion: "",
};

export function EmpresaFormModal({ open, onClose, onSave, empresa }: Props) {
  const [form, setForm] = useState<EmpresaCreateRequest>(EMPTY);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");
  const [lookupLoading, setLookupLoading] = useState(false);
  const [lookupMessage, setLookupMessage] = useState("");
  const [autofillFlash, setAutofillFlash] = useState(false);
  const [docLocked, setDocLocked] = useState(false);

  useEffect(() => {
    if (empresa) {
        setForm({
          ruc: empresa.ruc,
          nombre: empresa.nombre,
          nombre_comercial: empresa.nombre_comercial ?? "",
          telefono_contacto: empresa.telefono_contacto ?? "",
          direccion: empresa.direccion ?? "",
        });
    } else {
      setForm(EMPTY);
    }
    setError("");
    setLookupMessage("");
    setLookupLoading(false);
    setAutofillFlash(false);
    setDocLocked(false);
  }, [empresa, open]);

  useEffect(() => {
    if (!open || !!empresa) return;
    const ruc = form.ruc.trim();
    const isDocumentoValido = ruc.length === 8 || ruc.length === 11;
    if (!isDocumentoValido) {
      setLookupLoading(false);
      setLookupMessage("");
      setDocLocked(false);
      return;
    }

    const timer = setTimeout(async () => {
      setDocLocked(true);
      setLookupLoading(true);
      setLookupMessage("");
      try {
        const resp = await buscarClientePorDocumento(ruc);
        const nombre = (resp.cliente?.cliente ?? "").trim();
        const direccion = (resp.cliente?.direccion ?? "").trim();

        setForm((prev) => ({
          ...prev,
          nombre,
          nombre_comercial: nombre,
          direccion,
        }));
        setAutofillFlash(true);
        setLookupMessage("Datos completados automáticamente ✨");
        setTimeout(() => setAutofillFlash(false), 500);
      } catch {
        setLookupMessage("No se pudo autocompletar con este RUC todavía.");
      } finally {
        setLookupLoading(false);
        setDocLocked(false);
      }
    }, 800);

    return () => clearTimeout(timer);
  }, [form.ruc, open, empresa]);

  function set(field: keyof EmpresaCreateRequest, value: string) {
    setForm((prev) => ({ ...prev, [field]: value }));
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!form.ruc.trim() || !form.nombre.trim()) {
      setError("RUC y nombre son requeridos");
      return;
    }
    setSaving(true);
    setError("");
    try {
      await onSave(form);
      onClose();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Error inesperado");
    } finally {
      setSaving(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>
            {empresa ? "Editar Empresa" : "Nueva Empresa"}
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className={`overflow-hidden rounded-md border px-3 transition-all duration-300 ${lookupLoading || lookupMessage ? "max-h-14 py-2 opacity-100" : "max-h-0 py-0 opacity-0 border-transparent"}`}>
            <div className="flex items-center gap-2 text-sm text-muted-foreground">
              {lookupLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4 text-violet-500" />}
              <span>{lookupLoading ? "Consultando RUC..." : lookupMessage}</span>
            </div>
          </div>
          <div className="space-y-1">
            <label className="text-sm font-medium">RUC *</label>
            <Input
              value={form.ruc}
              onChange={(e) => set("ruc", e.target.value)}
              placeholder="20123456789"
              disabled={!!empresa || docLocked}
              maxLength={11}
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm font-medium">Nombre *</label>
            <Input
              value={form.nombre}
              onChange={(e) => set("nombre", e.target.value)}
              placeholder="Razón social"
              disabled={docLocked}
              className={autofillFlash ? "transition-colors duration-500 bg-emerald-50 border-emerald-300" : "transition-colors duration-300"}
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm font-medium">Nombre Comercial</label>
            <Input
              value={form.nombre_comercial}
              onChange={(e) => set("nombre_comercial", e.target.value)}
              placeholder="Nombre comercial (opcional)"
              className={autofillFlash ? "transition-colors duration-500 bg-emerald-50 border-emerald-300" : "transition-colors duration-300"}
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm font-medium">Teléfono de contacto</label>
            <Input
              value={form.telefono_contacto}
              onChange={(e) => set("telefono_contacto", e.target.value)}
              placeholder="+51 999 999 999"
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm font-medium">Dirección</label>
            <Input
              value={form.direccion}
              onChange={(e) => set("direccion", e.target.value)}
              placeholder="Dirección fiscal"
              disabled={docLocked}
              className={autofillFlash ? "transition-colors duration-500 bg-emerald-50 border-emerald-300" : "transition-colors duration-300"}
            />
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={onClose}
              disabled={saving}
            >
              Cancelar
            </Button>
            <Button type="submit" disabled={saving}>
              {saving
                ? "Guardando..."
                : empresa
                  ? "Guardar cambios"
                  : "Crear empresa"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

"use client";

import { useEffect, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Select } from "@/components/ui/select";
import type { AdminTelefono } from "@/lib/api";

export interface TelefonoFormData {
  codigo_pais: string;
  numero: string;
  status?: string;
}

interface Props {
  open: boolean;
  onClose: () => void;
  onSave: (data: TelefonoFormData) => Promise<void>;
  telefono?: AdminTelefono | null;
}

const EMPTY: TelefonoFormData = {
  codigo_pais: "+51",
  numero: "",
  status: "disconnected",
};

export function TelefonoFormModal({ open, onClose, onSave, telefono }: Props) {
  const [form, setForm] = useState<TelefonoFormData>(EMPTY);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (telefono) {
      setForm({
        codigo_pais: telefono.codigo_pais ?? "+51",
        numero: telefono.numero ?? "",
        status: telefono.status ?? "disconnected",
      });
    } else {
      setForm(EMPTY);
    }
    setError("");
  }, [telefono, open]);

  function setField(field: keyof TelefonoFormData, value: string) {
    setForm((prev) => ({ ...prev, [field]: value }));
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();

    if (!form.codigo_pais.trim() || !form.numero.trim()) {
      setError("Código país y número son requeridos");
      return;
    }

    setSaving(true);
    setError("");
    try {
      await onSave({
        codigo_pais: form.codigo_pais.trim(),
        numero: form.numero.trim(),
        status: form.status?.trim() || undefined,
      });
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
          <DialogTitle>{telefono ? "Editar teléfono" : "Nuevo teléfono"}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-1">
            <label className="text-sm font-medium">Código país *</label>
            <Input
              value={form.codigo_pais}
              onChange={(e) => setField("codigo_pais", e.target.value)}
              placeholder="+51"
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm font-medium">Número *</label>
            <Input
              value={form.numero}
              onChange={(e) => setField("numero", e.target.value)}
              placeholder="999999999"
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm font-medium">Estado</label>
            <Select
              value={form.status ?? "disconnected"}
              onChange={(e) => setField("status", e.target.value)}
            >
              <option value="disconnected">disconnected</option>
              <option value="qr_pending">qr_pending</option>
              <option value="active">active</option>
            </Select>
          </div>
          {error && <p className="text-sm text-destructive">{error}</p>}
          <DialogFooter>
            <Button type="button" variant="outline" onClick={onClose} disabled={saving}>
              Cancelar
            </Button>
            <Button type="submit" disabled={saving}>
              {saving ? "Guardando..." : telefono ? "Guardar cambios" : "Crear teléfono"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

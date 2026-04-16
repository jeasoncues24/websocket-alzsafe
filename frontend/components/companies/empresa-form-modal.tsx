"use client";

import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { Empresa, EmpresaCreateRequest } from "@/lib/api";

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
  telefono: "",
  direccion: "",
};

export function EmpresaFormModal({ open, onClose, onSave, empresa }: Props) {
  const [form, setForm] = useState<EmpresaCreateRequest>(EMPTY);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (empresa) {
      setForm({
        ruc: empresa.ruc,
        nombre: empresa.nombre,
        nombre_comercial: empresa.nombre_comercial ?? "",
        telefono: empresa.telefono ?? "",
        direccion: empresa.direccion ?? "",
      });
    } else {
      setForm(EMPTY);
    }
    setError("");
  }, [empresa, open]);

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
          <div className="space-y-1">
            <label className="text-sm font-medium">RUC *</label>
            <Input
              value={form.ruc}
              onChange={(e) => set("ruc", e.target.value)}
              placeholder="20123456789"
              disabled={!!empresa}
              maxLength={11}
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm font-medium">Nombre *</label>
            <Input
              value={form.nombre}
              onChange={(e) => set("nombre", e.target.value)}
              placeholder="Razón social"
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm font-medium">Nombre Comercial</label>
            <Input
              value={form.nombre_comercial}
              onChange={(e) => set("nombre_comercial", e.target.value)}
              placeholder="Nombre comercial (opcional)"
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm font-medium">Teléfono</label>
            <Input
              value={form.telefono}
              onChange={(e) => set("telefono", e.target.value)}
              placeholder="+51 999 999 999"
            />
          </div>
          <div className="space-y-1">
            <label className="text-sm font-medium">Dirección</label>
            <Input
              value={form.direccion}
              onChange={(e) => set("direccion", e.target.value)}
              placeholder="Dirección fiscal"
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

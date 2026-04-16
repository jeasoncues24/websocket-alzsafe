"use client";

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Building2, Phone, MapPin, Hash, Calendar } from "lucide-react";
import type { Empresa } from "@/lib/api";

interface Props {
  open: boolean;
  onClose: () => void;
  empresa: Empresa | null;
}

export function EmpresaDetailModal({ open, onClose, empresa }: Props) {
  if (!empresa) return null;

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Building2 className="h-5 w-5" />
            {empresa.nombre}
          </DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">Estado</span>
            <Badge
              variant={empresa.activo ? "default" : "secondary"}
              className={empresa.activo ? "bg-green-500" : ""}
            >
              {empresa.activo ? "Activa" : "Inactiva"}
            </Badge>
          </div>

          <div className="grid grid-cols-2 gap-4 text-sm">
            <div className="space-y-1">
              <div className="flex items-center gap-1 text-muted-foreground">
                <Hash className="h-3.5 w-3.5" />
                RUC
              </div>
              <p className="font-medium">{empresa.ruc}</p>
            </div>

            {empresa.nombre_comercial && (
              <div className="space-y-1">
                <div className="text-muted-foreground text-xs">
                  Nombre comercial
                </div>
                <p className="font-medium">{empresa.nombre_comercial}</p>
              </div>
            )}

            {empresa.telefono && (
              <div className="space-y-1">
                <div className="flex items-center gap-1 text-muted-foreground">
                  <Phone className="h-3.5 w-3.5" />
                  Teléfono
                </div>
                <p className="font-medium">{empresa.telefono}</p>
              </div>
            )}

            {empresa.direccion && (
              <div className="space-y-1 col-span-2">
                <div className="flex items-center gap-1 text-muted-foreground">
                  <MapPin className="h-3.5 w-3.5" />
                  Dirección
                </div>
                <p className="font-medium">{empresa.direccion}</p>
              </div>
            )}

            <div className="space-y-1">
              <div className="flex items-center gap-1 text-muted-foreground">
                <Calendar className="h-3.5 w-3.5" />
                Creada
              </div>
              <p className="font-medium">
                {new Date(empresa.created_at).toLocaleDateString("es-PE")}
              </p>
            </div>

            <div className="space-y-1">
              <div className="text-muted-foreground text-xs">Actualizada</div>
              <p className="font-medium">
                {new Date(empresa.updated_at).toLocaleDateString("es-PE")}
              </p>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

"use client"

import { type Empresa } from "@/lib/api"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Building2,
  Phone,
  Pencil,
  Eye,
  Trash2,
  KeyRound,
  RotateCcw,
  Loader2,
  ChevronLeft,
  ChevronRight,
  Inbox,
} from "lucide-react"

interface DataCardListProps {
  data: Empresa[]
  loading?: boolean
  onOpenTelefonos: (empresa: Empresa) => void
  onOpenDetail: (empresa: Empresa) => void
  onOpenEdit: (empresa: Empresa) => void
  onDelete: (empresa: Empresa) => void
  onRestore: (empresa: Empresa) => void
  deletingId: number | null
  restoringId: number | null
  emptyMessage?: string
  /** Paginación del servidor */
  page: number
  totalPages: number
  totalRows: number
  onPageChange: (newPage: number) => void
}

export function DataCardList({
  data,
  loading = false,
  onOpenTelefonos,
  onOpenDetail,
  onOpenEdit,
  onDelete,
  onRestore,
  deletingId,
  restoringId,
  emptyMessage = "No hay empresas registradas con estos filtros.",
  page,
  totalPages,
  totalRows,
  onPageChange,
}: DataCardListProps) {
  if (loading) {
    return (
      <div className="space-y-3">
        {Array.from({ length: 3 }).map((_, i) => (
          <div
            key={i}
            className="h-36 rounded-xl border bg-muted/20 animate-pulse"
          />
        ))}
      </div>
    )
  }

  if (data.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center rounded-xl border bg-card p-8 text-center text-muted-foreground">
        <Inbox className="h-10 w-10 opacity-30 mb-2" />
        <p className="text-sm font-medium">{emptyMessage}</p>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {data.map((empresa) => {
        const isDeleting = deletingId === empresa.id
        const isRestoring = restoringId === empresa.id

        return (
          <div
            key={empresa.id}
            className="flex flex-col rounded-xl border bg-card p-4 shadow-xs transition-colors hover:border-foreground/20"
          >
            {/* Header de la tarjeta */}
            <div className="flex items-start justify-between gap-2 border-b pb-3">
              <div className="flex items-center gap-2.5 min-w-0">
                <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg border bg-muted/30 text-muted-foreground">
                  <Building2 className="h-4 w-4" />
                </div>
                <div className="min-w-0">
                  <h3 className="truncate font-semibold text-sm text-foreground">
                    {empresa.nombre}
                  </h3>
                  <div className="flex items-center gap-2 text-xs text-muted-foreground">
                    <span className="font-mono">RUC: {empresa.ruc}</span>
                  </div>
                </div>
              </div>

              <Badge
                variant={empresa.activo ? "outline" : "secondary"}
                className={
                  empresa.activo
                    ? "border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 font-medium text-[11px]"
                    : "text-muted-foreground text-[11px]"
                }
              >
                {empresa.activo ? "Activa" : "Inactiva"}
              </Badge>
            </div>

            {/* Cuerpo de la tarjeta */}
            <div className="grid grid-cols-2 gap-2 py-3 text-xs">
              <div>
                <span className="text-muted-foreground">Nombre Comercial:</span>
                <p className="truncate font-medium text-foreground">
                  {empresa.nombre_comercial || "—"}
                </p>
              </div>

              <div>
                <span className="text-muted-foreground">Contacto:</span>
                <p className="flex items-center gap-1 font-medium text-foreground truncate">
                  {empresa.telefono_contacto ? (
                    <>
                      <Phone className="h-3 w-3 text-muted-foreground shrink-0" />
                      <span>{empresa.telefono_contacto}</span>
                    </>
                  ) : (
                    "—"
                  )}
                </p>
              </div>

              {empresa.direccion && (
                <div className="col-span-2">
                  <span className="text-muted-foreground">Dirección:</span>
                  <p className="truncate text-foreground">{empresa.direccion}</p>
                </div>
              )}
            </div>

            {/* Acciones de la tarjeta */}
            <div className="flex items-center justify-end gap-1.5 border-t pt-3">
              <Button
                variant="outline"
                size="sm"
                className="h-8 text-xs gap-1.5"
                onClick={() => onOpenTelefonos(empresa)}
              >
                <KeyRound className="h-3.5 w-3.5" />
                <span>Teléfonos</span>
              </Button>

              <Button
                variant="outline"
                size="sm"
                className="h-8 text-xs gap-1.5"
                onClick={() => onOpenDetail(empresa)}
              >
                <Eye className="h-3.5 w-3.5" />
                <span>Detalle</span>
              </Button>

              <Button
                variant="outline"
                size="sm"
                className="h-8 text-xs gap-1.5"
                onClick={() => onOpenEdit(empresa)}
              >
                <Pencil className="h-3.5 w-3.5" />
                <span>Editar</span>
              </Button>

              {empresa.activo ? (
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8 text-destructive hover:bg-destructive/10"
                  onClick={() => onDelete(empresa)}
                  disabled={isDeleting}
                  title="Eliminar"
                >
                  {isDeleting ? (
                    <Loader2 className="h-3.5 w-3.5 animate-spin" />
                  ) : (
                    <Trash2 className="h-3.5 w-3.5" />
                  )}
                </Button>
              ) : (
                <Button
                  variant="outline"
                  size="icon"
                  className="h-8 w-8 text-emerald-600 hover:bg-emerald-500/10 dark:text-emerald-400"
                  onClick={() => onRestore(empresa)}
                  disabled={isRestoring}
                  title="Restaurar"
                >
                  {isRestoring ? (
                    <Loader2 className="h-3.5 w-3.5 animate-spin" />
                  ) : (
                    <RotateCcw className="h-3.5 w-3.5" />
                  )}
                </Button>
              )}
            </div>
          </div>
        )
      })}

      {/* Paginación móvil */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between border-t pt-3 text-xs text-muted-foreground">
          <span>
            Pág. {page} de {totalPages} ({totalRows} empresas)
          </span>
          <div className="flex gap-1.5">
            <Button
              variant="outline"
              size="sm"
              className="h-7 px-2"
              disabled={page <= 1}
              onClick={() => onPageChange(page - 1)}
            >
              <ChevronLeft className="h-3.5 w-3.5 mr-1" />
              Ant.
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="h-7 px-2"
              disabled={page >= totalPages}
              onClick={() => onPageChange(page + 1)}
            >
              Sig.
              <ChevronRight className="h-3.5 w-3.5 ml-1" />
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}

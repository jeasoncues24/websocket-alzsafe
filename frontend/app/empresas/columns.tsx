"use client"

import { ColumnDef } from "@tanstack/react-table"
import { type Empresa } from "@/lib/api"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  ArrowUpDown,
  Building2,
  Phone,
  Pencil,
  Eye,
  Trash2,
  KeyRound,
  RotateCcw,
  Loader2,
  Copy,
  Check,
} from "lucide-react"
import { useState } from "react"

interface ColumnsOptions {
  onOpenTelefonos: (empresa: Empresa) => void
  onOpenDetail: (empresa: Empresa) => void
  onOpenEdit: (empresa: Empresa) => void
  onDelete: (empresa: Empresa) => void
  onRestore: (empresa: Empresa) => void
  deletingId: number | null
  restoringId: number | null
}

function CopyRucButton({ ruc }: { ruc: string }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = () => {
    navigator.clipboard.writeText(ruc)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  return (
    <button
      onClick={handleCopy}
      className="group inline-flex items-center gap-1.5 font-mono text-xs font-medium text-foreground hover:text-primary transition-colors"
      title="Copiar RUC"
    >
      <span>{ruc}</span>
      {copied ? (
        <Check className="h-3 w-3 text-emerald-500" />
      ) : (
        <Copy className="h-3 w-3 opacity-0 group-hover:opacity-100 transition-opacity text-muted-foreground" />
      )}
    </button>
  )
}

export function getColumns(options: ColumnsOptions): ColumnDef<Empresa>[] {
  const {
    onOpenTelefonos,
    onOpenDetail,
    onOpenEdit,
    onDelete,
    onRestore,
    deletingId,
    restoringId,
  } = options

  return [
    {
      accessorKey: "ruc",
      header: ({ column }) => (
        <Button
          variant="ghost"
          size="sm"
          className="-ml-3 h-8 text-xs font-medium hover:bg-transparent"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        >
          <span>RUC</span>
          <ArrowUpDown className="ml-1.5 h-3.5 w-3.5 text-muted-foreground" />
        </Button>
      ),
      cell: ({ row }) => <CopyRucButton ruc={row.original.ruc} />,
    },
    {
      accessorKey: "nombre",
      header: ({ column }) => (
        <Button
          variant="ghost"
          size="sm"
          className="-ml-3 h-8 text-xs font-medium hover:bg-transparent"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        >
          <span>Empresa</span>
          <ArrowUpDown className="ml-1.5 h-3.5 w-3.5 text-muted-foreground" />
        </Button>
      ),
      cell: ({ row }) => {
        const empresa = row.original
        return (
          <div className="flex items-center gap-2.5">
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border bg-muted/30 text-muted-foreground">
              <Building2 className="h-4 w-4" />
            </div>
            <div className="min-w-0">
              <span className="block truncate font-medium text-foreground text-sm">
                {empresa.nombre}
              </span>
              {empresa.direccion && (
                <span className="block truncate text-xs text-muted-foreground">
                  {empresa.direccion}
                </span>
              )}
            </div>
          </div>
        )
      },
    },
    {
      accessorKey: "nombre_comercial",
      header: "Nombre Comercial",
      cell: ({ row }) => (
        <span className="text-xs text-muted-foreground">
          {row.original.nombre_comercial || "—"}
        </span>
      ),
    },
    {
      accessorKey: "telefono_contacto",
      header: "Teléfono Contacto",
      cell: ({ row }) => {
        const tel = row.original.telefono_contacto
        if (!tel) return <span className="text-xs text-muted-foreground">—</span>
        return (
          <div className="inline-flex items-center gap-1.5 text-xs text-foreground">
            <Phone className="h-3 w-3 text-muted-foreground shrink-0" />
            <span>{tel}</span>
          </div>
        )
      },
    },
    {
      accessorKey: "activo",
      header: ({ column }) => (
        <Button
          variant="ghost"
          size="sm"
          className="-ml-3 h-8 text-xs font-medium hover:bg-transparent"
          onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
        >
          <span>Estado</span>
          <ArrowUpDown className="ml-1.5 h-3.5 w-3.5 text-muted-foreground" />
        </Button>
      ),
      cell: ({ row }) => {
        const activo = row.original.activo
        return (
          <Badge
            variant={activo ? "outline" : "secondary"}
            className={
              activo
                ? "border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 font-medium text-xs"
                : "text-muted-foreground text-xs"
            }
          >
            {activo ? "Activa" : "Inactiva"}
          </Badge>
        )
      },
    },
    {
      id: "actions",
      header: () => <div className="text-right">Acciones</div>,
      cell: ({ row }) => {
        const empresa = row.original
        const isDeleting = deletingId === empresa.id
        const isRestoring = restoringId === empresa.id

        return (
          <div className="flex items-center justify-end gap-1">
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => onOpenTelefonos(empresa)}
              title="Ver teléfonos / canales"
            >
              <KeyRound className="h-4 w-4 text-muted-foreground" />
              <span className="sr-only">Ver teléfonos</span>
            </Button>

            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => onOpenDetail(empresa)}
              title="Ver detalle"
            >
              <Eye className="h-4 w-4 text-muted-foreground" />
              <span className="sr-only">Ver detalle</span>
            </Button>

            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => onOpenEdit(empresa)}
              title="Editar empresa"
            >
              <Pencil className="h-4 w-4 text-muted-foreground" />
              <span className="sr-only">Editar</span>
            </Button>

            {empresa.activo ? (
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8 text-destructive hover:text-destructive hover:bg-destructive/10"
                onClick={() => onDelete(empresa)}
                disabled={isDeleting}
                title="Desactivar / Eliminar"
              >
                {isDeleting ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <Trash2 className="h-4 w-4" />
                )}
                <span className="sr-only">Eliminar</span>
              </Button>
            ) : (
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8 text-emerald-600 hover:text-emerald-700 hover:bg-emerald-500/10 dark:text-emerald-400"
                onClick={() => onRestore(empresa)}
                disabled={isRestoring}
                title="Restaurar empresa"
              >
                {isRestoring ? (
                  <Loader2 className="h-4 w-4 animate-spin" />
                ) : (
                  <RotateCcw className="h-4 w-4" />
                )}
                <span className="sr-only">Restaurar</span>
              </Button>
            )}
          </div>
        )
      },
    },
  ]
}

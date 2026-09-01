"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Pencil,
  Trash2,
  ChevronLeft,
  ChevronRight,
  Users,
  Shield,
  User,
} from "lucide-react"
import { type UserAdminRol } from "@/lib/api"

interface DataCardListProps {
  data: UserAdminRol[]
  loading?: boolean
  onEdit: (user: UserAdminRol) => void
  onDelete: (user: UserAdminRol) => void
  deletingId: number | null
}

const PAGE_SIZE = 10

export function DataCardList({
  data,
  loading = false,
  onEdit,
  onDelete,
  deletingId,
}: DataCardListProps) {
  const [page, setPage] = useState(1)
  const totalPages = Math.ceil(data.length / PAGE_SIZE) || 1
  const currentPage = Math.min(page, totalPages)
  const paged = data.slice((currentPage - 1) * PAGE_SIZE, currentPage * PAGE_SIZE)

  if (loading) {
    return (
      <div className="flex flex-col gap-3">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="rounded-xl p-4 ring-1 ring-foreground/10">
            <div className="flex flex-col gap-3">
              <div className="flex items-center justify-between">
                <Skeleton className="h-5 w-36" />
                <Skeleton className="h-5 w-20" />
              </div>
              <Skeleton className="h-4 w-48" />
              <div className="flex gap-2 pt-1">
                <Skeleton className="h-9 flex-1" />
                <Skeleton className="h-9 flex-1" />
              </div>
            </div>
          </div>
        ))}
      </div>
    )
  }

  if (data.length === 0) {
    return (
      <div className="rounded-xl px-6 py-12 text-center ring-1 ring-foreground/10">
        <div className="flex flex-col items-center gap-2 text-muted-foreground">
          <Users className="h-8 w-8 opacity-30" />
          <p className="text-sm font-medium">
            No se encontraron usuarios que coincidan con los filtros.
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-3">
      {paged.map((user) => {
        const isDeleting = deletingId === user.id

        return (
          <div
            key={user.id}
            className="flex flex-col gap-3 rounded-xl p-4 ring-1 ring-foreground/10"
          >
            {/* Encabezado: Avatar + Nombre + Estado */}
            <div className="flex items-start justify-between gap-2">
              <div className="flex items-center gap-2.5 min-w-0">
                <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-muted text-muted-foreground">
                  <User className="h-4 w-4" />
                </div>
                <div className="min-w-0">
                  <h3 className="truncate text-sm font-medium text-foreground">
                    {user.username}
                  </h3>
                  <p className="font-mono text-xs text-muted-foreground">
                    ID: #{user.id}
                  </p>
                </div>
              </div>
              <Badge
                variant={user.activo ? "outline" : "secondary"}
                className={
                  user.activo
                    ? "border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 font-medium text-xs"
                    : "text-muted-foreground text-xs"
                }
              >
                {user.activo ? "Activo" : "Inactivo"}
              </Badge>
            </div>

            {/* Info intermedia: Email, Rol, Root */}
            <div className="flex flex-col gap-2 border-y border-border py-2 text-xs">
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Email:</span>
                <span className="text-foreground">{user.email || "—"}</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Rol:</span>
                <Badge variant="outline" className="text-[11px]">
                  {user.rol}
                </Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Permisos Root:</span>
                {user.is_root ? (
                  <Badge variant="default" className="gap-1 text-[11px]">
                    <Shield className="h-3 w-3" />
                    Sí
                  </Badge>
                ) : (
                  <Badge variant="secondary" className="text-[11px] text-muted-foreground">
                    No
                  </Badge>
                )}
              </div>
            </div>

            {/* Acciones */}
            <div className="flex items-center gap-2 pt-1">
              <Button
                variant="outline"
                size="sm"
                className="flex-1 text-xs"
                onClick={() => onEdit(user)}
              >
                <Pencil className="mr-1.5 h-3.5 w-3.5" />
                Editar
              </Button>
              <Button
                variant="outline"
                size="sm"
                className="flex-1 text-xs text-destructive hover:bg-destructive/10 hover:text-destructive"
                disabled={isDeleting}
                onClick={() => onDelete(user)}
              >
                <Trash2 className="mr-1.5 h-3.5 w-3.5" />
                Eliminar
              </Button>
            </div>
          </div>
        )
      })}

      {/* Paginación móvil */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between px-1 py-2 text-xs text-muted-foreground">
          <span>
            Pág. <span className="font-medium text-foreground">{currentPage}</span> de{" "}
            <span className="font-medium text-foreground">{totalPages}</span> ·{" "}
            {data.length} usuarios
          </span>
          <div className="flex items-center gap-1">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={currentPage === 1}
            >
              <ChevronLeft className="mr-1 h-3.5 w-3.5" />
              Anterior
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={currentPage === totalPages}
            >
              Siguiente
              <ChevronRight className="ml-1 h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}

"use client"

import { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
  ArrowUpDown,
  Pencil,
  Trash2,
  User,
  Shield,
} from "lucide-react"
import { type UserAdminRol } from "@/lib/api"

function SortableHeader<T>({
  column,
  children,
}: {
  column: import("@tanstack/react-table").Column<T, unknown>
  children: React.ReactNode
}) {
  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
      className="-ml-3 h-8 text-xs font-semibold hover:bg-muted/50"
    >
      {children}
      <ArrowUpDown className="ml-1 h-3 w-3" />
    </Button>
  )
}

interface ColumnActionCallbacks {
  onEdit: (user: UserAdminRol) => void
  onDelete: (user: UserAdminRol) => void
  deletingId: number | null
}

export function getColumns(cb: ColumnActionCallbacks): ColumnDef<UserAdminRol>[] {
  return [
    {
      accessorKey: "username",
      header: ({ column }) => <SortableHeader column={column}>Usuario</SortableHeader>,
      cell: ({ row }) => {
        const user = row.original
        return (
          <div className="flex items-center gap-2">
            <div className="flex h-7 w-7 items-center justify-center rounded-full bg-muted text-muted-foreground">
              <User className="h-3.5 w-3.5" />
            </div>
            <div className="flex flex-col">
              <span className="font-medium text-foreground">{user.username}</span>
              <span className="font-mono text-[10px] text-muted-foreground">ID: #{user.id}</span>
            </div>
          </div>
        )
      },
    },
    {
      accessorKey: "email",
      header: ({ column }) => <SortableHeader column={column}>Email</SortableHeader>,
      cell: ({ row }) => (
        <span className="text-sm text-muted-foreground">
          {row.original.email || "—"}
        </span>
      ),
    },
    {
      accessorKey: "rol",
      header: ({ column }) => <SortableHeader column={column}>Rol</SortableHeader>,
      cell: ({ row }) => (
        <Badge variant="outline" className="font-medium">
          {row.original.rol}
        </Badge>
      ),
    },
    {
      accessorKey: "is_root",
      header: ({ column }) => <SortableHeader column={column}>Root</SortableHeader>,
      cell: ({ row }) => {
        const isRoot = row.original.is_root
        return isRoot ? (
          <Badge variant="default" className="gap-1 font-medium">
            <Shield className="h-3 w-3" />
            Sí
          </Badge>
        ) : (
          <Badge variant="secondary" className="font-normal text-muted-foreground">
            No
          </Badge>
        )
      },
    },
    {
      accessorKey: "activo",
      header: ({ column }) => <SortableHeader column={column}>Estado</SortableHeader>,
      cell: ({ row }) => {
        const activo = row.original.activo
        return (
          <Badge
            variant={activo ? "outline" : "secondary"}
            className={
              activo
                ? "border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 font-medium"
                : "text-muted-foreground"
            }
          >
            {activo ? "Activo" : "Inactivo"}
          </Badge>
        )
      },
    },
    {
      id: "actions",
      header: () => <div className="pr-2 text-right">Acciones</div>,
      cell: ({ row }) => {
        const user = row.original
        const isDeleting = cb.deletingId === user.id

        return (
          <div className="flex items-center justify-end gap-1">
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8 text-muted-foreground hover:text-foreground"
              onClick={() => cb.onEdit(user)}
              title="Editar usuario"
              aria-label="Editar usuario"
            >
              <Pencil className="h-3.5 w-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8 text-muted-foreground hover:text-destructive"
              onClick={() => cb.onDelete(user)}
              disabled={isDeleting}
              title="Eliminar usuario"
              aria-label="Eliminar usuario"
            >
              <Trash2 className="h-3.5 w-3.5" />
            </Button>
          </div>
        )
      },
    },
  ]
}

"use client"

import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Select } from "@/components/ui/select"
import { cn } from "@/lib/utils"
import {
  Search,
  X,
  RefreshCw,
  Layers,
  CheckCircle2,
  CircleSlash,
  Shield,
  ShieldCheck,
  Plus,
} from "lucide-react"
import { type Role } from "@/lib/api"

export type UserStatusFilter = "all" | "active" | "inactive" | "root"

interface DataTableToolbarProps {
  globalFilter: string
  onGlobalFilterChange: (value: string) => void
  statusFilter: UserStatusFilter
  onStatusFilterChange: (status: UserStatusFilter) => void
  roleFilter: string
  onRoleFilterChange: (role: string) => void
  roles: Role[]
  counts: Record<UserStatusFilter, number>
  loading: boolean
  onRefresh: () => void
  onNewUser: () => void
}

export function DataTableToolbar({
  globalFilter,
  onGlobalFilterChange,
  statusFilter,
  onStatusFilterChange,
  roleFilter,
  onRoleFilterChange,
  roles,
  counts,
  loading,
  onRefresh,
  onNewUser,
}: DataTableToolbarProps) {
  const tabs: {
    id: UserStatusFilter
    label: string
    icon: React.ReactNode
  }[] = [
    {
      id: "all",
      label: "Todos",
      icon: <Layers className="h-3.5 w-3.5" />,
    },
    {
      id: "active",
      label: "Activos",
      icon: <CheckCircle2 className="h-3.5 w-3.5" />,
    },
    {
      id: "inactive",
      label: "Inactivos",
      icon: <CircleSlash className="h-3.5 w-3.5" />,
    },
    {
      id: "root",
      label: "Root",
      icon: <Shield className="h-3.5 w-3.5" />,
    },
  ]

  return (
    <div className="space-y-3">
      {/* Control segmentado por estado: un único bloque, sin arcoíris de colores */}
      <div
        role="group"
        aria-label="Filtrar usuarios por estado"
        className="flex flex-wrap gap-0.5 rounded-md bg-muted p-0.5"
      >
        {tabs.map((tab) => {
          const isActive = statusFilter === tab.id
          return (
            <button
              key={tab.id}
              type="button"
              aria-pressed={isActive}
              onClick={() => onStatusFilterChange(tab.id)}
              className={cn(
                "inline-flex h-8 items-center gap-1.5 rounded-sm px-2.5 text-xs font-medium transition-colors",
                isActive
                  ? "bg-background text-foreground shadow-sm"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              {tab.icon}
              <span>{tab.label}</span>
              <span
                className={cn(
                  "tabular-nums text-[11px]",
                  isActive ? "text-foreground/70" : "text-muted-foreground/70",
                )}
              >
                {counts[tab.id] ?? 0}
              </span>
            </button>
          )
        })}
      </div>

      {/* Buscador + Selector de Rol + Acciones */}
      <div className="flex flex-col gap-2.5 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex flex-1 flex-col gap-2.5 sm:flex-row sm:items-center">
          {/* Buscador global */}
          <div className="relative w-full sm:max-w-sm">
            <Search className="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Buscar por usuario, email o rol..."
              value={globalFilter}
              onChange={(e) => onGlobalFilterChange(e.target.value)}
              className="h-9 pl-8 pr-8 text-xs sm:text-sm"
            />
            {globalFilter && (
              <button
                type="button"
                onClick={() => onGlobalFilterChange("")}
                className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-0.5 text-muted-foreground hover:text-foreground"
                title="Limpiar búsqueda"
              >
                <X className="h-3.5 w-3.5" />
              </button>
            )}
          </div>

          {/* Filtro por rol */}
          <div className="flex items-center gap-1.5 w-full sm:w-auto">
            <div className="relative w-full sm:w-[180px]">
              <ShieldCheck className="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Select
                value={roleFilter}
                onChange={(e) => onRoleFilterChange(e.target.value)}
                className="h-9 pl-8 text-xs"
              >
                <option value="">Todos los roles</option>
                {roles.map((r) => (
                  <option key={r.id} value={r.name}>
                    {r.name}
                  </option>
                ))}
              </Select>
            </div>
            {roleFilter && (
              <Button
                variant="ghost"
                size="icon"
                className="h-9 w-9 shrink-0 text-muted-foreground hover:text-foreground"
                onClick={() => onRoleFilterChange("")}
                title="Limpiar filtro de rol"
              >
                <X className="h-4 w-4" />
              </Button>
            )}
          </div>
        </div>

        {/* Botones de acción: Refresco y Nuevo */}
        <div className="flex items-center gap-2 self-end lg:self-auto">
          <Button
            variant="outline"
            size="sm"
            onClick={onRefresh}
            disabled={loading}
            className="h-9 text-xs"
          >
            <RefreshCw
              className={cn("mr-1.5 h-3.5 w-3.5", loading && "animate-spin")}
            />
            Actualizar
          </Button>

          <Button
            size="sm"
            onClick={onNewUser}
            className="h-9 text-xs"
          >
            <Plus className="mr-1.5 h-3.5 w-3.5" />
            Nuevo usuario
          </Button>
        </div>
      </div>
    </div>
  )
}

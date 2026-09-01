"use client"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Search, Plus, RefreshCw, X } from "lucide-react"

export type EmpresaStatusFilter = "all" | "active" | "inactive"

interface DataTableToolbarProps {
  search: string
  onSearchChange: (value: string) => void
  statusFilter: EmpresaStatusFilter
  onStatusFilterChange: (status: EmpresaStatusFilter) => void
  counts: {
    all: number
    active: number
    inactive: number
  }
  onOpenNew: () => void
  onRefresh: () => void
  refreshing?: boolean
}

export function DataTableToolbar({
  search,
  onSearchChange,
  statusFilter,
  onStatusFilterChange,
  counts,
  onOpenNew,
  onRefresh,
  refreshing = false,
}: DataTableToolbarProps) {
  const tabs: { id: EmpresaStatusFilter; label: string; count: number }[] = [
    { id: "all", label: "Todas", count: counts.all },
    { id: "active", label: "Activas", count: counts.active },
    { id: "inactive", label: "Inactivas", count: counts.inactive },
  ]

  return (
    <div className="flex flex-col gap-3">
      {/* Fila superior: Tabs de estado y acciones principales */}
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        {/* Control segmentado de estados */}
        <div
          role="group"
          aria-label="Filtro por estado de empresa"
          className="inline-flex rounded-lg border bg-muted/40 p-1 text-xs"
        >
          {tabs.map((tab) => {
            const isActive = statusFilter === tab.id
            return (
              <button
                key={tab.id}
                type="button"
                aria-pressed={isActive}
                onClick={() => onStatusFilterChange(tab.id)}
                className={`inline-flex items-center gap-1.5 rounded-md px-3 py-1.5 font-medium transition-all ${
                  isActive
                    ? "bg-background text-foreground shadow-xs"
                    : "text-muted-foreground hover:text-foreground"
                }`}
              >
                <span>{tab.label}</span>
                <Badge
                  variant={isActive ? "default" : "secondary"}
                  className="h-4.5 px-1.5 text-[10px] font-normal"
                >
                  {tab.count}
                </Badge>
              </button>
            )
          })}
        </div>

        {/* Acciones principales */}
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={onRefresh}
            disabled={refreshing}
            className="h-9 text-xs"
            title="Refrescar lista"
          >
            <RefreshCw
              className={`h-3.5 w-3.5 mr-1.5 ${refreshing ? "animate-spin" : ""}`}
            />
            <span>Actualizar</span>
          </Button>

          <Button onClick={onOpenNew} size="sm" className="h-9 text-xs">
            <Plus className="h-3.5 w-3.5 mr-1.5" />
            <span>Nueva Empresa</span>
          </Button>
        </div>
      </div>

      {/* Fila inferior: Buscador por texto */}
      <div className="flex items-center gap-2">
        <div className="relative flex-1 sm:max-w-md">
          <Search className="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Buscar por RUC, nombre o nombre comercial..."
            value={search}
            onChange={(e) => onSearchChange(e.target.value)}
            className="h-9 pl-8 pr-8 text-xs sm:text-sm"
          />
          {search && (
            <button
              onClick={() => onSearchChange("")}
              className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
              title="Limpiar búsqueda"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          )}
        </div>
      </div>
    </div>
  )
}

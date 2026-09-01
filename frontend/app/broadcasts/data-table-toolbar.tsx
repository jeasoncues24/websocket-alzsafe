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
  XCircle,
  Clock,
  Loader2,
  Building2,
} from "lucide-react"
import { type Empresa } from "@/lib/api"

export type BroadcastStatusFilter = "all" | "running" | "completed" | "failed" | "pending"

interface DataTableToolbarProps {
  globalFilter: string
  onGlobalFilterChange: (value: string) => void
  statusFilter: BroadcastStatusFilter
  onStatusFilterChange: (status: BroadcastStatusFilter) => void
  companyFilter: string
  onCompanyFilterChange: (ruc: string) => void
  companies: Empresa[]
  counts: Record<BroadcastStatusFilter, number>
  loading: boolean
  onRefresh: () => void
}

export function DataTableToolbar({
  globalFilter,
  onGlobalFilterChange,
  statusFilter,
  onStatusFilterChange,
  companyFilter,
  onCompanyFilterChange,
  companies,
  counts,
  loading,
  onRefresh,
}: DataTableToolbarProps) {
  const tabs: {
    id: BroadcastStatusFilter
    label: string
    icon: React.ReactNode
  }[] = [
    {
      id: "all",
      label: "Todos",
      icon: <Layers className="h-3.5 w-3.5" />,
    },
    {
      id: "running",
      label: "Enviando",
      icon: <Loader2 className="h-3.5 w-3.5 animate-spin" />,
    },
    {
      id: "completed",
      label: "Completados",
      icon: <CheckCircle2 className="h-3.5 w-3.5" />,
    },
    {
      id: "failed",
      label: "Fallidos",
      icon: <XCircle className="h-3.5 w-3.5" />,
    },
    {
      id: "pending",
      label: "Pendientes",
      icon: <Clock className="h-3.5 w-3.5" />,
    },
  ]

  return (
    <div className="space-y-3">
      {/* Control segmentado por estado: un único bloque, sin arcoíris de colores */}
      <div
        role="group"
        aria-label="Filtrar difusiones por estado"
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

      {/* Buscador + Selector de Empresa + Refresco */}
      <div className="flex flex-col gap-2.5 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex flex-1 flex-col gap-2.5 sm:flex-row sm:items-center">
          {/* Buscador global */}
          <div className="relative w-full sm:max-w-sm">
            <Search className="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="Buscar por ID de referencia, empresa o RUC..."
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

          {/* Filtro por empresa */}
          <div className="flex items-center gap-1.5 w-full sm:w-auto">
            <div className="relative w-full sm:w-[220px]">
              <Building2 className="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Select
                value={companyFilter}
                onChange={(e) => onCompanyFilterChange(e.target.value)}
                className="h-9 pl-8 text-xs"
              >
                <option value="">Todas las empresas</option>
                {companies.map((c) => (
                  <option key={c.ruc} value={c.ruc}>
                    {c.nombre ? `${c.nombre} (${c.ruc})` : c.ruc}
                  </option>
                ))}
              </Select>
            </div>
            {companyFilter && (
              <Button
                variant="ghost"
                size="icon"
                className="h-9 w-9 shrink-0 text-muted-foreground hover:text-foreground"
                onClick={() => onCompanyFilterChange("")}
                title="Limpiar filtro de empresa"
              >
                <X className="h-4 w-4" />
              </Button>
            )}
          </div>
        </div>

        {/* Botón de refresco */}
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
        </div>
      </div>
    </div>
  )
}

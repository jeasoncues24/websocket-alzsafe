"use client"

import { useEffect, useState } from "react"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import {
  Search,
  X,
  RefreshCw,
  Wifi,
  WifiOff,
  QrCode,
  AlertTriangle,
  Layers,
} from "lucide-react"
import { type SessionSummary } from "@/lib/api"

export type StatusTabFilter = "all" | "active" | "disconnected" | "qr_pending" | "mismatch"

interface DataTableToolbarProps {
  globalFilter: string
  onGlobalFilterChange: (value: string) => void
  statusFilter: StatusTabFilter
  onStatusFilterChange: (status: StatusTabFilter) => void
  summary: SessionSummary | null
  loading: boolean
  onRefresh: () => void
  lastUpdated?: Date | null
  /** Stream en tiempo real conectado. */
  live?: boolean
}

function formatUpdatedAgo(date: Date): string {
  const s = Math.floor((Date.now() - date.getTime()) / 1000)
  if (s < 10) return "hace instantes"
  if (s < 60) return `hace ${s}s`
  const m = Math.floor(s / 60)
  if (m < 60) return `hace ${m} min`
  const h = Math.floor(m / 60)
  return `hace ${h} h`
}

export function DataTableToolbar({
  globalFilter,
  onGlobalFilterChange,
  statusFilter,
  onStatusFilterChange,
  summary,
  loading,
  onRefresh,
  lastUpdated,
  live,
}: DataTableToolbarProps) {
  // Refresca el "actualizado hace Xs" sin recargar datos.
  const [, setTick] = useState(0)
  useEffect(() => {
    const id = setInterval(() => setTick((t) => t + 1), 15000)
    return () => clearInterval(id)
  }, [])

  const tabs: {
    id: StatusTabFilter
    label: string
    count: number | undefined
    icon: React.ReactNode
  }[] = [
    {
      id: "all",
      label: "Todas",
      count: summary?.total,
      icon: <Layers className="h-3.5 w-3.5" />,
    },
    {
      id: "active",
      label: "Activas",
      count: summary?.active,
      icon: <Wifi className="h-3.5 w-3.5" />,
    },
    {
      id: "disconnected",
      label: "Desconectadas",
      count: summary?.disconnected,
      icon: <WifiOff className="h-3.5 w-3.5" />,
    },
    {
      id: "qr_pending",
      label: "QR Pendiente",
      count: summary?.qr_pending,
      icon: <QrCode className="h-3.5 w-3.5" />,
    },
    {
      id: "mismatch",
      label: "Inconsistentes",
      count: summary?.mismatch,
      icon: <AlertTriangle className="h-3.5 w-3.5" />,
    },
  ]

  return (
    <div className="space-y-3">
      {/* Filtro por estado: un único control segmentado, sin color */}
      <div
        role="group"
        aria-label="Filtrar sesiones por estado"
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
                  ? "bg-background text-foreground"
                  : "text-muted-foreground hover:text-foreground"
              )}
            >
              {tab.icon}
              <span>{tab.label}</span>
              {tab.count !== undefined && (
                <span
                  className={cn(
                    "tabular-nums",
                    isActive ? "text-foreground/60" : "text-muted-foreground/70"
                  )}
                >
                  {tab.count}
                </span>
              )}
            </button>
          )
        })}
      </div>

      {/* Buscador + estado de actualización + refresco */}
      <div className="flex flex-col gap-2.5 sm:flex-row sm:items-center sm:justify-between">
        <div className="relative w-full sm:max-w-sm">
          <Search className="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Buscar por empresa, número o cuenta..."
            value={globalFilter}
            onChange={(e) => onGlobalFilterChange(e.target.value)}
            className="h-9 pl-8 pr-8"
          />
          {globalFilter && (
            <button
              onClick={() => onGlobalFilterChange("")}
              className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-0.5 text-muted-foreground hover:text-foreground"
              title="Limpiar búsqueda"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          )}
        </div>

        <div className="flex items-center gap-3 self-end sm:self-auto">
          <span
            className="hidden items-center gap-1.5 text-xs text-muted-foreground sm:inline-flex"
            title={live ? "Actualización en tiempo real activa" : "Sin stream: actualizando cada 5s"}
          >
            <span
              className={cn(
                "h-1.5 w-1.5 rounded-full",
                live ? "animate-pulse bg-emerald-500" : "bg-muted-foreground/40",
              )}
            />
            {live ? "En vivo" : "Reconectando…"}
          </span>
          {lastUpdated && (
            <span className="hidden text-xs text-muted-foreground sm:inline">
              Actualizado {formatUpdatedAgo(lastUpdated)}
            </span>
          )}
          <Button
            variant="outline"
            size="sm"
            onClick={onRefresh}
            disabled={loading}
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

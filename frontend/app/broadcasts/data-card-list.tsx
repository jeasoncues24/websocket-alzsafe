"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Eye,
  ChevronLeft,
  ChevronRight,
  MessageSquareMore,
} from "lucide-react"
import { type BroadcastInfo, type Empresa } from "@/lib/api"
import { BroadcastStatusBadge, ProgressBar } from "./broadcast-detail-sheet"
import { CopyRefButton } from "./columns"

interface DataCardListProps {
  data: BroadcastInfo[]
  loading?: boolean
  companyByRuc: Map<string, Empresa>
  onOpenDetails: (bc: BroadcastInfo) => void
}

const PAGE_SIZE = 10

function relativeTime(ts?: string | null): string {
  if (!ts) return "Nunca"
  const diff = Date.now() - new Date(ts).getTime()
  const m = Math.floor(diff / 60000)
  if (m < 1) return "ahora"
  if (m < 60) return `hace ${m}m`
  const h = Math.floor(m / 60)
  if (h < 24) return `hace ${h}h`
  return `hace ${Math.floor(h / 24)}d`
}

export function DataCardList({
  data,
  loading = false,
  companyByRuc,
  onOpenDetails,
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
              <Skeleton className="h-4 w-32" />
              <div className="pt-1">
                <Skeleton className="h-9 w-full" />
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
          <MessageSquareMore className="h-8 w-8 opacity-30" />
          <p className="text-sm font-medium">
            No se encontraron difusiones que coincidan con los filtros.
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-3">
      {paged.map((bc) => {
        const company = companyByRuc.get(bc.ruc_empresa)
        const empresaName = company?.nombre || bc.ruc_empresa

        return (
          <div
            key={bc.reference_id}
            className="flex flex-col gap-3 rounded-xl p-4 ring-1 ring-foreground/10"
          >
            {/* Encabezado: Ref ID + Empresa + Status Badge */}
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-1.5">
                  <CopyRefButton text={bc.reference_id} />
                </div>
                <h3
                  className="truncate text-sm font-medium text-foreground"
                  title={empresaName}
                >
                  {empresaName}
                </h3>
                <div className="font-mono text-xs text-muted-foreground">
                  {company?.ruc || bc.ruc_empresa}
                </div>
              </div>
              <BroadcastStatusBadge status={bc.status} />
            </div>

            {/* Progreso + Fecha */}
            <div className="flex flex-col gap-2 border-y border-border py-2 text-xs">
              <div className="flex items-center justify-between">
                <span className="text-muted-foreground">Progreso:</span>
                <ProgressBar sent={bc.success ?? 0} total={bc.total} />
              </div>
              <div className="flex items-center justify-between text-muted-foreground">
                <span>Fecha:</span>
                <span>{relativeTime(bc.created_at)}</span>
              </div>
            </div>

            {/* Acción */}
            <Button
              variant="outline"
              size="sm"
              className="w-full text-xs"
              onClick={() => onOpenDetails(bc)}
            >
              <Eye className="mr-1.5 h-3.5 w-3.5" />
              Ver detalle
            </Button>
          </div>
        )
      })}

      {/* Paginación móvil */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between px-1 py-2 text-xs text-muted-foreground">
          <span>
            Pág. <span className="font-medium text-foreground">{currentPage}</span> de{" "}
            <span className="font-medium text-foreground">{totalPages}</span> ·{" "}
            {data.length} difusiones
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

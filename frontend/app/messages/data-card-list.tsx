"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"
import {
  Eye,
  RefreshCw,
  Paperclip,
  ChevronLeft,
  ChevronRight,
  MessageSquareMore,
  Phone,
} from "lucide-react"
import { type AdminMessage, type Empresa } from "@/lib/api"
import { MessageStatusBadge, CopyNumberButton } from "./columns"

interface DataCardListProps {
  data: AdminMessage[]
  loading?: boolean
  companyByRuc: Map<string, Empresa>
  onOpenDetails: (msg: AdminMessage) => void
  onRetry: (msg: AdminMessage) => void
  retryingId: number | null
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
  onRetry,
  retryingId,
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
              <Skeleton className="h-8 w-full" />
              <div className="flex gap-2 pt-1">
                <Skeleton className="h-9 flex-1" />
                <Skeleton className="h-9 w-9" />
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
            No se encontraron mensajes que coincidan con los filtros.
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-3">
      {paged.map((msg) => {
        const company = companyByRuc.get(msg.account_id)
        const empresaName = company?.nombre || msg.account_id
        const attachmentCount = msg.adjuntos?.length ?? 0
        const isRetrying = retryingId === msg.id

        return (
          <div
            key={msg.id}
            className="flex flex-col gap-3 rounded-xl p-4 ring-1 ring-foreground/10"
          >
            {/* Encabezado: ID + Empresa + Badge */}
            <div className="flex items-start justify-between gap-2">
              <div className="min-w-0 flex-1">
                <div className="flex items-center gap-2">
                  <span className="font-mono text-xs font-semibold text-muted-foreground">
                    #{msg.id}
                  </span>
                  <h3
                    className="truncate text-sm font-medium text-foreground"
                    title={empresaName}
                  >
                    {empresaName}
                  </h3>
                </div>
                <div className="font-mono text-xs text-muted-foreground">
                  {company?.ruc || msg.account_id}
                </div>
              </div>
              <MessageStatusBadge
                status={msg.status}
                retryCount={msg.retry_count}
              />
            </div>

            {/* Info intermedia: Destino + Fecha */}
            <div className="flex flex-wrap items-center justify-between gap-2 border-y border-border py-2 text-xs">
              <div className="flex items-center gap-1.5">
                <Phone className="h-3.5 w-3.5 text-muted-foreground" />
                <span className="text-muted-foreground">Para:</span>
                <CopyNumberButton number={msg.to} />
              </div>
              <span className="text-muted-foreground">{relativeTime(msg.created_at)}</span>
            </div>

            {/* Contenido del mensaje */}
            <div className="flex flex-col gap-1.5">
              <p
                className="line-clamp-2 text-sm text-foreground"
                title={msg.content || "Sin contenido"}
              >
                {msg.content || "Sin contenido"}
              </p>
              <div className="flex flex-wrap items-center gap-2">
                {attachmentCount > 0 && (
                  <Badge variant="outline" className="h-5 gap-1 px-1.5 text-[10px]">
                    <Paperclip className="h-3 w-3" />
                    {attachmentCount === 1 ? "1 adjunto" : `${attachmentCount} adjuntos`}
                  </Badge>
                )}
                {msg.error_reason && (
                  <span
                    className="line-clamp-1 text-xs text-destructive"
                    title={msg.error_reason}
                  >
                    {msg.error_reason}
                  </span>
                )}
              </div>
            </div>

            {/* Acciones */}
            <div className="flex items-center gap-2 pt-1">
              <Button
                variant="outline"
                size="sm"
                className="flex-1 text-xs"
                onClick={() => onOpenDetails(msg)}
              >
                <Eye className="mr-1.5 h-3.5 w-3.5" />
                Ver detalles
              </Button>
              {msg.status === "failed" && msg.reference_id && (
                <Button
                  variant="secondary"
                  size="sm"
                  className="flex-1 text-xs"
                  disabled={isRetrying}
                  onClick={() => onRetry(msg)}
                >
                  <RefreshCw
                    className={cn("mr-1.5 h-3.5 w-3.5", isRetrying && "animate-spin")}
                  />
                  Reintentar
                </Button>
              )}
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
            {data.length} mensajes
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

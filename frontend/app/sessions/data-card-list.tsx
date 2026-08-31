"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Skeleton } from "@/components/ui/skeleton"
import { SessionStatusBadge } from "@/components/session/session-status-badge"
import { cn } from "@/lib/utils"
import {
  Phone,
  QrCode,
  LogOut,
  RefreshCw,
  ClipboardList,
  Copy,
  Check,
  ChevronLeft,
  ChevronRight,
  Smartphone,
} from "lucide-react"
import { type SessionInfo } from "@/lib/api"

interface DataCardListProps {
  data: SessionInfo[]
  loading?: boolean
  onOpenQR: (session: SessionInfo) => void
  onOpenDisconnect: (session: SessionInfo) => void
  onReconnect: (telefonoId: number) => void
  onOpenEvents: (session: SessionInfo) => void
  reconnectingId: number | null
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

function formatLocalTime(ts?: string | null): string {
  if (!ts) return "-"
  try {
    return new Date(ts).toLocaleString("es-PE", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      hour12: false,
    })
  } catch {
    return ts
  }
}

function CopyNumberButton({ number }: { number: string }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = (e: React.MouseEvent) => {
    e.stopPropagation()
    navigator.clipboard.writeText(number)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <button
      onClick={handleCopy}
      className="inline-flex items-center gap-1.5 rounded-md font-mono text-sm text-foreground transition-colors hover:text-muted-foreground"
      title="Copiar número"
    >
      <Phone className="h-3.5 w-3.5 text-muted-foreground" />
      <span>{number}</span>
      {copied ? (
        <Check className="h-3.5 w-3.5 text-emerald-600 dark:text-emerald-400" />
      ) : (
        <Copy className="h-3.5 w-3.5 text-muted-foreground" />
      )}
    </button>
  )
}

export function DataCardList({
  data,
  loading = false,
  onOpenQR,
  onOpenDisconnect,
  onReconnect,
  onOpenEvents,
  reconnectingId,
}: DataCardListProps) {
  const [page, setPage] = useState(1)
  const totalPages = Math.ceil(data.length / PAGE_SIZE) || 1
  const currentPage = Math.min(page, totalPages)
  const pagedData = data.slice((currentPage - 1) * PAGE_SIZE, currentPage * PAGE_SIZE)

  if (loading) {
    return (
      <div className="flex flex-col gap-3">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="rounded-xl p-4 ring-1 ring-foreground/10">
            <div className="flex flex-col gap-3">
              <div className="flex items-center justify-between">
                <Skeleton className="h-5 w-40" />
                <Skeleton className="h-5 w-20" />
              </div>
              <Skeleton className="h-4 w-28" />
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
          <Smartphone className="h-8 w-8 opacity-30" />
          <p className="text-sm font-medium">
            No se encontraron sesiones que coincidan con los filtros.
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-3">
      {pagedData.map((session) => {
        const empresaName = session.empresa_nombre || "Empresa sin nombre"
        const isReconnecting = reconnectingId === session.telefono_id
        const hasEvents = (session.events?.length ?? 0) > 0

        return (
          <div
            key={session.account_id}
            className="flex flex-col gap-3 rounded-xl p-4 ring-1 ring-foreground/10"
          >
            {/* Encabezado: empresa + estado (apilados para dar ancho al nombre) */}
            <div className="flex flex-col gap-2">
              <div className="min-w-0">
                <h3
                  className="truncate text-sm font-medium text-foreground"
                  title={empresaName}
                >
                  {empresaName}
                </h3>
                <span className="text-xs text-muted-foreground">
                  {session.empresa_id
                    ? `Empresa #${session.empresa_id}`
                    : `Teléfono #${session.telefono_id}`}
                </span>
              </div>
              <SessionStatusBadge
                status={session.status}
                reconnecting={session.reconnecting}
                mismatch={session.mismatch}
                runtimeConnected={session.runtime_connected}
                showRuntime
              />
            </div>

            {/* Detalles: número + última conexión */}
            <div className="flex flex-col gap-2 border-y border-border py-2.5">
              <div className="flex items-center justify-between gap-2">
                <span className="text-xs text-muted-foreground">Número / Cuenta</span>
                <CopyNumberButton number={session.account_id} />
              </div>
              <div className="flex items-center justify-between gap-2">
                <span className="text-xs text-muted-foreground">Última conexión</span>
                <span
                  className="text-sm font-medium text-foreground"
                  title={formatLocalTime(session.last_connected)}
                >
                  {relativeTime(session.last_connected)}
                </span>
              </div>
            </div>

            {/* Acciones */}
            <div className="flex items-center gap-2">
              {session.status === "qr_pending" && (
                <Button
                  variant="default"
                  size="sm"
                  className="flex-1"
                  onClick={() => onOpenQR(session)}
                >
                  <QrCode className="mr-1.5 h-4 w-4" />
                  Ver código QR
                </Button>
              )}

              {session.status === "active" && (
                <Button
                  variant="outline"
                  size="sm"
                  className="flex-1 text-destructive hover:text-destructive"
                  onClick={() => onOpenDisconnect(session)}
                >
                  <LogOut className="mr-1.5 h-4 w-4" />
                  Desconectar
                </Button>
              )}

              {session.status !== "active" &&
                session.telefono_id != null &&
                !session.reconnecting && (
                  <Button
                    variant="outline"
                    size="sm"
                    className="flex-1"
                    disabled={isReconnecting}
                    onClick={() => onReconnect(session.telefono_id!)}
                  >
                    <RefreshCw
                      className={cn("mr-1.5 h-4 w-4", isReconnecting && "animate-spin")}
                    />
                    Reconectar
                  </Button>
                )}

              {hasEvents && (
                <Button
                  variant="ghost"
                  size="icon"
                  className="size-9 shrink-0"
                  onClick={() => onOpenEvents(session)}
                  aria-label="Historial de eventos"
                >
                  <ClipboardList className="h-4 w-4 text-muted-foreground" />
                </Button>
              )}
            </div>
          </div>
        )
      })}

      {/* Paginación */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between px-1 py-2 text-xs text-muted-foreground">
          <span>
            Pág. <span className="font-medium text-foreground">{currentPage}</span> de{" "}
            <span className="font-medium text-foreground">{totalPages}</span> ·{" "}
            {data.length} sesiones
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

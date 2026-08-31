"use client"

import { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { SessionStatusBadge } from "@/components/session/session-status-badge"
import {
  ArrowUpDown,
  Phone,
  QrCode,
  LogOut,
  RefreshCw,
  ClipboardList,
  Copy,
  Check,
} from "lucide-react"
import { type SessionInfo } from "@/lib/api"
import { useState } from "react"

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
      second: "2-digit",
      hour12: false,
    })
  } catch {
    return ts
  }
}

interface ColumnActionCallbacks {
  onOpenQR: (session: SessionInfo) => void
  onOpenDisconnect: (session: SessionInfo) => void
  onReconnect: (telefonoId: number) => void
  onOpenEvents: (session: SessionInfo) => void
  reconnectingId: number | null
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
      className="inline-flex items-center gap-1 font-mono text-xs text-muted-foreground hover:text-foreground transition-colors group p-0.5 rounded cursor-pointer"
      title="Copiar número"
    >
      <span>{number}</span>
      {copied ? (
        <Check className="h-3 w-3 text-emerald-600 dark:text-emerald-400" />
      ) : (
        <Copy className="h-3 w-3 opacity-0 group-hover:opacity-100 transition-opacity" />
      )}
    </button>
  )
}

export function getColumns(callbacks: ColumnActionCallbacks): ColumnDef<SessionInfo>[] {
  return [
    {
      id: "empresa",
      accessorFn: (row) => row.empresa_nombre || row.account_id,
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            className="-ml-3 h-8 text-xs font-semibold hover:bg-muted/50"
          >
            Empresa
            <ArrowUpDown className="ml-1 h-3 w-3" />
          </Button>
        )
      },
      cell: ({ row }) => {
        const session = row.original
        const empresaName = session.empresa_nombre || "Empresa sin nombre"
        return (
          <div className="max-w-[240px]">
            <div className="truncate text-sm font-medium text-foreground" title={empresaName}>
              {empresaName}
            </div>
            <div className="text-xs text-muted-foreground">
              {session.empresa_id
                ? `Empresa #${session.empresa_id}`
                : `Teléfono #${session.telefono_id}`}
            </div>
          </div>
        )
      },
    },
    {
      accessorKey: "account_id",
      header: "Número / Cuenta",
      cell: ({ row }) => {
        const accountId = row.original.account_id
        return (
          <div className="flex items-center gap-1.5">
            <Phone className="h-3 w-3 text-muted-foreground shrink-0" />
            <CopyNumberButton number={accountId} />
          </div>
        )
      },
    },
    {
      accessorKey: "status",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            className="-ml-3 h-8 text-xs font-semibold hover:bg-muted/50"
          >
            Estado
            <ArrowUpDown className="ml-1 h-3 w-3" />
          </Button>
        )
      },
      cell: ({ row }) => {
        const session = row.original
        return (
          <SessionStatusBadge
            status={session.status}
            reconnecting={session.reconnecting}
            mismatch={session.mismatch}
            runtimeConnected={session.runtime_connected}
          />
        )
      },
    },
    {
      accessorKey: "runtime_connected",
      header: () => <div className="text-center">Runtime</div>,
      cell: ({ row }) => {
        const isOnline = row.original.runtime_connected
        return (
          <div className="flex justify-center">
            <span
              className={`inline-flex items-center gap-1.5 text-xs font-medium ${
                isOnline ? "text-emerald-600 dark:text-emerald-400" : "text-muted-foreground"
              }`}
              title={isOnline ? "Runtime conectado en memoria" : "Runtime desconectado"}
            >
              <span
                className={`h-1.5 w-1.5 rounded-full ${
                  isOnline ? "bg-emerald-500" : "bg-muted-foreground/40"
                }`}
              />
              {isOnline ? "Online" : "Offline"}
            </span>
          </div>
        )
      },
    },
    {
      accessorKey: "last_connected",
      header: ({ column }) => {
        return (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
            className="-ml-3 h-8 text-xs font-semibold hover:bg-muted/50"
          >
            Última Conexión
            <ArrowUpDown className="ml-1 h-3 w-3" />
          </Button>
        )
      },
      cell: ({ row }) => {
        const session = row.original
        if (!session.last_connected) {
          return <span className="text-xs text-muted-foreground/50">Nunca</span>
        }
        return (
          <div className="text-xs text-muted-foreground" title={formatLocalTime(session.last_connected)}>
            <span className="font-medium text-foreground">{relativeTime(session.last_connected)}</span>
            <span className="ml-1 hidden text-muted-foreground/70 xl:inline">
              · {formatLocalTime(session.last_connected)}
            </span>
          </div>
        )
      },
    },
    {
      id: "actions",
      header: () => <div className="text-right pr-2">Acciones</div>,
      cell: ({ row }) => {
        const session = row.original
        const isReconnecting = callbacks.reconnectingId === session.telefono_id
        const hasEvents = (session.events?.length ?? 0) > 0

        return (
          <div className="flex items-center justify-end gap-1.5">
            {/* Botón QR */}
            {session.status === "qr_pending" && (
              <Button
                variant="outline"
                size="sm"
                className="h-8 px-2.5 text-xs"
                onClick={() => callbacks.onOpenQR(session)}
              >
                <QrCode className="mr-1 h-3.5 w-3.5 text-muted-foreground" />
                QR
              </Button>
            )}

            {/* Botón Desconectar */}
            {session.status === "active" && (
              <Button
                variant="outline"
                size="sm"
                className="h-8 px-2.5 text-xs text-muted-foreground hover:border-destructive/30 hover:text-destructive"
                onClick={() => callbacks.onOpenDisconnect(session)}
              >
                <LogOut className="mr-1 h-3.5 w-3.5" />
                Desconectar
              </Button>
            )}

            {/* Botón Reconectar */}
            {session.status !== "active" && session.telefono_id != null && !session.reconnecting && (
              <Button
                variant="outline"
                size="sm"
                className="h-8 px-2.5 text-xs"
                disabled={isReconnecting}
                onClick={() => {
                  if (session.telefono_id != null) {
                    callbacks.onReconnect(session.telefono_id)
                  }
                }}
              >
                <RefreshCw
                  className={`mr-1 h-3.5 w-3.5 ${isReconnecting ? "animate-spin" : ""}`}
                />
                Reconectar
              </Button>
            )}

            {/* Botón Historial de eventos */}
            {hasEvents && (
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8 text-muted-foreground hover:text-foreground"
                onClick={() => callbacks.onOpenEvents(session)}
                title="Historial de eventos"
                aria-label="Historial de eventos"
              >
                <ClipboardList className="h-3.5 w-3.5" />
              </Button>
            )}
          </div>
        )
      },
    },
  ]
}

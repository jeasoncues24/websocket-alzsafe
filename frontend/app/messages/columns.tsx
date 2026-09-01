"use client"

import { useState } from "react"
import { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
  ArrowUpDown,
  Phone,
  Eye,
  RefreshCw,
  Paperclip,
  Check,
  Copy,
  AlertCircle,
  CheckCircle2,
  Clock,
  Send,
} from "lucide-react"
import { type AdminMessage, type Empresa } from "@/lib/api"

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

export function MessageStatusBadge({
  status,
  retryCount,
}: {
  status: string
  retryCount?: number
}) {
  switch (status) {
    case "sent":
      return (
        <Badge variant="secondary" className="gap-1 font-medium">
          <Send className="h-3 w-3" />
          <span>Enviado{retryCount ? ` (${retryCount})` : ""}</span>
        </Badge>
      )
    case "delivered":
      return (
        <Badge
          variant="outline"
          className="gap-1 border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400 font-medium"
        >
          <CheckCircle2 className="h-3 w-3 text-emerald-600 dark:text-emerald-400" />
          <span>Entregado</span>
        </Badge>
      )
    case "failed":
      return (
        <Badge variant="destructive" className="gap-1 font-medium">
          <AlertCircle className="h-3 w-3" />
          <span>Fallido{retryCount ? ` (${retryCount})` : ""}</span>
        </Badge>
      )
    case "pending":
      return (
        <Badge variant="outline" className="gap-1 font-medium text-muted-foreground">
          <Clock className="h-3 w-3" />
          <span>Pendiente</span>
        </Badge>
      )
    default:
      return <Badge variant="outline">{status}</Badge>
  }
}

export function CopyNumberButton({ number }: { number: string }) {
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
      type="button"
      className="group inline-flex items-center gap-1 rounded p-0.5 font-mono text-xs text-muted-foreground transition-colors hover:text-foreground cursor-pointer"
      title="Copiar número"
    >
      <span>{number}</span>
      {copied ? (
        <Check className="h-3 w-3 text-emerald-600 dark:text-emerald-400" />
      ) : (
        <Copy className="h-3 w-3 opacity-0 transition-opacity group-hover:opacity-100" />
      )}
    </button>
  )
}

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
  onOpenDetails: (msg: AdminMessage) => void
  onRetry: (msg: AdminMessage) => void
  retryingId: number | null
  companyByRuc: Map<string, Empresa>
}

export function getColumns(cb: ColumnActionCallbacks): ColumnDef<AdminMessage>[] {
  return [
    {
      accessorKey: "id",
      header: ({ column }) => <SortableHeader column={column}>ID</SortableHeader>,
      cell: ({ row }) => (
        <span className="font-mono text-xs font-medium text-foreground">
          #{row.original.id}
        </span>
      ),
    },
    {
      id: "empresa",
      accessorFn: (row) => {
        const company = cb.companyByRuc.get(row.account_id)
        return company?.nombre || row.account_id
      },
      header: ({ column }) => <SortableHeader column={column}>Empresa</SortableHeader>,
      cell: ({ row }) => {
        const msg = row.original
        const company = cb.companyByRuc.get(msg.account_id)
        const empresaName = company?.nombre || msg.account_id
        return (
          <div className="max-w-[200px]">
            <div className="truncate font-medium text-foreground" title={empresaName}>
              {empresaName}
            </div>
            <div className="font-mono text-xs text-muted-foreground">
              {company?.ruc || msg.account_id}
            </div>
          </div>
        )
      },
    },
    {
      accessorKey: "to",
      header: "Destino",
      cell: ({ row }) => (
        <div className="flex items-center gap-1.5">
          <Phone className="h-3 w-3 shrink-0 text-muted-foreground" />
          <CopyNumberButton number={row.original.to} />
        </div>
      ),
    },
    {
      accessorKey: "content",
      header: "Mensaje",
      cell: ({ row }) => {
        const msg = row.original
        const attachmentCount = msg.adjuntos?.length ?? 0
        return (
          <div className="max-w-[30rem] whitespace-normal">
            <div className="flex flex-col gap-1">
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
          </div>
        )
      },
    },
    {
      accessorKey: "status",
      header: ({ column }) => <SortableHeader column={column}>Estado</SortableHeader>,
      cell: ({ row }) => (
        <MessageStatusBadge
          status={row.original.status}
          retryCount={row.original.retry_count}
        />
      ),
    },
    {
      id: "created_at",
      accessorFn: (row) => (row.created_at ? new Date(row.created_at).getTime() : 0),
      header: ({ column }) => <SortableHeader column={column}>Fecha</SortableHeader>,
      cell: ({ row }) => {
        const msg = row.original
        return (
          <div
            className="text-xs text-muted-foreground"
            title={formatLocalTime(msg.created_at)}
          >
            <span className="font-medium text-foreground">
              {relativeTime(msg.created_at)}
            </span>
            <span className="ml-1 hidden text-muted-foreground/70 xl:inline">
              · {formatLocalTime(msg.created_at)}
            </span>
          </div>
        )
      },
    },
    {
      id: "actions",
      header: () => <div className="pr-2 text-right">Acciones</div>,
      cell: ({ row }) => {
        const msg = row.original
        const isRetrying = cb.retryingId === msg.id

        return (
          <div className="flex items-center justify-end gap-1.5">
            <Button
              variant="ghost"
              size="sm"
              className="h-8 px-2.5 text-xs text-muted-foreground hover:text-foreground"
              onClick={() => cb.onOpenDetails(msg)}
            >
              <Eye className="mr-1.5 h-3.5 w-3.5" />
              Detalles
            </Button>
            {msg.status === "failed" && msg.reference_id && (
              <Button
                variant="outline"
                size="sm"
                className="h-8 px-2.5 text-xs"
                disabled={isRetrying}
                onClick={() => cb.onRetry(msg)}
              >
                <RefreshCw
                  className={`mr-1.5 h-3.5 w-3.5 ${isRetrying ? "animate-spin" : ""}`}
                />
                Reintentar
              </Button>
            )}
          </div>
        )
      },
    },
  ]
}

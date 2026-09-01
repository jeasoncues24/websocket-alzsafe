"use client"

import { useState } from "react"
import { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import {
  ArrowUpDown,
  Eye,
  Check,
  Copy,
} from "lucide-react"
import { type BroadcastInfo, type Empresa } from "@/lib/api"
import { BroadcastStatusBadge, ProgressBar } from "./broadcast-detail-sheet"

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

export function CopyRefButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = (e: React.MouseEvent) => {
    e.stopPropagation()
    navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <button
      onClick={handleCopy}
      type="button"
      className="group inline-flex cursor-pointer items-center gap-1 rounded p-0.5 font-mono text-xs text-muted-foreground transition-colors hover:text-foreground"
      title="Copiar ID completo"
    >
      <span>{text.length > 12 ? `${text.slice(0, 8)}…` : text}</span>
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
  onOpenDetails: (bc: BroadcastInfo) => void
  companyByRuc: Map<string, Empresa>
}

export function getColumns(cb: ColumnActionCallbacks): ColumnDef<BroadcastInfo>[] {
  return [
    {
      accessorKey: "reference_id",
      header: ({ column }) => <SortableHeader column={column}>Ref ID</SortableHeader>,
      cell: ({ row }) => <CopyRefButton text={row.original.reference_id} />,
    },
    {
      id: "empresa",
      accessorFn: (row) => {
        const company = cb.companyByRuc.get(row.ruc_empresa)
        return company?.nombre || row.ruc_empresa
      },
      header: ({ column }) => <SortableHeader column={column}>Empresa</SortableHeader>,
      cell: ({ row }) => {
        const bc = row.original
        const company = cb.companyByRuc.get(bc.ruc_empresa)
        const empresaName = company?.nombre || bc.ruc_empresa
        return (
          <div className="max-w-[220px]">
            <div className="truncate font-medium text-foreground" title={empresaName}>
              {empresaName}
            </div>
            <div className="font-mono text-xs text-muted-foreground">
              {company?.ruc || bc.ruc_empresa}
            </div>
          </div>
        )
      },
    },
    {
      id: "progreso",
      accessorFn: (row) => (row.total > 0 ? (row.success ?? 0) / row.total : 0),
      header: ({ column }) => <SortableHeader column={column}>Progreso</SortableHeader>,
      cell: ({ row }) => (
        <ProgressBar
          sent={row.original.success ?? 0}
          total={row.original.total}
        />
      ),
    },
    {
      accessorKey: "status",
      header: ({ column }) => <SortableHeader column={column}>Estado</SortableHeader>,
      cell: ({ row }) => <BroadcastStatusBadge status={row.original.status} />,
    },
    {
      id: "created_at",
      accessorFn: (row) => (row.created_at ? new Date(row.created_at).getTime() : 0),
      header: ({ column }) => <SortableHeader column={column}>Fecha</SortableHeader>,
      cell: ({ row }) => {
        const bc = row.original
        return (
          <div
            className="text-xs text-muted-foreground"
            title={formatLocalTime(bc.created_at)}
          >
            <span className="font-medium text-foreground">
              {relativeTime(bc.created_at)}
            </span>
            <span className="ml-1 hidden text-muted-foreground/70 xl:inline">
              · {formatLocalTime(bc.created_at)}
            </span>
          </div>
        )
      },
    },
    {
      id: "actions",
      header: () => <div className="pr-2 text-right">Acciones</div>,
      cell: ({ row }) => (
        <div className="flex items-center justify-end">
          <Button
            variant="ghost"
            size="sm"
            className="h-8 px-2.5 text-xs text-muted-foreground hover:text-foreground"
            onClick={() => cb.onOpenDetails(row.original)}
          >
            <Eye className="mr-1.5 h-3.5 w-3.5" />
            Detalle
          </Button>
        </div>
      ),
    },
  ]
}

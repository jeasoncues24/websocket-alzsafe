"use client"

import { useCallback, useEffect, useMemo, useRef, useState } from "react"
import {
  CheckCircle2,
  Clock,
  FileText,
  Loader2,
  XCircle,
} from "lucide-react"
import { Badge } from "@/components/ui/badge"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import {
  getAdminBroadcastDetail,
  type BroadcastDetail,
  type BroadcastItemResult,
} from "@/lib/api"

function formatDate(value?: string) {
  if (!value) return "-"
  try {
    return new Date(value).toLocaleString("es-PE")
  } catch {
    return value
  }
}

function formatPhone(phone: string) {
  if (phone.length > 6) {
    return `+${phone.slice(0, phone.length - 4).replace(/./g, "·")}${phone.slice(-4)}`
  }
  return phone
}

function formatSecondsRemaining(estimatedSeconds: number, createdAt: string): string {
  const elapsed = Math.floor((Date.now() - new Date(createdAt).getTime()) / 1000)
  const remaining = Math.max(0, estimatedSeconds - elapsed)
  if (remaining === 0) return "completando..."
  const mins = Math.floor(remaining / 60)
  const secs = remaining % 60
  return mins > 0 ? `~${mins} min ${secs}s restantes` : `~${secs}s restantes`
}

const JOB_STATUS = {
  completed: { label: "Completado", variant: "default" as const, icon: CheckCircle2, pulse: false },
  running: { label: "Enviando", variant: "secondary" as const, icon: Loader2, pulse: true },
  pending: { label: "Pendiente", variant: "outline" as const, icon: Clock, pulse: false },
  failed: { label: "Fallido", variant: "destructive" as const, icon: XCircle, pulse: false },
  cancelled: { label: "Cancelado", variant: "outline" as const, icon: XCircle, pulse: false },
}

export function BroadcastStatusBadge({ status }: { status: string }) {
  const cfg =
    JOB_STATUS[status as keyof typeof JOB_STATUS] ?? {
      label: status,
      variant: "outline" as const,
      icon: Clock,
      pulse: false,
    }
  const Icon = cfg.icon
  return (
    <Badge variant={cfg.variant} className="gap-1.5 font-medium">
      <Icon className={`h-3 w-3 ${cfg.pulse ? "animate-spin" : ""}`} />
      {cfg.label}
    </Badge>
  )
}

const ITEM_STATUS = {
  sent: { label: "Enviado", variant: "default" as const },
  failed: { label: "Fallido", variant: "destructive" as const },
  pending: { label: "Pendiente", variant: "outline" as const },
  skipped: { label: "Omitido", variant: "secondary" as const },
}

function ItemBadge({ status }: { status: BroadcastItemResult["status"] }) {
  const cfg = ITEM_STATUS[status] ?? { label: status, variant: "outline" as const }
  return <Badge variant={cfg.variant} className="text-xs">{cfg.label}</Badge>
}

export function ProgressBar({ sent, total }: { sent: number; total: number }) {
  const pct = total > 0 ? Math.round((sent / total) * 100) : 0
  return (
    <div className="flex items-center gap-2">
      <div className="h-1.5 w-20 overflow-hidden rounded-full bg-muted">
        <div
          className="h-full rounded-full bg-emerald-500 transition-all duration-500"
          style={{ width: `${pct}%` }}
        />
      </div>
      <span className="tabular-nums text-xs text-muted-foreground">
        {sent}/{total}
      </span>
    </div>
  )
}

type FilterTab = "all" | "sent" | "failed" | "pending"

function useBroadcastDetail(referenceId: string | null) {
  const [detail, setDetail] = useState<BroadcastDetail | null>(null)
  const [loading, setLoading] = useState(false)
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const fetch = useCallback(async (id: string) => {
    try {
      const res = await getAdminBroadcastDetail(id)
      if (res.ok) setDetail(res.data)
    } catch {
      // silencioso — mantener el último estado conocido
    }
  }, [])

  useEffect(() => {
    if (!referenceId) {
      setDetail(null)
      return
    }

    setLoading(true)
    fetch(referenceId).finally(() => setLoading(false))

    // Polling para difusiones activas
    intervalRef.current = setInterval(async () => {
      setDetail((prev) => {
        if (!prev) return null
        const active = prev.status === "running" || prev.status === "pending"
        if (!active && intervalRef.current) {
          clearInterval(intervalRef.current)
        }
        return prev
      })
      await fetch(referenceId)
    }, 3000)

    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current)
    }
  }, [referenceId, fetch])

  return { detail, loading }
}

interface BroadcastDetailSheetProps {
  open: boolean
  onOpenChange: (v: boolean) => void
  referenceId: string | null
  empresaNombre: string
}

export function BroadcastDetailSheet({
  open,
  onOpenChange,
  referenceId,
  empresaNombre,
}: BroadcastDetailSheetProps) {
  const { detail, loading } = useBroadcastDetail(open ? referenceId : null)
  const [activeTab, setActiveTab] = useState<FilterTab>("all")

  const items = useMemo(() => detail?.items ?? [], [detail])
  const sent = useMemo(() => items.filter((i) => i.status === "sent").length, [items])
  const failed = useMemo(() => items.filter((i) => i.status === "failed").length, [items])
  const pending = useMemo(() => items.filter((i) => i.status === "pending").length, [items])

  const filteredItems = useMemo(() => {
    if (activeTab === "all") return items
    if (activeTab === "sent") return items.filter((i) => i.status === "sent")
    if (activeTab === "failed") return items.filter((i) => i.status === "failed")
    if (activeTab === "pending")
      return items.filter((i) => i.status === "pending" || i.status === "skipped")
    return items
  }, [items, activeTab])

  const isActive = detail?.status === "running" || detail?.status === "pending"
  const total = detail?.total ?? 0
  const pct = total > 0 ? Math.round((sent / total) * 100) : 0

  const TABS: { key: FilterTab; label: string; count: number }[] = [
    { key: "all", label: "Todos", count: items.length },
    { key: "sent", label: "Enviados", count: sent },
    { key: "failed", label: "Fallidos", count: failed },
    { key: "pending", label: "Pendientes", count: pending },
  ]

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-full overflow-y-auto sm:max-w-3xl">
        <SheetHeader className="text-left">
          <SheetTitle>Detalle de difusión</SheetTitle>
          <SheetDescription>
            {referenceId ? `Ref: ${referenceId}` : "Selecciona una difusión"}
          </SheetDescription>
        </SheetHeader>

        {loading && !detail ? (
          <div className="mt-10 flex justify-center">
            <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
          </div>
        ) : detail ? (
          <div className="mt-6 space-y-5">
            {/* Empresa + Estado */}
            <div className="grid gap-3 sm:grid-cols-2">
              <div className="rounded-lg border p-4">
                <p className="text-xs uppercase tracking-wide text-muted-foreground">Empresa</p>
                <p className="mt-1 font-medium">{empresaNombre || detail.ruc_empresa}</p>
                <p className="font-mono text-xs text-muted-foreground">{detail.ruc_empresa}</p>
              </div>
              <div className="flex flex-col gap-2 rounded-lg border p-4">
                <p className="text-xs uppercase tracking-wide text-muted-foreground">Estado</p>
                <div className="flex flex-wrap items-center gap-2">
                  <BroadcastStatusBadge status={detail.status} />
                  {isActive && (
                    <span className="text-xs text-muted-foreground">
                      {detail.estimated_seconds
                        ? formatSecondsRemaining(detail.estimated_seconds, detail.created_at)
                        : "procesando…"}
                    </span>
                  )}
                </div>
              </div>
            </div>

            {/* Barra de progreso */}
            <div className="space-y-3 rounded-lg border p-4">
              <div className="flex items-center justify-between">
                <p className="text-xs uppercase tracking-wide text-muted-foreground">Progreso</p>
                <span className="tabular-nums text-sm font-semibold">{pct}%</span>
              </div>
              <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
                <div
                  className="h-full rounded-full bg-emerald-500 transition-all duration-700"
                  style={{ width: `${pct}%` }}
                />
              </div>
              <div className="grid grid-cols-3 gap-3 pt-1">
                <div className="text-center">
                  <p className="tabular-nums text-2xl font-semibold text-emerald-600 dark:text-emerald-400">
                    {sent}
                  </p>
                  <p className="text-xs text-muted-foreground">Enviados</p>
                </div>
                <div className="text-center">
                  <p className="tabular-nums text-2xl font-semibold text-destructive">
                    {failed}
                  </p>
                  <p className="text-xs text-muted-foreground">Fallidos</p>
                </div>
                <div className="text-center">
                  <p className="tabular-nums text-2xl font-semibold text-foreground">
                    {pending}
                  </p>
                  <p className="text-xs text-muted-foreground">Pendientes</p>
                </div>
              </div>
            </div>

            {/* Adjuntos */}
            {detail.adjuntos && detail.adjuntos.length > 0 && (
              <div className="space-y-2 rounded-lg border p-4">
                <p className="text-xs uppercase tracking-wide text-muted-foreground">Adjuntos</p>
                {detail.adjuntos.map((att) => (
                  <div
                    key={att.sha256_hash}
                    className="flex items-center gap-2 rounded-md border bg-muted/30 p-2"
                  >
                    <FileText className="h-4 w-4 shrink-0 text-muted-foreground" />
                    <div className="min-w-0">
                      <p className="truncate text-sm font-medium">{att.nombre}</p>
                      <p className="text-xs text-muted-foreground">
                        {att.tamano_bytes.toLocaleString()} bytes
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {/* Tabla de destinatarios */}
            <div className="overflow-hidden rounded-lg border">
              {/* Tabs filtro */}
              <div className="flex overflow-x-auto border-b">
                {TABS.map((tab) => (
                  <button
                    key={tab.key}
                    type="button"
                    onClick={() => setActiveTab(tab.key)}
                    className={`flex shrink-0 cursor-pointer items-center gap-1.5 px-4 py-2.5 text-sm transition-colors ${
                      activeTab === tab.key
                        ? "border-b-2 border-primary font-medium text-foreground"
                        : "text-muted-foreground hover:text-foreground"
                    }`}
                  >
                    {tab.label}
                    {tab.count > 0 && (
                      <span
                        className={`tabular-nums rounded-full px-1.5 py-0.5 text-xs ${
                          activeTab === tab.key
                            ? "bg-primary text-primary-foreground"
                            : "bg-muted"
                        }`}
                      >
                        {tab.count}
                      </span>
                    )}
                  </button>
                ))}
              </div>

              {filteredItems.length === 0 ? (
                <p className="py-8 text-center text-sm text-muted-foreground">
                  No hay destinatarios en este filtro.
                </p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-sm">
                    <thead>
                      <tr className="border-b bg-muted/40">
                        <th className="px-4 py-2.5 text-left font-medium text-muted-foreground">#</th>
                        <th className="px-4 py-2.5 text-left font-medium text-muted-foreground">Destino</th>
                        <th className="px-4 py-2.5 text-left font-medium text-muted-foreground">Estado</th>
                        <th className="px-4 py-2.5 text-left font-medium text-muted-foreground">Hora</th>
                        <th className="px-4 py-2.5 text-left font-medium text-muted-foreground">Error</th>
                      </tr>
                    </thead>
                    <tbody>
                      {filteredItems.map((item) => (
                        <tr
                          key={item.id}
                          className="border-b transition-colors last:border-0 hover:bg-muted/20"
                        >
                          <td className="tabular-nums px-4 py-2.5 text-xs text-muted-foreground">
                            {item.sequence_order + 1}
                          </td>
                          <td className="font-mono px-4 py-2.5 text-xs">
                            {formatPhone(item.destino)}
                          </td>
                          <td className="px-4 py-2.5">
                            <ItemBadge status={item.status} />
                          </td>
                          <td className="whitespace-nowrap px-4 py-2.5 text-xs text-muted-foreground">
                            {item.processed_at
                              ? new Date(item.processed_at).toLocaleTimeString("es-PE")
                              : "—"}
                          </td>
                          <td
                            className="max-w-[160px] truncate px-4 py-2.5 text-xs text-destructive"
                            title={item.error_text ?? ""}
                          >
                            {item.error_text ?? "—"}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>

            <p className="text-xs text-muted-foreground">
              Creado: {formatDate(detail.created_at)}
            </p>
          </div>
        ) : null}
      </SheetContent>
    </Sheet>
  )
}

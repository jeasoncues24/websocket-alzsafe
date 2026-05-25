"use client"

import { useCallback, useEffect, useMemo, useRef, useState } from "react"
import { CheckCircle2, Clock, Eye, FileText, Loader2, MessageSquareMore, XCircle } from "lucide-react"
import { DataEmptyState } from "@/components/feedback/data-empty-state"
import { TableLoadingRows } from "@/components/feedback/table-loading-rows"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import {
  getAdminBroadcasts,
  getAdminBroadcastDetail,
  getEmpresas,
  type BroadcastDetail,
  type BroadcastInfo,
  type BroadcastItemResult,
  type Empresa,
} from "@/lib/api"

// ─── helpers ────────────────────────────────────────────────────────────────

function formatDate(value?: string) {
  return value ? new Date(value).toLocaleString("es-PE") : "-"
}

function formatPhone(phone: string) {
  // Muestra solo los últimos 4 dígitos por privacidad
  if (phone.length > 6) return `+${phone.slice(0, phone.length - 4).replace(/./g, "·")}${phone.slice(-4)}`
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

// ─── status visuals ──────────────────────────────────────────────────────────

const JOB_STATUS = {
  completed: { label: "Completado",  variant: "default"      as const, icon: CheckCircle2, pulse: false },
  running:   { label: "Enviando",    variant: "secondary"    as const, icon: Loader2,      pulse: true  },
  pending:   { label: "Pendiente",   variant: "outline"      as const, icon: Clock,        pulse: false },
  failed:    { label: "Fallido",     variant: "destructive"  as const, icon: XCircle,      pulse: false },
  cancelled: { label: "Cancelado",   variant: "outline"      as const, icon: XCircle,      pulse: false },
}

function StatusBadge({ status }: { status: string }) {
  const cfg = JOB_STATUS[status as keyof typeof JOB_STATUS] ?? { label: status, variant: "outline" as const, icon: Clock, pulse: false }
  const Icon = cfg.icon
  return (
    <Badge variant={cfg.variant} className="gap-1.5">
      <Icon className={`h-3 w-3 ${cfg.pulse ? "animate-spin" : ""}`} />
      {cfg.label}
    </Badge>
  )
}

const ITEM_STATUS = {
  sent:    { label: "Enviado",   variant: "default"     as const },
  failed:  { label: "Fallido",   variant: "destructive" as const },
  pending: { label: "Pendiente", variant: "outline"     as const },
  skipped: { label: "Omitido",   variant: "secondary"   as const },
}

function ItemBadge({ status }: { status: BroadcastItemResult["status"] }) {
  const cfg = ITEM_STATUS[status] ?? { label: status, variant: "outline" as const }
  return <Badge variant={cfg.variant} className="text-xs">{cfg.label}</Badge>
}

// ─── progress bar ────────────────────────────────────────────────────────────

function ProgressBar({ sent, total }: { sent: number; total: number }) {
  const pct = total > 0 ? Math.round((sent / total) * 100) : 0
  return (
    <div className="flex items-center gap-2">
      <div className="h-1.5 w-20 rounded-full bg-muted overflow-hidden">
        <div
          className="h-full rounded-full bg-emerald-500 transition-all duration-500"
          style={{ width: `${pct}%` }}
        />
      </div>
      <span className="text-xs text-muted-foreground tabular-nums">{sent}/{total}</span>
    </div>
  )
}

// ─── hooks ───────────────────────────────────────────────────────────────────

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
    if (!referenceId) { setDetail(null); return }

    setLoading(true)
    fetch(referenceId).finally(() => setLoading(false))

    // polling para jobs activos
    intervalRef.current = setInterval(async () => {
      setDetail(prev => {
        if (!prev) return null
        const active = prev.status === "running" || prev.status === "pending"
        if (!active) {
          clearInterval(intervalRef.current!)
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

// ─── detail sheet ────────────────────────────────────────────────────────────

function BroadcastDetailSheet({
  open,
  onOpenChange,
  referenceId,
  empresaNombre,
}: {
  open: boolean
  onOpenChange: (v: boolean) => void
  referenceId: string | null
  empresaNombre: string
}) {
  const { detail, loading } = useBroadcastDetail(open ? referenceId : null)
  const [activeTab, setActiveTab] = useState<FilterTab>("all")

  const items = useMemo(() => detail?.items ?? [], [detail])
  const sent    = useMemo(() => items.filter(i => i.status === "sent").length,    [items])
  const failed  = useMemo(() => items.filter(i => i.status === "failed").length,  [items])
  const pending = useMemo(() => items.filter(i => i.status === "pending").length, [items])

  const filteredItems = useMemo(() => {
    if (activeTab === "all")     return items
    if (activeTab === "sent")    return items.filter(i => i.status === "sent")
    if (activeTab === "failed")  return items.filter(i => i.status === "failed")
    if (activeTab === "pending") return items.filter(i => i.status === "pending" || i.status === "skipped")
    return items
  }, [items, activeTab])

  const isActive = detail?.status === "running" || detail?.status === "pending"
  const total = detail?.total ?? 0
  const pct = total > 0 ? Math.round((sent / total) * 100) : 0

  const TABS: { key: FilterTab; label: string; count: number }[] = [
    { key: "all",     label: "Todos",    count: items.length },
    { key: "sent",    label: "Enviados", count: sent },
    { key: "failed",  label: "Fallidos", count: failed },
    { key: "pending", label: "Pendientes", count: pending },
  ]

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent side="right" className="w-full overflow-y-auto sm:max-w-3xl">
        <SheetHeader className="text-left">
          <SheetTitle>Detalle de difusión</SheetTitle>
          <SheetDescription>
            {referenceId ? `Ref: ${referenceId.slice(0, 16)}…` : "Selecciona una difusión"}
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
                <p className="mt-1 font-medium">{empresaNombre}</p>
                <p className="text-xs text-muted-foreground">{detail.ruc_empresa}</p>
              </div>
              <div className="rounded-lg border p-4 flex flex-col gap-2">
                <p className="text-xs uppercase tracking-wide text-muted-foreground">Estado</p>
                <div className="flex items-center gap-2 flex-wrap">
                  <StatusBadge status={detail.status} />
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
            <div className="rounded-lg border p-4 space-y-3">
              <div className="flex items-center justify-between">
                <p className="text-xs uppercase tracking-wide text-muted-foreground">Progreso</p>
                <span className="text-sm font-semibold tabular-nums">{pct}%</span>
              </div>
              <div className="h-2 w-full rounded-full bg-muted overflow-hidden">
                <div
                  className="h-full rounded-full bg-emerald-500 transition-all duration-700"
                  style={{ width: `${pct}%` }}
                />
              </div>
              <div className="grid grid-cols-3 gap-3 pt-1">
                <div className="text-center">
                  <p className="text-2xl font-semibold tabular-nums text-emerald-600">{sent}</p>
                  <p className="text-xs text-muted-foreground">Enviados</p>
                </div>
                <div className="text-center">
                  <p className="text-2xl font-semibold tabular-nums text-destructive">{failed}</p>
                  <p className="text-xs text-muted-foreground">Fallidos</p>
                </div>
                <div className="text-center">
                  <p className="text-2xl font-semibold tabular-nums">{pending}</p>
                  <p className="text-xs text-muted-foreground">Pendientes</p>
                </div>
              </div>
            </div>

            {/* Adjuntos */}
            {detail.adjuntos && detail.adjuntos.length > 0 && (
              <div className="rounded-lg border p-4 space-y-2">
                <p className="text-xs uppercase tracking-wide text-muted-foreground">Adjuntos</p>
                {detail.adjuntos.map(att => (
                  <div key={att.sha256_hash} className="flex items-center gap-2 rounded-md border bg-muted/30 p-2">
                    <FileText className="h-4 w-4 shrink-0 text-muted-foreground" />
                    <div className="min-w-0">
                      <p className="truncate text-sm font-medium">{att.nombre}</p>
                      <p className="text-xs text-muted-foreground">{att.tamano_bytes.toLocaleString()} bytes</p>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {/* Tabla de destinatarios */}
            <div className="rounded-lg border overflow-hidden">
              {/* Tabs filtro */}
              <div className="flex border-b overflow-x-auto">
                {TABS.map(tab => (
                  <button
                    key={tab.key}
                    onClick={() => setActiveTab(tab.key)}
                    className={`flex shrink-0 items-center gap-1.5 px-4 py-2.5 text-sm transition-colors cursor-pointer
                      ${activeTab === tab.key
                        ? "border-b-2 border-primary font-medium text-foreground"
                        : "text-muted-foreground hover:text-foreground"}`}
                  >
                    {tab.label}
                    {tab.count > 0 && (
                      <span className={`rounded-full px-1.5 py-0.5 text-xs tabular-nums
                        ${activeTab === tab.key ? "bg-primary text-primary-foreground" : "bg-muted"}`}>
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
                      {filteredItems.map(item => (
                        <tr key={item.id} className="border-b last:border-0 hover:bg-muted/20 transition-colors">
                          <td className="px-4 py-2.5 tabular-nums text-muted-foreground text-xs">
                            {item.sequence_order + 1}
                          </td>
                          <td className="px-4 py-2.5 font-mono text-xs">
                            {formatPhone(item.destino)}
                          </td>
                          <td className="px-4 py-2.5">
                            <ItemBadge status={item.status} />
                          </td>
                          <td className="px-4 py-2.5 text-xs text-muted-foreground whitespace-nowrap">
                            {item.processed_at ? new Date(item.processed_at).toLocaleTimeString("es-PE") : "—"}
                          </td>
                          <td className="px-4 py-2.5 text-xs text-destructive max-w-[140px] truncate" title={item.error_text ?? ""}>
                            {item.error_text ?? "—"}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>

            <p className="text-xs text-muted-foreground">Creado: {formatDate(detail.created_at)}</p>
          </div>
        ) : null}
      </SheetContent>
    </Sheet>
  )
}

// ─── page ────────────────────────────────────────────────────────────────────

export default function BroadcastsPage() {
  const [broadcasts, setBroadcasts] = useState<BroadcastInfo[]>([])
  const [companies, setCompanies] = useState<Empresa[]>([])
  const [loading, setLoading] = useState(true)
  const [filterRuc, setFilterRuc] = useState("")
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [detailsOpen, setDetailsOpen] = useState(false)

  const companyByRuc = useMemo(() => {
    const map = new Map<string, Empresa>()
    if (!Array.isArray(companies)) return map
    companies.forEach(c => map.set(c.ruc, c))
    return map
  }, [companies])

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const [bcData, compsData] = await Promise.all([
        getAdminBroadcasts(filterRuc || undefined),
        getEmpresas({ limit: 1000 }),
      ])
      setBroadcasts(bcData.broadcasts ?? [])
      setCompanies(compsData.empresas ?? [])
    } catch (error) {
      console.error("Failed to load data:", error)
    } finally {
      setLoading(false)
    }
  }, [filterRuc])

  useEffect(() => { loadData() }, [loadData])

  const openDetails = (bc: BroadcastInfo) => {
    setSelectedId(bc.reference_id)
    setDetailsOpen(true)
  }

  const selectedCompanyName = useMemo(() => {
    if (!selectedId) return ""
    const bc = broadcasts.find(b => b.reference_id === selectedId)
    if (!bc) return ""
    return companyByRuc.get(bc.ruc_empresa)?.nombre ?? bc.ruc_empresa
  }, [selectedId, broadcasts, companyByRuc])

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <h1 className="text-3xl font-bold tracking-tight">Broadcasts</h1>
        <p className="text-muted-foreground">Historial de difusiones masivas con estado en tiempo real.</p>
      </div>

      <Card>
        <CardHeader className="flex flex-col gap-4">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div className="flex flex-col gap-1">
              <CardTitle>Lista de broadcasts</CardTitle>
              <CardDescription>
                {loading ? "Cargando…" : `${broadcasts.length} difusión(es) encontrada(s)`}
              </CardDescription>
            </div>
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
              <select
                className="h-10 rounded-md border border-input bg-background px-3 py-2 text-sm cursor-pointer"
                value={filterRuc}
                onChange={e => setFilterRuc(e.target.value)}
              >
                <option value="">Todas las empresas</option>
                {companies.map(c => (
                  <option key={c.ruc} value={c.ruc}>{c.nombre || c.ruc}</option>
                ))}
              </select>
              {filterRuc && (
                <Button variant="outline" onClick={() => setFilterRuc("")}>Limpiar filtro</Button>
              )}
            </div>
          </div>
        </CardHeader>

        <CardContent className="flex flex-col gap-4">
          {!loading && broadcasts.length === 0 && (
            <DataEmptyState
              icon={MessageSquareMore}
              title="No hay broadcasts para mostrar"
              description="Prueba con otra empresa o revisa si ya se enviaron difusiones."
            />
          )}

          {(loading || broadcasts.length > 0) && (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Reference ID</TableHead>
                  <TableHead>Empresa</TableHead>
                  <TableHead>Progreso</TableHead>
                  <TableHead>Estado</TableHead>
                  <TableHead>Fecha</TableHead>
                  <TableHead className="text-right">Acciones</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {loading ? (
                  <TableLoadingRows columns={["w-28", "w-32", "w-24", "w-24", "w-28"]} />
                ) : (
                  broadcasts.map(bc => {
                    const company = companyByRuc.get(bc.ruc_empresa)
                    return (
                      <TableRow key={bc.reference_id}>
                        <TableCell className="font-mono text-sm">
                          {bc.reference_id.slice(0, 8)}…
                        </TableCell>
                        <TableCell>
                          <div className="flex flex-col gap-0.5">
                            <span className="font-medium">{company?.nombre ?? bc.ruc_empresa}</span>
                            <span className="text-xs text-muted-foreground">{company?.ruc ?? bc.ruc_empresa}</span>
                          </div>
                        </TableCell>
                        <TableCell>
                          <ProgressBar sent={bc.success ?? 0} total={bc.total} />
                        </TableCell>
                        <TableCell>
                          <StatusBadge status={bc.status} />
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground whitespace-nowrap">
                          {formatDate(bc.created_at)}
                        </TableCell>
                        <TableCell>
                          <div className="flex justify-end">
                            <Button variant="outline" size="sm" onClick={() => openDetails(bc)} className="cursor-pointer">
                              <Eye className="mr-2 h-4 w-4" />
                              Ver detalle
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    )
                  })
                )}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <BroadcastDetailSheet
        open={detailsOpen}
        onOpenChange={setDetailsOpen}
        referenceId={selectedId}
        empresaNombre={selectedCompanyName}
      />
    </div>
  )
}

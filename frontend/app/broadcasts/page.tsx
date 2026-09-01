"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import { AlertCircle, Radio } from "lucide-react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { DataTable } from "@/components/data-table"
import {
  getAdminBroadcasts,
  getEmpresas,
  type BroadcastInfo,
  type Empresa,
} from "@/lib/api"
import { getColumns } from "./columns"
import {
  DataTableToolbar,
  type BroadcastStatusFilter,
} from "./data-table-toolbar"
import { DataCardList } from "./data-card-list"
import { BroadcastDetailSheet } from "./broadcast-detail-sheet"

export default function BroadcastsPage() {
  const [broadcasts, setBroadcasts] = useState<BroadcastInfo[]>([])
  const [companies, setCompanies] = useState<Empresa[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Filtros
  const [globalFilter, setGlobalFilter] = useState("")
  const [statusFilter, setStatusFilter] = useState<BroadcastStatusFilter>("all")
  const [companyFilter, setCompanyFilter] = useState("")

  // Estado del modal de detalle
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [detailsOpen, setDetailsOpen] = useState(false)

  const companyByRuc = useMemo(() => {
    const map = new Map<string, Empresa>()
    if (!Array.isArray(companies)) return map
    companies.forEach((c) => map.set(c.ruc, c))
    return map
  }, [companies])

  const loadData = useCallback(async (opts?: { silent?: boolean }) => {
    if (!opts?.silent) setRefreshing(true)
    try {
      const [bcData, compsData] = await Promise.all([
        getAdminBroadcasts(),
        getEmpresas({ limit: 1000 }),
      ])
      setBroadcasts(bcData.broadcasts ?? [])
      setCompanies(compsData.empresas ?? [])
      setError(null)
    } catch (err) {
      console.error("Failed to load broadcasts data:", err)
      setError("No se pudieron cargar las difusiones. Revisa la conexión con el servidor.")
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }, [])

  useEffect(() => {
    loadData({ silent: true })
  }, [loadData])

  const handleOpenDetails = useCallback((bc: BroadcastInfo) => {
    setSelectedId(bc.reference_id)
    setDetailsOpen(true)
  }, [])

  // Conteo para los botones segmentados de estado
  const counts = useMemo<Record<BroadcastStatusFilter, number>>(() => {
    const scoped = companyFilter
      ? broadcasts.filter((b) => b.ruc_empresa === companyFilter)
      : broadcasts

    return {
      all: scoped.length,
      running: scoped.filter((b) => b.status === "running").length,
      completed: scoped.filter((b) => b.status === "completed").length,
      failed: scoped.filter((b) => b.status === "failed" || b.status === "cancelled").length,
      pending: scoped.filter((b) => b.status === "pending").length,
    }
  }, [broadcasts, companyFilter])

  // Filtrado de negocio en el contenedor
  const filteredBroadcasts = useMemo(() => {
    return broadcasts.filter((bc) => {
      // 1. Filtro por empresa
      if (companyFilter && bc.ruc_empresa !== companyFilter) {
        return false
      }

      // 2. Filtro por estado
      if (statusFilter !== "all") {
        if (statusFilter === "failed") {
          if (bc.status !== "failed" && bc.status !== "cancelled") return false
        } else if (bc.status !== statusFilter) {
          return false
        }
      }

      // 3. Búsqueda por texto global
      if (!globalFilter.trim()) return true
      const query = globalFilter.toLowerCase().trim()
      const company = companyByRuc.get(bc.ruc_empresa)
      const companyName = (company?.nombre || "").toLowerCase()
      const ruc = (bc.ruc_empresa || "").toLowerCase()
      const refId = (bc.reference_id || "").toLowerCase()
      const status = (bc.status || "").toLowerCase()

      return (
        refId.includes(query) ||
        companyName.includes(query) ||
        ruc.includes(query) ||
        status.includes(query)
      )
    })
  }, [broadcasts, companyFilter, statusFilter, globalFilter, companyByRuc])

  // Columnas para TanStack Table
  const columns = useMemo(
    () =>
      getColumns({
        onOpenDetails: handleOpenDetails,
        companyByRuc,
      }),
    [handleOpenDetails, companyByRuc],
  )

  const selectedCompanyName = useMemo(() => {
    if (!selectedId) return ""
    const bc = broadcasts.find((b) => b.reference_id === selectedId)
    if (!bc) return ""
    return companyByRuc.get(bc.ruc_empresa)?.nombre ?? bc.ruc_empresa
  }, [selectedId, broadcasts, companyByRuc])

  return (
    <div className="space-y-4 pb-12">
      {/* Encabezado */}
      <div className="space-y-1">
        <h1 className="flex items-center gap-2 text-xl font-semibold tracking-tight text-foreground sm:text-2xl">
          <Radio className="h-5 w-5 text-muted-foreground" />
          <span>Difusiones Masivas</span>
        </h1>
        <p className="text-sm text-muted-foreground">
          Supervisa el progreso, entregas y errores de las campañas de difusión masiva.
        </p>
      </div>

      {/* Alerta de error */}
      {error && (
        <Alert variant="destructive">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div className="flex items-start gap-3">
              <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
              <div>
                <AlertTitle>Error al cargar difusiones</AlertTitle>
                <AlertDescription>{error}</AlertDescription>
              </div>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={() => loadData()}
              disabled={refreshing}
              className="shrink-0 self-start"
            >
              Reintentar
            </Button>
          </div>
        </Alert>
      )}

      {/* Barra de herramientas: control segmentado + buscador + selector de empresa + refresco */}
      <DataTableToolbar
        globalFilter={globalFilter}
        onGlobalFilterChange={setGlobalFilter}
        statusFilter={statusFilter}
        onStatusFilterChange={setStatusFilter}
        companyFilter={companyFilter}
        onCompanyFilterChange={setCompanyFilter}
        companies={companies}
        counts={counts}
        loading={refreshing}
        onRefresh={() => loadData()}
      />

      {/* Escritorio: tabla TanStack Table */}
      <div className="hidden md:flex md:flex-col">
        <DataTable
          columns={columns}
          data={filteredBroadcasts}
          loading={loading}
          itemLabel="difusiones"
          initialPageSize={50}
          pageSizeOptions={[50, 100, 150, 200]}
          emptyMessage="No se encontraron difusiones que coincidan con los filtros."
        />
      </div>

      {/* Móvil: lista de tarjetas accionables */}
      <div className="md:hidden">
        <DataCardList
          data={filteredBroadcasts}
          loading={loading}
          companyByRuc={companyByRuc}
          onOpenDetails={handleOpenDetails}
        />
      </div>

      {/* Sheet de Detalle del Broadcast */}
      <BroadcastDetailSheet
        open={detailsOpen}
        onOpenChange={setDetailsOpen}
        referenceId={selectedId}
        empresaNombre={selectedCompanyName}
      />
    </div>
  )
}

"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import { useRouter } from "next/navigation"
import { DataTable } from "@/components/data-table"
import { getColumns } from "./columns"
import { DataTableToolbar, type EmpresaStatusFilter } from "./data-table-toolbar"
import { DataCardList } from "./data-card-list"
import { EmpresaFormModal } from "@/components/companies/empresa-form-modal"
import { EmpresaDetailModal } from "@/components/companies/empresa-detail-modal"
import {
  getEmpresas,
  createEmpresa,
  updateEmpresa,
  deleteEmpresa,
  restoreEmpresa,
  type Empresa,
  type EmpresaCreateRequest,
} from "@/lib/api"
import { Building2, AlertCircle, CheckCircle2 } from "lucide-react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { PaginationState } from "@tanstack/react-table"

export default function CompaniesPage() {
  const router = useRouter()
  const [empresas, setEmpresas] = useState<Empresa[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [errorMessage, setErrorMessage] = useState<string | null>(null)
  const [actionMessage, setActionMessage] = useState<string | null>(null)

  // Filtros y paginación de servidor
  const [search, setSearch] = useState("")
  const [debouncedSearch, setDebouncedSearch] = useState("")
  const [statusFilter, setStatusFilter] = useState<EmpresaStatusFilter>("all")
  const [pagination, setPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: 10,
  })

  // Modales y estados de fila
  const [formOpen, setFormOpen] = useState(false)
  const [editTarget, setEditTarget] = useState<Empresa | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)
  const [detailTarget, setDetailTarget] = useState<Empresa | null>(null)
  const [deletingId, setDeletingId] = useState<number | null>(null)
  const [restoringId, setRestoringId] = useState<number | null>(null)

  // Debounce del buscador de texto
  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search)
      setPagination((prev) => ({ ...prev, pageIndex: 0 }))
    }, 300)
    return () => clearTimeout(timer)
  }, [search])

  // Carga de empresas desde el servidor
  const load = useCallback(
    async (opts?: { silent?: boolean }) => {
      if (!opts?.silent) {
        setRefreshing(true)
      }
      try {
        const resp = await getEmpresas({
          page: pagination.pageIndex + 1,
          limit: pagination.pageSize,
          busqueda: debouncedSearch.trim() || undefined,
          estado: statusFilter !== "all" ? statusFilter : undefined,
        })
        setEmpresas(resp.empresas ?? [])
        setTotal(resp.total ?? 0)
        setErrorMessage(null)
      } catch (err: unknown) {
        console.error("Error cargando empresas:", err)
        setErrorMessage(
          err instanceof Error ? err.message : "Error al cargar el listado de empresas",
        )
      } finally {
        setLoading(false)
        setRefreshing(false)
      }
    },
    [pagination.pageIndex, pagination.pageSize, debouncedSearch, statusFilter],
  )

  useEffect(() => {
    load({ silent: !loading })
  }, [load, loading])

  // Conteo de empresas en la vista actual
  const counts = useMemo(() => {
    let active = 0
    let inactive = 0
    empresas.forEach((e) => {
      if (e.activo) active++
      else inactive++
    })
    return {
      all: total,
      active: statusFilter === "active" ? total : active,
      inactive: statusFilter === "inactive" ? total : inactive,
    }
  }, [empresas, total, statusFilter])

  // Handlers de modales y acciones
  const handleOpenTelefonos = useCallback(
    (empresa: Empresa) => {
      router.push(`/empresas/${empresa.id}/telefonos`)
    },
    [router],
  )

  const handleOpenDetail = useCallback((empresa: Empresa) => {
    setDetailTarget(empresa)
    setDetailOpen(true)
  }, [])

  const handleOpenEdit = useCallback((empresa: Empresa) => {
    setEditTarget(empresa)
    setFormOpen(true)
  }, [])

  const handleOpenNew = useCallback(() => {
    setEditTarget(null)
    setFormOpen(true)
  }, [])

  const handleSave = useCallback(
    async (data: EmpresaCreateRequest) => {
      try {
        if (editTarget) {
          await updateEmpresa(editTarget.id, data)
          setActionMessage(`Empresa "${data.nombre}" actualizada con éxito.`)
        } else {
          await createEmpresa(data)
          setActionMessage(`Empresa "${data.nombre}" creada con éxito.`)
        }
        await load({ silent: true })
      } catch (err: unknown) {
        throw err
      }
    },
    [editTarget, load],
  )

  const handleDelete = useCallback(
    async (empresa: Empresa) => {
      if (!confirm(`¿Estás seguro de inhabilitar la empresa "${empresa.nombre}"?`)) {
        return
      }
      setDeletingId(empresa.id)
      setActionMessage(null)
      try {
        await deleteEmpresa(empresa.id)
        setActionMessage(`Empresa "${empresa.nombre}" inhabilitada con éxito.`)
        await load({ silent: true })
      } catch (err: unknown) {
        setErrorMessage(
          err instanceof Error ? err.message : "Error al inhabilitar la empresa",
        )
      } finally {
        setDeletingId(null)
      }
    },
    [load],
  )

  const handleRestore = useCallback(
    async (empresa: Empresa) => {
      if (!confirm(`¿Deseas restaurar la empresa "${empresa.nombre}"?`)) {
        return
      }
      setRestoringId(empresa.id)
      setActionMessage(null)
      try {
        await restoreEmpresa(empresa.id)
        setActionMessage(`Empresa "${empresa.nombre}" restaurada con éxito.`)
        await load({ silent: true })
      } catch (err: unknown) {
        setErrorMessage(
          err instanceof Error ? err.message : "Error al restaurar la empresa",
        )
      } finally {
        setRestoringId(null)
      }
    },
    [load],
  )

  // Columnas para TanStack Table
  const columns = useMemo(
    () =>
      getColumns({
        onOpenTelefonos: handleOpenTelefonos,
        onOpenDetail: handleOpenDetail,
        onOpenEdit: handleOpenEdit,
        onDelete: handleDelete,
        onRestore: handleRestore,
        deletingId,
        restoringId,
      }),
    [
      handleOpenTelefonos,
      handleOpenDetail,
      handleOpenEdit,
      handleDelete,
      handleRestore,
      deletingId,
      restoringId,
    ],
  )

  const totalPages = Math.ceil(total / pagination.pageSize)

  return (
    <div className="space-y-4 pb-12">
      {/* Encabezado */}
      <div className="space-y-1">
        <h1 className="flex items-center gap-2 text-xl font-semibold tracking-tight text-foreground sm:text-2xl">
          <Building2 className="h-5 w-5 text-muted-foreground" />
          <span>Empresas</span>
        </h1>
        <p className="text-sm text-muted-foreground">
          Gestiona las empresas registradas en el sistema, canales WhatsApp asociados y configuración.
        </p>
      </div>

      {/* Alertas de resultado / errores */}
      {actionMessage && (
        <Alert>
          <div className="flex items-start gap-3">
            <CheckCircle2 className="mt-0.5 h-5 w-5 text-emerald-500 shrink-0" />
            <div>
              <AlertTitle>Operación completada</AlertTitle>
              <AlertDescription>{actionMessage}</AlertDescription>
            </div>
          </div>
        </Alert>
      )}

      {errorMessage && (
        <Alert variant="destructive">
          <div className="flex items-start gap-3">
            <AlertCircle className="mt-0.5 h-5 w-5 shrink-0" />
            <div>
              <AlertTitle>Error</AlertTitle>
              <AlertDescription>{errorMessage}</AlertDescription>
            </div>
          </div>
        </Alert>
      )}

      {/* Barra de herramientas */}
      <DataTableToolbar
        search={search}
        onSearchChange={setSearch}
        statusFilter={statusFilter}
        onStatusFilterChange={(st) => {
          setStatusFilter(st)
          setPagination((prev) => ({ ...prev, pageIndex: 0 }))
        }}
        counts={counts}
        onOpenNew={handleOpenNew}
        onRefresh={() => load()}
        refreshing={refreshing}
      />

      {/* Tabla Desktop (md+) con TanStack Table */}
      <div className="hidden md:block">
        <DataTable
          columns={columns}
          data={empresas}
          loading={loading}
          itemLabel="empresas"
          pageSizeOptions={[10, 20, 30, 50]}
          manualPagination={true}
          pageCount={totalPages}
          pagination={pagination}
          onPaginationChange={setPagination}
          totalRows={total}
          emptyMessage="No se encontraron empresas con los filtros aplicados."
        />
      </div>

      {/* Vista Móvil (< md) */}
      <div className="md:hidden">
        <DataCardList
          data={empresas}
          loading={loading}
          onOpenTelefonos={handleOpenTelefonos}
          onOpenDetail={handleOpenDetail}
          onOpenEdit={handleOpenEdit}
          onDelete={handleDelete}
          onRestore={handleRestore}
          deletingId={deletingId}
          restoringId={restoringId}
          page={pagination.pageIndex + 1}
          totalPages={totalPages}
          totalRows={total}
          onPageChange={(newPage) =>
            setPagination((prev) => ({ ...prev, pageIndex: newPage - 1 }))
          }
        />
      </div>

      {/* Modales */}
      <EmpresaFormModal
        open={formOpen}
        onClose={() => setFormOpen(false)}
        onSave={handleSave}
        empresa={editTarget}
      />

      <EmpresaDetailModal
        open={detailOpen}
        onClose={() => setDetailOpen(false)}
        empresa={detailTarget}
      />
    </div>
  )
}

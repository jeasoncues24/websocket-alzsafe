"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import {
  AlertCircle,
  CheckCircle2,
  FileText,
  MessageSquare,
  RefreshCw,
} from "lucide-react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import { DataTable } from "@/components/data-table"
import {
  getAdminMessages,
  getEmpresas,
  retryMessageAdmin,
  type AdminMessage,
  type Empresa,
} from "@/lib/api"
import { getColumns, MessageStatusBadge } from "./columns"
import {
  DataTableToolbar,
  type MessageStatusFilter,
} from "./data-table-toolbar"
import { DataCardList } from "./data-card-list"

function formatDate(value?: string) {
  if (!value) return "-"
  try {
    return new Date(value).toLocaleString("es-PE", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
    })
  } catch {
    return value
  }
}

function formatBytes(bytes: number) {
  if (!bytes) return "0 B"
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function attachmentLabel(count: number) {
  return count === 1 ? "1 adjunto" : `${count} adjuntos`
}

export default function MessagesPage() {
  const [messages, setMessages] = useState<AdminMessage[]>([])
  const [companies, setCompanies] = useState<Empresa[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Filtros
  const [globalFilter, setGlobalFilter] = useState("")
  const [statusFilter, setStatusFilter] = useState<MessageStatusFilter>("all")
  const [companyFilter, setCompanyFilter] = useState("")

  // Estado de acciones y detalles
  const [retryingId, setRetryingId] = useState<number | null>(null)
  const [resultAlert, setResultAlert] = useState<{ ok: boolean; message: string } | null>(null)
  const [selectedMessage, setSelectedMessage] = useState<AdminMessage | null>(null)
  const [detailsOpen, setDetailsOpen] = useState(false)

  const companyByRuc = useMemo(() => {
    const map = new Map<string, Empresa>()
    if (!Array.isArray(companies)) return map
    companies.forEach((company) => map.set(company.ruc, company))
    return map
  }, [companies])

  const loadData = useCallback(async (opts?: { silent?: boolean }) => {
    if (!opts?.silent) setRefreshing(true)
    try {
      const [msgsData, compsData] = await Promise.all([
        getAdminMessages({ limit: 500 }),
        getEmpresas({ limit: 1000 }),
      ])
      setMessages(msgsData.messages ?? [])
      setCompanies(compsData.empresas ?? [])
      setError(null)
    } catch (err) {
      console.error("Failed to load messages data:", err)
      setError("No se pudieron cargar los mensajes. Revisa la conexión con el servidor.")
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }, [])

  useEffect(() => {
    loadData({ silent: true })
  }, [loadData])

  const handleRetry = useCallback(
    async (msg: AdminMessage) => {
      if (!msg.reference_id) return
      setRetryingId(msg.id)
      setResultAlert(null)
      try {
        const result = await retryMessageAdmin(msg.reference_id)
        setResultAlert({
          ok: result.ok,
          message: result.error || "Mensaje reenviado exitosamente",
        })
        if (result.ok) {
          await loadData({ silent: true })
        }
      } catch (err) {
        const errMsg = err instanceof Error ? err.message : "Error desconocido al reintentar"
        setResultAlert({ ok: false, message: errMsg })
      } finally {
        setRetryingId(null)
      }
    },
    [loadData],
  )

  const handleOpenDetails = useCallback((msg: AdminMessage) => {
    setSelectedMessage(msg)
    setDetailsOpen(true)
  }, [])

  // Conteo por estado para los botones segmentados
  const counts = useMemo<Record<MessageStatusFilter, number>>(() => {
    const scoped = companyFilter
      ? messages.filter((m) => m.account_id === companyFilter)
      : messages

    return {
      all: scoped.length,
      pending: scoped.filter((m) => m.status === "pending").length,
      sent: scoped.filter((m) => m.status === "sent").length,
      delivered: scoped.filter((m) => m.status === "delivered").length,
      failed: scoped.filter((m) => m.status === "failed").length,
    }
  }, [messages, companyFilter])

  // Filtrado de negocio en el contenedor
  const filteredMessages = useMemo(() => {
    return messages.filter((msg) => {
      // 1. Filtro por empresa seleccionada
      if (companyFilter && msg.account_id !== companyFilter) {
        return false
      }

      // 2. Filtro por estado
      if (statusFilter !== "all" && msg.status !== statusFilter) {
        return false
      }

      // 3. Búsqueda por texto global
      if (!globalFilter.trim()) return true
      const query = globalFilter.toLowerCase().trim()
      const company = companyByRuc.get(msg.account_id)
      const companyName = (company?.nombre || "").toLowerCase()
      const accountId = (msg.account_id || "").toLowerCase()
      const to = (msg.to || "").toLowerCase()
      const content = (msg.content || "").toLowerCase()
      const errorReason = (msg.error_reason || "").toLowerCase()
      const idStr = String(msg.id)

      return (
        idStr.includes(query) ||
        companyName.includes(query) ||
        accountId.includes(query) ||
        to.includes(query) ||
        content.includes(query) ||
        errorReason.includes(query)
      )
    })
  }, [messages, companyFilter, statusFilter, globalFilter, companyByRuc])

  // Definición de columnas memoizadas
  const columns = useMemo(
    () =>
      getColumns({
        onOpenDetails: handleOpenDetails,
        onRetry: handleRetry,
        retryingId,
        companyByRuc,
      }),
    [handleOpenDetails, handleRetry, retryingId, companyByRuc],
  )

  const selectedCompany = selectedMessage
    ? companyByRuc.get(selectedMessage.account_id)
    : undefined

  return (
    <div className="space-y-4 pb-12">
      {/* Encabezado */}
      <div className="space-y-1">
        <h1 className="flex items-center gap-2 text-xl font-semibold tracking-tight text-foreground sm:text-2xl">
          <MessageSquare className="h-5 w-5 text-muted-foreground" />
          <span>Historial de Mensajes</span>
        </h1>
        <p className="text-sm text-muted-foreground">
          Consulta el estado de entrega, adjuntos y registros de reintentos de mensajería.
        </p>
      </div>

      {/* Alerta de resultado tras reintento */}
      {resultAlert && (
        <Alert variant={resultAlert.ok ? "default" : "destructive"}>
          <div className="flex items-start gap-3">
            {resultAlert.ok ? (
              <CheckCircle2 className="mt-0.5 h-5 w-5 text-emerald-500" />
            ) : (
              <AlertCircle className="mt-0.5 h-5 w-5" />
            )}
            <div>
              <AlertTitle>{resultAlert.ok ? "Operación exitosa" : "Error en el reintento"}</AlertTitle>
              <AlertDescription>{resultAlert.message}</AlertDescription>
            </div>
          </div>
        </Alert>
      )}

      {/* Alerta de error de conexión */}
      {error && (
        <Alert variant="destructive">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div className="flex items-start gap-3">
              <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
              <div>
                <AlertTitle>Error al cargar mensajes</AlertTitle>
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
          data={filteredMessages}
          loading={loading}
          itemLabel="mensajes"
          initialPageSize={50}
          pageSizeOptions={[50, 100, 150, 200]}
          emptyMessage="No se encontraron mensajes que coincidan con los filtros."
        />
      </div>

      {/* Móvil: lista de tarjetas accionables */}
      <div className="md:hidden">
        <DataCardList
          data={filteredMessages}
          loading={loading}
          companyByRuc={companyByRuc}
          onOpenDetails={handleOpenDetails}
          onRetry={handleRetry}
          retryingId={retryingId}
        />
      </div>

      {/* Sheet de Detalle del Mensaje */}
      <Sheet open={detailsOpen} onOpenChange={setDetailsOpen}>
        <SheetContent side="right" className="w-full overflow-y-auto sm:max-w-2xl">
          <SheetHeader className="text-left">
            <SheetTitle>Detalle del mensaje</SheetTitle>
            <SheetDescription>
              {selectedMessage?.reference_id
                ? `Referencia: ${selectedMessage.reference_id}`
                : selectedMessage
                  ? `Mensaje #${selectedMessage.id}`
                  : "Detalles del mensaje seleccionado"}
            </SheetDescription>
          </SheetHeader>

          {selectedMessage && (
            <div className="mt-6 space-y-6">
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="rounded-lg border p-4">
                  <p className="text-xs uppercase tracking-wide text-muted-foreground">
                    Empresa
                  </p>
                  <p className="mt-1 font-medium">
                    {selectedCompany?.nombre || selectedMessage.account_id}
                  </p>
                  <p className="font-mono text-sm text-muted-foreground">
                    {selectedCompany?.ruc || selectedMessage.account_id}
                  </p>
                </div>
                <div className="rounded-lg border p-4">
                  <p className="text-xs uppercase tracking-wide text-muted-foreground">
                    Estado
                  </p>
                  <div className="mt-2">
                    <MessageStatusBadge
                      status={selectedMessage.status}
                      retryCount={selectedMessage.retry_count}
                    />
                  </div>
                </div>
              </div>

              <div className="rounded-lg border p-4">
                <p className="text-xs uppercase tracking-wide text-muted-foreground">
                  Destino
                </p>
                <p className="mt-1 font-mono text-sm">{selectedMessage.to}</p>
              </div>

              <div className="rounded-lg border p-4">
                <p className="text-xs uppercase tracking-wide text-muted-foreground">
                  Contenido
                </p>
                <p className="mt-2 whitespace-pre-wrap text-sm leading-6">
                  {selectedMessage.content || "Sin contenido"}
                </p>
              </div>

              <div className="rounded-lg border p-4">
                <div className="flex items-center justify-between gap-3">
                  <p className="text-xs uppercase tracking-wide text-muted-foreground">
                    Adjuntos
                  </p>
                  <Badge variant="outline">
                    {attachmentLabel(selectedMessage.adjuntos?.length ?? 0)}
                  </Badge>
                </div>
                <div className="mt-3 space-y-3">
                  {selectedMessage.adjuntos?.length ? (
                    selectedMessage.adjuntos.map((att) => (
                      <div
                        key={`${att.sha256_hash}-${att.nombre}`}
                        className="rounded-md border bg-muted/30 p-3"
                      >
                        <div className="flex items-center gap-2">
                          <FileText className="h-4 w-4 text-muted-foreground" />
                          <p className="font-medium">{att.nombre}</p>
                        </div>
                        <p className="mt-1 text-sm text-muted-foreground">
                          {formatBytes(att.tamano_bytes)}
                        </p>
                        <p className="mt-1 break-all font-mono text-xs text-muted-foreground">
                          {att.sha256_hash}
                        </p>
                      </div>
                    ))
                  ) : (
                    <p className="text-sm text-muted-foreground">
                      Este mensaje no tiene adjuntos.
                    </p>
                  )}
                </div>
              </div>

              {selectedMessage.error_reason && (
                <Alert variant="destructive">
                  <AlertCircle className="h-4 w-4" />
                  <AlertTitle>Motivo del error</AlertTitle>
                  <AlertDescription>{selectedMessage.error_reason}</AlertDescription>
                </Alert>
              )}

              <div className="grid gap-4 sm:grid-cols-2">
                <div className="rounded-lg border p-4">
                  <p className="text-xs uppercase tracking-wide text-muted-foreground">
                    Fecha de creación
                  </p>
                  <p className="mt-1 text-sm">{formatDate(selectedMessage.created_at)}</p>
                </div>
                <div className="rounded-lg border p-4">
                  <p className="text-xs uppercase tracking-wide text-muted-foreground">
                    Reintentos
                  </p>
                  <p className="mt-1 text-sm">{selectedMessage.retry_count ?? 0}</p>
                </div>
              </div>

              {selectedMessage.status === "failed" && selectedMessage.reference_id && (
                <Button
                  className="w-full"
                  onClick={() => handleRetry(selectedMessage)}
                  disabled={retryingId === selectedMessage.id}
                >
                  {retryingId === selectedMessage.id && (
                    <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                  )}
                  Reintentar mensaje
                </Button>
              )}
            </div>
          )}
        </SheetContent>
      </Sheet>
    </div>
  )
}

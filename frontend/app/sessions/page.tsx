"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import { SessionQRDialog } from "@/components/session/session-qr-dialog"
import { SessionDisconnectDialog } from "@/components/session/session-disconnect-dialog"
import { SessionEventsSheet } from "@/components/session/session-events-sheet"
import { DataTable } from "./data-table"
import { DataCardList } from "./data-card-list"
import { DataTableToolbar, type StatusTabFilter } from "./data-table-toolbar"
import { getColumns } from "./columns"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import {
  getAdminSessions,
  postAdminSession,
  reconnectAdminSession,
  type SessionInfo,
  type SessionSummary,
} from "@/lib/api"
import { ShieldCheck, AlertTriangle } from "lucide-react"

export default function SessionsPage() {
  const [sessions, setSessions] = useState<SessionInfo[]>([])
  const [summary, setSummary] = useState<SessionSummary | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null)

  // Filtros
  const [globalFilter, setGlobalFilter] = useState("")
  const [statusFilter, setStatusFilter] = useState<StatusTabFilter>("all")

  // Modales y Sheets
  const [selectedQR, setSelectedQR] = useState<SessionInfo | null>(null)
  const [qrOpen, setQrOpen] = useState(false)

  const [selectedDisconnect, setSelectedDisconnect] = useState<SessionInfo | null>(null)
  const [disconnectOpen, setDisconnectOpen] = useState(false)
  const [disconnectLoading, setDisconnectLoading] = useState(false)

  const [reconnectingId, setReconnectingId] = useState<number | null>(null)
  const [eventsSession, setEventsSession] = useState<SessionInfo | null>(null)

  const loadSessions = useCallback(async () => {
    setLoading(true)
    try {
      const data = await getAdminSessions()
      setSessions(data.sessions || [])
      setSummary(data.summary ?? null)
      setError(null)
      setLastUpdated(new Date())
    } catch (err) {
      // Se conserva el último dato bueno: una caída del backend no debe
      // parecer una flota vacía.
      console.error("Failed to load sessions:", err)
      setError("No se pudo cargar el estado de las sesiones.")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadSessions()
  }, [loadSessions])

  const handleOpenQR = (session: SessionInfo) => {
    setSelectedQR(session)
    setQrOpen(true)
  }

  const handleOpenDisconnect = (session: SessionInfo) => {
    setSelectedDisconnect(session)
    setDisconnectOpen(true)
  }

  const handleConfirmDisconnect = async () => {
    if (!selectedDisconnect) return
    setDisconnectLoading(true)
    try {
      await postAdminSession("disconnect", selectedDisconnect.account_id)
      await loadSessions()
    } catch (error) {
      console.error("Failed to disconnect session:", error)
    } finally {
      setDisconnectLoading(false)
      setDisconnectOpen(false)
      setSelectedDisconnect(null)
    }
  }

  const handleReconnect = async (telefonoId: number) => {
    setReconnectingId(telefonoId)
    try {
      await reconnectAdminSession(telefonoId)
      await loadSessions()
    } catch (error) {
      console.error("Failed to reconnect session:", error)
    } finally {
      setReconnectingId(null)
    }
  }

  const handleOpenEvents = (session: SessionInfo) => {
    setEventsSession(session)
  }

  // Filtrado compuesto
  const filteredSessions = useMemo(() => {
    return sessions.filter((s) => {
      // 1. Filtro por pestaña de estado
      if (statusFilter === "active" && s.status !== "active") return false
      if (statusFilter === "disconnected" && s.status !== "disconnected") return false
      if (statusFilter === "qr_pending" && s.status !== "qr_pending") return false
      if (statusFilter === "mismatch" && !s.mismatch) return false

      // 2. Filtro global por texto
      if (!globalFilter.trim()) return true
      const query = globalFilter.toLowerCase().trim()
      const empresa = (s.empresa_nombre ?? "").toLowerCase()
      const account = (s.account_id ?? "").toLowerCase()
      const status = (s.status ?? "").toLowerCase()
      const empresaId = String(s.empresa_id || "")

      return (
        empresa.includes(query) ||
        account.includes(query) ||
        status.includes(query) ||
        empresaId.includes(query)
      )
    })
  }, [sessions, statusFilter, globalFilter])

  // Columnas para TanStack Table
  const columns = useMemo(
    () =>
      getColumns({
        onOpenQR: handleOpenQR,
        onOpenDisconnect: handleOpenDisconnect,
        onReconnect: handleReconnect,
        onOpenEvents: handleOpenEvents,
        reconnectingId,
      }),
    [reconnectingId]
  )

  return (
    <div className="space-y-4 pb-12">
      {/* Encabezado */}
      <div className="space-y-1">
        <h1 className="flex items-center gap-2 text-xl font-semibold tracking-tight text-foreground sm:text-2xl">
          <ShieldCheck className="h-5 w-5 text-muted-foreground" />
          <span>Sesiones WhatsApp</span>
        </h1>
        <p className="text-sm text-muted-foreground">
          Supervisa el estado y reconecta los canales de mensajería.
        </p>
      </div>

      {/* Barra de herramientas: filtro por estado + buscador + refresco */}
      <DataTableToolbar
        globalFilter={globalFilter}
        onGlobalFilterChange={setGlobalFilter}
        statusFilter={statusFilter}
        onStatusFilterChange={setStatusFilter}
        summary={summary}
        loading={loading}
        onRefresh={loadSessions}
        lastUpdated={lastUpdated}
      />

      {/* Estado de error: nunca dejar que un fallo parezca "flota vacía" */}
      {error && (
        <Alert variant="destructive">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div className="flex items-start gap-3">
              <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
              <div>
                <AlertTitle>No se pudo cargar el estado de las sesiones</AlertTitle>
                <AlertDescription>
                  {sessions.length > 0
                    ? "Se muestran los últimos datos disponibles. Reintenta para actualizar."
                    : "Revisa la conexión con el servidor y reintenta."}
                </AlertDescription>
              </div>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={loadSessions}
              disabled={loading}
              className="shrink-0 self-start"
            >
              Reintentar
            </Button>
          </div>
        </Alert>
      )}

      {/* Escritorio: tabla densa. Móvil: lista de tarjetas accionables. */}
      <div className="hidden md:flex md:flex-col">
        <DataTable columns={columns} data={filteredSessions} loading={loading} />
      </div>
      <div className="md:hidden">
        <DataCardList
          data={filteredSessions}
          loading={loading}
          onOpenQR={handleOpenQR}
          onOpenDisconnect={handleOpenDisconnect}
          onReconnect={handleReconnect}
          onOpenEvents={handleOpenEvents}
          reconnectingId={reconnectingId}
        />
      </div>

      {/* Diálogo de Código QR */}
      <SessionQRDialog
        session={selectedQR}
        open={qrOpen}
        onOpenChange={setQrOpen}
      />

      {/* Diálogo de Confirmación de Desconexión */}
      <SessionDisconnectDialog
        open={disconnectOpen}
        onOpenChange={setDisconnectOpen}
        onConfirm={handleConfirmDisconnect}
        targetName={selectedDisconnect?.empresa_nombre || selectedDisconnect?.account_id || ""}
        loading={disconnectLoading}
      />

      {/* Sheet de Historial de Eventos */}
      <SessionEventsSheet
        session={eventsSession}
        open={!!eventsSession}
        onOpenChange={(open) => {
          if (!open) setEventsSession(null)
        }}
      />
    </div>
  )
}

"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import { SessionQRDialog } from "@/components/session/session-qr-dialog"
import { SessionDisconnectDialog } from "@/components/session/session-disconnect-dialog"
import { SessionEventsSheet } from "@/components/session/session-events-sheet"
import { DataTable } from "@/components/data-table"
import { DataCardList } from "./data-card-list"
import { DataTableToolbar, type StatusTabFilter } from "./data-table-toolbar"
import { getColumns } from "./columns"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import {
  buildAdminSessionsStreamUrl,
  getAdminSessions,
  postAdminSession,
  reconnectAdminSession,
  type SessionInfo,
  type SessionSummary,
} from "@/lib/api"
import { ShieldCheck, AlertTriangle } from "lucide-react"

// Cache local del último snapshot: al recargar la página se pinta al instante con
// el último dato conocido (aunque sea de hace un rato) y el stream lo refresca en
// segundo plano. Así una recarga nunca muestra el esqueleto.
const SNAPSHOT_CACHE_KEY = "admin_sessions_snapshot"

type CachedSnapshot = {
  sessions: SessionInfo[]
  summary: SessionSummary | null
  ts: number
}

function readCachedSnapshot(): CachedSnapshot | null {
  if (typeof window === "undefined") return null
  try {
    const raw = window.localStorage.getItem(SNAPSHOT_CACHE_KEY)
    if (!raw) return null
    const parsed = JSON.parse(raw) as CachedSnapshot
    if (!Array.isArray(parsed.sessions)) return null
    return parsed
  } catch {
    return null
  }
}

function writeCachedSnapshot(sessions: SessionInfo[], summary: SessionSummary | null) {
  if (typeof window === "undefined") return
  try {
    window.localStorage.setItem(
      SNAPSHOT_CACHE_KEY,
      JSON.stringify({ sessions, summary, ts: Date.now() } satisfies CachedSnapshot),
    )
  } catch {
    // Sin storage (modo privado, cuota): la vista sigue funcionando sin cache.
  }
}

export default function SessionsPage() {
  const [sessions, setSessions] = useState<SessionInfo[]>([])
  const [summary, setSummary] = useState<SessionSummary | null>(null)
  // `loading` solo cubre el primer render sin ningún dato. Tras la primera carga
  // —o tras hidratar la cache local en el montaje— las actualizaciones (stream,
  // polling de respaldo o refresco manual) reemplazan los datos en silencio.
  const [loading, setLoading] = useState(true)
  // Refresco manual en curso: solo anima el icono del botón, no la tabla.
  const [refreshing, setRefreshing] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null)
  // Stream SSE conectado: cuando es false, un polling de respaldo mantiene el dato fresco.
  const [live, setLive] = useState(false)

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
  const [eventsOpen, setEventsOpen] = useState(false)

  const loadSessions = useCallback(async (opts?: { silent?: boolean }) => {
    if (!opts?.silent) setRefreshing(true)
    try {
      const data = await getAdminSessions()
      setSessions(data.sessions || [])
      setSummary(data.summary ?? null)
      writeCachedSnapshot(data.sessions || [], data.summary ?? null)
      setError(null)
      setLastUpdated(new Date())
    } catch (err) {
      // Se conserva el último dato bueno: una caída del backend no debe
      // parecer una flota vacía.
      console.error("Failed to load sessions:", err)
      setError("No se pudo cargar el estado de las sesiones.")
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }, [])

  // Aplica un snapshot (venga del stream SSE o del polling de respaldo) sin
  // vaciar la tabla ante un fallo posterior.
  const applySnapshot = useCallback(
    (data: { sessions?: SessionInfo[]; summary?: SessionSummary | null }) => {
      setSessions(data.sessions || [])
      setSummary(data.summary ?? null)
      writeCachedSnapshot(data.sessions || [], data.summary ?? null)
      setError(null)
      setLastUpdated(new Date())
      setLoading(false)
    },
    [],
  )

  // Hidratación desde cache local en el montaje: si hay un snapshot previo se
  // pinta al instante (sin esqueleto) y el stream lo refresca en segundo plano.
  // Se hace en efecto —no en el estado inicial— para no romper la hidratación SSR.
  useEffect(() => {
    const c = readCachedSnapshot()
    if (!c) return
    setSessions(c.sessions)
    setSummary(c.summary)
    setLastUpdated(new Date(c.ts))
    setLoading(false)
  }, [])

  // Carga inicial por REST: primer dato fresco y respaldo si el navegador no
  // puede abrir el stream. Silenciosa: el esqueleto del primer render lo cubre
  // `loading`; aquí no se anima el botón de refresco.
  useEffect(() => {
    loadSessions({ silent: true })
  }, [loadSessions])

  // Stream en tiempo real. El backend emite un evento `snapshot` con el mismo
  // payload que el GET REST ante cada transición de sesión (más un latido).
  useEffect(() => {
    if (typeof window === "undefined" || typeof EventSource === "undefined") return

    let es: EventSource | null = null
    let cleanedUp = false
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null

    const connect = () => {
      es = new EventSource(buildAdminSessionsStreamUrl())

      es.addEventListener("open", () => setLive(true))

      es.addEventListener("snapshot", (evt) => {
        try {
          applySnapshot(JSON.parse((evt as MessageEvent).data))
          setLive(true)
        } catch (err) {
          console.error("Snapshot de sesiones inválido:", err)
        }
      })

      es.onerror = () => {
        setLive(false)
        // EventSource reintenta solo mientras la conexión siga viva. Si el server
        // la cerró (p. ej. token expirado → 401), queda en CLOSED: reintentamos
        // de forma espaciada tras recargar el token.
        if (es && es.readyState === EventSource.CLOSED && !cleanedUp && !reconnectTimer) {
          es.close()
          reconnectTimer = setTimeout(() => {
            reconnectTimer = null
            connect()
          }, 5000)
        }
      }
    }

    connect()

    return () => {
      cleanedUp = true
      if (reconnectTimer) clearTimeout(reconnectTimer)
      es?.close()
    }
  }, [applySnapshot])

  // Polling de respaldo: solo activo mientras el stream esté caído. Silencioso:
  // reemplaza los datos sin animar nada.
  useEffect(() => {
    if (live) return
    const id = setInterval(() => {
      loadSessions({ silent: true })
    }, 5000)
    return () => clearInterval(id)
  }, [live, loadSessions])

  const handleOpenQR = useCallback((session: SessionInfo) => {
    setSelectedQR(session)
    setQrOpen(true)
  }, [])

  const handleOpenDisconnect = useCallback((session: SessionInfo) => {
    setSelectedDisconnect(session)
    setDisconnectOpen(true)
  }, [])

  const handleConfirmDisconnect = async () => {
    if (!selectedDisconnect) return
    setDisconnectLoading(true)
    try {
      await postAdminSession("disconnect", selectedDisconnect.account_id)
      // El stream ya empuja el snapshot nuevo; este refresco silencioso es el
      // respaldo si no está conectado.
      await loadSessions({ silent: true })
    } catch (error) {
      console.error("Failed to disconnect session:", error)
    } finally {
      setDisconnectLoading(false)
      setDisconnectOpen(false)
      setSelectedDisconnect(null)
    }
  }

  const handleReconnect = useCallback(async (telefonoId: number) => {
    setReconnectingId(telefonoId)
    try {
      await reconnectAdminSession(telefonoId)
      await loadSessions({ silent: true })
    } catch (error) {
      console.error("Failed to reconnect session:", error)
    } finally {
      setReconnectingId(null)
    }
  }, [loadSessions])

  const handleOpenEvents = useCallback((session: SessionInfo) => {
    setEventsSession(session)
    setEventsOpen(true)
  }, [])

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
    [handleOpenQR, handleOpenDisconnect, handleReconnect, handleOpenEvents, reconnectingId]
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
        loading={refreshing}
        onRefresh={() => loadSessions()}
        lastUpdated={lastUpdated}
        live={live}
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
              onClick={() => loadSessions()}
              disabled={refreshing}
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
        open={eventsOpen}
        onOpenChange={setEventsOpen}
      />
    </div>
  )
}

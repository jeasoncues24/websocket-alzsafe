"use client"

import { useCallback, useEffect, useState } from "react"
import { Card, CardContent } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription } from "@/components/ui/sheet"
import { Skeleton } from "@/components/ui/skeleton"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { QRRender } from "@/components/qr/qr-render"
import {
  QrCode, LogOut, RefreshCw, Wifi, WifiOff, Loader2,
  Search, Building2, Hash, ClipboardList, AlertTriangle,
} from "lucide-react"
import {
  getAdminSessions, postAdminSession, reconnectAdminSession,
  type SessionInfo, type SessionSummary,
} from "@/lib/api"

function relativeTime(ts: string): string {
  const diff = Date.now() - new Date(ts).getTime()
  const m = Math.floor(diff / 60000)
  if (m < 1) return "ahora"
  if (m < 60) return `hace ${m}m`
  const h = Math.floor(m / 60)
  if (h < 24) return `hace ${h}h`
  return `hace ${Math.floor(h / 24)}d`
}

function formatLocalTime(ts: string): string {
  return new Date(ts).toLocaleString("es-PE", {
    day: "2-digit", month: "2-digit",
    hour: "2-digit", minute: "2-digit", second: "2-digit",
    hour12: false,
  })
}

function StatusBadge({ status }: { status: string }) {
  switch (status) {
    case "active":
      return <Badge className="bg-green-500 hover:bg-green-500 text-white">Activa</Badge>
    case "qr_pending":
      return <Badge variant="secondary">QR Pendiente</Badge>
    case "initializing":
      return <Badge variant="secondary">Conectando</Badge>
    case "disconnected":
      return <Badge variant="destructive">Desconectada</Badge>
    default:
      return <Badge variant="outline">Inactiva</Badge>
  }
}

function MetricsTile({ label, value, colorClass }: { label: string; value: number; colorClass: string }) {
  return (
    <Card>
      <CardContent className="pt-4 pb-4">
        <div className={`text-2xl font-bold ${colorClass}`}>{value}</div>
        <div className="text-sm text-muted-foreground">{label}</div>
      </CardContent>
    </Card>
  )
}

function EventTypeIcon({ type }: { type: string }) {
  switch (type) {
    case "connected": return <Wifi className="h-3 w-3 text-green-500 shrink-0" />
    case "disconnected": return <WifiOff className="h-3 w-3 text-red-400 shrink-0" />
    case "initializing": return <Loader2 className="h-3 w-3 text-gray-400 shrink-0" />
    default: return <QrCode className="h-3 w-3 text-blue-400 shrink-0" />
  }
}

export default function SessionsPage() {
  const [sessions, setSessions] = useState<SessionInfo[]>([])
  const [summary, setSummary] = useState<SessionSummary | null>(null)
  const [loading, setLoading] = useState(true)

  const [empresaFilter, setEmpresaFilter] = useState("")
  const [numeroFilter, setNumeroFilter] = useState("")

  const [selectedSession, setSelectedSession] = useState<SessionInfo | null>(null)
  const [qrOpen, setQrOpen] = useState(false)
  const [disconnectOpen, setDisconnectOpen] = useState(false)
  const [disconnectingId, setDisconnectingId] = useState<string | null>(null)
  const [reconnectingId, setReconnectingId] = useState<number | null>(null)
  const [eventsSession, setEventsSession] = useState<SessionInfo | null>(null)

  const loadSessions = useCallback(async () => {
    setLoading(true)
    try {
      const data = await getAdminSessions()
      setSessions(data.sessions)
      setSummary(data.summary ?? null)
    } catch (error) {
      console.error("Failed to load sessions:", error)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { loadSessions() }, [loadSessions])

  const openDisconnect = (accountId: string) => {
    setDisconnectingId(accountId)
    setDisconnectOpen(true)
  }

  const confirmDisconnect = async () => {
    if (!disconnectingId) return
    try {
      await postAdminSession("disconnect", disconnectingId)
      await loadSessions()
    } catch (error) {
      console.error("Failed to disconnect:", error)
    } finally {
      setDisconnectOpen(false)
      setDisconnectingId(null)
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

  const filteredSessions = sessions.filter((s) => {
    const empresaOk = !empresaFilter ||
      (s.empresa_nombre ?? s.account_id).toLowerCase().includes(empresaFilter.toLowerCase())
    const numeroOk = !numeroFilter ||
      s.account_id.includes(numeroFilter.replace(/\D/g, ""))
    return empresaOk && numeroOk
  })

  const disconnectingName = sessions.find(s => s.account_id === disconnectingId)?.empresa_nombre ?? disconnectingId

  return (
    <div className="space-y-6">
      {/* Encabezado */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Sesiones WhatsApp</h1>
          <p className="text-muted-foreground">Administra las conexiones de WhatsApp por empresa</p>
        </div>
        <Button variant="outline" onClick={loadSessions} disabled={loading}>
          <RefreshCw className={`h-4 w-4 mr-2 ${loading ? "animate-spin" : ""}`} />
          Actualizar
        </Button>
      </div>

      {/* Métricas */}
      {summary && (
        <div className="grid gap-4 grid-cols-2 md:grid-cols-4">
          <MetricsTile label="Activas" value={summary.active} colorClass="text-green-600" />
          <MetricsTile label="Desconectadas" value={summary.disconnected} colorClass="text-red-600" />
          <MetricsTile label="Inconsistentes" value={summary.mismatch} colorClass="text-yellow-600" />
          <MetricsTile label="QR Pendiente" value={summary.qr_pending} colorClass="text-blue-600" />
        </div>
      )}

      {/* Buscadores */}
      <div className="flex flex-col sm:flex-row gap-3">
        <div className="relative flex-1">
          <Building2 className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
          <Input
            placeholder="Buscar por empresa..."
            value={empresaFilter}
            onChange={(e) => setEmpresaFilter(e.target.value)}
            className="pl-9"
          />
        </div>
        <div className="relative flex-1">
          <Hash className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
          <Input
            placeholder="Buscar por número..."
            value={numeroFilter}
            onChange={(e) => setNumeroFilter(e.target.value)}
            className="pl-9"
            inputMode="tel"
          />
        </div>
        {(empresaFilter || numeroFilter) && (
          <Button
            variant="ghost"
            size="icon"
            onClick={() => { setEmpresaFilter(""); setNumeroFilter("") }}
            title="Limpiar filtros"
          >
            <Search className="h-4 w-4" />
          </Button>
        )}
      </div>

      {/* Tabla */}
      <Card>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-[200px]">Empresa</TableHead>
              <TableHead>Número</TableHead>
              <TableHead>Estado</TableHead>
              <TableHead className="text-center">Runtime</TableHead>
              <TableHead>Última Conexión</TableHead>
              <TableHead className="text-right pr-4">Acciones</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {loading ? (
              Array.from({ length: 5 }).map((_, i) => (
                <TableRow key={i}>
                  {Array.from({ length: 6 }).map((_, j) => (
                    <TableCell key={j}><Skeleton className="h-5 w-full" /></TableCell>
                  ))}
                </TableRow>
              ))
            ) : filteredSessions.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center py-12 text-muted-foreground">
                  {sessions.length === 0
                    ? "No hay sesiones configuradas"
                    : "Ninguna sesión coincide con los filtros"}
                </TableCell>
              </TableRow>
            ) : (
              filteredSessions.map((session) => {
                const isReconnecting = reconnectingId === session.telefono_id
                const hasEvents = (session.events?.length ?? 0) > 0
                return (
                  <TableRow key={session.account_id}>
                    {/* Empresa */}
                    <TableCell className="font-medium max-w-[200px]">
                      <div className="flex items-center gap-2 min-w-0">
                        <Building2 className="h-4 w-4 shrink-0 text-muted-foreground" />
                        <span className="truncate" title={session.empresa_nombre ?? session.account_id}>
                          {session.empresa_nombre ?? "—"}
                        </span>
                      </div>
                    </TableCell>

                    {/* Número */}
                    <TableCell>
                      <span className="font-mono text-xs text-muted-foreground select-all">
                        {session.account_id}
                      </span>
                    </TableCell>

                    {/* Estado */}
                    <TableCell>
                      <div className="flex items-center gap-1.5 flex-wrap">
                        <StatusBadge status={session.status} />
                        {session.reconnecting && (
                          <Badge variant="outline" className="border-blue-400 text-blue-500 text-xs">
                            Reconectando
                          </Badge>
                        )}
                        {session.mismatch && !session.reconnecting && (
                          <span title="Estado inconsistente">
                            <AlertTriangle className="h-4 w-4 text-yellow-500" />
                          </span>
                        )}
                      </div>
                    </TableCell>

                    {/* Runtime */}
                    <TableCell className="text-center">
                      {session.runtime_connected ? (
                        <span className="inline-flex items-center justify-center" title="Runtime conectado">
                          <span className="h-2.5 w-2.5 rounded-full bg-green-500 block" />
                        </span>
                      ) : (
                        <span className="inline-flex items-center justify-center" title="Runtime desconectado">
                          <span className="h-2.5 w-2.5 rounded-full bg-red-400 block" />
                        </span>
                      )}
                    </TableCell>

                    {/* Última conexión */}
                    <TableCell className="text-sm text-muted-foreground">
                      {session.last_connected ? (
                        <span title={formatLocalTime(session.last_connected)}>
                          {relativeTime(session.last_connected)}
                          <span className="hidden md:inline text-xs ml-1 opacity-70">
                            · {formatLocalTime(session.last_connected)}
                          </span>
                        </span>
                      ) : (
                        <span className="opacity-50">Nunca</span>
                      )}
                    </TableCell>

                    {/* Acciones */}
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-1.5">
                        {session.status === "qr_pending" && (
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => { setSelectedSession(session); setQrOpen(true) }}
                          >
                            <QrCode className="h-3.5 w-3.5 mr-1.5" />
                            Ver QR
                          </Button>
                        )}
                        {session.status === "active" && (
                          <Button
                            variant="outline"
                            size="sm"
                            onClick={() => openDisconnect(session.account_id)}
                          >
                            <LogOut className="h-3.5 w-3.5 mr-1.5" />
                            Desconectar
                          </Button>
                        )}
                        {session.status !== "active" && session.telefono_id != null && !session.reconnecting && (
                          <Button
                            variant="outline"
                            size="sm"
                            disabled={isReconnecting}
                            onClick={() => handleReconnect(session.telefono_id!)}
                          >
                            <RefreshCw className={`h-3.5 w-3.5 mr-1.5 ${isReconnecting ? "animate-spin" : ""}`} />
                            Reconectar
                          </Button>
                        )}
                        {hasEvents && (
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => setEventsSession(session)}
                            title="Ver historial de eventos"
                          >
                            <ClipboardList className="h-3.5 w-3.5 mr-1.5" />
                            Eventos
                          </Button>
                        )}
                      </div>
                    </TableCell>
                  </TableRow>
                )
              })
            )}
          </TableBody>
        </Table>
      </Card>

      {/* Conteo de resultados cuando hay filtro activo */}
      {(empresaFilter || numeroFilter) && !loading && (
        <p className="text-sm text-muted-foreground">
          {filteredSessions.length} de {sessions.length} sesiones
        </p>
      )}

      {/* Dialog: QR */}
      <Dialog open={qrOpen} onOpenChange={setQrOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>QR — {selectedSession?.empresa_nombre ?? selectedSession?.account_id}</DialogTitle>
            <DialogDescription>Escanea este código con WhatsApp en tu teléfono</DialogDescription>
          </DialogHeader>
          {selectedSession?.qr_string && (
            <div className="flex flex-col items-center gap-4 p-4">
              <QRRender value={selectedSession.qr_string} size={200} title={`QR ${selectedSession.account_id}`} />
              <p className="text-xs text-muted-foreground text-center">
                Válido por 60 segundos. Recarga la página si expira.
              </p>
            </div>
          )}
        </DialogContent>
      </Dialog>

      {/* Dialog: Confirmar desconexión */}
      <Dialog open={disconnectOpen} onOpenChange={setDisconnectOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Confirmar desconexión</DialogTitle>
            <DialogDescription>
              ¿Deseas desconectar la sesión de <strong>{disconnectingName}</strong>?
              El dispositivo deberá escanear el código QR nuevamente para reconectar.
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-end gap-2 mt-4">
            <Button variant="outline" onClick={() => setDisconnectOpen(false)}>Cancelar</Button>
            <Button variant="destructive" onClick={confirmDisconnect}>Desconectar</Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* Sheet: Historial de eventos */}
      <Sheet open={!!eventsSession} onOpenChange={(open) => { if (!open) setEventsSession(null) }}>
        <SheetContent className="w-full sm:max-w-md overflow-y-auto">
          <SheetHeader className="mb-4">
            <SheetTitle>Historial de eventos</SheetTitle>
            <SheetDescription>
              {eventsSession?.empresa_nombre ?? eventsSession?.account_id}
              <span className="block font-mono text-xs mt-0.5 text-muted-foreground">
                {eventsSession?.account_id}
              </span>
            </SheetDescription>
          </SheetHeader>
          <div className="space-y-1">
            {eventsSession?.events && [...eventsSession.events].reverse().map((evt, i) => (
              <div
                key={i}
                className="flex items-start gap-2.5 py-2 border-b last:border-0 text-sm"
              >
                <span className="mt-0.5">
                  {evt.type === "connected" && <Wifi className="h-3.5 w-3.5 text-green-500" />}
                  {evt.type === "disconnected" && <WifiOff className="h-3.5 w-3.5 text-red-400" />}
                  {evt.type === "initializing" && <Loader2 className="h-3.5 w-3.5 text-gray-400" />}
                  {!["connected", "disconnected", "initializing"].includes(evt.type) && (
                    <QrCode className="h-3.5 w-3.5 text-blue-400" />
                  )}
                </span>
                <div className="flex-1 min-w-0">
                  <span className="capitalize font-medium">{evt.type}</span>
                  {evt.details && (
                    <span className="block text-xs text-muted-foreground truncate" title={evt.details}>
                      {evt.details}
                    </span>
                  )}
                </div>
                <span className="font-mono text-xs text-muted-foreground shrink-0" title={formatLocalTime(evt.timestamp)}>
                  {formatLocalTime(evt.timestamp)}
                </span>
              </div>
            ))}
          </div>
        </SheetContent>
      </Sheet>
    </div>
  )
}

"use client"

import { useCallback, useEffect, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog"
import { QRRender } from "@/components/qr/qr-render"
import { Building2, QrCode, LogOut, RefreshCw, Wifi, WifiOff, Loader2, ChevronDown, ChevronUp } from "lucide-react"
import { getAdminSessions, postAdminSession, reconnectAdminSession, type SessionInfo, type SessionSummary } from "@/lib/api"

function relativeTime(ts: string): string {
  const diff = Date.now() - new Date(ts).getTime()
  const m = Math.floor(diff / 60000)
  if (m < 1) return "ahora"
  if (m < 60) return `hace ${m}m`
  const h = Math.floor(m / 60)
  if (h < 24) return `hace ${h}h`
  return `hace ${Math.floor(h / 24)}d`
}

function EventIcon({ type }: { type: string }) {
  switch (type) {
    case "connected": return <Wifi className="h-3 w-3 text-green-500" />
    case "disconnected": return <WifiOff className="h-3 w-3 text-red-400" />
    case "initializing": return <Loader2 className="h-3 w-3 text-gray-400" />
    default: return <QrCode className="h-3 w-3 text-blue-400" />
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

export default function SessionsPage() {
  const [sessions, setSessions] = useState<SessionInfo[]>([])
  const [summary, setSummary] = useState<SessionSummary | null>(null)
  const [loading, setLoading] = useState(true)
  const [selectedSession, setSelectedSession] = useState<SessionInfo | null>(null)
  const [qrOpen, setQrOpen] = useState(false)
  const [disconnectOpen, setDisconnectOpen] = useState(false)
  const [disconnectingId, setDisconnectingId] = useState<string | null>(null)
  const [reconnectingId, setReconnectingId] = useState<number | null>(null)
  const [expandedEvents, setExpandedEvents] = useState<Set<string>>(new Set())

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

  useEffect(() => {
    loadSessions()
  }, [loadSessions])

  const disconnect = async (accountId: string) => {
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
      if (reconnectingId === telefonoId) setReconnectingId(null)
    }
  }

  const toggleEvents = (accountId: string) => {
    setExpandedEvents(prev => {
      const next = new Set(prev)
      if (next.has(accountId)) next.delete(accountId)
      else next.add(accountId)
      return next
    })
  }

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "active":
        return <Badge variant="default" className="bg-green-500">Activa</Badge>
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

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Sesiones WhatsApp</h1>
          <p className="text-muted-foreground">
            Administra las conexiones de WhatsApp por empresa
          </p>
        </div>
        <Button variant="outline" onClick={loadSessions}>
          <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""} mr-2`} />
          Actualizar
        </Button>
      </div>

      {summary && (
        <div className="grid gap-4 grid-cols-2 md:grid-cols-4">
          <MetricsTile label="Activas" value={summary.active} colorClass="text-green-600" />
          <MetricsTile label="Desconectadas" value={summary.disconnected} colorClass="text-red-600" />
          <MetricsTile label="Inconsistentes" value={summary.mismatch} colorClass="text-yellow-600" />
          <MetricsTile label="QR Pendiente" value={summary.qr_pending} colorClass="text-blue-600" />
        </div>
      )}

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {loading ? (
          <Card>
            <CardContent className="pt-6">
              <p className="text-muted-foreground">Cargando...</p>
            </CardContent>
          </Card>
        ) : sessions.length === 0 ? (
          <Card>
            <CardContent className="pt-6">
              <p className="text-muted-foreground">No hay sesiones configuradas</p>
            </CardContent>
          </Card>
        ) : (
          sessions.map((session) => {
            const hasEvents = session.events && session.events.length > 0
            const eventsOpen = expandedEvents.has(session.account_id)
            const isReconnecting = reconnectingId === session.telefono_id
            return (
              <Card key={session.account_id}>
                <CardHeader className="pb-2">
                  <div className="flex items-center justify-between gap-2">
                    <div className="flex items-center gap-2 min-w-0">
                      <Building2 className="h-5 w-5 shrink-0" />
                      <CardTitle className="text-lg truncate">
                        {session.empresa_nombre ?? session.account_id}
                      </CardTitle>
                    </div>
                    <div className="flex items-center gap-1 shrink-0">
                      {getStatusBadge(session.status)}
                      {session.mismatch && (
                        <Badge variant="outline" className="border-yellow-500 text-yellow-600">
                          Inconsistente
                        </Badge>
                      )}
                    </div>
                  </div>
                  <CardDescription className="flex items-center gap-2">
                    <span className={session.runtime_connected ? "text-green-500" : "text-red-400"}>●</span>
                    <span className="font-mono text-xs">{session.account_id}</span>
                  </CardDescription>
                  <CardDescription>
                    Última conexión:{" "}
                    {session.last_connected
                      ? relativeTime(session.last_connected)
                      : "Nunca"}
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-2">
                  <div className="flex gap-2 flex-wrap">
                    {session.status === "qr_pending" && (
                      <Button variant="outline" size="sm" onClick={() => {
                        setSelectedSession(session)
                        setQrOpen(true)
                      }}>
                        <QrCode className="h-4 w-4 mr-2" />
                        Ver QR
                      </Button>
                    )}
                    {session.status === "active" && (
                      <Button variant="outline" size="sm" onClick={() => disconnect(session.account_id)}>
                        <LogOut className="h-4 w-4 mr-2" />
                        Desconectar
                      </Button>
                    )}
                    {session.status !== "active" && session.telefono_id != null && (
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={isReconnecting}
                        onClick={() => handleReconnect(session.telefono_id!)}
                      >
                        <RefreshCw className={`h-4 w-4 mr-2 ${isReconnecting ? "animate-spin" : ""}`} />
                        Reconectar
                      </Button>
                    )}
                    {hasEvents && (
                      <Button variant="ghost" size="sm" onClick={() => toggleEvents(session.account_id)}>
                        {eventsOpen ? <ChevronUp className="h-4 w-4 mr-1" /> : <ChevronDown className="h-4 w-4 mr-1" />}
                        Ver eventos
                      </Button>
                    )}
                  </div>

                  {eventsOpen && hasEvents && (
                    <div className="border-t pt-2 space-y-1">
                      {[...session.events!].reverse().map((evt, i) => (
                        <div key={i} className="flex items-center gap-2 text-xs text-muted-foreground">
                          <EventIcon type={evt.type} />
                          <span className="capitalize">{evt.type}</span>
                          <span className="ml-auto">{relativeTime(evt.timestamp)}</span>
                          {evt.details && (
                            <span className="text-xs opacity-60">{evt.details}</span>
                          )}
                        </div>
                      ))}
                    </div>
                  )}
                </CardContent>
              </Card>
            )
          })
        )}
      </div>

      <Dialog open={qrOpen} onOpenChange={setQrOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>QR - {selectedSession?.empresa_nombre ?? selectedSession?.account_id}</DialogTitle>
            <DialogDescription>
              Escanea este código con WhatsApp en tu teléfono
            </DialogDescription>
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

      <Dialog open={disconnectOpen} onOpenChange={setDisconnectOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Confirmar desconexión</DialogTitle>
            <DialogDescription>
              ¿Estás seguro de que deseas desconectar la sesión de {disconnectingId}?
              El dispositivo deberá escanear el código QR nuevamente para reconectar.
            </DialogDescription>
          </DialogHeader>
          <div className="flex justify-end gap-2 mt-4">
            <Button variant="outline" onClick={() => setDisconnectOpen(false)}>
              Cancelar
            </Button>
            <Button variant="destructive" onClick={confirmDisconnect}>
              Desconectar
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}

"use client"

import { useCallback, useEffect, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog"
import { QRRender } from "@/components/qr/qr-render"
import { Building2, QrCode, LogOut, RefreshCw } from "lucide-react"
import { getAdminSessions, postAdminSession, type SessionInfo } from "@/lib/api"

export default function SessionsPage() {
  const [sessions, setSessions] = useState<SessionInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedSession, setSelectedSession] = useState<SessionInfo | null>(null)
  const [qrOpen, setQrOpen] = useState(false)
  const [disconnectOpen, setDisconnectOpen] = useState(false)
  const [disconnectingId, setDisconnectingId] = useState<string | null>(null)

  const loadSessions = useCallback(async () => {
    setLoading(true)
    try {
      const data = await getAdminSessions()
      setSessions(data.sessions)
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
          sessions.map((session) => (
            <Card key={session.account_id}>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <Building2 className="h-5 w-5" />
                    <CardTitle className="text-lg">{session.account_id}</CardTitle>
                  </div>
                  {getStatusBadge(session.status)}
                </div>
                <CardDescription>
                  Última actualización: {session.updated_at ? new Date(session.updated_at).toLocaleString() : "-"}
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="flex gap-2">
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
                </div>
              </CardContent>
            </Card>
          ))
        )}
      </div>

      <Dialog open={qrOpen} onOpenChange={setQrOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>QR - {selectedSession?.account_id}</DialogTitle>
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

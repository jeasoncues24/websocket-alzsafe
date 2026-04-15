"use client"

import { useEffect, useState } from "react"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog"
import { Building2, QrCode, LogOut, RefreshCw, AlertCircle, CheckCircle } from "lucide-react"
import { getAdminSessions, postAdminSession, type SessionInfo } from "@/lib/api"

export default function SessionsPage() {
  const [sessions, setSessions] = useState<SessionInfo[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedSession, setSelectedSession] = useState<SessionInfo | null>(null)
  const [qrOpen, setQrOpen] = useState(false)

  async function loadSessions() {
    setLoading(true)
    try {
      const data = await getAdminSessions()
      setSessions(data.sessions)
    } catch (error) {
      console.error("Failed to load sessions:", error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadSessions()
  }, [])

  const disconnect = async (accountId: string) => {
    if (!confirm(`¿Desconectar sesión de ${accountId}?`)) return
    try {
      await postAdminSession("disconnect", accountId)
      await loadSessions()
    } catch (error) {
      console.error("Failed to disconnect:", error)
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
              <img
                src={`https://api.qrserver.com/v1/create/?size=200x200&data=${encodeURIComponent(selectedSession.qr_string)}`}
                alt="QR Code"
                className="border rounded-lg"
              />
              <p className="text-xs text-muted-foreground text-center">
                Válido por 60 segundos. Recarga la página si expira.
              </p>
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  )
}
"use client"

import { useEffect, useState } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Skeleton } from "@/components/ui/skeleton"
import { Button } from "@/components/ui/button"
import { Building2, MessageSquare, Send, CheckCircle2, AlertCircle, RefreshCw } from "lucide-react"
import { getMetrics, type DashboardMetrics } from "@/lib/api"
import Link from "next/link"

export default function DashboardPage() {
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null)
  const [loading, setLoading] = useState(true)
  const [lastUpdate, setLastUpdate] = useState<string>("")

  async function loadMetrics() {
    setLoading(true)
    try {
      const data = await getMetrics()
      setMetrics(data)
      setLastUpdate(data.last_update)
    } catch (error) {
      console.error("Failed to load metrics:", error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadMetrics()
  }, [])

  const formatNumber = (num: number) => {
    if (num >= 1000) {
      return (num / 1000).toFixed(1) + "k"
    }
    return num.toLocaleString()
  }

  const formatPercent = (num: number) => {
    return num.toFixed(1) + "%"
  }

  const getAlertColor = (level: string) => {
    switch (level) {
      case "warning":
        return "text-yellow-500"
      case "error":
        return "text-red-500"
      default:
        return "text-blue-500"
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
          <p className="text-muted-foreground">
            Resumen global del sistema WhatsApp API
          </p>
        </div>
        <div className="flex items-center gap-2">
          {lastUpdate && (
            <span className="text-xs text-muted-foreground">
              Actualizado: {new Date(lastUpdate).toLocaleTimeString()}
            </span>
          )}
          <Button variant="outline" size="sm" onClick={loadMetrics}>
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
          </Button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Link href="/companies">
          <Card className="hover:bg-muted/50 transition-colors cursor-pointer">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Empresas Activas
              </CardTitle>
              <Building2 className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              {loading ? (
                <Skeleton className="h-8 w-20" />
              ) : (
                <>
                  <div className="text-2xl font-bold">
                    {formatNumber(metrics?.sessions_active || 0)}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Sesiones activas
                  </p>
                </>
              )}
            </CardContent>
          </Card>
        </Link>

        <Link href="/messages">
          <Card className="hover:bg-muted/50 transition-colors cursor-pointer">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Mensajes Hoy
              </CardTitle>
              <MessageSquare className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              {loading ? (
                <Skeleton className="h-8 w-20" />
              ) : (
                <>
                  <div className="text-2xl font-bold">
                    {formatNumber(metrics?.messages_today || 0)}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Enviados
                  </p>
                </>
              )}
            </CardContent>
          </Card>
        </Link>

        <Link href="/broadcasts">
          <Card className="hover:bg-muted/50 transition-colors cursor-pointer">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                Broadcasts Hoy
              </CardTitle>
              <Send className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              {loading ? (
                <Skeleton className="h-8 w-20" />
              ) : (
                <>
                  <div className="text-2xl font-bold">
                    {formatNumber(metrics?.broadcasts_today || 0)}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    Creados
                  </p>
                </>
              )}
            </CardContent>
          </Card>
        </Link>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Tasa de Éxito
            </CardTitle>
            <CheckCircle2 className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {loading ? (
              <Skeleton className="h-8 w-20" />
            ) : (
              <>
                <div className="text-2xl font-bold">
                  {formatPercent(metrics?.success_rate || 0)}
                </div>
                <p className="text-xs text-muted-foreground">
                  Mensajes entregados
                </p>
              </>
            )}
          </CardContent>
        </Card>
      </div>

      {metrics?.alerts && metrics.alerts.length > 0 && (
        <div className="grid gap-4 md:grid-cols-2">
          <Card>
            <CardHeader>
              <CardTitle>Alertas</CardTitle>
              <CardDescription>
                Recomendaciones del sistema
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {metrics.alerts.map((alert, i) => (
                  <div key={i} className="flex items-center gap-3">
                    <AlertCircle className={`h-4 w-4 ${getAlertColor(alert.level)}`} />
                    <span className="text-sm">{alert.message}</span>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle>Métricas</CardTitle>
              <CardDescription>
                Detalles adicionales
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Mensajes fallidos</span>
                  <span className="text-sm font-medium">
                    {formatNumber(metrics?.messages_failed || 0)}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-muted-foreground">Broadcasts completados</span>
                  <span className="text-sm font-medium">
                    {formatNumber(metrics?.broadcasts_created || 0)}
                  </span>
                </div>
              </div>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  )
}
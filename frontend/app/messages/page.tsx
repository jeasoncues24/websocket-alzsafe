"use client"

import { useCallback, useEffect, useState } from "react"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { Skeleton } from "@/components/ui/skeleton"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { getAdminMessages, getEmpresas, retryMessageAdmin, type AdminMessage, type Empresa } from "@/lib/api"
import { RefreshCw, AlertCircle, CheckCircle2 } from "lucide-react"

export default function MessagesPage() {
  const [messages, setMessages] = useState<AdminMessage[]>([])
  const [companies, setCompanies] = useState<Empresa[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState("all")
  const [filterAccount, setFilterAccount] = useState("")
  const [total, setTotal] = useState(0)
  const [retryingId, setRetryingId] = useState<number | null>(null)
  const [resultAlert, setResultAlert] = useState<{ok: boolean; message: string} | null>(null)

  const handleRetry = async (msg: AdminMessage) => {
    if (!msg.reference_id) return
    setRetryingId(msg.id)
    setResultAlert(null)
    try {
      const result = await retryMessageAdmin(msg.reference_id)
      console.log("Retry result:", result)
      setResultAlert({ ok: result.ok, message: result.error || "Mensaje reenviado exitosamente" })
      if (result.ok) {
        await loadData()
      }
    } catch (error) {
      const errMsg = error instanceof Error ? error.message : "Error desconocido"
      setResultAlert({ ok: false, message: errMsg })
      console.error("Retry failed:", error)
    } finally {
      setRetryingId(null)
    }
  }

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const statusFilter = activeTab === "all" ? "" : activeTab
      const [msgsData, compsData] = await Promise.all([
        getAdminMessages({ status: statusFilter, account_id: filterAccount, limit: 50 }),
        getEmpresas({ limit: 1000 }),
      ])
      setMessages(msgsData.messages)
      setTotal(msgsData.total)
      setCompanies(compsData.empresas)
    } catch (error) {
      console.error("Failed to load data:", error)
    } finally {
      setLoading(false)
    }
  }, [activeTab, filterAccount])

  useEffect(() => {
    loadData()
  }, [loadData])

  const getStatusBadge = (status: string, msg: AdminMessage) => {
    const retryCount = msg.retry_count
    switch (status) {
      case "sent":
        return (
          <Badge variant="default" className="bg-green-500">
            Enviado {retryCount && retryCount > 0 ? `(${retryCount})` : ""}
          </Badge>
        )
      case "delivered":
        return <Badge variant="default" className="bg-blue-500">Entregado</Badge>
      case "failed":
        return (
          <Badge variant="destructive">
            Fallido {retryCount && retryCount > 0 ? `(${retryCount})` : ""}
          </Badge>
        )
      case "pending":
        return <Badge variant="secondary">Pendiente</Badge>
      default:
        return <Badge variant="outline">{status}</Badge>
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Mensajes</h1>
        <p className="text-muted-foreground">
          Historial de todos los mensajes enviados
        </p>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Historial de Mensajes</CardTitle>
              <CardDescription>
                {total} mensaje(s) encontrado(s)
              </CardDescription>
            </div>
            <div className="flex gap-2">
              <select
                className="h-10 rounded-md border border-input bg-background px-3 py-2 text-sm"
                value={filterAccount}
                onChange={(e) => setFilterAccount(e.target.value)}
              >
                <option value="">Todas las empresas</option>
                {companies.map((c) => (
                  <option key={c.ruc} value={c.ruc}>
                    {c.ruc}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="all" onValueChange={setActiveTab}>
            <TabsList>
              <TabsTrigger value="all">Todos</TabsTrigger>
              <TabsTrigger value="pending">Pendiente</TabsTrigger>
              <TabsTrigger value="sent">Enviado</TabsTrigger>
              <TabsTrigger value="delivered">Entregado</TabsTrigger>
              <TabsTrigger value="failed">Fallido</TabsTrigger>
            </TabsList>
          </Tabs>
          {resultAlert && (
            <Alert variant={resultAlert.ok ? "default" : "destructive"} className="mt-4 flex items-start gap-3">
              <div className="mt-0.5">
                {resultAlert.ok ? <CheckCircle2 className="h-5 w-5 text-green-500" /> : <AlertCircle className="h-5 w-5 text-red-500" />}
              </div>
              <div>
                <AlertTitle>{resultAlert.ok ? "Éxito" : "Error"}</AlertTitle>
                <AlertDescription className={resultAlert.ok ? "text-green-600" : "text-red-600"}>
                  {resultAlert.message}
                </AlertDescription>
              </div>
            </Alert>
          )}
          <Table className="mt-4">
            <TableHeader>
              <TableRow>
                <TableHead>ID</TableHead>
                <TableHead>Empresa</TableHead>
                <TableHead>Destino</TableHead>
                <TableHead>Contenido</TableHead>
                <TableHead>Estado</TableHead>
                <TableHead>Error</TableHead>
                <TableHead>Fecha</TableHead>
                <TableHead>Acciones</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                Array.from({ length: 5 }).map((_, i) => (
                  <TableRow key={i}>
                    <TableCell><Skeleton className="h-4 w-12" /></TableCell>
                    <TableCell><Skeleton className="h-4 w-24" /></TableCell>
                    <TableCell><Skeleton className="h-4 w-28" /></TableCell>
                    <TableCell><Skeleton className="h-4 w-40" /></TableCell>
                    <TableCell><Skeleton className="h-5 w-20" /></TableCell>
                    <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                  </TableRow>
                ))
              ) : messages.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={8} className="text-center text-muted-foreground">
                    No hay mensajes
                  </TableCell>
                </TableRow>
              ) : (
                messages.map((msg) => (
                  <TableRow key={msg.id}>
                    <TableCell className="font-medium whitespace-nowrap">#{msg.id}</TableCell>
                    <TableCell className="whitespace-nowrap">{msg.account_id}</TableCell>
                    <TableCell className="whitespace-nowrap font-mono text-sm">{msg.to}</TableCell>
                    <TableCell className="max-w-md">
                      <div className="whitespace-pre-wrap text-sm">{msg.content}</div>
                    </TableCell>
                    <TableCell className="whitespace-nowrap">{getStatusBadge(msg.status, msg)}</TableCell>
                    <TableCell className="max-w-xs">
                      {msg.error_reason && (
                        <div className="whitespace-pre-wrap text-red-500 text-sm">{msg.error_reason}</div>
                      )}
                    </TableCell>
                    <TableCell className="whitespace-nowrap text-muted-foreground text-sm">
                      {msg.created_at ? new Date(msg.created_at).toLocaleString() : "-"}
                    </TableCell>
                    <TableCell className="whitespace-nowrap">
                      {msg.status === "failed" && msg.reference_id && (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => handleRetry(msg)}
                          disabled={retryingId === msg.id}
                        >
                          {retryingId === msg.id ? (
                            <RefreshCw className="h-4 w-4 animate-spin" />
                          ) : (
                            "Reintentar"
                          )}
                        </Button>
                      )}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}

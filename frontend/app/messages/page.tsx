"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import { AlertCircle, CheckCircle2, Eye, FileText, Paperclip, RefreshCw, MessageSquareMore } from "lucide-react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet"
import { Skeleton } from "@/components/ui/skeleton"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { getAdminMessages, getEmpresas, retryMessageAdmin, type AdminMessage, type Empresa } from "@/lib/api"

function formatDate(value?: string) {
  return value ? new Date(value).toLocaleString() : "-"
}

function formatBytes(bytes: number) {
  if (!bytes) return "0 B"
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function statusBadge(status: string, retryCount?: number) {
  switch (status) {
    case "sent":
      return <Badge className="bg-emerald-500 text-white hover:bg-emerald-500">Enviado{retryCount ? ` (${retryCount})` : ""}</Badge>
    case "delivered":
      return <Badge className="bg-sky-500 text-white hover:bg-sky-500">Entregado</Badge>
    case "failed":
      return <Badge variant="destructive">Fallido{retryCount ? ` (${retryCount})` : ""}</Badge>
    case "pending":
      return <Badge variant="secondary">Pendiente</Badge>
    default:
      return <Badge variant="outline">{status}</Badge>
  }
}

function attachmentLabel(count: number) {
  return count === 1 ? "1 adjunto" : `${count} adjuntos`
}

export default function MessagesPage() {
  const [messages, setMessages] = useState<AdminMessage[]>([])
  const [companies, setCompanies] = useState<Empresa[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState("all")
  const [filterAccount, setFilterAccount] = useState("")
  const [total, setTotal] = useState(0)
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

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const statusFilter = activeTab === "all" ? "" : activeTab
      const [msgsData, compsData] = await Promise.all([
        getAdminMessages({ status: statusFilter, account_id: filterAccount, limit: 50 }),
        getEmpresas({ limit: 1000 }),
      ])
      setMessages(msgsData.messages ?? [])
      setTotal(msgsData.total ?? 0)
      setCompanies(compsData.empresas ?? [])
    } catch (error) {
      console.error("Failed to load data:", error)
    } finally {
      setLoading(false)
    }
  }, [activeTab, filterAccount])

  useEffect(() => {
    loadData()
  }, [loadData])

  const handleRetry = async (msg: AdminMessage) => {
    if (!msg.reference_id) return
    setRetryingId(msg.id)
    setResultAlert(null)
    try {
      const result = await retryMessageAdmin(msg.reference_id)
      setResultAlert({ ok: result.ok, message: result.error || "Mensaje reenviado exitosamente" })
      if (result.ok) {
        await loadData()
      }
    } catch (error) {
      const errMsg = error instanceof Error ? error.message : "Error desconocido"
      setResultAlert({ ok: false, message: errMsg })
    } finally {
      setRetryingId(null)
    }
  }

  const openDetails = (msg: AdminMessage) => {
    setSelectedMessage(msg)
    setDetailsOpen(true)
  }

  const selectedCompany = selectedMessage ? companyByRuc.get(selectedMessage.account_id) : undefined

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-3xl font-bold tracking-tight">Mensajes</h1>
        <p className="text-muted-foreground">Historial claro, con adjuntos y errores visibles solo cuando hacen falta.</p>
      </div>

      <Card>
        <CardHeader className="space-y-4">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div className="space-y-1">
              <CardTitle>Historial de mensajes</CardTitle>
              <CardDescription>
                {loading ? "Cargando mensajes..." : `${total} mensaje(s) encontrados`}
              </CardDescription>
            </div>
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
              <select
                className="h-10 rounded-md border border-input bg-background px-3 py-2 text-sm"
                value={filterAccount}
                onChange={(e) => setFilterAccount(e.target.value)}
              >
                <option value="">Todas las empresas</option>
                {companies.map((c) => (
                  <option key={c.ruc} value={c.ruc}>
                    {c.nombre || c.ruc}
                  </option>
                ))}
              </select>
              {filterAccount ? (
                <Button variant="outline" onClick={() => setFilterAccount("")}>Limpiar filtro</Button>
              ) : null}
            </div>
          </div>
          <Tabs defaultValue="all" onValueChange={setActiveTab} className="space-y-3">
            <TabsList className="flex h-auto flex-wrap gap-2 bg-transparent p-0">
              <TabsTrigger value="all" className="rounded-md border border-border data-[state=active]:bg-primary data-[state=active]:text-primary-foreground">Todos</TabsTrigger>
              <TabsTrigger value="pending" className="rounded-md border border-border data-[state=active]:bg-primary data-[state=active]:text-primary-foreground">Pendiente</TabsTrigger>
              <TabsTrigger value="sent" className="rounded-md border border-border data-[state=active]:bg-primary data-[state=active]:text-primary-foreground">Enviado</TabsTrigger>
              <TabsTrigger value="delivered" className="rounded-md border border-border data-[state=active]:bg-primary data-[state=active]:text-primary-foreground">Entregado</TabsTrigger>
              <TabsTrigger value="failed" className="rounded-md border border-border data-[state=active]:bg-primary data-[state=active]:text-primary-foreground">Fallido</TabsTrigger>
            </TabsList>
          </Tabs>
          {resultAlert ? (
            <Alert variant={resultAlert.ok ? "default" : "destructive"}>
              <div className="flex items-start gap-3">
                {resultAlert.ok ? <CheckCircle2 className="mt-0.5 h-5 w-5 text-emerald-500" /> : <AlertCircle className="mt-0.5 h-5 w-5" />}
                <div>
                  <AlertTitle>{resultAlert.ok ? "Hecho" : "Error"}</AlertTitle>
                  <AlertDescription>{resultAlert.message}</AlertDescription>
                </div>
              </div>
            </Alert>
          ) : null}
        </CardHeader>
        <CardContent className="space-y-4">
          {!loading && messages.length === 0 ? (
            <div className="rounded-lg border border-dashed p-8 text-center">
              <MessageSquareMore className="mx-auto h-10 w-10 text-muted-foreground" />
              <h3 className="mt-4 text-lg font-medium">No hay mensajes para mostrar</h3>
              <p className="mt-1 text-sm text-muted-foreground">
                Prueba cambiando el filtro de empresa o estado.
              </p>
            </div>
          ) : null}

          {loading || messages.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>Empresa</TableHead>
                  <TableHead>Destino</TableHead>
                  <TableHead>Mensaje</TableHead>
                  <TableHead>Estado</TableHead>
                  <TableHead>Fecha</TableHead>
                  <TableHead className="text-right">Acciones</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {loading ? (
                  Array.from({ length: 5 }).map((_, i) => (
                    <TableRow key={i}>
                      <TableCell><Skeleton className="h-4 w-12" /></TableCell>
                      <TableCell><Skeleton className="h-4 w-28" /></TableCell>
                      <TableCell><Skeleton className="h-4 w-28" /></TableCell>
                      <TableCell><Skeleton className="h-4 w-80" /></TableCell>
                      <TableCell><Skeleton className="h-5 w-24" /></TableCell>
                      <TableCell><Skeleton className="h-4 w-28" /></TableCell>
                      <TableCell><Skeleton className="ml-auto h-8 w-28" /></TableCell>
                    </TableRow>
                  ))
                ) : messages.map((msg) => {
                  const attachmentCount = msg.adjuntos?.length ?? 0
                  const company = companyByRuc.get(msg.account_id)

                  return (
                    <TableRow key={msg.id}>
                      <TableCell className="font-medium">#{msg.id}</TableCell>
                      <TableCell>
                        <div className="space-y-0.5">
                          <div className="font-medium">{company?.nombre || msg.account_id}</div>
                          <div className="text-xs text-muted-foreground">{company?.ruc || msg.account_id}</div>
                        </div>
                      </TableCell>
                      <TableCell className="font-mono text-sm">{msg.to}</TableCell>
                      <TableCell className="max-w-[34rem] whitespace-normal">
                        <div className="space-y-2">
                          <p className="line-clamp-2 text-sm text-foreground">{msg.content || "Sin contenido"}</p>
                          <div className="flex flex-wrap gap-2">
                            {attachmentCount > 0 ? (
                              <Badge variant="outline" className="gap-1">
                                <Paperclip className="h-3.5 w-3.5" />
                                {attachmentLabel(attachmentCount)}
                              </Badge>
                            ) : null}
                            {msg.error_reason ? (
                              <span className="line-clamp-1 text-xs text-destructive">{msg.error_reason}</span>
                            ) : null}
                          </div>
                        </div>
                      </TableCell>
                      <TableCell>{statusBadge(msg.status, msg.retry_count)}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{formatDate(msg.created_at)}</TableCell>
                      <TableCell>
                        <div className="flex justify-end gap-2">
                          <Button variant="outline" size="sm" onClick={() => openDetails(msg)}>
                            <Eye className="mr-2 h-4 w-4" />
                            Ver detalles
                          </Button>
                          {msg.status === "failed" && msg.reference_id ? (
                            <Button
                              variant="secondary"
                              size="sm"
                              onClick={() => handleRetry(msg)}
                              disabled={retryingId === msg.id}
                            >
                              {retryingId === msg.id ? (
                                <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
                              ) : null}
                              Reintentar
                            </Button>
                          ) : null}
                        </div>
                      </TableCell>
                    </TableRow>
                  )
                })}
              </TableBody>
            </Table>
          ) : null}
        </CardContent>
      </Card>

      <Sheet open={detailsOpen} onOpenChange={setDetailsOpen}>
        <SheetContent side="right" className="w-full overflow-y-auto sm:max-w-2xl">
          <SheetHeader className="text-left">
            <SheetTitle>Detalle del mensaje</SheetTitle>
            <SheetDescription>
              {selectedMessage?.reference_id ? `Referencia ${selectedMessage.reference_id}` : "Selecciona un mensaje para ver más"}
            </SheetDescription>
          </SheetHeader>

          {selectedMessage ? (
            <div className="mt-6 space-y-6">
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="rounded-lg border p-4">
                  <p className="text-xs uppercase tracking-wide text-muted-foreground">Empresa</p>
                  <p className="mt-1 font-medium">{selectedCompany?.nombre || selectedMessage.account_id}</p>
                  <p className="text-sm text-muted-foreground">{selectedCompany?.ruc || selectedMessage.account_id}</p>
                </div>
                <div className="rounded-lg border p-4">
                  <p className="text-xs uppercase tracking-wide text-muted-foreground">Estado</p>
                  <div className="mt-2">{statusBadge(selectedMessage.status, selectedMessage.retry_count)}</div>
                </div>
              </div>

              <div className="rounded-lg border p-4">
                <p className="text-xs uppercase tracking-wide text-muted-foreground">Destino</p>
                <p className="mt-1 font-mono text-sm">{selectedMessage.to}</p>
              </div>

              <div className="rounded-lg border p-4">
                <p className="text-xs uppercase tracking-wide text-muted-foreground">Contenido</p>
                <p className="mt-2 whitespace-pre-wrap text-sm leading-6">{selectedMessage.content || "Sin contenido"}</p>
              </div>

              <div className="rounded-lg border p-4">
                <div className="flex items-center justify-between gap-3">
                  <p className="text-xs uppercase tracking-wide text-muted-foreground">Adjuntos</p>
                  <Badge variant="outline">{attachmentLabel(selectedMessage.adjuntos?.length ?? 0)}</Badge>
                </div>
                <div className="mt-3 space-y-3">
                  {selectedMessage.adjuntos?.length ? (
                    selectedMessage.adjuntos.map((att) => (
                      <div key={`${att.sha256_hash}-${att.nombre}`} className="rounded-md border bg-muted/30 p-3">
                        <div className="flex items-center gap-2">
                          <FileText className="h-4 w-4 text-muted-foreground" />
                          <p className="font-medium">{att.nombre}</p>
                        </div>
                        <p className="mt-1 text-sm text-muted-foreground">{formatBytes(att.tamano_bytes)}</p>
                        <p className="mt-1 break-all text-xs text-muted-foreground">{att.sha256_hash}</p>
                      </div>
                    ))
                  ) : (
                    <p className="text-sm text-muted-foreground">Este mensaje no tiene adjuntos.</p>
                  )}
                </div>
              </div>

              {selectedMessage.error_reason ? (
                <Alert variant="destructive">
                  <AlertCircle className="h-4 w-4" />
                  <AlertTitle>Error</AlertTitle>
                  <AlertDescription>{selectedMessage.error_reason}</AlertDescription>
                </Alert>
              ) : null}

              <div className="grid gap-4 sm:grid-cols-2">
                <div className="rounded-lg border p-4">
                  <p className="text-xs uppercase tracking-wide text-muted-foreground">Fecha</p>
                  <p className="mt-1 text-sm">{formatDate(selectedMessage.created_at)}</p>
                </div>
                <div className="rounded-lg border p-4">
                  <p className="text-xs uppercase tracking-wide text-muted-foreground">Reintentos</p>
                  <p className="mt-1 text-sm">{selectedMessage.retry_count ?? 0}</p>
                </div>
              </div>

              {selectedMessage.status === "failed" && selectedMessage.reference_id ? (
                <Button className="w-full" onClick={() => handleRetry(selectedMessage)} disabled={retryingId === selectedMessage.id}>
                  {retryingId === selectedMessage.id ? <RefreshCw className="mr-2 h-4 w-4 animate-spin" /> : null}
                  Reintentar mensaje
                </Button>
              ) : null}
            </div>
          ) : null}
        </SheetContent>
      </Sheet>
    </div>
  )
}

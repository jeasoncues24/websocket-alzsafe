"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import { Eye, FileText, MessageSquareMore } from "lucide-react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from "@/components/ui/sheet"
import { Skeleton } from "@/components/ui/skeleton"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { getAdminBroadcasts, getEmpresas, type BroadcastInfo, type Empresa } from "@/lib/api"

function formatDate(value?: string) {
  return value ? new Date(value).toLocaleString() : "-"
}

function statusBadge(status: string) {
  switch (status) {
    case "completed":
      return <Badge className="bg-emerald-500 text-white hover:bg-emerald-500">Completado</Badge>
    case "processing":
      return <Badge variant="secondary">Procesando</Badge>
    case "pending":
      return <Badge variant="outline">Pendiente</Badge>
    case "failed":
      return <Badge variant="destructive">Fallido</Badge>
    default:
      return <Badge variant="outline">{status}</Badge>
  }
}

function attachmentLabel(count: number) {
  return count === 1 ? "1 adjunto" : `${count} adjuntos`
}

export default function BroadcastsPage() {
  const [broadcasts, setBroadcasts] = useState<BroadcastInfo[]>([])
  const [companies, setCompanies] = useState<Empresa[]>([])
  const [loading, setLoading] = useState(true)
  const [filterRuc, setFilterRuc] = useState("")
  const [selectedBroadcast, setSelectedBroadcast] = useState<BroadcastInfo | null>(null)
  const [detailsOpen, setDetailsOpen] = useState(false)

  const companyByRuc = useMemo(() => {
    const map = new Map<string, Empresa>()
    companies.forEach((company) => map.set(company.ruc, company))
    return map
  }, [companies])

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const [bcData, compsData] = await Promise.all([
        getAdminBroadcasts(filterRuc || undefined),
        getEmpresas({ limit: 1000 }),
      ])
      setBroadcasts(bcData.broadcasts)
      setCompanies(compsData.empresas)
    } catch (error) {
      console.error("Failed to load data:", error)
    } finally {
      setLoading(false)
    }
  }, [filterRuc])

  useEffect(() => {
    loadData()
  }, [loadData])

  const openDetails = (broadcast: BroadcastInfo) => {
    setSelectedBroadcast(broadcast)
    setDetailsOpen(true)
  }

  const selectedCompany = selectedBroadcast ? companyByRuc.get(selectedBroadcast.ruc_empresa) : undefined

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-3xl font-bold tracking-tight">Broadcasts</h1>
        <p className="text-muted-foreground">Historial de difusiones masivas con una vista más limpia y fácil de leer.</p>
      </div>

      <Card>
        <CardHeader className="space-y-4">
          <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div className="space-y-1">
              <CardTitle>Lista de broadcasts</CardTitle>
              <CardDescription>{loading ? "Cargando broadcasts..." : `${broadcasts.length} broadcast(s) encontrado(s)`}</CardDescription>
            </div>
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
              <select
                className="h-10 rounded-md border border-input bg-background px-3 py-2 text-sm"
                value={filterRuc}
                onChange={(e) => setFilterRuc(e.target.value)}
              >
                <option value="">Todas las empresas</option>
                {companies.map((c) => (
                  <option key={c.ruc} value={c.ruc}>
                    {c.nombre || c.ruc}
                  </option>
                ))}
              </select>
              {filterRuc ? <Button variant="outline" onClick={() => setFilterRuc("")}>Limpiar filtro</Button> : null}
            </div>
          </div>
        </CardHeader>

        <CardContent className="space-y-4">
          {!loading && broadcasts.length === 0 ? (
            <div className="rounded-lg border border-dashed p-8 text-center">
              <MessageSquareMore className="mx-auto h-10 w-10 text-muted-foreground" />
              <h3 className="mt-4 text-lg font-medium">No hay broadcasts para mostrar</h3>
              <p className="mt-1 text-sm text-muted-foreground">Prueba con otra empresa o revisa si ya se enviaron difusiones.</p>
            </div>
          ) : null}

          {loading || broadcasts.length > 0 ? (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Reference ID</TableHead>
                  <TableHead>Empresa</TableHead>
                  <TableHead>Total</TableHead>
                  <TableHead>Estado</TableHead>
                  <TableHead>Fecha</TableHead>
                  <TableHead className="text-right">Acciones</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {loading ? (
                  Array.from({ length: 5 }).map((_, i) => (
                    <TableRow key={i}>
                      <TableCell><Skeleton className="h-4 w-28" /></TableCell>
                      <TableCell><Skeleton className="h-4 w-32" /></TableCell>
                      <TableCell><Skeleton className="h-4 w-12" /></TableCell>
                      <TableCell><Skeleton className="h-5 w-24" /></TableCell>
                      <TableCell><Skeleton className="h-4 w-28" /></TableCell>
                      <TableCell><Skeleton className="ml-auto h-8 w-28" /></TableCell>
                    </TableRow>
                  ))
                ) : broadcasts.map((bc) => {
                  const company = companyByRuc.get(bc.ruc_empresa)
                  const attachmentCount = bc.adjuntos?.length ?? 0

                  return (
                    <TableRow key={bc.reference_id}>
                      <TableCell className="font-mono text-sm">{bc.reference_id.slice(0, 8)}...</TableCell>
                      <TableCell>
                        <div className="space-y-0.5">
                          <div className="font-medium">{company?.nombre || bc.ruc_empresa}</div>
                          <div className="text-xs text-muted-foreground">{company?.ruc || bc.ruc_empresa}</div>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="space-y-1">
                          <div className="font-medium">{bc.total}</div>
                          {attachmentCount > 0 ? (
                            <Badge variant="outline" className="gap-1">
                              <FileText className="h-3.5 w-3.5" />
                              {attachmentLabel(attachmentCount)}
                            </Badge>
                          ) : null}
                        </div>
                      </TableCell>
                      <TableCell>{statusBadge(bc.status)}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{formatDate(bc.created_at)}</TableCell>
                      <TableCell>
                        <div className="flex justify-end">
                          <Button variant="outline" size="sm" onClick={() => openDetails(bc)}>
                            <Eye className="mr-2 h-4 w-4" />
                            Ver detalles
                          </Button>
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
            <SheetTitle>Detalle del broadcast</SheetTitle>
            <SheetDescription>
              {selectedBroadcast?.reference_id ? `Referencia ${selectedBroadcast.reference_id}` : "Selecciona una difusión para ver más"}
            </SheetDescription>
          </SheetHeader>

          {selectedBroadcast ? (
            <div className="mt-6 space-y-6">
              <div className="grid gap-4 sm:grid-cols-2">
                <div className="rounded-lg border p-4">
                  <p className="text-xs uppercase tracking-wide text-muted-foreground">Empresa</p>
                  <p className="mt-1 font-medium">{selectedCompany?.nombre || selectedBroadcast.ruc_empresa}</p>
                  <p className="text-sm text-muted-foreground">{selectedCompany?.ruc || selectedBroadcast.ruc_empresa}</p>
                </div>
                <div className="rounded-lg border p-4">
                  <p className="text-xs uppercase tracking-wide text-muted-foreground">Estado</p>
                  <div className="mt-2">{statusBadge(selectedBroadcast.status)}</div>
                </div>
              </div>

              <div className="grid gap-4 sm:grid-cols-3">
                <div className="rounded-lg border p-4">
                  <p className="text-xs uppercase tracking-wide text-muted-foreground">Total</p>
                  <p className="mt-1 text-2xl font-semibold">{selectedBroadcast.total}</p>
                </div>
                <div className="rounded-lg border p-4">
                  <p className="text-xs uppercase tracking-wide text-muted-foreground">Exitosos</p>
                  <p className="mt-1 text-2xl font-semibold text-emerald-600">{selectedBroadcast.success}</p>
                </div>
                <div className="rounded-lg border p-4">
                  <p className="text-xs uppercase tracking-wide text-muted-foreground">Fallidos</p>
                  <p className="mt-1 text-2xl font-semibold text-destructive">{selectedBroadcast.failed}</p>
                </div>
              </div>

              <div className="rounded-lg border p-4">
                <p className="text-xs uppercase tracking-wide text-muted-foreground">Adjuntos</p>
                <div className="mt-3 space-y-3">
                  {selectedBroadcast.adjuntos?.length ? (
                    selectedBroadcast.adjuntos.map((att) => (
                      <div key={`${att.sha256_hash}-${att.nombre}`} className="rounded-md border bg-muted/30 p-3">
                        <div className="flex items-center gap-2">
                          <FileText className="h-4 w-4 text-muted-foreground" />
                          <p className="font-medium">{att.nombre}</p>
                        </div>
                        <p className="mt-1 text-sm text-muted-foreground">{att.tamano_bytes.toLocaleString()} bytes</p>
                        <p className="mt-1 break-all text-xs text-muted-foreground">{att.sha256_hash}</p>
                      </div>
                    ))
                  ) : (
                    <p className="text-sm text-muted-foreground">Esta difusión no incluye adjuntos.</p>
                  )}
                </div>
              </div>

              <div className="rounded-lg border p-4">
                <p className="text-xs uppercase tracking-wide text-muted-foreground">Fecha</p>
                <p className="mt-1 text-sm">{formatDate(selectedBroadcast.created_at)}</p>
              </div>
            </div>
          ) : null}
        </SheetContent>
      </Sheet>
    </div>
  )
}

"use client"

import { useEffect, useState } from "react"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetDescription } from "@/components/ui/sheet"
import { Send, Check, X, ExternalLink } from "lucide-react"
import { getAdminBroadcasts, getCompanies, type BroadcastInfo, type Company } from "@/lib/api"

export default function BroadcastsPage() {
  const [broadcasts, setBroadcasts] = useState<BroadcastInfo[]>([])
  const [companies, setCompanies] = useState<Company[]>([])
  const [loading, setLoading] = useState(true)
  const [filterRuc, setFilterRuc] = useState("")
  const [selectedBroadcast, setSelectedBroadcast] = useState<BroadcastInfo | null>(null)
  const [detailsOpen, setDetailsOpen] = useState(false)

  async function loadData() {
    setLoading(true)
    try {
      const [bcData, compsData] = await Promise.all([
        getAdminBroadcasts(filterRuc || undefined),
        getCompanies(),
      ])
      setBroadcasts(bcData.broadcasts)
      setCompanies(compsData.companies)
    } catch (error) {
      console.error("Failed to load data:", error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [filterRuc])

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "completed":
        return <Badge variant="default" className="bg-green-500">Completado</Badge>
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

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Broadcasts</h1>
        <p className="text-muted-foreground">
          Historial de difusiones masivas
        </p>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Lista de Broadcasts</CardTitle>
              <CardDescription>
                {broadcasts.length} broadcast(s) encontrado(s)
              </CardDescription>
            </div>
            <select
              className="h-10 rounded-md border border-input bg-background px-3 py-2 text-sm"
              value={filterRuc}
              onChange={(e) => setFilterRuc(e.target.value)}
            >
              <option value="">Todas las empresas</option>
              {companies.map((c) => (
                <option key={c.account_id} value={c.account_id}>
                  {c.account_id}
                </option>
              ))}
            </select>
          </div>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Reference ID</TableHead>
                <TableHead>Empresa</TableHead>
                <TableHead>Total</TableHead>
                <TableHead>Exitosos</TableHead>
                <TableHead>Fallidos</TableHead>
                <TableHead>Estado</TableHead>
                <TableHead>Fecha</TableHead>
                <TableHead className="text-right">Acciones</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={8} className="text-center text-muted-foreground">
                    Cargando...
                  </TableCell>
                </TableRow>
              ) : broadcasts.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={8} className="text-center text-muted-foreground">
                    No hay broadcasts
                  </TableCell>
                </TableRow>
              ) : (
                broadcasts.map((bc) => (
                  <TableRow key={bc.reference_id}>
                    <TableCell className="font-mono text-sm">{bc.reference_id.slice(0, 8)}...</TableCell>
                    <TableCell>{bc.ruc_empresa}</TableCell>
                    <TableCell>{bc.total}</TableCell>
                    <TableCell>
                      <span className="text-green-600">{bc.success}</span>
                    </TableCell>
                    <TableCell>
                      <span className="text-red-600">{bc.failed}</span>
                    </TableCell>
                    <TableCell>{getStatusBadge(bc.status)}</TableCell>
                    <TableCell className="text-muted-foreground">
                      {bc.created_at ? new Date(bc.created_at).toLocaleString() : "-"}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button variant="ghost" size="icon" onClick={() => {
                        setSelectedBroadcast(bc)
                        setDetailsOpen(true)
                      }}>
                        <ExternalLink className="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Sheet open={detailsOpen} onOpenChange={setDetailsOpen}>
        <SheetContent side="right">
          <SheetHeader>
            <SheetTitle>Detalles del Broadcast</SheetTitle>
            <SheetDescription>
              Reference: {selectedBroadcast?.reference_id}
            </SheetDescription>
          </SheetHeader>
          <div className="mt-4 space-y-4">
            <div className="grid grid-cols-2 gap-4">
              <div>
                <p className="text-sm text-muted-foreground">Empresa</p>
                <p className="font-medium">{selectedBroadcast?.ruc_empresa}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Estado</p>
                {selectedBroadcast && getStatusBadge(selectedBroadcast.status)}
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Total</p>
                <p className="font-medium">{selectedBroadcast?.total}</p>
              </div>
              <div>
                <p className="text-sm text-muted-foreground">Fecha</p>
                <p className="font-medium">
                  {selectedBroadcast?.created_at 
                    ? new Date(selectedBroadcast.created_at).toLocaleString() 
                    : "-"}
                </p>
              </div>
            </div>
            <div className="border-t pt-4">
              <p className="text-sm font-medium mb-2">Resultados</p>
              <div className="flex gap-4">
                <div className="flex items-center gap-2">
                  <span className="text-green-600 font-bold">{selectedBroadcast?.success}</span>
                  <span className="text-sm text-muted-foreground">exitosos</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-red-600 font-bold">{selectedBroadcast?.failed}</span>
                  <span className="text-sm text-muted-foreground">fallidos</span>
                </div>
              </div>
            </div>
          </div>
        </SheetContent>
      </Sheet>
    </div>
  )
}
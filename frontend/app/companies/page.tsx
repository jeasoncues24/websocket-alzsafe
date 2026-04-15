"use client"

import { useEffect, useState } from "react"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Search, Building2, Circle, MoreVertical } from "lucide-react"
import { getCompanies, type Company } from "@/lib/api"

export default function CompaniesPage() {
  const [companies, setCompanies] = useState<Company[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState("")
  const [filtered, setFiltered] = useState<Company[]>([])

  async function loadCompanies() {
    setLoading(true)
    try {
      const data = await getCompanies()
      setCompanies(data.companies)
      setFiltered(data.companies)
    } catch (error) {
      console.error("Failed to load companies:", error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadCompanies()
  }, [])

  useEffect(() => {
    const timer = setTimeout(() => {
      if (search.trim() === "") {
        setFiltered(companies)
      } else {
        const q = search.toLowerCase()
        setFiltered(companies.filter(c => c.account_id.toLowerCase().includes(q)))
      }
    }, 300)
    return () => clearTimeout(timer)
  }, [search, companies])

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "active":
        return <Badge variant="default" className="bg-green-500">Activa</Badge>
      case "qr_pending":
        return <Badge variant="secondary">QR Pendiente</Badge>
      case "initializing":
        return <Badge variant="secondary">Inicializando</Badge>
      case "disconnected":
        return <Badge variant="destructive">Desconectada</Badge>
      default:
        return <Badge variant="outline">{status}</Badge>
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Empresas</h1>
        <p className="text-muted-foreground">
          Gestiona las conexiones de WhatsApp por empresa
        </p>
      </div>

      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>Lista de Empresas</CardTitle>
              <CardDescription>
                {filtered.length} empresa(s) encontrada(s)
              </CardDescription>
            </div>
            <div className="relative w-64">
              <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Buscar por RUC..."
                className="pl-8"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
              />
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>RUC</TableHead>
                <TableHead>Estado</TableHead>
                <TableHead>Última Actualización</TableHead>
                <TableHead className="text-right">Acciones</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={4} className="text-center text-muted-foreground">
                    Cargando...
                  </TableCell>
                </TableRow>
              ) : filtered.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={4} className="text-center text-muted-foreground">
                    No hay empresas registradas
                  </TableCell>
                </TableRow>
              ) : (
                filtered.map((company) => (
                  <TableRow key={company.account_id}>
                    <TableCell className="font-medium">
                      <div className="flex items-center gap-2">
                        <Building2 className="h-4 w-4" />
                        {company.account_id}
                      </div>
                    </TableCell>
                    <TableCell>{getStatusBadge(company.status)}</TableCell>
                    <TableCell className="text-muted-foreground">
                      {company.updated_at ? new Date(company.updated_at).toLocaleString() : "-"}
                    </TableCell>
                    <TableCell className="text-right">
                      <Button variant="ghost" size="icon">
                        <MoreVertical className="h-4 w-4" />
                      </Button>
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
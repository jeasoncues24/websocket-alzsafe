"use client"

import { useEffect, useState } from "react"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs"
import { Skeleton } from "@/components/ui/skeleton"
import { getAdminMessages, getCompanies, type AdminMessage, type Company } from "@/lib/api"

export default function MessagesPage() {
  const [messages, setMessages] = useState<AdminMessage[]>([])
  const [companies, setCompanies] = useState<Company[]>([])
  const [loading, setLoading] = useState(true)
  const [activeTab, setActiveTab] = useState("all")
  const [filterAccount, setFilterAccount] = useState("")
  const [total, setTotal] = useState(0)

  async function loadData() {
    setLoading(true)
    try {
      const statusFilter = activeTab === "all" ? "" : activeTab
      const [msgsData, compsData] = await Promise.all([
        getAdminMessages({ status: statusFilter, account_id: filterAccount, limit: 50 }),
        getCompanies(),
      ])
      setMessages(msgsData.messages)
      setTotal(msgsData.total)
      setCompanies(compsData.companies)
    } catch (error) {
      console.error("Failed to load data:", error)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadData()
  }, [activeTab, filterAccount])

  const getStatusBadge = (status: string) => {
    switch (status) {
      case "sent":
        return <Badge variant="default" className="bg-green-500">Enviado</Badge>
      case "delivered":
        return <Badge variant="default" className="bg-blue-500">Entregado</Badge>
      case "failed":
        return <Badge variant="destructive">Fallido</Badge>
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
                  <option key={c.account_id} value={c.account_id}>
                    {c.account_id}
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
          <Table className="mt-4">
            <TableHeader>
              <TableRow>
                <TableHead>ID</TableHead>
                <TableHead>Empresa</TableHead>
                <TableHead>Destino</TableHead>
                <TableHead>Contenido</TableHead>
                <TableHead>Estado</TableHead>
                <TableHead>Fecha</TableHead>
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
                  <TableCell colSpan={6} className="text-center text-muted-foreground">
                    No hay mensajes
                  </TableCell>
                </TableRow>
              ) : (
                messages.map((msg) => (
                  <TableRow key={msg.id}>
                    <TableCell className="font-medium">#{msg.id}</TableCell>
                    <TableCell>{msg.account_id}</TableCell>
                    <TableCell>{msg.to}</TableCell>
                    <TableCell className="max-w-xs truncate">{msg.content}</TableCell>
                    <TableCell>{getStatusBadge(msg.status)}</TableCell>
                    <TableCell className="text-muted-foreground">
                      {msg.created_at ? new Date(msg.created_at).toLocaleString() : "-"}
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
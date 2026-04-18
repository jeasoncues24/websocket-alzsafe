'use client'

import { useCallback, useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'

interface Phone {
  id: number
  numeroCompleto: string
  status: string
  lastConnected: string
}

interface Metrics {
  messages_sent: number
  messages_failed: number
  success_rate: number
  active_phones: number
}

export default function EmpresaDashboardPage() {
  const router = useRouter()
  const [phones, setPhones] = useState<Phone[]>([])
  const [metrics, setMetrics] = useState<Metrics | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const fetchData = useCallback(async (token: string) => {
    try {
      const headers = { 'Authorization': `Bearer ${token}` }

      const [phonesRes, metricsRes] = await Promise.all([
        fetch('/api/telefonos', { headers }),
        fetch('/api/metricas', { headers }),
      ])

      const phonesData = await phonesRes.json()
      const metricsData = await metricsRes.json()

      if (phonesData.ok) {
        setPhones(phonesData.data.sesiones || phonesData.data.sessions || [])
      }
      if (metricsData.ok) {
        setMetrics(metricsData.data)
      }
    } catch {
      setError('Error cargando datos')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    const token = localStorage.getItem('empresa_token')
    if (!token) {
      router.push('/empresa/login')
      return
    }
    fetchData(token)
  }, [fetchData, router])

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'bg-green-500'
      case 'qr_pending': return 'bg-yellow-500'
      case 'disconnected': return 'bg-red-500'
      default: return 'bg-gray-500'
    }
  }

  if (loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <p>Cargando...</p>
      </div>
    )
  }

  return (
    <div className="container mx-auto p-6">
      <h1 className="text-2xl font-bold mb-6">Dashboard Empresa</h1>

      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Mensajes Enviados</CardDescription>
          </CardHeader>
          <CardContent>
            <CardTitle className="text-3xl">{metrics?.messages_sent || 0}</CardTitle>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Éxito</CardDescription>
          </CardHeader>
          <CardContent>
            <CardTitle className="text-3xl">{metrics?.success_rate?.toFixed(1) || 0}%</CardTitle>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Teléfonos Activos</CardDescription>
          </CardHeader>
          <CardContent>
            <CardTitle className="text-3xl">{metrics?.active_phones || 0}</CardTitle>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardDescription>Total Teléfonos</CardDescription>
          </CardHeader>
          <CardContent>
            <CardTitle className="text-3xl">{phones.length}</CardTitle>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Mis Teléfonos</CardTitle>
          <CardDescription>Lista de números conectados</CardDescription>
        </CardHeader>
        <CardContent>
          {phones.length === 0 ? (
            <p className="text-gray-500">No hay teléfonos registrados</p>
          ) : (
            <div className="space-y-2">
              {phones.map((phone) => (
                <div key={phone.id} className="flex items-center justify-between p-3 border rounded">
                  <div>
                    <p className="font-medium">{phone.numeroCompleto}</p>
                    <p className="text-sm text-gray-500">
                     Última conexión: {phone.lastConnected || 'N/A'}
                    </p>
                  </div>
                  <Badge className={getStatusColor(phone.status)}>
                    {phone.status}
                  </Badge>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

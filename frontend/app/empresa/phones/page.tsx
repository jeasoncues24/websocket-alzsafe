'use client'

import { useCallback, useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'

interface Phone {
  id: number
  numeroCompleto: string
  status: string
}

export default function PhonesPage() {
  const router = useRouter()
  const [phones, setPhones] = useState<Phone[]>([])
  const [loading, setLoading] = useState(true)

  const fetchPhones = useCallback(async (token: string) => {
    try {
      const res = await fetch('/api/telefonos', {
        headers: { 'Authorization': `Bearer ${token}` },
      })
      const data = await res.json()
      if (data.ok) {
        setPhones(data.data.sesiones || data.data.sessions || [])
      }
    } catch {
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
    fetchPhones(token)
  }, [fetchPhones, router])

  const disconnectPhone = async (id: number) => {
    const token = localStorage.getItem('empresa_token')
    if (!token) return

    if (!confirm('¿Desconectar este teléfono?')) return

    try {
      const res = await fetch(`/api/telefonos/${id}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${token}` },
      })
      const data = await res.json()
      if (data.ok) {
        fetchPhones(token)
      }
    } catch {
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active': return 'bg-green-500'
      case 'qr_pending': return 'bg-yellow-500'
      case 'disconnected': return 'bg-red-500'
      default: return 'bg-gray-500'
    }
  }

  return (
    <div className="container mx-auto p-6">
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Mis Teléfonos</h1>
        <Button onClick={() => router.push('/empresa/phones/new')}>
          Agregar Teléfono
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Lista de Teléfonos</CardTitle>
          <CardDescription>Gestiona tus números de WhatsApp</CardDescription>
        </CardHeader>
        <CardContent>
          {loading ? (
            <p>Cargando...</p>
          ) : phones.length === 0 ? (
            <p className="text-gray-500">No hay teléfonos</p>
          ) : (
            <div className="space-y-3">
              {phones.map((phone) => (
                <div key={phone.id} className="flex items-center justify-between p-4 border rounded">
                  <div>
                    <p className="font-medium">{phone.numeroCompleto}</p>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge className={getStatusColor(phone.status)}>
                      {phone.status}
                    </Badge>
                      {phone.status === 'qr_pending' && (
                        <Button size="sm" onClick={() => router.push(`/empresa/phones/qr?id=${phone.id}`)}>
                          Ver QR
                        </Button>
                      )}
                    {phone.status === 'active' && (
                      <Button size="sm" variant="destructive" onClick={() => disconnectPhone(phone.id)}>
                        Desconectar
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

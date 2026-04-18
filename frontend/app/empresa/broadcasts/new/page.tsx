'use client'

import { useCallback, useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

export default function BroadcastPage() {
  const router = useRouter()
  const [phones, setPhones] = useState<{id: number, numeroCompleto: string}[]>([])
  const [telefonoId, setTelefonoId] = useState('')
  const [destinos, setDestinos] = useState('')
  const [mensaje, setMensaje] = useState('')
  const [sending, setSending] = useState(false)
  const [success, setSuccess] = useState('')
  const [error, setError] = useState('')

  const fetchPhones = useCallback(async (token: string) => {
    try {
      const res = await fetch('/api/telefonos', {
        headers: { 'Authorization': `Bearer ${token}` },
      })
      const data = await res.json()
      if (data.ok) {
        const active = (data.data.sesiones || data.data.sessions || []).filter((p: { status: string }) => p.status === 'active')
        setPhones(active)
      }
    } catch {
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

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault()
    const token = localStorage.getItem('empresa_token')
    if (!token) return

    const lista = destinos.split(/[\n,]/).map(d => d.trim()).filter(d => d)
    if (lista.length === 0) {
      setError('Ingresa destinos')
      return
    }

    setSending(true)
    setError('')
    setSuccess('')

    try {
      const res = await fetch('/api/difusiones', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          telefono_id: Number(telefonoId),
          destinos: lista,
          mensaje: mensaje,
        }),
      })
      const data = await res.json()
      if (data.ok) {
        setSuccess(`Difusión creada: ${data.data.reference_id}`)
        setDestinos('')
        setMensaje('')
      } else {
        setError(data.message || 'Error')
      }
    } catch {
      setError('Error de conexión')
    } finally {
      setSending(false)
    }
  }

  return (
    <div className="container mx-auto p-6">
      <Card>
        <CardHeader>
          <CardTitle>Crear Difusión</CardTitle>
          <CardDescription>Envía mensaje a múltiples destinos</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSend} className="space-y-4">
            <div className="space-y-2">
              <Label>Desde Teléfono</Label>
              <select
                className="w-full p-2 border rounded"
                value={telefonoId}
                onChange={(e) => setTelefonoId(e.target.value)}
                required
              >
                <option value="">Seleccionar...</option>
                {phones.map((p) => (
                  <option key={p.id} value={p.id}>{p.numeroCompleto}</option>
                ))}
              </select>
            </div>

            <div className="space-y-2">
              <Label>Destinos (uno por línea o separados por coma)</Label>
              <Textarea
                placeholder="+51 999 999 999&#10;+51 888 888 888"
                value={destinos}
                onChange={(e) => setDestinos(e.target.value)}
                required
              />
            </div>

            <div className="space-y-2">
              <Label>Mensaje</Label>
              <Textarea
                placeholder="Tu mensaje..."
                value={mensaje}
                onChange={(e) => setMensaje(e.target.value)}
                required
              />
            </div>

            {error && <div className="text-red-500 text-sm">{error}</div>}
            {success && <div className="text-green-500 text-sm">{success}</div>}

            <Button type="submit" className="w-full" disabled={sending}>
              {sending ? 'Creando...' : 'Crear Difusión'}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}

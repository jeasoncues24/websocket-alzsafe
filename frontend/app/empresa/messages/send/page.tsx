'use client'

import { useCallback, useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'

export default function SendMessagePage() {
  const router = useRouter()
  const [phones, setPhones] = useState<{id: number, numeroCompleto: string}[]>([])
  const [telefonoId, setTelefonoId] = useState('')
  const [destino, setDestino] = useState('')
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
        setPhones(data.data.sesiones || data.data.sessions || [])
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

    setSending(true)
    setError('')
    setSuccess('')

    try {
      const res = await fetch('/api/mensajes', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          telefono_id: Number(telefonoId),
          destino: destino,
          contenido: mensaje,
        }),
      })
      const data = await res.json()
      if (data.ok) {
        setSuccess('Mensaje enviado')
        setDestino('')
        setMensaje('')
      } else {
        setError(data.message || 'Error enviando mensaje')
      }
    } catch {
      setError('Error de conexión')
    } finally {
      setSending(false)
    }
  }

  return (
    <div className="container mx-auto p-6 max-w-lg">
      <Card>
        <CardHeader>
          <CardTitle>Enviar Mensaje</CardTitle>
          <CardDescription>Envía un mensaje de WhatsApp</CardDescription>
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
                  <option key={p.id} value={p.id}>
                    {p.numeroCompleto}
                  </option>
                ))}
              </select>
            </div>

            <div className="space-y-2">
              <Label>Destino</Label>
              <Input
                placeholder="+51 999 999 999"
                value={destino}
                onChange={(e) => setDestino(e.target.value)}
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

            {error && (
              <div className="text-red-500 text-sm">{error}</div>
            )}
            {success && (
              <div className="text-green-500 text-sm">{success}</div>
            )}

            <Button type="submit" className="w-full" disabled={sending}>
              {sending ? 'Enviando...' : 'Enviar'}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}

'use client'

import { useCallback, useEffect, useState } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription } from '@/components/ui/alert'
import Image from 'next/image'

export default function PhoneQRPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [phoneId, setPhoneId] = useState('')
  const [qrString, setQrString] = useState('')
  const [status, setStatus] = useState('qr_pending')
  const [countdown, setCountdown] = useState(300)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const fetchQR = useCallback(async (token: string, id: string) => {
    setLoading(true)
    try {
      const res = await fetch(`/api/telefonos/${id}`, {
        headers: { 'Authorization': `Bearer ${token}` },
      })
      const data = await res.json()
      if (data.ok) {
        setQrString(data.data.qr_string || '')
        setStatus(data.data.status)
      } else {
        setError(data.message || 'Error obteniendo QR')
      }
    } catch {
      setError('Error de conexión')
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

    const id = searchParams.get('id')
    if (id) {
      setPhoneId(id)
      fetchQR(token, id)
    }
  }, [fetchQR, router, searchParams])

  useEffect(() => {
    if (countdown > 0 && qrString) {
      const timer = setTimeout(() => setCountdown(c => c - 1), 1000)
      return () => clearTimeout(timer)
    }
  }, [countdown, qrString])

  const formatTime = (seconds: number) => {
    const m = Math.floor(seconds / 60)
    const s = seconds % 60
    return `${m}:${s.toString().padStart(2, '0')}`
  }

  if (!phoneId) {
    return <div className="p-6">ID de teléfono requerido</div>
  }

  return (
    <div className="container mx-auto p-6 max-w-md">
      <Card>
        <CardHeader>
          <CardTitle>Conectar WhatsApp</CardTitle>
          <CardDescription>
            Escanea el código QR con tu teléfono
          </CardDescription>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {qrString ? (
            <div className="text-center">
              <div className="bg-white p-4 inline-block border rounded mb-4">
                <Image src={`https://api.qrserver.com/create/?data=${encodeURIComponent(qrString)}`} alt="QR Code" width={200} height={200} />
              </div>
              <p className="text-sm text-gray-500 mb-4">
                Código QR válido por {formatTime(countdown)}
              </p>
              <p className="text-xs text-gray-400 mb-4">Estado: {status}</p>
            </div>
          ) : (
            <p className="text-center text-gray-500">
              {loading ? 'Generando QR...' : 'No hay QR disponible'}
            </p>
          )}

          <Button
            onClick={() => {
              const token = localStorage.getItem('empresa_token')
              if (token && phoneId) fetchQR(token, phoneId)
            }}
            className="w-full mt-4"
            variant="outline"
          >
            Regenerar QR
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}

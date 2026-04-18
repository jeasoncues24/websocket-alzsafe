'use client'

import { useCallback, useEffect, useState } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Alert, AlertDescription } from '@/components/ui/alert'

export default function EmpresaLoginPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [token, setToken] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const validateToken = useCallback(async (jwt: string) => {
    setLoading(true)
    setError('')

    try {
      const res = await fetch('/api/auth/empresa/validate', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${jwt}`,
        },
      })

      const data = await res.json()

      if (data.ok) {
        localStorage.setItem('empresa_token', jwt)
        router.push('/empresa/dashboard')
      } else {
        setError(data.message || 'Credencial inválida')
      }
    } catch {
      setError('Error validando credencial')
    } finally {
      setLoading(false)
    }
  }, [router])

  useEffect(() => {
    const tokenParam = searchParams.get('token')
    if (tokenParam) {
      setToken(tokenParam)
      validateToken(tokenParam)
    }
  }, [searchParams, validateToken])

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (token) {
      validateToken(token)
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <Card className="w-full max-w-md">
        <CardHeader>
	          <CardTitle>Acceso Empresa</CardTitle>
	          <CardDescription>
	            Ingresa tu credencial de empresa para acceder
	          </CardDescription>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Input
	                placeholder="Credencial JWT"
                value={token}
                onChange={(e) => setToken(e.target.value)}
              />
            </div>
            <Button type="submit" className="w-full" disabled={loading}>
              {loading ? 'Validando...' : 'Ingresar'}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}

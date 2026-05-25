"use client"

import { useState, useEffect } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { useTheme } from "next-themes"
import { useAppStore } from "@/stores/useAppStore"
import { updateMeProfile, updateMePassword } from "@/lib/api"
import { 
  Moon, 
  Sun, 
  Monitor, 
  Info, 
  UserRound, 
  KeyRound, 
  Loader2, 
  CheckCircle2, 
  AlertCircle 
} from "lucide-react"

export default function SettingsPage() {
  const { theme, setTheme } = useTheme()
  const { refreshInterval, setRefreshInterval, itemsPerPage, setItemsPerPage, user, setUser } = useAppStore()

  // Profile Form States
  const [username, setUsername] = useState("")
  const [email, setEmail] = useState("")
  const [profileLoading, setProfileLoading] = useState(false)
  const [profileSuccess, setProfileSuccess] = useState("")
  const [profileError, setProfileError] = useState("")

  // Password Form States
  const [currentPassword, setCurrentPassword] = useState("")
  const [newPassword, setNewPassword] = useState("")
  const [confirmPassword, setConfirmPassword] = useState("")
  const [passwordLoading, setPasswordLoading] = useState(false)
  const [passwordSuccess, setPasswordSuccess] = useState("")
  const [passwordError, setPasswordError] = useState("")

  // Load User Data
  useEffect(() => {
    if (user) {
      setUsername(user.username || "")
      setEmail(user.email || "")
    }
  }, [user])

  // Profile Form Submit
  async function handleProfileSubmit(e: React.FormEvent) {
    e.preventDefault()
    setProfileSuccess("")
    setProfileError("")

    if (!username.trim() || !email.trim()) {
      setProfileError("Todos los campos son obligatorios")
      return
    }

    setProfileLoading(true)
    try {
      const res = await updateMeProfile(username.trim(), email.trim())
      if (res.ok) {
        setProfileSuccess("Perfil actualizado con éxito")
        if (user) {
          setUser({
            ...user,
            username: username.trim(),
            email: email.trim()
          })
        }
      } else {
        setProfileError("No se pudo actualizar el perfil")
      }
    } catch (err: unknown) {
      const errorMsg = err instanceof Error ? err.message : "El nombre de usuario o correo ya está en uso"
      setProfileError(errorMsg)
    } finally {
      setProfileLoading(false)
    }
  }

  // Password Form Submit
  async function handlePasswordSubmit(e: React.FormEvent) {
    e.preventDefault()
    setPasswordSuccess("")
    setPasswordError("")

    if (!currentPassword || !newPassword || !confirmPassword) {
      setPasswordError("Todos los campos son obligatorios")
      return
    }

    if (newPassword !== confirmPassword) {
      setPasswordError("La nueva contraseña y la confirmación no coinciden")
      return
    }

    if (newPassword.length < 6) {
      setPasswordError("La nueva contraseña debe tener al menos 6 caracteres")
      return
    }

    setPasswordLoading(true)
    try {
      const res = await updateMePassword(currentPassword, newPassword)
      if (res.ok) {
        setPasswordSuccess("Contraseña actualizada con éxito")
        setCurrentPassword("")
        setNewPassword("")
        setConfirmPassword("")
      } else {
        setPasswordError("No se pudo actualizar la contraseña")
      }
    } catch (err: unknown) {
      const errorMsg = err instanceof Error ? err.message : "La contraseña actual es incorrecta"
      setPasswordError(errorMsg)
    } finally {
      setPasswordLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Configuración</h1>
        <p className="text-muted-foreground">
          Personaliza tu experiencia y gestiona tu seguridad
        </p>
      </div>

      <Tabs defaultValue="account">
        <TabsList>
          <TabsTrigger value="account">Mi cuenta</TabsTrigger>
          <TabsTrigger value="appearance">Apariencia</TabsTrigger>
          <TabsTrigger value="general">General</TabsTrigger>
          <TabsTrigger value="about">Acerca de</TabsTrigger>
        </TabsList>

        <TabsContent value="account" className="mt-4">
          <div className="grid gap-6 md:grid-cols-2">
            
            {/* PROFILE CARD */}
            <Card className="shadow-sm">
              <CardHeader>
                <div className="flex items-center gap-2">
                  <UserRound className="h-5 w-5 text-primary" />
                  <CardTitle>Datos de Perfil</CardTitle>
                </div>
                <CardDescription>
                  Actualiza tu nombre de usuario y dirección de correo electrónico
                </CardDescription>
              </CardHeader>
              <CardContent>
                <form onSubmit={handleProfileSubmit} className="space-y-4">
                  {profileSuccess && (
                    <div className="flex items-center gap-2 rounded-lg bg-emerald-500/15 p-3 text-sm text-emerald-500 border border-emerald-500/30 motion-enter-up">
                      <CheckCircle2 className="h-4 w-4 flex-shrink-0" />
                      <span>{profileSuccess}</span>
                    </div>
                  )}

                  {profileError && (
                    <div className="flex items-center gap-2 rounded-lg bg-destructive/15 p-3 text-sm text-destructive border border-destructive/30 motion-enter-up">
                      <AlertCircle className="h-4 w-4 flex-shrink-0" />
                      <span>{profileError}</span>
                    </div>
                  )}

                  <div className="space-y-1">
                    <Label htmlFor="username">Nombre de usuario</Label>
                    <Input
                      id="username"
                      value={username}
                      onChange={(e) => setUsername(e.target.value)}
                      placeholder="Ingrese su usuario"
                      disabled={profileLoading}
                    />
                  </div>

                  <div className="space-y-1">
                    <Label htmlFor="email">Correo electrónico</Label>
                    <Input
                      id="email"
                      type="email"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      placeholder="Ingrese su correo"
                      disabled={profileLoading}
                    />
                  </div>

                  <Button type="submit" disabled={profileLoading} className="w-full">
                    {profileLoading ? (
                      <>
                        <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                        Guardando...
                      </>
                    ) : (
                      "Guardar Cambios"
                    )}
                  </Button>
                </form>
              </CardContent>
            </Card>

            {/* PASSWORD CARD */}
            <Card className="shadow-sm">
              <CardHeader>
                <div className="flex items-center gap-2">
                  <KeyRound className="h-5 w-5 text-primary" />
                  <CardTitle>Cambiar Contraseña</CardTitle>
                </div>
                <CardDescription>
                  Protege tu cuenta actualizando tu contraseña periódicamente
                </CardDescription>
              </CardHeader>
              <CardContent>
                <form onSubmit={handlePasswordSubmit} className="space-y-4">
                  {passwordSuccess && (
                    <div className="flex items-center gap-2 rounded-lg bg-emerald-500/15 p-3 text-sm text-emerald-500 border border-emerald-500/30 motion-enter-up">
                      <CheckCircle2 className="h-4 w-4 flex-shrink-0" />
                      <span>{passwordSuccess}</span>
                    </div>
                  )}

                  {passwordError && (
                    <div className="flex items-center gap-2 rounded-lg bg-destructive/15 p-3 text-sm text-destructive border border-destructive/30 motion-enter-up">
                      <AlertCircle className="h-4 w-4 flex-shrink-0" />
                      <span>{passwordError}</span>
                    </div>
                  )}

                  <div className="space-y-1">
                    <Label htmlFor="currentPassword">Contraseña actual</Label>
                    <Input
                      id="currentPassword"
                      type="password"
                      value={currentPassword}
                      onChange={(e) => setCurrentPassword(e.target.value)}
                      placeholder="••••••••"
                      disabled={passwordLoading}
                    />
                  </div>

                  <div className="space-y-1">
                    <Label htmlFor="newPassword">Nueva contraseña</Label>
                    <Input
                      id="newPassword"
                      type="password"
                      value={newPassword}
                      onChange={(e) => setNewPassword(e.target.value)}
                      placeholder="••••••••"
                      disabled={passwordLoading}
                    />
                  </div>

                  <div className="space-y-1">
                    <Label htmlFor="confirmPassword">Confirmar nueva contraseña</Label>
                    <Input
                      id="confirmPassword"
                      type="password"
                      value={confirmPassword}
                      onChange={(e) => setConfirmPassword(e.target.value)}
                      placeholder="••••••••"
                      disabled={passwordLoading}
                    />
                  </div>

                  <Button type="submit" disabled={passwordLoading} className="w-full">
                    {passwordLoading ? (
                      <>
                        <Loader2 className="h-4 w-4 mr-2 animate-spin" />
                        Actualizando...
                      </>
                    ) : (
                      "Actualizar Contraseña"
                    )}
                  </Button>
                </form>
              </CardContent>
            </Card>

          </div>
        </TabsContent>

        <TabsContent value="appearance" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle>Tema</CardTitle>
              <CardDescription>
                Selecciona el tema de color para el panel
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex gap-2">
                <Button
                  variant={theme === "light" ? "default" : "outline"}
                  onClick={() => setTheme("light")}
                >
                  <Sun className="h-4 w-4 mr-2" />
                  Claro
                </Button>
                <Button
                  variant={theme === "dark" ? "default" : "outline"}
                  onClick={() => setTheme("dark")}
                >
                  <Moon className="h-4 w-4 mr-2" />
                  Oscuro
                </Button>
                <Button
                  variant={theme === "system" ? "default" : "outline"}
                  onClick={() => setTheme("system")}
                >
                  <Monitor className="h-4 w-4 mr-2" />
                  Sistema
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="general" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle>General</CardTitle>
              <CardDescription>
                Configura opciones generales del panel
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <label className="text-sm font-medium">Intervalo de auto-actualización</label>
                <select
                  className="mt-1 h-10 rounded-md border border-input bg-background px-3 py-2 text-sm w-full animate-transition focus:ring-2 focus:ring-primary/20"
                  value={refreshInterval}
                  onChange={(e) => setRefreshInterval(Number(e.target.value))}
                >
                  <option value={15000}>15 segundos</option>
                  <option value={30000}>30 segundos</option>
                  <option value={60000}>1 minuto</option>
                  <option value={300000}>5 minutos</option>
                </select>
              </div>

              <div>
                <label className="text-sm font-medium">Elementos por página</label>
                <select
                  className="mt-1 h-10 rounded-md border border-input bg-background px-3 py-2 text-sm w-full animate-transition focus:ring-2 focus:ring-primary/20"
                  value={itemsPerPage}
                  onChange={(e) => setItemsPerPage(Number(e.target.value))}
                >
                  <option value={10}>10</option>
                  <option value={20}>20</option>
                  <option value={50}>50</option>
                  <option value={100}>100</option>
                </select>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="about" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle>Acerca de</CardTitle>
              <CardDescription>
                Información sobre WhatsApp API
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center gap-3">
                <Info className="h-5 w-5 text-muted-foreground" />
                <div>
                  <p className="font-medium">WhatsApp API</p>
                  <p className="text-sm text-muted-foreground">Versión 1.0.0</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}

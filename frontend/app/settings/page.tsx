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
  AlertCircle,
  Eye,
  EyeOff
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

  // Password Visibility
  const [showCurrentPassword, setShowCurrentPassword] = useState(false)
  const [showNewPassword, setShowNewPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)

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

  function getPasswordStrength(password: string): { level: number; label: string; barColor: string; textColor: string } {
    if (!password) return { level: 0, label: "", barColor: "", textColor: "" }
    let score = 0
    if (password.length >= 8) score++
    if (password.length >= 12) score++
    if (/[A-Z]/.test(password) && /[a-z]/.test(password)) score++
    if (/[0-9]/.test(password)) score++
    if (/[^A-Za-z0-9]/.test(password)) score++
    if (score <= 1) return { level: 1, label: "Débil",   barColor: "bg-red-500",     textColor: "text-red-500" }
    if (score === 2) return { level: 2, label: "Regular", barColor: "bg-amber-500",   textColor: "text-amber-500" }
    if (score === 3) return { level: 3, label: "Buena",   barColor: "bg-blue-500",    textColor: "text-blue-500" }
    return             { level: 4, label: "Fuerte",  barColor: "bg-emerald-500", textColor: "text-emerald-500" }
  }

  const passwordStrength = getPasswordStrength(newPassword)
  const passwordsMatch = confirmPassword.length > 0 && newPassword === confirmPassword
  const passwordsMismatch = confirmPassword.length > 0 && newPassword !== confirmPassword

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

                  {/* Contraseña actual */}
                  <div className="space-y-1">
                    <Label htmlFor="currentPassword">Contraseña actual</Label>
                    <div className="relative">
                      <Input
                        id="currentPassword"
                        type={showCurrentPassword ? "text" : "password"}
                        value={currentPassword}
                        onChange={(e) => setCurrentPassword(e.target.value)}
                        placeholder="••••••••"
                        disabled={passwordLoading}
                        className="pr-10"
                      />
                      <button
                        type="button"
                        onClick={() => setShowCurrentPassword((v) => !v)}
                        disabled={passwordLoading}
                        aria-label={showCurrentPassword ? "Ocultar contraseña" : "Mostrar contraseña"}
                        className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors cursor-pointer disabled:pointer-events-none"
                      >
                        {showCurrentPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                      </button>
                    </div>
                  </div>

                  {/* Nueva contraseña + fortaleza */}
                  <div className="space-y-1">
                    <Label htmlFor="newPassword">Nueva contraseña</Label>
                    <div className="relative">
                      <Input
                        id="newPassword"
                        type={showNewPassword ? "text" : "password"}
                        value={newPassword}
                        onChange={(e) => setNewPassword(e.target.value)}
                        placeholder="••••••••"
                        disabled={passwordLoading}
                        className="pr-10"
                      />
                      <button
                        type="button"
                        onClick={() => setShowNewPassword((v) => !v)}
                        disabled={passwordLoading}
                        aria-label={showNewPassword ? "Ocultar contraseña" : "Mostrar contraseña"}
                        className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors cursor-pointer disabled:pointer-events-none"
                      >
                        {showNewPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                      </button>
                    </div>
                    {/* Barra de fortaleza */}
                    {newPassword && (
                      <div className="space-y-1 pt-1">
                        <div className="flex gap-1">
                          {[1, 2, 3, 4].map((i) => (
                            <div
                              key={i}
                              className={`h-1.5 flex-1 rounded-full transition-colors duration-300 ${
                                passwordStrength.level >= i ? passwordStrength.barColor : "bg-muted"
                              }`}
                            />
                          ))}
                        </div>
                        <p className={`text-xs font-medium ${passwordStrength.textColor}`}>
                          Fortaleza: {passwordStrength.label}
                        </p>
                      </div>
                    )}
                  </div>

                  {/* Confirmar contraseña + indicador de coincidencia */}
                  <div className="space-y-1">
                    <Label htmlFor="confirmPassword">Confirmar nueva contraseña</Label>
                    <div className="relative">
                      <Input
                        id="confirmPassword"
                        type={showConfirmPassword ? "text" : "password"}
                        value={confirmPassword}
                        onChange={(e) => setConfirmPassword(e.target.value)}
                        placeholder="••••••••"
                        disabled={passwordLoading}
                        className={`pr-10 transition-colors ${
                          passwordsMismatch ? "border-red-500 focus-visible:ring-red-500/30" :
                          passwordsMatch   ? "border-emerald-500 focus-visible:ring-emerald-500/30" : ""
                        }`}
                      />
                      <button
                        type="button"
                        onClick={() => setShowConfirmPassword((v) => !v)}
                        disabled={passwordLoading}
                        aria-label={showConfirmPassword ? "Ocultar contraseña" : "Mostrar contraseña"}
                        className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors cursor-pointer disabled:pointer-events-none"
                      >
                        {showConfirmPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                      </button>
                    </div>
                    {/* Indicador de coincidencia en tiempo real */}
                    {passwordsMatch && (
                      <div className="flex items-center gap-1.5 text-xs text-emerald-500 motion-enter-up">
                        <CheckCircle2 className="h-3.5 w-3.5" />
                        Las contraseñas coinciden
                      </div>
                    )}
                    {passwordsMismatch && (
                      <div className="flex items-center gap-1.5 text-xs text-red-500 motion-enter-up">
                        <AlertCircle className="h-3.5 w-3.5" />
                        Las contraseñas no coinciden
                      </div>
                    )}
                  </div>

                  <Button
                    type="submit"
                    disabled={passwordLoading || passwordsMismatch}
                    className="w-full"
                  >
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

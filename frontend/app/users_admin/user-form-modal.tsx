"use client"

import { useEffect, useState } from "react"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { cn } from "@/lib/utils"
import {
  User,
  Mail,
  Lock,
  Shield,
  ShieldCheck,
  Layers,
  Check,
  Loader2,
  Calendar,
  AlertCircle,
  CheckCircle2,
  Eye,
  EyeOff,
} from "lucide-react"
import {
  type UserAdminRol,
  type Role,
  type Module,
  type CreateUserRequest,
  type UpdateUserRequest,
} from "@/lib/api"

interface UserFormModalProps {
  open: boolean
  onClose: () => void
  onSave: (
    data: CreateUserRequest | UpdateUserRequest,
    modules: number[],
  ) => Promise<void>
  user: UserAdminRol | null
  roles: Role[]
  modules: Module[]
  userModules?: number[]
}

function formatDate(value?: string) {
  if (!value) return "—"
  try {
    return new Date(value).toLocaleDateString("es-PE", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
    })
  } catch {
    return value
  }
}

function getInitials(name: string): string {
  if (!name) return "U"
  const parts = name.trim().split(/[._\s-]+/)
  if (parts.length >= 2) {
    return `${parts[0][0]}${parts[1][0]}`.toUpperCase()
  }
  return name.slice(0, 2).toUpperCase()
}

export function UserFormModal({
  open,
  onClose,
  onSave,
  user,
  roles,
  modules,
  userModules = [],
}: UserFormModalProps) {
  const [loading, setLoading] = useState(false)
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [showPassword, setShowPassword] = useState(false)
  const [email, setEmail] = useState("")
  const [roleId, setRoleId] = useState<number | undefined>()
  const [isActive, setIsActive] = useState<boolean>(true)
  const [selectedModules, setSelectedModules] = useState<number[]>([])
  const [error, setError] = useState("")

  useEffect(() => {
    if (open) {
      if (user) {
        setUsername(user.username)
        setPassword("")
        setShowPassword(false)
        setEmail(user.email || "")
        setRoleId(user.role_id)
        setIsActive(user.activo ?? true)
        setSelectedModules(userModules)
      } else {
        setUsername("")
        setPassword("")
        setShowPassword(false)
        setEmail("")
        setRoleId(roles[0]?.id)
        setIsActive(true)
        setSelectedModules([])
      }
      setError("")
    }
  }, [open, user, roles, userModules])

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError("")
    setLoading(true)

    try {
      if (!user && !password) {
        setError("La contraseña es requerida para nuevos usuarios")
        return
      }

      if (!roleId) {
        setError("Debes seleccionar un rol para el usuario")
        return
      }

      const data = user
        ? { email, role_id: roleId, is_active: isActive }
        : { username, password, email, role_id: roleId }

      await onSave(
        data as CreateUserRequest | UpdateUserRequest,
        selectedModules,
      )
      onClose()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Error al guardar los cambios")
    } finally {
      setLoading(false)
    }
  }

  function toggleModule(moduleId: number) {
    setSelectedModules((prev) =>
      prev.includes(moduleId)
        ? prev.filter((id) => id !== moduleId)
        : [...prev, moduleId],
    )
  }

  function handleSelectAllModules() {
    setSelectedModules(modules.map((m) => m.id))
  }

  function handleClearAllModules() {
    setSelectedModules([])
  }

  const selectedRole = roles.find((r) => r.id === roleId)

  if (!open) return null

  return (
    <Dialog open={open} onOpenChange={(next) => !next && onClose()}>
      <DialogContent className="flex max-h-[88vh] max-w-2xl flex-col gap-0 overflow-hidden p-0 border bg-background shadow-2xl sm:rounded-2xl">
        {/* Cabecera Fija del Modal */}
        <DialogHeader className="shrink-0 border-b border-border bg-muted/30 px-6 py-4 pr-14">
          {user ? (
            <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
              <div className="flex items-center gap-3">
                <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-primary font-semibold text-sm text-primary-foreground shadow-sm">
                  {getInitials(user.username)}
                </div>
                <div>
                  <DialogTitle className="text-lg font-bold text-foreground">
                    {user.username}
                  </DialogTitle>
                  <div className="flex flex-wrap items-center gap-2 pt-0.5 text-xs text-muted-foreground">
                    <span className="font-mono font-medium text-foreground/80">
                      ID #{user.id}
                    </span>
                    <span>·</span>
                    <span className="inline-flex items-center gap-1">
                      <Calendar className="h-3 w-3" />
                      Registrado el {formatDate(user.created_at)}
                    </span>
                  </div>
                </div>
              </div>

              <div className="flex items-center gap-2">
                <Badge variant="outline" className="gap-1 font-medium text-xs">
                  <Shield className="h-3 w-3 text-muted-foreground" />
                  {user.rol}
                </Badge>
                <Badge
                  variant={isActive ? "outline" : "secondary"}
                  className={
                    isActive
                      ? "border-emerald-500/30 bg-emerald-500/10 font-medium text-xs text-emerald-700 dark:text-emerald-400"
                      : "text-xs text-muted-foreground"
                  }
                >
                  {isActive ? "Activo" : "Inactivo"}
                </Badge>
              </div>
            </div>
          ) : (
            <div>
              <DialogTitle className="flex items-center gap-2 text-lg font-bold text-foreground">
                <User className="h-5 w-5 text-muted-foreground" />
                <span>Nuevo usuario</span>
              </DialogTitle>
              <p className="mt-1 text-xs text-muted-foreground">
                Crea una cuenta de usuario con asignación de rol y módulos permitidos.
              </p>
            </div>
          )}
        </DialogHeader>

        {/* Formulario con Body Scrolleable Interno y Footer Fijo */}
        <form onSubmit={handleSubmit} className="flex min-h-0 flex-1 flex-col overflow-hidden m-0 p-0">
          {/* Mensaje de error si falla la petición */}
          {error && (
            <div className="mx-6 mt-4 flex shrink-0 items-start gap-2.5 rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-xs text-destructive">
              <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
              <div className="flex-1">{error}</div>
            </div>
          )}

          {/* Modal Body: Este contenedor tiene el scrollbar adentro */}
          <div className="flex-1 overflow-y-auto px-6 py-5 space-y-6 overscroll-contain">
            {/* SECCIÓN 1: Información de la Cuenta */}
            <div className="space-y-4">
              <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                <User className="h-3.5 w-3.5" />
                <span>Información de la Cuenta</span>
              </div>

              <div className="grid gap-4 sm:grid-cols-2">
                {/* Usuario */}
                <div className="space-y-1.5">
                  <label className="block text-xs font-medium text-foreground">
                    Nombre de usuario <span className="text-destructive">*</span>
                  </label>
                  <div className="relative">
                    <User className="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                    <Input
                      value={username}
                      onChange={(e) => setUsername(e.target.value)}
                      required
                      disabled={!!user}
                      placeholder="ej. admin_ventas"
                      className={cn(
                        "h-9 pl-8 text-sm",
                        user && "cursor-not-allowed bg-muted/40 text-muted-foreground",
                      )}
                    />
                  </div>
                  {user && (
                    <p className="flex items-center gap-1 text-[11px] text-muted-foreground">
                      <Lock className="h-3 w-3" />
                      El nombre de usuario es inmutable tras su creación.
                    </p>
                  )}
                </div>

                {/* Contraseña (solo requerida en creación) */}
                {!user && (
                  <div className="space-y-1.5">
                    <label className="block text-xs font-medium text-foreground">
                      Contraseña <span className="text-destructive">*</span>
                    </label>
                    <div className="relative">
                      <Lock className="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                      <Input
                        type={showPassword ? "text" : "password"}
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        required
                        placeholder="••••••••••••"
                        className="h-9 pl-8 pr-9 text-sm"
                      />
                      <button
                        type="button"
                        onClick={() => setShowPassword(!showPassword)}
                        className="absolute right-2.5 top-1/2 -translate-y-1/2 rounded p-0.5 text-muted-foreground hover:text-foreground"
                        title={showPassword ? "Ocultar" : "Mostrar"}
                      >
                        {showPassword ? (
                          <EyeOff className="h-3.5 w-3.5" />
                        ) : (
                          <Eye className="h-3.5 w-3.5" />
                        )}
                      </button>
                    </div>
                  </div>
                )}

                {/* Email */}
                <div className={cn("space-y-1.5", !user ? "sm:col-span-2" : "sm:col-span-1")}>
                  <label className="block text-xs font-medium text-foreground">
                    Correo Electrónico
                  </label>
                  <div className="relative">
                    <Mail className="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                    <Input
                      type="email"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      placeholder="usuario@empresa.com"
                      className="h-9 pl-8 text-sm"
                    />
                  </div>
                </div>
              </div>

              {/* Switch de Estado de Cuenta (solo en edición) */}
              {user && (
                <div className="flex items-center justify-between rounded-xl border bg-muted/20 p-3.5 transition-colors">
                  <div className="space-y-0.5">
                    <span className="text-xs font-semibold text-foreground">
                      Estado de la Cuenta
                    </span>
                    <p className="text-xs text-muted-foreground">
                      {isActive
                        ? "El usuario tiene acceso activo y puede autenticarse en el panel."
                        : "El usuario está inhabilitado y no podrá iniciar sesión."}
                    </p>
                  </div>
                  <button
                    type="button"
                    role="switch"
                    aria-checked={isActive}
                    onClick={() => setIsActive(!isActive)}
                    className={cn(
                      "relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus-visible:ring-2 focus-visible:ring-primary",
                      isActive ? "bg-emerald-600" : "bg-muted-foreground/30",
                    )}
                  >
                    <span
                      className={cn(
                        "pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow-sm ring-0 transition duration-200 ease-in-out",
                        isActive ? "translate-x-5" : "translate-x-0",
                      )}
                    />
                  </button>
                </div>
              )}
            </div>

            {/* SECCIÓN 2: Rol y Seguridad */}
            <div className="space-y-4 border-t pt-5">
              <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                <ShieldCheck className="h-3.5 w-3.5" />
                <span>Rol y Permisos de Acceso</span>
              </div>

              <div className="space-y-2">
                <label className="block text-xs font-medium text-foreground">
                  Rol del Usuario <span className="text-destructive">*</span>
                </label>
                <div className="relative">
                  <Shield className="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <select
                    className="h-9 w-full rounded-md border border-input bg-background pl-8 pr-3 text-xs sm:text-sm"
                    value={roleId || ""}
                    onChange={(e) =>
                      setRoleId(e.target.value ? Number(e.target.value) : undefined)
                    }
                  >
                    <option value="">Seleccionar rol...</option>
                    {roles.map((role) => (
                      <option key={role.id} value={role.id}>
                        {role.name} {role.is_root ? "(Root / SuperAdmin)" : ""}
                      </option>
                    ))}
                  </select>
                </div>

                {selectedRole && (
                  <div className="flex items-center justify-between rounded-xl border border-border bg-muted/20 p-3.5 transition-all">
                    <div className="flex items-center gap-2.5 min-w-0">
                      <div
                        className={cn(
                          "flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border",
                          selectedRole.is_root
                            ? "border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400"
                            : "border-border bg-background text-muted-foreground",
                        )}
                      >
                        {selectedRole.is_root ? (
                          <ShieldCheck className="h-4 w-4" />
                        ) : (
                          <Shield className="h-4 w-4" />
                        )}
                      </div>
                      <div className="min-w-0">
                        <p className="truncate text-xs font-semibold text-foreground">
                          {selectedRole.description}
                        </p>
                        <p className="text-[11px] text-muted-foreground">
                          {selectedRole.is_root
                            ? "Privilegios completos de administración y configuración global"
                            : "Permisos limitados a las acciones y módulos asignados"}
                        </p>
                      </div>
                    </div>

                    <div className="shrink-0 pl-3">
                      {selectedRole.is_root ? (
                        <Badge
                          variant="outline"
                          className="border-emerald-500/30 bg-emerald-500/10 font-medium text-xs text-emerald-700 dark:text-emerald-400"
                        >
                          Acceso total
                        </Badge>
                      ) : (
                        <Badge
                          variant="secondary"
                          className="font-normal text-xs text-muted-foreground"
                        >
                          Acceso restringido
                        </Badge>
                      )}
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* SECCIÓN 3: Módulos Permitidos */}
            {modules.length > 0 && (
              <div className="space-y-3.5 border-t pt-5">
                <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                  <div className="flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
                    <Layers className="h-3.5 w-3.5" />
                    <span>Módulos de Aplicación</span>
                    <span className="font-normal normal-case text-muted-foreground/80">
                      ({selectedModules.length} de {modules.length} seleccionados)
                    </span>
                  </div>

                  <div className="flex items-center gap-1.5">
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={handleSelectAllModules}
                      className="h-7 px-2 text-[11px] text-muted-foreground hover:text-foreground"
                    >
                      Seleccionar todos
                    </Button>
                    <span className="text-muted-foreground/40">·</span>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={handleClearAllModules}
                      className="h-7 px-2 text-[11px] text-muted-foreground hover:text-foreground"
                    >
                      Limpiar
                    </Button>
                  </div>
                </div>

                {/* Grid de tarjetas interactivas de módulos */}
                <div className="grid gap-2.5 sm:grid-cols-2">
                  {modules.map((mod) => {
                    const isChecked = selectedModules.includes(mod.id)
                    return (
                      <div
                        key={mod.id}
                        onClick={() => toggleModule(mod.id)}
                        role="checkbox"
                        aria-checked={isChecked}
                        className={cn(
                          "group relative flex cursor-pointer items-start gap-3 rounded-xl border p-3 text-left transition-all duration-150 select-none",
                          isChecked
                            ? "border-primary/60 bg-primary/[0.03] ring-1 ring-primary/40 shadow-xs"
                            : "border-input bg-card hover:border-foreground/20 hover:bg-muted/20",
                        )}
                      >
                        {/* Checkbox visual personalizado */}
                        <div
                          className={cn(
                            "mt-0.5 flex h-4 w-4 shrink-0 items-center justify-center rounded-sm border transition-colors",
                            isChecked
                              ? "border-primary bg-primary text-primary-foreground"
                              : "border-muted-foreground/40 bg-background group-hover:border-foreground/60",
                          )}
                        >
                          {isChecked && <Check className="h-3 w-3 stroke-[2.5]" />}
                        </div>

                        {/* Información del módulo */}
                        <div className="min-w-0 flex-1 space-y-0.5">
                          <div className="flex items-center justify-between gap-1.5">
                            <span className="truncate text-xs font-semibold text-foreground">
                              {mod.name}
                            </span>
                            <span className="font-mono text-[10px] text-muted-foreground">
                              [{mod.slug}]
                            </span>
                          </div>
                          {mod.description && (
                            <p className="line-clamp-2 text-[11px] leading-tight text-muted-foreground">
                              {mod.description}
                            </p>
                          )}
                        </div>
                      </div>
                    )
                  })}
                </div>
              </div>
            )}
          </div>

          {/* Footer Fijo Persistente al fondo del Modal */}
          <DialogFooter className="shrink-0 border-t border-border bg-muted/15 px-6 py-3.5 m-0">
            <Button
              type="button"
              variant="outline"
              onClick={onClose}
              disabled={loading}
              className="text-xs"
            >
              Cancelar
            </Button>
            <Button
              type="submit"
              disabled={loading}
              className="text-xs font-medium"
            >
              {loading ? (
                <>
                  <Loader2 className="mr-1.5 h-3.5 w-3.5 animate-spin" />
                  Guardando cambios...
                </>
              ) : user ? (
                <>
                  <CheckCircle2 className="mr-1.5 h-3.5 w-3.5" />
                  Guardar cambios
                </>
              ) : (
                "Crear usuario"
              )}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

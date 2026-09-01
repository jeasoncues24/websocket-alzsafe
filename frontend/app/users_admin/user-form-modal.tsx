"use client"

import { useEffect, useState } from "react"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
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
  const [email, setEmail] = useState("")
  const [roleId, setRoleId] = useState<number | undefined>()
  const [selectedModules, setSelectedModules] = useState<number[]>([])
  const [error, setError] = useState("")

  useEffect(() => {
    if (open) {
      if (user) {
        setUsername(user.username)
        setPassword("")
        setEmail(user.email || "")
        setRoleId(user.role_id)
        setSelectedModules(userModules)
      } else {
        setUsername("")
        setPassword("")
        setEmail("")
        setRoleId(roles[0]?.id)
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

      const data = user
        ? { email, role_id: roleId }
        : { username, password, email, role_id: roleId }

      await onSave(
        data as CreateUserRequest | UpdateUserRequest,
        selectedModules,
      )
      onClose()
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Error al guardar")
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

  if (!open) return null

  return (
    <Dialog open={open} onOpenChange={(next) => !next && onClose()}>
      <DialogContent className="max-h-[90vh] max-w-2xl overflow-y-auto">
        <DialogHeader>
          <DialogTitle>
            {user ? "Editar usuario_admin" : "Nuevo usuario_admin"}
          </DialogTitle>
          <DialogDescription>
            {user
              ? "Ajusta datos, rol y módulos del usuario administrativo."
              : "Crea un nuevo usuario administrativo con permisos iniciales."}
          </DialogDescription>
        </DialogHeader>

        {error && (
          <div className="rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="grid gap-4 md:grid-cols-2">
          <div className="space-y-2">
            <label className="block text-sm font-medium">Usuario</label>
            <Input
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              disabled={!!user}
              placeholder="Nombre de usuario"
            />
          </div>

          {!user && (
            <div className="space-y-2">
              <label className="block text-sm font-medium">Contraseña</label>
              <Input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                placeholder="Contraseña"
              />
            </div>
          )}

          <div className="space-y-2 md:col-span-2">
            <label className="block text-sm font-medium">Email</label>
            <Input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="email@ejemplo.com"
            />
          </div>

          <div className="space-y-2 md:col-span-2">
            <label className="block text-sm font-medium">Rol</label>
            <select
              className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm"
              value={roleId || ""}
              onChange={(e) =>
                setRoleId(e.target.value ? Number(e.target.value) : undefined)
              }
            >
              <option value="">Seleccionar rol</option>
              {roles.map((role) => (
                <option key={role.id} value={role.id}>
                  {role.name}
                </option>
              ))}
            </select>
          </div>

          {modules.length > 0 && (
            <div className="space-y-2 md:col-span-2">
              <label className="block text-sm font-medium">Módulos permitidos</label>
              <div className="grid gap-2 rounded-md border p-3 sm:grid-cols-2">
                {modules.map((mod) => (
                  <label
                    key={mod.id}
                    className="flex cursor-pointer items-center gap-2 rounded-md border bg-background px-3 py-2 transition-colors hover:bg-muted/30"
                  >
                    <input
                      type="checkbox"
                      checked={selectedModules.includes(mod.id)}
                      onChange={() => toggleModule(mod.id)}
                      className="rounded border-input"
                    />
                    <span className="text-sm">
                      {mod.name}
                      <span className="ml-2 font-mono text-xs text-muted-foreground">
                        {mod.slug}
                      </span>
                    </span>
                  </label>
                ))}
              </div>
            </div>
          )}

          <DialogFooter className="md:col-span-2">
            <Button type="button" variant="outline" onClick={onClose}>
              Cancelar
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? "Guardando..." : "Guardar"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

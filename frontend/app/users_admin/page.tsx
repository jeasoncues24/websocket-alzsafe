"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { DataTable } from "@/components/data-table"
import {
  AlertCircle,
  CheckCircle2,
  Users,
} from "lucide-react"
import {
  getUsuarioAdmins,
  createUsuarioAdmin,
  updateUsuarioAdmin,
  deleteUsuarioAdmin,
  assignUsuarioAdminModules,
  getUsuarioAdminModules,
  getRoles,
  getModules,
  type UserAdminRol,
  type Role,
  type Module,
  type CreateUserRequest,
  type UpdateUserRequest,
} from "@/lib/api"
import { getColumns } from "./columns"
import {
  DataTableToolbar,
  type UserStatusFilter,
} from "./data-table-toolbar"
import { DataCardList } from "./data-card-list"
import { UserFormModal } from "./user-form-modal"

async function loadUserModules(userId: number): Promise<number[]> {
  try {
    const json = await getUsuarioAdminModules(userId)
    return json.module_ids || []
  } catch {
    return []
  }
}

export default function UsuarioAdminPage() {
  const [users, setUsers] = useState<UserAdminRol[]>([])
  const [roles, setRoles] = useState<Role[]>([])
  const [modules, setModules] = useState<Module[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)

  // Filtros
  const [globalFilter, setGlobalFilter] = useState("")
  const [statusFilter, setStatusFilter] = useState<UserStatusFilter>("all")
  const [roleFilter, setRoleFilter] = useState("")

  // Formulario y eliminación
  const [formOpen, setFormOpen] = useState(false)
  const [editTarget, setEditTarget] = useState<UserAdminRol | null>(null)
  const [editModules, setEditModules] = useState<number[]>([])
  const [deletingId, setDeletingId] = useState<number | null>(null)

  // Notificaciones y alertas
  const [loadError, setLoadError] = useState("")
  const [deleteError, setDeleteError] = useState("")
  const [actionMessage, setActionMessage] = useState("")

  const loadData = useCallback(async (opts?: { silent?: boolean }) => {
    if (!opts?.silent) setRefreshing(true)
    try {
      const [usersResp, rolesResp, modulesResp] = await Promise.all([
        getUsuarioAdmins(1, 500),
        getRoles(),
        getModules(),
      ])
      setUsers(usersResp.users || [])
      setRoles(rolesResp.roles || [])
      setModules(modulesResp.modules || [])
      setLoadError("")
    } catch (err: unknown) {
      setLoadError(
        err instanceof Error ? err.message : "Error al cargar usuarios administrativos",
      )
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }, [])

  useEffect(() => {
    loadData({ silent: true })
  }, [loadData])

  async function handleSave(
    data: CreateUserRequest | UpdateUserRequest,
    selectedModules: number[],
  ) {
    if (editTarget) {
      await updateUsuarioAdmin(editTarget.id, data)
      await assignUsuarioAdminModules(editTarget.id, selectedModules)
      setActionMessage(`Usuario ${editTarget.username} actualizado correctamente`)
    } else {
      const newUser = await createUsuarioAdmin(data as CreateUserRequest)
      if (selectedModules.length > 0) {
        await assignUsuarioAdminModules(newUser.id, selectedModules)
      }
      setActionMessage(`Usuario creado correctamente`)
    }
    await loadData({ silent: true })
  }

  const handleDelete = useCallback(
    async (user: UserAdminRol) => {
      if (!confirm(`¿Eliminar o deshabilitar el usuario "${user.username}"?`)) return
      setDeletingId(user.id)
      setDeleteError("")
      setActionMessage("")
      try {
        const result = await deleteUsuarioAdmin(user.id)
        setActionMessage(
          result.status === "disabled"
            ? `Usuario ${user.username} deshabilitado (tiene dependencias activas)`
            : `Usuario ${user.username} eliminado correctamente`,
        )
        await loadData({ silent: true })
      } catch (err: unknown) {
        setDeleteError(
          err instanceof Error ? err.message : "Error al eliminar usuario",
        )
      } finally {
        setDeletingId(null)
      }
    },
    [loadData],
  )

  const openNew = useCallback(() => {
    setEditTarget(null)
    setEditModules([])
    setFormOpen(true)
  }, [])

  const openEdit = useCallback(async (user: UserAdminRol) => {
    const userMods = await loadUserModules(user.id)
    setEditTarget(user)
    setEditModules(userMods)
    setFormOpen(true)
  }, [])

  // Conteo para los botones segmentados de estado
  const counts = useMemo<Record<UserStatusFilter, number>>(() => {
    const scoped = roleFilter
      ? users.filter((u) => u.rol === roleFilter)
      : users

    return {
      all: scoped.length,
      active: scoped.filter((u) => u.activo).length,
      inactive: scoped.filter((u) => !u.activo).length,
      root: scoped.filter((u) => u.is_root).length,
    }
  }, [users, roleFilter])

  // Filtrado de negocio en el contenedor
  const filteredUsers = useMemo(() => {
    return users.filter((user) => {
      // 1. Filtro por rol
      if (roleFilter && user.rol !== roleFilter) {
        return false
      }

      // 2. Filtro por estado
      if (statusFilter === "active" && !user.activo) return false
      if (statusFilter === "inactive" && user.activo) return false
      if (statusFilter === "root" && !user.is_root) return false

      // 3. Búsqueda por texto global
      if (!globalFilter.trim()) return true
      const query = globalFilter.toLowerCase().trim()
      const username = (user.username || "").toLowerCase()
      const email = (user.email || "").toLowerCase()
      const rol = (user.rol || "").toLowerCase()
      const idStr = String(user.id)

      return (
        idStr.includes(query) ||
        username.includes(query) ||
        email.includes(query) ||
        rol.includes(query)
      )
    })
  }, [users, roleFilter, statusFilter, globalFilter])

  // Columnas para TanStack Table
  const columns = useMemo(
    () =>
      getColumns({
        onEdit: openEdit,
        onDelete: handleDelete,
        deletingId,
      }),
    [openEdit, handleDelete, deletingId],
  )

  return (
    <div className="space-y-4 pb-12">
      {/* Encabezado */}
      <div className="space-y-1">
        <h1 className="flex items-center gap-2 text-xl font-semibold tracking-tight text-foreground sm:text-2xl">
          <Users className="h-5 w-5 text-muted-foreground" />
          <span>Usuarios Administrativos</span>
        </h1>
        <p className="text-sm text-muted-foreground">
          Administra los usuarios con acceso al panel, asignación de roles y permisos modulares.
        </p>
      </div>

      {/* Alertas de resultado / errores */}
      {actionMessage && (
        <Alert>
          <div className="flex items-start gap-3">
            <CheckCircle2 className="mt-0.5 h-5 w-5 text-emerald-500" />
            <div>
              <AlertTitle>Operación completada</AlertTitle>
              <AlertDescription>{actionMessage}</AlertDescription>
            </div>
          </div>
        </Alert>
      )}

      {deleteError && (
        <Alert variant="destructive">
          <div className="flex items-start gap-3">
            <AlertCircle className="mt-0.5 h-5 w-5" />
            <div>
              <AlertTitle>Error al eliminar usuario</AlertTitle>
              <AlertDescription>{deleteError}</AlertDescription>
            </div>
          </div>
        </Alert>
      )}

      {loadError && (
        <Alert variant="destructive">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div className="flex items-start gap-3">
              <AlertCircle className="mt-0.5 h-4 w-4 shrink-0" />
              <div>
                <AlertTitle>Error al cargar usuarios</AlertTitle>
                <AlertDescription>{loadError}</AlertDescription>
              </div>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={() => loadData()}
              disabled={refreshing}
              className="shrink-0 self-start"
            >
              Reintentar
            </Button>
          </div>
        </Alert>
      )}

      {/* Barra de herramientas: control segmentado + buscador + selector de rol + botón nuevo */}
      <DataTableToolbar
        globalFilter={globalFilter}
        onGlobalFilterChange={setGlobalFilter}
        statusFilter={statusFilter}
        onStatusFilterChange={setStatusFilter}
        roleFilter={roleFilter}
        onRoleFilterChange={setRoleFilter}
        roles={roles}
        counts={counts}
        loading={refreshing}
        onRefresh={() => loadData()}
        onNewUser={openNew}
      />

      {/* Escritorio: tabla TanStack Table */}
      <div className="hidden md:flex md:flex-col">
        <DataTable
          columns={columns}
          data={filteredUsers}
          loading={loading}
          itemLabel="usuarios"
          initialPageSize={10}
          pageSizeOptions={[10, 20, 30, 50]}
          emptyMessage="No se encontraron usuarios que coincidan con los filtros."
        />
      </div>

      {/* Móvil: lista de tarjetas accionables */}
      <div className="md:hidden">
        <DataCardList
          data={filteredUsers}
          loading={loading}
          onEdit={openEdit}
          onDelete={handleDelete}
          deletingId={deletingId}
        />
      </div>

      {/* Modal para Crear / Editar Usuario */}
      <UserFormModal
        open={formOpen}
        onClose={() => setFormOpen(false)}
        onSave={handleSave}
        user={editTarget}
        roles={roles}
        modules={modules}
        userModules={editModules}
      />
    </div>
  )
}

"use client";

import { useEffect, useState, useCallback } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  CardDescription,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Search, Users as UsersIcon, Plus, Pencil, Trash2 } from "lucide-react";
import {
  getUsers,
  createUser,
  updateUser,
  deleteUser,
  assignUserModules,
  getRoles,
  getModules,
  type UserAdminRol,
  type Role,
  type Module,
  type CreateUserRequest,
  type UpdateUserRequest,
  API_BASE,
} from "@/lib/api";

function UserFormModal({
  open,
  onClose,
  onSave,
  user,
  roles,
  modules,
  userModules = [],
}: {
  open: boolean;
  onClose: () => void;
  onSave: (
    data: CreateUserRequest | UpdateUserRequest,
    modules: number[],
  ) => Promise<void>;
  user: UserAdminRol | null;
  roles: Role[];
  modules: Module[];
  userModules?: number[];
}) {
  const [loading, setLoading] = useState(false);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [email, setEmail] = useState("");
  const [roleId, setRoleId] = useState<number | undefined>();
  const [selectedModules, setSelectedModules] = useState<number[]>([]);
  const [empresaId, setEmpresaId] = useState<number | undefined>();
  const [error, setError] = useState("");

  useEffect(() => {
    if (open) {
      if (user) {
        setUsername(user.username);
        setPassword("");
        setEmail(user.email || "");
        setRoleId(user.role_id);
        setSelectedModules(userModules);
        setEmpresaId(user.empresa_id);
      } else {
        setUsername("");
        setPassword("");
        setEmail("");
        setRoleId(roles[0]?.id);
        setSelectedModules([]);
        setEmpresaId(undefined);
      }
      setError("");
    }
  }, [open, user, roles, userModules]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      if (!user && !password) {
        setError("La contraseña es requerida para nuevos usuarios");
        return;
      }

      const data = user
        ? { email, role_id: roleId, empresa_id: empresaId }
        : { username, password, email, role_id: roleId, empresa_id: empresaId };

      await onSave(
        data as CreateUserRequest | UpdateUserRequest,
        selectedModules,
      );
      onClose();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Error al guardar");
    } finally {
      setLoading(false);
    }
  }

  function toggleModule(moduleId: number) {
    setSelectedModules((prev) =>
      prev.includes(moduleId)
        ? prev.filter((id) => id !== moduleId)
        : [...prev, moduleId],
    );
  }

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="bg-background rounded-lg shadow-lg w-full max-w-md mx-4 max-h-[90vh] overflow-y-auto">
        <div className="p-6">
          <h2 className="text-xl font-semibold mb-4">
            {user ? "Editar Usuario" : "Nuevo Usuario"}
          </h2>

          {error && (
            <div className="mb-4 p-3 bg-destructive/10 text-destructive rounded-md text-sm">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium mb-1">Usuario</label>
              <Input
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                required
                disabled={!!user}
                placeholder="Nombre de usuario"
              />
            </div>

            {!user && (
              <div>
                <label className="block text-sm font-medium mb-1">
                  Contraseña
                </label>
                <Input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                  placeholder="Contraseña"
                />
              </div>
            )}

            <div>
              <label className="block text-sm font-medium mb-1">Email</label>
              <Input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="email@ejemplo.com"
              />
            </div>

            <div>
              <label className="block text-sm font-medium mb-1">Rol</label>
              <select
                className="w-full h-10 px-3 rounded-md border border-input bg-background text-sm"
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
              <div>
                <label className="block text-sm font-medium mb-2">
                  Módulos Adicionales
                </label>
                <div className="space-y-2 max-h-40 overflow-y-auto border rounded-md p-3">
                  {modules.map((mod) => (
                    <label
                      key={mod.id}
                      className="flex items-center gap-2 cursor-pointer"
                    >
                      <input
                        type="checkbox"
                        checked={selectedModules.includes(mod.id)}
                        onChange={() => toggleModule(mod.id)}
                        className="rounded border-input"
                      />
                      <span className="text-sm">{mod.name}</span>
                    </label>
                  ))}
                </div>
              </div>
            )}

            <div className="flex justify-end gap-2 pt-4">
              <Button type="button" variant="outline" onClick={onClose}>
                Cancelar
              </Button>
              <Button type="submit" disabled={loading}>
                {loading ? "Guardando..." : "Guardar"}
              </Button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}

async function loadUserModules(userId: number): Promise<number[]> {
  try {
    const res = await fetch(`${API_BASE}/admin/users/${userId}/modules`, {
      headers: {
        Authorization: `Bearer ${localStorage.getItem("admin_token")}`,
      },
    });
    if (!res.ok) return [];
    const json = await res.json();
    return json.module_ids || [];
  } catch {
    return [];
  }
}

export default function UsersPage() {
  const [users, setUsers] = useState<UserAdminRol[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const limit = 20;

  const [roles, setRoles] = useState<Role[]>([]);
  const [modules, setModules] = useState<Module[]>([]);
  const [formOpen, setFormOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<UserAdminRol | null>(null);
  const [editModules, setEditModules] = useState<number[]>([]);
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [loadError, setLoadError] = useState("");
  const [deleteError, setDeleteError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const resp = await getUsers(page, limit);
      console.log(resp);
      const filtered = search
        ? resp.users.filter(
            (u) =>
              u.username.toLowerCase().includes(search.toLowerCase()) ||
              u.email?.toLowerCase().includes(search.toLowerCase()),
          )
        : resp.users;
      setUsers(filtered);
      setTotal(resp.total);
      setLoadError("");
    } catch (err: unknown) {
      setUsers([]);
      setLoadError(
        err instanceof Error ? err.message : "Error al cargar usuarios",
      );
    } finally {
      setLoading(false);
    }
  }, [page, search]);

  const loadRolesModules = useCallback(async () => {
    try {
      const [rolesResp, modulesResp] = await Promise.all([
        getRoles(),
        getModules(),
      ]);
      setRoles(rolesResp.roles || []);
      setModules(modulesResp.modules || []);
    } catch (err) {
      console.error("Error loading roles/modules:", err);
    }
  }, []);

  useEffect(() => {
    const timer = setTimeout(load, 300);
    return () => clearTimeout(timer);
  }, [load]);

  useEffect(() => {
    loadRolesModules();
  }, [loadRolesModules]);

  async function handleSave(
    data: CreateUserRequest | UpdateUserRequest,
    selectedModules: number[],
  ) {
    if (editTarget) {
      await updateUser(editTarget.id, data);
      await assignUserModules(editTarget.id, selectedModules);
    } else {
      const newUser = await createUser(data as CreateUserRequest);
      if (selectedModules.length > 0) {
        await assignUserModules(newUser.id, selectedModules);
      }
    }
    setPage(1);
    await load();
  }

  async function handleDelete(user: UserAdminRol) {
    if (!confirm(`¿Eliminar el usuario "${user.username}"?`)) return;
    setDeletingId(user.id);
    setDeleteError("");
    try {
      await deleteUser(user.id);
      await load();
    } catch (err: unknown) {
      setDeleteError(
        err instanceof Error ? err.message : "Error al eliminar usuario",
      );
    } finally {
      setDeletingId(null);
    }
  }

  function openNew() {
    setEditTarget(null);
    setEditModules([]);
    setFormOpen(true);
  }

  async function openEdit(user: UserAdminRol) {
    const userMods = await loadUserModules(user.id);
    setEditTarget(user);
    setEditModules(userMods);
    setFormOpen(true);
  }

  const totalPages = Math.ceil(total / limit);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Usuarios</h1>
          <p className="text-muted-foreground">
            Gestiona los usuarios del panel de administración
          </p>
        </div>
        <Button onClick={openNew}>
          <Plus className="h-4 w-4 mr-2" />
          Nuevo Usuario
        </Button>
      </div>

      {loadError && <p className="text-sm text-destructive">{loadError}</p>}
      {deleteError && <p className="text-sm text-destructive">{deleteError}</p>}

      <Card>
        <CardHeader>
          <div className="flex flex-col sm:flex-row sm:items-center gap-3 justify-between">
            <div>
              <CardTitle>Lista de Usuarios</CardTitle>
              <CardDescription>{total} usuario(s) en total</CardDescription>
            </div>
            <div className="relative w-52">
              <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Buscar por usuario o email"
                className="pl-8"
                value={search}
                onChange={(e) => {
                  setSearch(e.target.value);
                  setPage(1);
                }}
              />
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Usuario</TableHead>
                <TableHead>Email</TableHead>
                <TableHead>Rol</TableHead>
                <TableHead>Estado</TableHead>
                <TableHead className="text-right">Acciones</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell
                    colSpan={5}
                    className="text-center text-muted-foreground py-8"
                  >
                    Cargando...
                  </TableCell>
                </TableRow>
              ) : users.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={5}
                    className="text-center text-muted-foreground py-8"
                  >
                    <UsersIcon className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    No hay usuarios registrados
                  </TableCell>
                </TableRow>
              ) : (
                users.map((user) => (
                  <TableRow key={user.id}>
                    <TableCell className="font-medium">
                      {user.username}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {user.email || "—"}
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{user.rol}</Badge>
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={user.activo ? "default" : "secondary"}
                        className={user.activo ? "bg-green-500" : ""}
                      >
                        {user.activo ? "Activo" : "Inactivo"}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-1">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => openEdit(user)}
                          title="Editar"
                        >
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleDelete(user)}
                          disabled={deletingId === user.id}
                          title="Eliminar"
                          className="text-destructive hover:text-destructive"
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>

          {totalPages > 1 && (
            <div className="flex justify-between items-center mt-4 text-sm text-muted-foreground">
              <span>
                Página {page} de {totalPages}
              </span>
              <div className="flex gap-2">
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page === 1}
                  onClick={() => setPage((p) => p - 1)}
                >
                  Anterior
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  disabled={page === totalPages}
                  onClick={() => setPage((p) => p + 1)}
                >
                  Siguiente
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>

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
  );
}

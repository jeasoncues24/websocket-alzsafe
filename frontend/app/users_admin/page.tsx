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
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Search, Users as UsersIcon, Plus, Pencil, Trash2 } from "lucide-react";
import {
  getEmpresas,
  getUsuarioAdmins,
  createUsuarioAdmin,
  updateUsuarioAdmin,
  deleteUsuarioAdmin,
  assignUsuarioAdminModules,
  getUsuarioAdminModules,
  getRoles,
  getModules,
  type Empresa,
  type UserAdminRol,
  type Role,
  type Module,
  type CreateUserRequest,
  type UpdateUserRequest,
} from "@/lib/api";

function UserFormModal({
  open,
  onClose,
  onSave,
  user,
  roles,
  modules,
  companies,
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
  companies: Empresa[];
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

  const companyOptions = companies;

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
    <Dialog open={open} onOpenChange={(next) => !next && onClose()}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>
            {user ? "Editar usuario_admin" : "Nuevo usuario_admin"}
          </DialogTitle>
          <DialogDescription>
            {user
              ? "Ajusta datos, rol, empresa y módulos del usuario."
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

          <div className="space-y-2">
            <label className="block text-sm font-medium">Rol</label>
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

          <div className="space-y-2">
            <label className="block text-sm font-medium">Empresa</label>
            <select
              className="w-full h-10 px-3 rounded-md border border-input bg-background text-sm"
              value={empresaId || ""}
              onChange={(e) =>
                setEmpresaId(e.target.value ? Number(e.target.value) : undefined)
              }
            >
              <option value="">Sin empresa</option>
              {companyOptions.map((company) => (
                <option key={company.id} value={company.id}>
                  {company.nombre}
                </option>
              ))}
            </select>
          </div>

          {modules.length > 0 && (
            <div className="md:col-span-2 space-y-2">
              <label className="block text-sm font-medium">Módulos</label>
              <div className="grid gap-2 rounded-md border p-3 sm:grid-cols-2">
                {modules.map((mod) => (
                  <label
                    key={mod.id}
                    className="flex items-center gap-2 rounded-md border bg-background px-3 py-2 cursor-pointer"
                  >
                    <input
                      type="checkbox"
                      checked={selectedModules.includes(mod.id)}
                      onChange={() => toggleModule(mod.id)}
                      className="rounded border-input"
                    />
                    <span className="text-sm">
                      {mod.name}
                      <span className="ml-2 text-xs text-muted-foreground">
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
  );
}

async function loadUserModules(userId: number): Promise<number[]> {
  try {
    const json = await getUsuarioAdminModules(userId);
    return json.module_ids || [];
  } catch {
    return [];
  }
}

export default function UsuarioAdminPage() {
  const [users, setUsers] = useState<UserAdminRol[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const limit = 20;

  const [roles, setRoles] = useState<Role[]>([]);
  const [modules, setModules] = useState<Module[]>([]);
  const [companies, setCompanies] = useState<Empresa[]>([]);
  const [formOpen, setFormOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<UserAdminRol | null>(null);
  const [editModules, setEditModules] = useState<number[]>([]);
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [loadError, setLoadError] = useState("");
  const [deleteError, setDeleteError] = useState("");
  const [actionMessage, setActionMessage] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const resp = await getUsuarioAdmins(page, limit);
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
        err instanceof Error ? err.message : "Error al cargar usuario_admin",
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

  const loadCompanies = useCallback(async () => {
    try {
      const resp = await getEmpresas({ limit: 1000 });
      setCompanies(resp.empresas || []);
    } catch (err) {
      console.error("Error loading companies:", err);
    }
  }, []);

  useEffect(() => {
    const timer = setTimeout(load, 300);
    return () => clearTimeout(timer);
  }, [load]);

  useEffect(() => {
    loadRolesModules();
  }, [loadRolesModules]);

  useEffect(() => {
    loadCompanies();
  }, [loadCompanies]);

  async function handleSave(
    data: CreateUserRequest | UpdateUserRequest,
    selectedModules: number[],
  ) {
    if (editTarget) {
      await updateUsuarioAdmin(editTarget.id, data);
      await assignUsuarioAdminModules(editTarget.id, selectedModules);
    } else {
      const newUser = await createUsuarioAdmin(data as CreateUserRequest);
      if (selectedModules.length > 0) {
        await assignUsuarioAdminModules(newUser.id, selectedModules);
      }
    }
    setPage(1);
    await load();
  }

  async function handleDelete(user: UserAdminRol) {
    if (!confirm(`¿Eliminar el usuario "${user.username}"?`)) return;
    setDeletingId(user.id);
    setDeleteError("");
    setActionMessage("");
    try {
      const result = await deleteUsuarioAdmin(user.id);
      setActionMessage(
        result.status === "disabled"
          ? `Usuario ${user.username} deshabilitado`
          : `Usuario ${user.username} eliminado`,
      );
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
          <h1 className="text-3xl font-bold tracking-tight">Usuario Admin</h1>
          <p className="text-muted-foreground">
            Gestiona los usuario_admin del panel de administración
          </p>
        </div>
        <Button onClick={openNew}>
          <Plus className="h-4 w-4 mr-2" />
          Nuevo usuario_admin
        </Button>
      </div>

      <Alert>
        <AlertTitle>Delete behavior</AlertTitle>
        <AlertDescription>
          If a user has dependencies, the backend disables it instead of removing it.
        </AlertDescription>
      </Alert>

      {loadError && (
        <Alert variant="destructive">
          <AlertTitle>Error al cargar</AlertTitle>
          <AlertDescription>{loadError}</AlertDescription>
        </Alert>
      )}
      {deleteError && (
        <Alert variant="destructive">
          <AlertTitle>Error de eliminación</AlertTitle>
          <AlertDescription>{deleteError}</AlertDescription>
        </Alert>
      )}
      {actionMessage && (
        <Alert>
          <AlertTitle>Acción completada</AlertTitle>
          <AlertDescription>{actionMessage}</AlertDescription>
        </Alert>
      )}

      <Card>
        <CardHeader>
          <div className="flex flex-col sm:flex-row sm:items-center gap-3 justify-between">
            <div>
              <CardTitle>Lista de usuario_admin</CardTitle>
              <CardDescription>{total} usuario_admin(s) en total</CardDescription>
            </div>
            <div className="relative w-52">
              <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input
                placeholder="Buscar por usuario_admin o email"
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
                <TableHead>Empresa</TableHead>
                <TableHead>Rol</TableHead>
                <TableHead>Root</TableHead>
                <TableHead>Estado</TableHead>
                <TableHead className="text-right">Acciones</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell
                    colSpan={7}
                    className="text-center text-muted-foreground py-8"
                  >
                    Cargando...
                  </TableCell>
                </TableRow>
              ) : users.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={7}
                    className="text-center text-muted-foreground py-8"
                  >
                    <UsersIcon className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    No hay usuario_admin registrados
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
                      <Badge variant="outline">
                        {user.empresa_id
                          ? companies.find((c) => c.id === user.empresa_id)
                              ?.nombre ?? `Empresa #${user.empresa_id}`
                          : "Global"}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{user.rol}</Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant={user.is_root ? "default" : "secondary"}>
                        {user.is_root ? "Sí" : "No"}
                      </Badge>
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
        companies={companies}
        userModules={editModules}
      />
    </div>
  );
}

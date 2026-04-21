"use client";

import { useEffect, useMemo, useState } from "react";
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
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Shield, ShieldCheck, Plus, Pencil, Trash2 } from "lucide-react";
import {
  createRole,
  deleteRole,
  getModules,
  getRoles,
  updateRole,
  type Module,
  type Role,
  type RoleRequest,
} from "@/lib/api";

function RoleFormDialog({
  open,
  onClose,
  onSave,
  role,
  modules,
}: {
  open: boolean;
  onClose: () => void;
  onSave: (data: RoleRequest) => Promise<void>;
  role: Role | null;
  modules: Module[];
}) {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [isRoot, setIsRoot] = useState(false);
  const [permissions, setPermissions] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!open) return;
    setName(role?.name ?? "");
    setDescription(role?.description ?? "");
    setIsRoot(role?.is_root ?? false);
    setPermissions(role?.permissions ?? []);
    setError("");
  }, [open, role]);

  const moduleSlugs = useMemo(
    () => modules.map((module) => module.slug).filter(Boolean),
    [modules],
  );

  function togglePermission(slug: string) {
    setPermissions((current) =>
      current.includes(slug)
        ? current.filter((item) => item !== slug)
        : [...current, slug],
    );
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      await onSave({
        name,
        description,
        is_root: isRoot,
        permissions,
      });
      onClose();
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Error al guardar rol");
    } finally {
      setLoading(false);
    }
  }

  if (!open) return null;

  return (
    <Dialog open={open} onOpenChange={(next) => !next && onClose()}>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{role ? "Editar rol" : "Nuevo rol"}</DialogTitle>
          <DialogDescription>
            Define el nombre, el estado root y los permisos por módulo.
          </DialogDescription>
        </DialogHeader>

        {error && (
          <div className="rounded-md border border-destructive/40 bg-destructive/10 p-3 text-sm text-destructive">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="grid gap-4 md:grid-cols-2">
          <div className="space-y-2">
            <label className="block text-sm font-medium">Nombre</label>
            <Input
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="Nombre del rol"
              required
            />
          </div>

          <div className="space-y-2">
            <label className="block text-sm font-medium">Tipo</label>
            <label className="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
              <input
                type="checkbox"
                checked={isRoot}
                onChange={(e) => setIsRoot(e.target.checked)}
              />
              Root / super admin
            </label>
          </div>

          <div className="space-y-2 md:col-span-2">
            <label className="block text-sm font-medium">Descripción</label>
            <Input
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Descripción del rol"
            />
          </div>

          <div className="space-y-2 md:col-span-2">
            <div className="flex items-center justify-between">
              <label className="block text-sm font-medium">Permisos</label>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => setPermissions(moduleSlugs)}
              >
                Marcar todos
              </Button>
            </div>
            <div className="grid gap-2 rounded-md border p-3 sm:grid-cols-2">
              {modules.map((module) => (
                <label
                  key={module.id}
                  className="flex items-center gap-2 rounded-md border bg-background px-3 py-2 cursor-pointer"
                >
                  <input
                    type="checkbox"
                    checked={permissions.includes(module.slug)}
                    onChange={() => togglePermission(module.slug)}
                    className="rounded border-input"
                  />
                  <span className="text-sm">
                    {module.name}
                    <span className="ml-2 text-xs text-muted-foreground">
                      {module.slug}
                    </span>
                  </span>
                </label>
              ))}
            </div>
          </div>

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

export default function RolesPage() {
  const [roles, setRoles] = useState<Role[]>([]);
  const [modules, setModules] = useState<Module[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [formOpen, setFormOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<Role | null>(null);
  const [deleteMessage, setDeleteMessage] = useState("");
  const [deletingId, setDeletingId] = useState<number | null>(null);

  async function load() {
    setLoading(true);
    try {
      const [rolesResp, modulesResp] = await Promise.all([getRoles(), getModules()]);
      setRoles(rolesResp.roles || []);
      setModules(modulesResp.modules || []);
      setError("");
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Error al cargar roles");
      setRoles([]);
      setModules([]);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    load();
  }, []);

  async function handleSave(data: RoleRequest) {
    if (editTarget) {
      await updateRole(editTarget.id, {
        name: data.name,
        description: data.description,
        is_root: data.is_root,
        permissions: data.permissions,
      });
    } else {
      await createRole({
        name: data.name,
        description: data.description,
        is_root: data.is_root,
        permissions: data.permissions,
      });
    }
    await load();
  }

  async function handleDelete(role: Role) {
    if (role.is_root || (role.usage_count ?? 0) > 0) return;
    if (!confirm(`¿Eliminar el rol "${role.name}"?`)) return;
    setDeletingId(role.id);
    setDeleteMessage("");
    try {
      const result = await deleteRole(role.id);
      setDeleteMessage(
        result.status === "deleted"
          ? `Rol ${role.name} eliminado`
          : `Rol ${role.name} procesado`,
      );
      await load();
    } catch (err: unknown) {
      setDeleteMessage(err instanceof Error ? err.message : "No se pudo eliminar el rol");
    } finally {
      setDeletingId(null);
    }
  }

  const openCreate = () => {
    setEditTarget(null);
    setFormOpen(true);
  };

  const openEdit = (role: Role) => {
    setEditTarget(role);
    setFormOpen(true);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Roles</h1>
          <p className="text-muted-foreground">Gestiona roles y permisos por módulo</p>
        </div>
        <Button onClick={openCreate}>
          <Plus className="h-4 w-4 mr-2" />
          Nuevo rol
        </Button>
      </div>

      <Alert>
        <AlertTitle>Riesgo de eliminación</AlertTitle>
        <AlertDescription>
          Los roles root o en uso no se pueden eliminar. El catálogo sirve como referencia para permisos.
        </AlertDescription>
      </Alert>

      {error && (
        <Alert variant="destructive">
          <AlertTitle>Error al cargar</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}
      {deleteMessage && (
        <Alert>
          <AlertTitle>Acción completada</AlertTitle>
          <AlertDescription>{deleteMessage}</AlertDescription>
        </Alert>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Lista de roles</CardTitle>
          <CardDescription>{roles.length} rol(es) configurado(s)</CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nombre</TableHead>
                <TableHead>Descripción</TableHead>
                <TableHead>Permisos</TableHead>
                <TableHead>Root</TableHead>
                <TableHead>Uso</TableHead>
                <TableHead className="text-right">Acciones</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center text-muted-foreground py-8">
                    Cargando...
                  </TableCell>
                </TableRow>
              ) : roles.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={6} className="text-center text-muted-foreground py-8">
                    <Shield className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    No hay roles configurados
                  </TableCell>
                </TableRow>
              ) : (
                roles.map((role) => (
                  <TableRow key={role.id}>
                    <TableCell className="font-medium flex items-center gap-2">
                      {role.is_root ? (
                        <ShieldCheck className="h-4 w-4 text-purple-500" />
                      ) : (
                        <Shield className="h-4 w-4" />
                      )}
                      {role.name}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {role.description || "—"}
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap gap-1">
                        {(role.permissions || []).slice(0, 3).map((perm) => (
                          <Badge key={perm} variant="outline">
                            {perm}
                          </Badge>
                        ))}
                        {(role.permissions?.length || 0) > 3 && (
                          <Badge variant="secondary">+{role.permissions.length - 3}</Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant={role.is_root ? "default" : "secondary"}>
                        {role.is_root ? "Root" : "Normal"}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Badge variant={(role.usage_count ?? 0) > 0 ? "default" : "secondary"}>
                        {(role.usage_count ?? 0) > 0 ? `En uso (${role.usage_count})` : "Disponible"}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-1">
                        <Button variant="ghost" size="icon" onClick={() => openEdit(role)} title="Editar">
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleDelete(role)}
                          disabled={deletingId === role.id || role.is_root || (role.usage_count ?? 0) > 0}
                          title={role.is_root ? "Rol root protegido" : (role.usage_count ?? 0) > 0 ? "Rol en uso" : "Eliminar"}
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
        </CardContent>
      </Card>

      <RoleFormDialog
        open={formOpen}
        onClose={() => setFormOpen(false)}
        onSave={handleSave}
        role={editTarget}
        modules={modules}
      />
    </div>
  );
}

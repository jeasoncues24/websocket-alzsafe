"use client";

import { useEffect, useState, useCallback } from "react";
import { useRouter } from "next/navigation";
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
import { Select } from "@/components/ui/select";
import { Search, Building2, Plus, Pencil, Eye, Trash2, KeyRound } from "lucide-react";
import {
  getEmpresas,
  createEmpresa,
  updateEmpresa,
  deleteEmpresa,
  type Empresa,
  type EmpresaCreateRequest,
} from "@/lib/api";
import { EmpresaFormModal } from "@/components/companies/empresa-form-modal";
import { EmpresaDetailModal } from "@/components/companies/empresa-detail-modal";

export default function CompaniesPage() {
  const router = useRouter();
  const [empresas, setEmpresas] = useState<Empresa[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [search, setSearch] = useState("");
  const [estado, setEstado] = useState<string>("todos");
  const [page, setPage] = useState(1);
  const limit = 20;

  const [formOpen, setFormOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<Empresa | null>(null);
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailTarget, setDetailTarget] = useState<Empresa | null>(null);
  const [deletingId, setDeletingId] = useState<number | null>(null);
  const [loadError, setLoadError] = useState("");
  const [deleteError, setDeleteError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const resp = await getEmpresas({
        page,
        limit,
        busqueda: search || undefined,
        estado: estado !== "todos" ? estado : undefined,
      });
      setEmpresas(resp.empresas ?? []);
      setTotal(resp.total);
      setLoadError("");
    } catch (err: unknown) {
      setEmpresas([]);
      setLoadError(
        err instanceof Error ? err.message : "Error al cargar empresas",
      );
    } finally {
      setLoading(false);
    }
  }, [page, search, estado]);

  useEffect(() => {
    const timer = setTimeout(load, 300);
    return () => clearTimeout(timer);
  }, [load]);

  async function handleSave(data: EmpresaCreateRequest) {
    if (editTarget) {
      await updateEmpresa(editTarget.id, data);
    } else {
      await createEmpresa(data);
    }
    setPage(1);
    await load();
  }

  async function handleDelete(empresa: Empresa) {
    if (!confirm(`¿Eliminar la empresa "${empresa.nombre}"?`)) return;
    setDeletingId(empresa.id);
    setDeleteError("");
    try {
      await deleteEmpresa(empresa.id);
      await load();
    } catch (err: unknown) {
      setDeleteError(
        err instanceof Error ? err.message : "Error al eliminar empresa",
      );
    } finally {
      setDeletingId(null);
    }
  }

  function openNew() {
    setEditTarget(null);
    setFormOpen(true);
  }

  function openEdit(empresa: Empresa) {
    setEditTarget(empresa);
    setFormOpen(true);
  }

  function openDetail(empresa: Empresa) {
    setDetailTarget(empresa);
    setDetailOpen(true);
  }

  const totalPages = Math.ceil(total / limit);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Empresas</h1>
          <p className="text-muted-foreground">
            Gestiona las empresas registradas en el sistema
          </p>
        </div>
        <Button onClick={openNew}>
          <Plus className="h-4 w-4 mr-2" />
          Nueva Empresa
        </Button>
      </div>

      {loadError && <p className="text-sm text-destructive">{loadError}</p>}
      {deleteError && <p className="text-sm text-destructive">{deleteError}</p>}

      <Card>
        <CardHeader>
          <div className="flex flex-col sm:flex-row sm:items-center gap-3 justify-between">
            <div>
              <CardTitle>Lista de Empresas</CardTitle>
              <CardDescription>{total} empresa(s) en total</CardDescription>
            </div>
            <div className="flex gap-2">
              <div className="relative w-52">
                <Search className="absolute left-2 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder="Buscar por nombre o RUC"
                  className="pl-8"
                  value={search}
                  onChange={(e) => {
                    setSearch(e.target.value);
                    setPage(1);
                  }}
                />
              </div>
              <Select
                className="w-36"
                value={estado}
                onChange={(e) => {
                  setEstado(e.target.value);
                  setPage(1);
                }}
              >
                <option value="todos">Todos</option>
                <option value="activo">Activo</option>
                <option value="inactivo">Inactivo</option>
              </Select>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>RUC</TableHead>
                <TableHead>Nombre</TableHead>
                <TableHead>Nombre Comercial</TableHead>
                <TableHead>Teléfono contacto</TableHead>
                <TableHead>Estado</TableHead>
                <TableHead className="text-right">Acciones</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell
                    colSpan={6}
                    className="text-center text-muted-foreground py-8"
                  >
                    Cargando...
                  </TableCell>
                </TableRow>
              ) : empresas.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={6}
                    className="text-center text-muted-foreground py-8"
                  >
                    <Building2 className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    No hay empresas registradas
                  </TableCell>
                </TableRow>
              ) : (
                empresas.map((empresa) => (
                  <TableRow key={empresa.id}>
                    <TableCell className="font-mono text-sm">
                      {empresa.ruc}
                    </TableCell>
                    <TableCell className="font-medium">
                      {empresa.nombre}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {empresa.nombre_comercial ?? "—"}
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {empresa.telefono_contacto ?? "—"}
                    </TableCell>
                    <TableCell>
                      <Badge
                        variant={empresa.activo ? "default" : "secondary"}
                        className={empresa.activo ? "bg-green-500" : ""}
                      >
                        {empresa.activo ? "Activa" : "Inactiva"}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-1">
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => router.push(`/empresas/${empresa.id}/telefonos`)}
                          title="Ver teléfonos"
                        >
                          <KeyRound className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => openDetail(empresa)}
                          title="Ver detalle"
                        >
                          <Eye className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => openEdit(empresa)}
                          title="Editar"
                        >
                          <Pencil className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          onClick={() => handleDelete(empresa)}
                          disabled={deletingId === empresa.id}
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

      <EmpresaFormModal
        open={formOpen}
        onClose={() => setFormOpen(false)}
        onSave={handleSave}
        empresa={editTarget}
      />

      <EmpresaDetailModal
        open={detailOpen}
        onClose={() => setDetailOpen(false)}
        empresa={detailTarget}
      />
    </div>
  );
}

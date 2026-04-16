"use client";

import { useEffect, useState } from "react";
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
import { LayoutGrid } from "lucide-react";
import { getModules, type Module } from "@/lib/api";

export default function ModulesPage() {
  const [modules, setModules] = useState<Module[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    async function load() {
      try {
        const resp = await getModules();
        setModules(resp.modules || []);
      } catch (err: unknown) {
        setError(err instanceof Error ? err.message : "Error al cargar módulos");
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Módulos</h1>
        <p className="text-muted-foreground">
          Lista de módulos disponibles en el sistema
        </p>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      <Card>
        <CardHeader>
          <CardTitle>Lista de Módulos</CardTitle>
          <CardDescription>
            {modules.length} módulo(s) configurado(s)
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nombre</TableHead>
                <TableHead>Slug</TableHead>
                <TableHead>Descripción</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {loading ? (
                <TableRow>
                  <TableCell
                    colSpan={3}
                    className="text-center text-muted-foreground py-8"
                  >
                    Cargando...
                  </TableCell>
                </TableRow>
              ) : modules.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={3}
                    className="text-center text-muted-foreground py-8"
                  >
                    <LayoutGrid className="h-8 w-8 mx-auto mb-2 opacity-40" />
                    No hay módulos configurados
                  </TableCell>
                </TableRow>
              ) : (
                modules.map((mod) => (
                  <TableRow key={mod.id}>
                    <TableCell className="font-medium">{mod.name}</TableCell>
                    <TableCell>
                      <code className="bg-muted px-2 py-1 rounded text-sm">
                        {mod.slug}
                      </code>
                    </TableCell>
                    <TableCell className="text-muted-foreground">
                      {mod.description || "—"}
                    </TableCell>
                  </TableRow>
                ))
              )}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  );
}
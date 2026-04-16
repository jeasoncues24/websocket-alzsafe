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
import { Badge } from "@/components/ui/badge";
import { Shield, ShieldCheck } from "lucide-react";
import { getRoles, type Role } from "@/lib/api";

export default function RolesPage() {
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    async function load() {
      try {
        const resp = await getRoles();
        setRoles(resp.roles || []);
      } catch (err: unknown) {
        setError(err instanceof Error ? err.message : "Error al cargar roles");
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Roles</h1>
        <p className="text-muted-foreground">
          Lista de roles disponibles en el sistema
        </p>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      <Card>
        <CardHeader>
          <CardTitle>Lista de Roles</CardTitle>
          <CardDescription>
            {roles.length} rol(es) configurado(s)
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Nombre</TableHead>
                <TableHead>Descripción</TableHead>
                <TableHead>Tipo</TableHead>
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
              ) : roles.length === 0 ? (
                <TableRow>
                  <TableCell
                    colSpan={3}
                    className="text-center text-muted-foreground py-8"
                  >
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
                      {role.is_root && (
                        <Badge variant="secondary" className="bg-purple-500/10 text-purple-500">
                          Super Admin
                        </Badge>
                      )}
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
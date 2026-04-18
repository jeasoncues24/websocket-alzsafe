"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Building2,
  Calendar,
  ArrowRight,
  Hash,
  KeyRound,
  MapPin,
  MessageSquare,
  Phone,
  Waves,
} from "lucide-react";
import {
  getAdminMessages,
  getAdminSessions,
  getAdminEmpresaTelefonos,
  type AdminMessage,
  type AdminTelefono,
  type Empresa,
  type SessionInfo,
} from "@/lib/api";

interface Props {
  open: boolean;
  onClose: () => void;
  empresa: Empresa | null;
}

function formatDateTime(value: string | null) {
  if (!value) return "—";
  return new Date(value).toLocaleString("es-PE", {
    dateStyle: "medium",
    timeStyle: "short",
  });
}

export function EmpresaDetailModal({ open, onClose, empresa }: Props) {
  const router = useRouter();
  const [session, setSession] = useState<SessionInfo | null>(null);
  const [messages, setMessages] = useState<AdminMessage[]>([]);
  const [telefonos, setTelefonos] = useState<AdminTelefono[]>([]);
  const [loading, setLoading] = useState(false);
  const [tab, setTab] = useState("overview");

  useEffect(() => {
    if (!open || !empresa) return;

    const currentEmpresa = empresa;

    let cancelled = false;

    async function load() {
      setLoading(true);
      setSession(null);
      setMessages([]);
      setTelefonos([]);
      setTab("overview");

      try {
        const [sessionsData, messagesData, telefonosData] = await Promise.all([
          getAdminSessions(),
          getAdminMessages({ account_id: currentEmpresa.ruc, limit: 5 }),
          getAdminEmpresaTelefonos(currentEmpresa.id),
        ]);

        if (cancelled) return;

        setSession(
          sessionsData.sessions.find((item) => item.account_id === currentEmpresa.ruc) ?? null,
        );
        setMessages(messagesData.messages ?? []);
        setTelefonos(telefonosData.telefonos ?? []);
      } catch {
        if (!cancelled) {
          setSession(null);
          setMessages([]);
          setTelefonos([]);
        }
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    load();

    return () => {
      cancelled = true;
    };
  }, [open, empresa]);

  if (!empresa) return null;

  return (
    <>
      <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
        <DialogContent className="sm:max-w-3xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Building2 className="h-5 w-5" />
              {empresa.nombre}
            </DialogTitle>
            <DialogDescription>
              Detalle de empresa, actividad operativa y acceso a teléfonos/API keys.
            </DialogDescription>
            <div className="flex justify-end">
              <Button
                variant="outline"
                size="sm"
                onClick={() => router.push(`/empresas/${empresa.id}/telefonos`)}
              >
                <KeyRound className="mr-2 h-4 w-4" />
                Ver teléfonos
                <ArrowRight className="ml-2 h-4 w-4" />
              </Button>
            </div>
          </DialogHeader>

          <Tabs key={`${empresa.id}-${tab}`} defaultValue={tab} onValueChange={setTab} className="space-y-4">
            <TabsList className="grid w-full grid-cols-3">
              <TabsTrigger value="overview">Resumen</TabsTrigger>
              <TabsTrigger value="phones">Teléfonos</TabsTrigger>
              <TabsTrigger value="activity">Actividad</TabsTrigger>
            </TabsList>

            <TabsContent value="overview" className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle>Información general</CardTitle>
                  <CardDescription>Datos principales de la empresa seleccionada.</CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="grid gap-4 sm:grid-cols-2 text-sm">
                    <div className="space-y-1">
                      <div className="flex items-center gap-1 text-muted-foreground">
                        <Hash className="h-3.5 w-3.5" />
                        RUC
                      </div>
                      <p className="font-medium font-mono">{empresa.ruc}</p>
                    </div>

                    <div className="space-y-1">
                      <div className="text-muted-foreground text-xs">Estado</div>
                      <Badge
                        variant={empresa.activo ? "default" : "secondary"}
                        className={empresa.activo ? "bg-green-500" : ""}
                      >
                        {empresa.activo ? "Activa" : "Inactiva"}
                      </Badge>
                    </div>

                    <div className="space-y-1">
                      <div className="text-muted-foreground text-xs">Nombre comercial</div>
                      <p className="font-medium">{empresa.nombre_comercial ?? "—"}</p>
                    </div>

                    <div className="space-y-1">
                      <div className="flex items-center gap-1 text-muted-foreground">
                        <Phone className="h-3.5 w-3.5" />
                        Teléfono contacto
                      </div>
                      <p className="font-medium">{empresa.telefono_contacto ?? "—"}</p>
                    </div>

                    <div className="space-y-1 sm:col-span-2">
                      <div className="flex items-center gap-1 text-muted-foreground">
                        <MapPin className="h-3.5 w-3.5" />
                        Dirección
                      </div>
                      <p className="font-medium">{empresa.direccion ?? "—"}</p>
                    </div>

                    <div className="space-y-1">
                      <div className="flex items-center gap-1 text-muted-foreground">
                        <KeyRound className="h-3.5 w-3.5" />
                        Teléfonos registrados
                      </div>
                      <p className="font-medium">{telefonos.length}</p>
                    </div>

                    <div className="space-y-1">
                      <div className="flex items-center gap-1 text-muted-foreground">
                        <Calendar className="h-3.5 w-3.5" />
                        Actualizada
                      </div>
                      <p className="font-medium">
                        {new Date(empresa.updated_at).toLocaleDateString("es-PE")}
                      </p>
                    </div>
                  </div>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="phones" className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle>Teléfonos y API keys</CardTitle>
                  <CardDescription>
                    Cada teléfono administra sus propias API keys y consumo.
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-3">
                  {loading ? (
                    <div className="space-y-3">
                      <Skeleton className="h-16 w-full" />
                      <Skeleton className="h-16 w-full" />
                    </div>
                  ) : telefonos.length > 0 ? (
                    telefonos.map((telefono) => (
                      <div
                        key={telefono.id}
                        className="flex flex-col gap-3 rounded-xl border p-4 sm:flex-row sm:items-center sm:justify-between"
                      >
                        <div className="space-y-1">
                          <div className="flex items-center gap-2">
                            <Phone className="h-4 w-4 text-muted-foreground" />
                            <span className="font-medium">{telefono.numero_completo}</span>
                            <Badge variant={telefono.status === "active" ? "default" : "secondary"}>
                              {telefono.status}
                            </Badge>
                          </div>
                          <p className="text-sm text-muted-foreground">
                            {telefono.codigo_pais} {telefono.numero}
                          </p>
                        </div>

                        <Button
                          variant="outline"
                          onClick={() => router.push(`/empresas/${empresa.id}/telefonos/${telefono.id}/api-keys`)}
                        >
                          <KeyRound className="mr-2 h-4 w-4" />
                          Gestionar API keys
                        </Button>
                      </div>
                    ))
                  ) : (
                    <p className="text-sm text-muted-foreground">No hay teléfonos registrados para esta empresa.</p>
                  )}
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="activity" className="space-y-4">
              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <Waves className="h-4 w-4" />
                    Sesión WhatsApp
                  </CardTitle>
                  <CardDescription>Estado operativo y conexión actual.</CardDescription>
                </CardHeader>
                <CardContent>
                  {loading ? (
                    <div className="space-y-3">
                      <Skeleton className="h-10 w-full" />
                      <Skeleton className="h-10 w-2/3" />
                    </div>
                  ) : session ? (
                    <div className="flex items-center justify-between rounded-md border px-3 py-2 text-sm">
                      <div className="space-y-0.5">
                        <p className="font-medium">{session.account_id}</p>
                        <p className="text-xs text-muted-foreground">
                          Actualizado {formatDateTime(session.updated_at)}
                        </p>
                      </div>
                      <Badge variant={session.status === "active" ? "default" : "secondary"}>
                        {session.status}
                      </Badge>
                    </div>
                  ) : (
                    <p className="text-sm text-muted-foreground">Sin sesión registrada</p>
                  )}
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle className="flex items-center gap-2">
                    <MessageSquare className="h-4 w-4" />
                    Últimos mensajes
                  </CardTitle>
                  <CardDescription>Vista rápida de la actividad reciente.</CardDescription>
                </CardHeader>
                <CardContent>
                  {loading ? (
                    <div className="space-y-3">
                      <Skeleton className="h-16 w-full" />
                      <Skeleton className="h-16 w-full" />
                      <Skeleton className="h-16 w-5/6" />
                    </div>
                  ) : messages.length > 0 ? (
                    <div className="space-y-3">
                      {messages.map((msg) => (
                        <div key={msg.id} className="rounded-md border px-3 py-2 text-xs">
                          <div className="flex items-center justify-between gap-2">
                            <span className="font-medium">{msg.to}</span>
                            <Badge variant={msg.status === "sent" ? "default" : "outline"}>
                              {msg.status}
                            </Badge>
                          </div>
                          <p className="mt-1 text-muted-foreground">{msg.content}</p>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className="text-sm text-muted-foreground">Sin mensajes recientes</p>
                  )}
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        </DialogContent>
      </Dialog>

    </>
  );
}

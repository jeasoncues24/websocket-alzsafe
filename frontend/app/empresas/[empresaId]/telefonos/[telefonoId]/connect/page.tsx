"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, CheckCircle2, Loader2, QrCode, Smartphone, Wifi, WifiOff } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { QRRender } from "@/components/qr/qr-render";
import { buildAdminWsUrl, connectEmpresaTelefono, type EmpresaTelefonoSessionData } from "@/lib/api";

type WsPayload = {
  event?: string;
  data?: Record<string, unknown>;
};

export default function AdminPhoneConnectPage() {
  const router = useRouter();
  const params = useParams<{ empresaId: string; telefonoId: string }>();
  const empresaId = Number(params?.empresaId);
  const telefonoId = Number(params?.telefonoId);

  const wsRef = useRef<WebSocket | null>(null);

  const [phone, setPhone] = useState<EmpresaTelefonoSessionData | null>(null);
  const [countdown, setCountdown] = useState(300);
  const [wsConnected, setWsConnected] = useState(false);
  const [starting, setStarting] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");

  const closeSocket = useCallback(() => {
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
  }, []);

  const mergePhone = useCallback((next: Partial<EmpresaTelefonoSessionData>) => {
    setPhone((current) => ({
      telefono_id: next.telefono_id ?? current?.telefono_id ?? telefonoId,
      numeroCompleto: next.numeroCompleto ?? current?.numeroCompleto ?? "",
      status: next.status ?? current?.status ?? "initializing",
      lastConnected: next.lastConnected ?? current?.lastConnected ?? null,
      qr_string: next.qr_string ?? current?.qr_string,
      expires_in: next.expires_in ?? current?.expires_in,
    }));
  }, [telefonoId]);

  const openSocket = useCallback(() => {
    const token = localStorage.getItem("admin_token");
    if (!token) {
      router.push("/login");
      return;
    }

    if (!Number.isFinite(telefonoId) || telefonoId <= 0) {
      setError("ID de teléfono inválido");
      return;
    }

    closeSocket();
    setError("");

    try {
      const ws = new WebSocket(
        buildAdminWsUrl(`/api/admin/telefonos/${telefonoId}/connect/ws`, token),
      );
      wsRef.current = ws;

      ws.onopen = () => {
        setWsConnected(true);
        setStarting(true);
        setSuccess("");
      };

      ws.onmessage = (event) => {
        try {
          const payload = JSON.parse(event.data as string) as WsPayload;
          const type = payload.event || "";
          const data = payload.data ?? {};

          if (type === "phone-info") {
            mergePhone({
              telefono_id: Number(data.telefono_id ?? telefonoId),
              numeroCompleto: String(data.numeroCompleto ?? ""),
              status: String(data.status ?? "initializing"),
              qr_string: String(data.qr_string ?? ""),
              lastConnected: data.lastConnected ? String(data.lastConnected) : null,
            });
            setStarting(true);
            return;
          }

          if (type.startsWith("qr-")) {
            mergePhone({
              telefono_id: telefonoId,
              status: "qr_pending",
              qr_string: String(data.qrString ?? data.qr_string ?? ""),
            });
            setCountdown(300);
            setStarting(true);
            setSuccess("");
            setError("");
            return;
          }

          if (type.startsWith("active-")) {
            const isActive = Boolean(data.isActive);
            const requiresNewQR = Boolean(data.requiresNewQR);
            mergePhone({
              telefono_id: telefonoId,
              status: isActive ? "active" : String(data.reason ?? "disconnected"),
              qr_string: isActive ? undefined : "",
            });
            if (isActive) {
              setCountdown(0);
              setSuccess(String(data.message ?? "Teléfono conectado"));
              setError("");
              setStarting(false);
            } else {
              const detail = data.detail ? ` (${data.detail})` : "";
              setError(String(data.message ?? "Conexión cerrada") + detail);
              setSuccess("");
              setStarting(false);
              // If a new QR is needed, reopen the socket after a brief pause
              if (requiresNewQR) {
                setTimeout(() => {
                  openSocket();
                }, 2000);
              }
            }
            return;
          }

          if (type === "error" || type === "error-event") {
            setError(String(data.message ?? "Error de conexión"));
            setSuccess("");
            return;
          }
        } catch {
          setError("Respuesta WS inválida");
        }
      };

      ws.onerror = () => {
        setError("Error de conexión WS");
      };

      ws.onclose = () => {
        setWsConnected(false);
        setStarting(false);
        if (wsRef.current === ws) {
          wsRef.current = null;
        }
      };
    } catch {
      setError("No se pudo abrir el WebSocket");
    }
  }, [closeSocket, mergePhone, router, telefonoId]);

  useEffect(() => {
    const token = localStorage.getItem("admin_token");
    if (!token) {
      router.push("/login");
      return;
    }

    openSocket();
    return () => closeSocket();
  }, [closeSocket, openSocket, router]);

  useEffect(() => {
    if (!phone?.qr_string || phone.status !== "qr_pending" || countdown <= 0) {
      return;
    }

    const timer = window.setTimeout(() => setCountdown((value) => value - 1), 1000);
    return () => window.clearTimeout(timer);
  }, [countdown, phone?.qr_string, phone?.status]);

  const startFallback = async () => {
    setStarting(true);
    setError("");
    try {
      const response = await connectEmpresaTelefono(telefonoId);
      if (response.ok && response.data) {
        mergePhone(response.data);
        setCountdown(response.data.expires_in ?? 300);
        setSuccess("Conexión iniciada manualmente");
        if (!wsRef.current) {
          openSocket();
        }
      } else {
        setError(response.message || "Error iniciando la conexión");
      }
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Error de conexión");
    } finally {
      setStarting(false);
    }
  };

  const formatTime = (seconds: number) => {
    const m = Math.floor(seconds / 60);
    const s = seconds % 60;
    return `${m}:${s.toString().padStart(2, "0")}`;
  };

  if (!Number.isFinite(telefonoId) || telefonoId <= 0) {
    return <div className="p-6">ID de teléfono requerido</div>;
  }

  const status = phone?.status || (wsConnected ? "initializing" : "disconnected");
  const qrString = phone?.qr_string || "";

  return (
    <div className="container mx-auto max-w-xl p-6">
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Smartphone className="h-5 w-5" />
            Conectar WhatsApp
          </CardTitle>
          <CardDescription>
            La sesión se inicia por WebSocket y el QR aparece en tiempo real.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {error && (
            <Alert variant="destructive">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {success && !error && (
            <Alert>
              <CheckCircle2 className="h-4 w-4" />
              <AlertDescription>{success}</AlertDescription>
            </Alert>
          )}

          <div className="flex flex-wrap items-center gap-2 text-sm">
            <Badge variant={wsConnected ? "default" : "secondary"}>
              {wsConnected ? "WS activo" : "WS inactivo"}
            </Badge>
            <Badge variant={status === "active" ? "default" : "secondary"}>{status}</Badge>
            <span className="text-muted-foreground">Empresa #{empresaId}</span>
            <span className="text-muted-foreground">Teléfono #{telefonoId}</span>
            {phone?.numeroCompleto && <span className="text-muted-foreground">{phone.numeroCompleto}</span>}
          </div>

          {starting && !qrString && status !== "active" ? (
            <div className="text-center">
              <Loader2 className="mx-auto h-8 w-8 animate-spin text-muted-foreground" />
              <p className="mt-2 text-sm text-muted-foreground">Iniciando conexión...</p>
            </div>
          ) : status === "active" ? (
            <div className="rounded-lg border p-4 text-center">
              <CheckCircle2 className="mx-auto h-8 w-8 text-green-500" />
              <p className="mt-2 font-medium">Teléfono conectado</p>
            </div>
          ) : qrString ? (
            <div className="space-y-4 text-center">
              <QRRender value={qrString} size={220} title={`QR ${phone?.numeroCompleto || telefonoId}`} />
              <p className="text-sm text-muted-foreground">Código QR válido por {formatTime(countdown)}</p>
              <p className="text-xs text-muted-foreground">Estado: {status}</p>
              <p className="text-sm text-muted-foreground">
                Escanea el código con WhatsApp y espera a que cambie a activo.
              </p>
            </div>
          ) : (
            <div className="rounded-lg border border-dashed p-6 text-center text-sm text-muted-foreground">
              {wsConnected ? "Esperando QR..." : "No hay conexión WS activa."}
            </div>
          )}

          <div className="grid gap-2 sm:grid-cols-2">
            <Button onClick={startFallback} disabled={starting}>
              {starting ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <QrCode className="mr-2 h-4 w-4" />}
              Iniciar / regenerar
            </Button>
            <Button variant="outline" onClick={openSocket}>
              {wsConnected ? <Wifi className="mr-2 h-4 w-4" /> : <WifiOff className="mr-2 h-4 w-4" />}
              Reconectar WS
            </Button>
          </div>

          <Button variant="ghost" className="w-full" onClick={() => router.push(`/empresas/${empresaId}/telefonos`)}>
            <ArrowLeft className="mr-2 h-4 w-4" />
            Volver a teléfonos
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}

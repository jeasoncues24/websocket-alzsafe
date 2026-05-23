"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { ArrowLeft, CheckCircle2, Loader2, RefreshCw, Share2, Smartphone } from "lucide-react";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { QRRender } from "@/components/qr/qr-render";
import { buildAdminWsUrl, connectEmpresaTelefono, generateQRLink, getAdminSessions, type EmpresaTelefonoSessionData } from "@/lib/api";

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
  const statusPollRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const [phone, setPhone] = useState<EmpresaTelefonoSessionData | null>(null);
  const [countdown, setCountdown] = useState(60);
  const [wsConnected, setWsConnected] = useState(false);
  const [starting, setStarting] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState("");
  const [shareUrl, setShareUrl] = useState("");
  const [sharing, setSharing] = useState(false);
  const [copied, setCopied] = useState(false);

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

  const stopStatusPoll = useCallback(() => {
    if (statusPollRef.current) {
      clearInterval(statusPollRef.current);
      statusPollRef.current = null;
    }
  }, []);

  const startStatusPoll = useCallback(() => {
    stopStatusPoll();
    statusPollRef.current = setInterval(async () => {
      try {
        const res = await getAdminSessions();
        const session = res.sessions?.find((s) => s.telefono_id === telefonoId);
        if (session?.status === "active") {
          mergePhone({ status: "active" });
          setSuccess("Teléfono conectado");
          setError("");
          stopStatusPoll();
        }
      } catch {
        // ignorar errores de poll
      }
    }, 3000);
  }, [telefonoId, mergePhone, stopStatusPoll]);

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
        stopStatusPoll();
        setWsConnected(true);
        setStarting(true);
        setSuccess("");
      };

      ws.onmessage = (event) => {
        try {
          const payload = JSON.parse(event.data as string) as WsPayload;
          const type = payload.event || "";
          const data = payload.data ?? {};

          // Keepalive ping del servidor — ignorar silenciosamente
          if (type === "ping") return;

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
            const expiresIn = typeof data.expires_in === "number" ? data.expires_in : 60;
            setCountdown(expiresIn);
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
          // Si el teléfono no está activo todavía, hacer poll REST para detectar
          // cuando se conecta por otra vía (ej: enlace QR compartido).
          startStatusPoll();
        }
      };
    } catch {
      setError("No se pudo abrir el WebSocket");
    }
  }, [closeSocket, mergePhone, router, telefonoId, startStatusPoll, stopStatusPoll]);

  useEffect(() => {
    const token = localStorage.getItem("admin_token");
    if (!token) {
      router.push("/login");
      return;
    }

    openSocket();
    return () => {
      closeSocket();
      stopStatusPoll();
    };
  }, [closeSocket, openSocket, router, stopStatusPoll]);

  // Detener el poll cuando el teléfono llega a activo (por WS o por poll)
  useEffect(() => {
    if (phone?.status === "active") {
      stopStatusPoll();
    }
  }, [phone?.status, stopStatusPoll]);

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
        setCountdown(response.data.expires_in ?? 60);
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

  const handleShare = async () => {
    setSharing(true);
    try {
      const data = await generateQRLink(telefonoId);
      if (data.ok && data.token) {
        setShareUrl(`${window.location.origin}/qr?token=${data.token}`);
      }
    } finally {
      setSharing(false);
    }
  };

  const handleCopy = async () => {
    await navigator.clipboard.writeText(shareUrl);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
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

          {phone?.numeroCompleto && (
            <p className="font-mono text-sm text-muted-foreground">{phone.numeroCompleto}</p>
          )}

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

          <div className="flex flex-col gap-2">
            {status !== "active" && (
              wsConnected ? (
                <Button variant="outline" onClick={() => {
                  closeSocket();
                  router.push(`/empresas/${empresaId}/telefonos`);
                }}>
                  Cancelar
                </Button>
              ) : (
                <Button onClick={openSocket} disabled={starting}>
                  {starting
                    ? <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    : <RefreshCw className="mr-2 h-4 w-4" />}
                  {status === "disconnected" ? "Iniciar conexión" : "Reconectar"}
                </Button>
              )
            )}
            {error && (
              <Button variant="ghost" size="sm" onClick={startFallback} disabled={starting}>
                Forzar conexión REST
              </Button>
            )}
            {status !== "active" && (
              <Button variant="outline" size="sm" onClick={handleShare} disabled={sharing}>
                {sharing
                  ? <Loader2 className="mr-2 h-3 w-3 animate-spin" />
                  : <Share2 className="mr-2 h-3 w-3" />}
                Compartir enlace QR
              </Button>
            )}
            {shareUrl && (
              <div className="rounded border p-2 text-xs space-y-1">
                <p className="break-all text-muted-foreground">{shareUrl}</p>
                <Button size="sm" variant="ghost" onClick={handleCopy} className="h-6 text-xs">
                  {copied ? "✓ Copiado" : "Copiar enlace"}
                </Button>
              </div>
            )}
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

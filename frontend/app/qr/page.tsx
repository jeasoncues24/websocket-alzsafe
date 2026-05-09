"use client";

import { Suspense, useCallback, useEffect, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { QRRender } from "@/components/qr/qr-render";
import { buildAdminWsUrl } from "@/lib/api";
import { CheckCircle2, Loader2, WifiOff } from "lucide-react";

function QRPageContent() {
  const searchParams = useSearchParams();
  const token = searchParams.get("token") ?? "";

  const wsRef = useRef<WebSocket | null>(null);
  const [qrString, setQrString] = useState("");
  const [countdown, setCountdown] = useState(60);
  const [status, setStatus] = useState<"connecting" | "qr" | "connected" | "error" | "closed">("connecting");
  const [errorMsg, setErrorMsg] = useState("");

  const connect = useCallback(() => {
    if (!token) {
      setStatus("error");
      setErrorMsg("Enlace inválido o expirado");
      return;
    }

    const ws = new WebSocket(buildAdminWsUrl("/api/service/v1/ws", token));
    wsRef.current = ws;

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data as string) as { type: string; data?: Record<string, unknown> };
        if (msg.type === "ping") return;
        if (msg.type === "qr") {
          const d = msg.data ?? {};
          setQrString(String(d.qrString ?? d.qr_string ?? ""));
          setCountdown(typeof d.expires_in === "number" ? d.expires_in : 60);
          setStatus("qr");
          return;
        }
        if (msg.type === "connected") {
          const d = msg.data ?? {};
          if (d.isActive) {
            setStatus("connected");
          } else {
            setStatus("error");
            setErrorMsg(String(d.message ?? d.reason ?? "Sesión cerrada"));
          }
          return;
        }
        if (msg.type === "error") {
          setStatus("error");
          setErrorMsg(String(msg.data?.message ?? "Error de conexión"));
        }
      } catch {
        // ignorar mensajes malformados
      }
    };

    ws.onerror = () => {
      setStatus("error");
      setErrorMsg("Error de conexión WebSocket");
    };

    ws.onclose = () => {
      wsRef.current = null;
      setStatus((prev) => (prev === "connected" ? "connected" : "closed"));
    };
  }, [token]);

  useEffect(() => {
    connect();
    return () => {
      wsRef.current?.close();
    };
  }, [connect]);

  useEffect(() => {
    if (status !== "qr" || !qrString || countdown <= 0) return;
    const t = setTimeout(() => setCountdown((n) => n - 1), 1000);
    return () => clearTimeout(t);
  }, [countdown, qrString, status]);

  const formatTime = (s: number) => `${Math.floor(s / 60)}:${(s % 60).toString().padStart(2, "0")}`;

  return (
    <div className="min-h-screen flex items-center justify-center bg-muted/30 p-4">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <CardTitle>Conectar WhatsApp</CardTitle>
          <CardDescription>Escanea el código QR con tu WhatsApp</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {status === "connecting" && (
            <div className="text-center py-8">
              <Loader2 className="mx-auto h-8 w-8 animate-spin text-muted-foreground" />
              <p className="mt-2 text-sm text-muted-foreground">Conectando...</p>
            </div>
          )}

          {status === "qr" && qrString && (
            <div className="space-y-3 text-center">
              <QRRender value={qrString} size={220} title="QR WhatsApp" />
              <p className="text-sm text-muted-foreground">Válido por {formatTime(countdown)}</p>
              <p className="text-xs text-muted-foreground">
                Abre WhatsApp → Dispositivos vinculados → Vincular dispositivo
              </p>
            </div>
          )}

          {status === "connected" && (
            <div className="text-center py-8">
              <CheckCircle2 className="mx-auto h-10 w-10 text-green-500" />
              <p className="mt-2 font-medium">¡Teléfono conectado!</p>
              <p className="text-sm text-muted-foreground mt-1">Puedes cerrar esta página</p>
            </div>
          )}

          {(status === "error" || status === "closed") && (
            <Alert variant="destructive">
              <WifiOff className="h-4 w-4" />
              <AlertDescription>
                {errorMsg || (status === "closed" ? "Enlace expirado o cerrado" : "Error desconocido")}
              </AlertDescription>
            </Alert>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export default function QRPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen flex items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      }
    >
      <QRPageContent />
    </Suspense>
  );
}

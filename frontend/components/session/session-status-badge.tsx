"use client"

import { Badge } from "@/components/ui/badge"
import { AlertTriangle, Loader2 } from "lucide-react"

interface SessionStatusBadgeProps {
  status: string
  reconnecting?: boolean
  mismatch?: boolean
  runtimeConnected?: boolean
  showRuntime?: boolean
  className?: string
}

export function SessionStatusBadge({
  status,
  reconnecting = false,
  mismatch = false,
  runtimeConnected = false,
  showRuntime = false,
  className = "",
}: SessionStatusBadgeProps) {
  const getStatusBadge = () => {
    switch (status) {
      case "active":
        return (
          <Badge
            variant="outline"
            className="border-emerald-600/20 bg-emerald-500/10 text-emerald-700 dark:border-emerald-500/20 dark:bg-emerald-950/40 dark:text-emerald-300 font-medium px-2 py-0.5 text-xs"
          >
            <span className="h-1.5 w-1.5 rounded-full bg-emerald-500 mr-1.5 animate-pulse" />
            Activa
          </Badge>
        )
      case "qr_pending":
        return (
          <Badge
            variant="outline"
            className="border-border bg-muted/60 text-foreground font-medium px-2 py-0.5 text-xs"
          >
            QR Pendiente
          </Badge>
        )
      case "initializing":
        return (
          <Badge
            variant="outline"
            className="border-border bg-muted/60 text-muted-foreground font-medium px-2 py-0.5 text-xs"
          >
            <Loader2 className="h-3 w-3 mr-1 animate-spin text-muted-foreground" />
            Conectando
          </Badge>
        )
      case "disconnected":
        return (
          <Badge
            variant="outline"
            className="border-border bg-muted/30 text-muted-foreground font-normal px-2 py-0.5 text-xs"
          >
            Desconectada
          </Badge>
        )
      case "client_outdated":
        return (
          <Badge
            variant="outline"
            className="border-border bg-muted/50 text-foreground font-medium px-2 py-0.5 text-xs"
          >
            Librería desactualizada
          </Badge>
        )
      default:
        return (
          <Badge variant="outline" className="text-muted-foreground px-2 py-0.5 text-xs">
            {status || "Inactiva"}
          </Badge>
        )
    }
  }

  return (
    <div className={`inline-flex items-center gap-1.5 flex-wrap ${className}`}>
      {getStatusBadge()}

      {reconnecting && (
        <Badge
          variant="outline"
          className="border-emerald-600/30 text-emerald-700 dark:text-emerald-400 text-[11px] px-1.5 py-0"
        >
          <Loader2 className="h-2.5 w-2.5 mr-1 animate-spin" />
          Reconectando
        </Badge>
      )}

      {mismatch && !reconnecting && (
        <Badge
          variant="outline"
          className="border-border text-muted-foreground text-[11px] px-1.5 py-0 flex items-center gap-1"
          title="Inconsistencia detectada entre el estado de BD y la conexión en memoria"
        >
          <AlertTriangle className="h-3 w-3 text-muted-foreground" />
          Inconsistente
        </Badge>
      )}

      {showRuntime && (
        <span
          className={`inline-flex items-center gap-1 text-[11px] font-medium ${
            runtimeConnected ? "text-emerald-600 dark:text-emerald-400" : "text-muted-foreground"
          }`}
          title={runtimeConnected ? "Runtime conectado en memoria" : "Runtime desconectado"}
        >
          <span
            className={`h-1.5 w-1.5 rounded-full ${
              runtimeConnected ? "bg-emerald-500" : "bg-muted-foreground/40"
            }`}
          />
          {runtimeConnected ? "Online" : "Offline"}
        </span>
      )}
    </div>
  )
}
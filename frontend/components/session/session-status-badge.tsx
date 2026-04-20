"use client";

import { Badge } from "@/components/ui/badge";
import { AlertTriangle, Wifi, WifiOff } from "lucide-react";

interface SessionStatusBadgeProps {
  statusDb: string;
  statusRuntime?: string;
  runtimeConnected?: boolean;
  mismatch?: boolean;
  mismatchReason?: string;
}

export function SessionStatusBadge({
  statusDb,
  statusRuntime,
  runtimeConnected = false,
  mismatch = false,
  mismatchReason,
}: SessionStatusBadgeProps) {
  const getDbBadge = () => {
    switch (statusDb) {
      case "active":
        return <Badge variant="default" className="bg-green-500">DB: Activo</Badge>;
      case "qr_pending":
        return <Badge variant="secondary">DB: QR Pendiente</Badge>;
      case "initializing":
        return <Badge variant="secondary">DB: Conectando</Badge>;
      case "disconnected":
        return <Badge variant="outline">DB: Desconectado</Badge>;
      default:
        return <Badge variant="outline">DB: {statusDb}</Badge>;
    }
  };

  const getRuntimeBadge = () => {
    if (runtimeConnected) {
      return (
        <Badge variant="default" className="bg-blue-500">
          <Wifi className="h-3 w-3 mr-1" />
          Runtime: Conectado
        </Badge>
      );
    }
    return (
      <Badge variant="outline" className="text-muted-foreground">
        <WifiOff className="h-3 w-3 mr-1" />
        Runtime: Desconectado
      </Badge>
    );
  };

  const getMismatchAlert = () => {
    if (!mismatch) return null;
    return (
      <div className="flex items-center gap-1 text-amber-600 text-xs">
        <AlertTriangle className="h-3 w-3" />
        <span>Inconsistencia: {mismatchReason}</span>
      </div>
    );
  };

  if (statusRuntime !== undefined || runtimeConnected) {
    return (
      <div className="flex flex-col gap-1">
        <div className="flex items-center gap-2">
          {getDbBadge()}
          {getRuntimeBadge()}
        </div>
        {getMismatchAlert()}
      </div>
    );
  }

  return getDbBadge();
}

export function getConnectionStatusLabel(statusDb: string, runtimeConnected?: boolean): string {
  if (runtimeConnected === undefined) {
    return statusDb === "active" ? "Conectado" : "Desconectado";
  }

  if (statusDb === "active" && runtimeConnected) {
    return "Conectado";
  }
  if (statusDb === "active" && !runtimeConnected) {
    return "Activo (DB) sin conexión runtime";
  }
  if (statusDb === "disconnected" && runtimeConnected) {
    return "Sin registro DB pero conectado";
  }
  return statusDb;
}
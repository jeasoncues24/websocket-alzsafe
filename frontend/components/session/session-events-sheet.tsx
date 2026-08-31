"use client"

import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import { Badge } from "@/components/ui/badge"
import { type SessionInfo } from "@/lib/api"
import { Wifi, WifiOff, Loader2, QrCode, Clock, Info, AlertTriangle } from "lucide-react"

interface SessionEventsSheetProps {
  session: SessionInfo | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

function formatLocalTime(ts: string): string {
  try {
    return new Date(ts).toLocaleString("es-PE", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
    })
  } catch {
    return ts
  }
}

function EventTypeIcon({ type }: { type: string }) {
  // Grey-State Rule: solo "conectado" lleva verde; el resto es neutro.
  switch (type) {
    case "connected":
      return <Wifi className="h-4 w-4 shrink-0 text-emerald-600 dark:text-emerald-400" />
    case "disconnected":
      return <WifiOff className="h-4 w-4 shrink-0 text-muted-foreground" />
    case "client_outdated":
      return <AlertTriangle className="h-4 w-4 shrink-0 text-muted-foreground" />
    case "initializing":
      return <Loader2 className="h-4 w-4 shrink-0 animate-spin text-muted-foreground" />
    default:
      return <QrCode className="h-4 w-4 shrink-0 text-muted-foreground" />
  }
}

export function SessionEventsSheet({
  session,
  open,
  onOpenChange,
}: SessionEventsSheetProps) {
  const events = session?.events ? [...session.events].reverse() : []
  const title = session?.empresa_nombre || session?.account_id || ""

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        className="w-full sm:max-w-md overflow-y-auto flex flex-col"
        onCloseAutoFocus={(e) => e.preventDefault()}
      >
        <SheetHeader className="pb-4 border-b">
          <SheetTitle className="text-lg font-semibold flex items-center gap-2">
            <Clock className="h-5 w-5 text-muted-foreground" />
            <span>Historial de Eventos</span>
          </SheetTitle>
          <SheetDescription asChild>
            <div className="text-left space-y-1 pt-1">
              <span className="font-medium text-foreground block">{title}</span>
              {session?.account_id && (
                <span className="font-mono text-xs text-muted-foreground block select-all">
                  {session.account_id}
                </span>
              )}
            </div>
          </SheetDescription>
        </SheetHeader>

        <div className="flex-1 py-4">
          {!session || events.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-center text-muted-foreground">
              <Info className="h-8 w-8 mb-2 opacity-40" />
              <p className="text-sm">No hay eventos registrados recientemente.</p>
            </div>
          ) : (
            <div className="space-y-3">
              {events.map((evt, idx) => (
                <div
                  key={idx}
                  className="flex items-start gap-3 rounded-lg border border-border bg-card p-3 transition-colors hover:bg-muted/40"
                >
                  <div className="mt-0.5 rounded-full bg-muted/60 p-1.5">
                    <EventTypeIcon type={evt.type} />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center justify-between gap-2">
                      <Badge variant="outline" className="text-xs capitalize font-medium">
                        {evt.type}
                      </Badge>
                      <span className="shrink-0 font-mono text-xs text-muted-foreground">
                        {formatLocalTime(evt.timestamp)}
                      </span>
                    </div>
                    {evt.details && (
                      <p className="mt-1 text-xs text-muted-foreground break-words">
                        {evt.details}
                      </p>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </SheetContent>
    </Sheet>
  )
}

"use client"

import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { QRRender } from "@/components/qr/qr-render"
import { type SessionInfo } from "@/lib/api"
import { QrCode } from "lucide-react"

interface SessionQRDialogProps {
  session: SessionInfo | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function SessionQRDialog({ session, open, onOpenChange }: SessionQRDialogProps) {
  const title = session?.empresa_nombre || session?.account_id || ""

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md" onCloseAutoFocus={(e) => e.preventDefault()}>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <QrCode className="h-5 w-5 text-muted-foreground" />
            <span>Código QR — {title}</span>
          </DialogTitle>
          <DialogDescription>
            Escanea este código QR con la aplicación WhatsApp en tu teléfono para vincular la sesión.
          </DialogDescription>
        </DialogHeader>

        {session && session.qr_string ? (
          <div className="flex flex-col items-center gap-4 py-4">
            {/* Fondo blanco fijo: requisito de contraste para que el QR sea escaneable. */}
            <div className="rounded-xl border border-border bg-white p-4">
              <QRRender
                value={session.qr_string}
                size={220}
                title={`QR ${session.account_id}`}
              />
            </div>
            <div className="space-y-1 text-center">
              <p className="font-mono text-xs font-medium text-muted-foreground">
                Número: {session.account_id}
              </p>
              <p className="text-xs text-muted-foreground">
                El código expira automáticamente. Si caduca, actualiza la página.
              </p>
            </div>
          </div>
        ) : (
          <div className="py-8 text-center text-sm text-muted-foreground">
            No hay código QR disponible en este momento para esta sesión.
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}

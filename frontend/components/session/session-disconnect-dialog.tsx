"use client"

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { AlertTriangle, LogOut } from "lucide-react"

interface SessionDisconnectDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: () => Promise<void>
  targetName: string
  loading?: boolean
}

export function SessionDisconnectDialog({
  open,
  onOpenChange,
  onConfirm,
  targetName,
  loading = false,
}: SessionDisconnectDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <div className="flex items-center gap-2 text-destructive">
            <AlertTriangle className="h-5 w-5" />
            <DialogTitle>Confirmar Desconexión</DialogTitle>
          </div>
          <DialogDescription className="pt-2 text-left">
            ¿Estás seguro de que deseas desconectar la sesión de WhatsApp de{" "}
            <strong className="text-foreground font-semibold">{targetName}</strong>?
          </DialogDescription>
        </DialogHeader>

        <div className="rounded-md bg-muted/50 p-3 text-xs text-muted-foreground">
          Al desconectar, el teléfono dejará de enviar y recibir mensajes a través de la API hasta que se vuelva a escanear el código QR.
        </div>

        <DialogFooter className="gap-2 sm:gap-0 mt-4">
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={loading}
          >
            Cancelar
          </Button>
          <Button
            variant="destructive"
            onClick={onConfirm}
            disabled={loading}
          >
            <LogOut className="h-4 w-4 mr-1.5" />
            {loading ? "Desconectando..." : "Desconectar"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

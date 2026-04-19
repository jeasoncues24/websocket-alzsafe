"use client"

import { QRCodeSVG } from "qrcode.react"

import { cn } from "@/lib/utils"

type QRRenderProps = {
  value: string
  size?: number
  title?: string
  className?: string
}

export function QRRender({ value, size = 220, title = "Codigo QR", className }: QRRenderProps) {
  const content = value.trim()
  if (!content) {
    return null
  }

  return (
    <div className={cn("inline-flex rounded-xl border bg-white p-4", className)}>
      <QRCodeSVG
        value={content}
        size={size}
        level="M"
        marginSize={2}
        bgColor="#FFFFFF"
        fgColor="#111111"
        title={title}
      />
    </div>
  )
}

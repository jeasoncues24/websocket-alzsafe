"use client"

import * as React from "react"
import { cn } from "@/lib/utils"

interface SelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
  children: React.ReactNode
}

function Select({ className, children, ...props }: SelectProps) {
  return (
    <select
      className={cn(
        "flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
        className
      )}
      {...props}
    >
      {children}
    </select>
  )
}

interface SelectTriggerProps extends React.HTMLAttributes<HTMLDivElement> {
  children: React.ReactNode
}

function SelectTrigger({ className, children, ...props }: SelectTriggerProps) {
  return (
    <div className={cn("relative", className)} {...props}>
      {children}
    </div>
  )
}

interface SelectValueProps extends React.HTMLAttributes<HTMLSpanElement> {
  placeholder?: string
}

function SelectValue({ className, placeholder, ...props }: SelectValueProps) {
  return (
    <span className={cn("text-muted-foreground", className)} {...props}>
      {placeholder}
    </span>
  )
}

interface SelectContentProps extends React.HTMLAttributes<HTMLDivElement> {
  children: React.ReactNode
}

function SelectContent({ className, children, ...props }: SelectContentProps) {
  return (
    <div
      className={cn(
        "absolute z-50 mt-1 max-h-60 w-full overflow-auto rounded-md border bg-background py-1 shadow-md",
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
}

interface SelectItemProps extends React.HTMLAttributes<HTMLDivElement> {
  value: string
  children: React.ReactNode
}

function SelectItem({ className, children, ...props }: SelectItemProps) {
  return (
    <div
      className={cn(
        "relative flex w-full cursor-default select-none items-center rounded-sm py-1.5 pl-8 pr-2 text-sm outline-none hover:bg-accent hover:text-accent-foreground",
        className
      )}
      {...props}
    >
      {children}
    </div>
  )
}

interface SelectLabelProps extends React.HTMLAttributes<HTMLDivElement> {
  children: React.ReactNode
}

function SelectLabel({ className, children, ...props }: SelectLabelProps) {
  return (
    <div
      className={cn("py-1.5 pl-8 pr-2 text-sm font-semibold", className)}
      {...props}
    >
      {children}
    </div>
  )
}

export { Select, SelectTrigger, SelectValue, SelectContent, SelectItem, SelectLabel }
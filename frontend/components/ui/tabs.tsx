"use client"

import * as React from "react"
import { cn } from "@/lib/utils"

interface TabsProps extends React.HTMLAttributes<HTMLDivElement> {
  defaultValue?: string
  onValueChange?: (value: string) => void
}

function Tabs({ className, defaultValue, onValueChange, children, ...props }: TabsProps) {
  const [value, setValue] = React.useState(defaultValue || "")

  const handleValueChange = (newValue: string) => {
    setValue(newValue)
    onValueChange?.(newValue)
  }

  return (
    <TabsContext.Provider value={{ value, onValueChange: handleValueChange }}>
      <div className={cn("space-y-4", className)} {...props}>
        {children}
      </div>
    </TabsContext.Provider>
  )
}

interface TabsListProps extends React.HTMLAttributes<HTMLDivElement> {}

function TabsList({ className, children, ...props }: TabsListProps) {
  return (
    <div className={cn("inline-flex h-10 items-center justify-center rounded-md bg-muted p-1 text-muted-foreground", className)} {...props}>
      {children}
    </div>
  )
}

interface TabsTriggerProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  value: string
}

function TabsTrigger({ className, value: triggerValue, children, ...props }: TabsTriggerProps) {
  const parent = React.useContext(TabsContext)
  const isActive = parent?.value === triggerValue
  
  return (
    <button
      type="button"
      className={cn(
        "inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50",
        isActive ? "bg-background text-foreground shadow-sm" : "text-muted-foreground hover:text-foreground",
        className
      )}
      onClick={() => parent?.onValueChange?.(triggerValue)}
      {...props}
    >
      {children}
    </button>
  )
}

interface TabsContentProps extends React.HTMLAttributes<HTMLDivElement> {
  value: string
}

function TabsContent({ className, value: contentValue, children, ...props }: TabsContentProps) {
  const parent = React.useContext(TabsContext)
  
  if (parent?.value !== contentValue) return null
  
  return (
    <div className={cn("mt-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2", className)} {...props}>
      {children}
    </div>
  )
}

const TabsContext = React.createContext<{value: string; onValueChange?: (v: string) => void} | null>(null)

export { Tabs, TabsList, TabsTrigger, TabsContent }
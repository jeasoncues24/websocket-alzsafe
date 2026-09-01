"use client"

import * as React from "react"
import { Check, ChevronsUpDown, Loader2, Building2, X } from "lucide-react"
import { Button } from "@/components/ui/button"
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover"
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/ui/command"
import { getEmpresas, type Empresa } from "@/lib/api"
import { cn } from "@/lib/utils"

export interface EmpresaComboboxProps {
  value: string
  onChange: (ruc: string, empresa?: Empresa) => void
  placeholder?: string
  className?: string
  disabled?: boolean
  selectedEmpresaName?: string
}

export function EmpresaCombobox({
  value,
  onChange,
  placeholder = "Todas las empresas",
  className,
  disabled = false,
  selectedEmpresaName,
}: EmpresaComboboxProps) {
  const [open, setOpen] = React.useState(false)
  const [search, setSearch] = React.useState("")
  const [debouncedSearch, setDebouncedSearch] = React.useState("")
  const [empresas, setEmpresas] = React.useState<Empresa[]>([])
  const [currentEmpresa, setCurrentEmpresa] = React.useState<Empresa | null>(null)
  const [loading, setLoading] = React.useState(false)

  // Debounce búsqueda por 300ms
  React.useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedSearch(search)
    }, 300)
    return () => clearTimeout(timer)
  }, [search])

  // Cargar empresas al abrir o al cambiar el texto de búsqueda
  React.useEffect(() => {
    let cancel = false

    async function fetchEmpresas() {
      setLoading(true)
      try {
        const res = await getEmpresas({
          busqueda: debouncedSearch.trim() || undefined,
          estado: "activo",
          limit: 20,
        })
        if (!cancel) {
          const list = res.empresas ?? []
          setEmpresas(list)

          // Si hay un valor seleccionado y no tenemos el objeto empresa completo, buscarlo en la lista
          if (value && !currentEmpresa) {
            const found = list.find((e) => e.ruc === value)
            if (found) setCurrentEmpresa(found)
          }
        }
      } catch (err) {
        console.error("Error al buscar empresas:", err)
      } finally {
        if (!cancel) setLoading(false)
      }
    }

    if (open || debouncedSearch.trim()) {
      fetchEmpresas()
    }

    return () => {
      cancel = true
    }
  }, [open, debouncedSearch, value, currentEmpresa])

  // Si el valor cambia externamente a vacío, resetear currentEmpresa
  React.useEffect(() => {
    if (!value) {
      setCurrentEmpresa(null)
    }
  }, [value])

  const displayName = currentEmpresa?.nombre || selectedEmpresaName || value

  const handleSelect = (empresa: Empresa) => {
    if (value === empresa.ruc) {
      // Deseleccionar
      setCurrentEmpresa(null)
      onChange("", undefined)
    } else {
      setCurrentEmpresa(empresa)
      onChange(empresa.ruc, empresa)
    }
    setOpen(false)
  }

  const handleClear = (e: React.MouseEvent) => {
    e.stopPropagation()
    setCurrentEmpresa(null)
    onChange("", undefined)
  }

  return (
    <div className={cn("flex items-center gap-1.5", className)}>
      <Popover open={open} onOpenChange={setOpen}>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            role="combobox"
            aria-expanded={open}
            disabled={disabled}
            className={cn(
              "h-9 w-full sm:w-[240px] justify-between text-xs font-normal",
              !value && "text-muted-foreground"
            )}
          >
            <div className="flex items-center gap-2 truncate">
              <Building2 className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
              <span className="truncate">
                {value ? displayName : placeholder}
              </span>
            </div>
            <div className="flex items-center gap-1 shrink-0 ml-1">
              <ChevronsUpDown className="h-3.5 w-3.5 opacity-50" />
            </div>
          </Button>
        </PopoverTrigger>

        <PopoverContent
          className="w-[300px] sm:w-[340px] p-0"
          align="start"
          sideOffset={4}
        >
          {/* shouldFilter={false} desactiva el filtrado local de cmdk para permitir búsqueda Server-Side */}
          <Command shouldFilter={false}>
            <CommandInput
              placeholder="Buscar por RUC o nombre..."
              value={search}
              onValueChange={setSearch}
              className="h-9 text-xs"
            />
            <CommandList>
              {loading && (
                <div className="flex items-center justify-center gap-2 py-6 text-xs text-muted-foreground">
                  <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
                  <span>Buscando empresas...</span>
                </div>
              )}

              {!loading && empresas.length === 0 && (
                <CommandEmpty className="py-6 text-center text-xs text-muted-foreground">
                  No se encontraron empresas activas.
                </CommandEmpty>
              )}

              {!loading && empresas.length > 0 && (
                <CommandGroup heading="Empresas activas">
                  {empresas.map((empresa) => {
                    const isSelected = value === empresa.ruc
                    return (
                      <CommandItem
                        key={empresa.id || empresa.ruc}
                        value={empresa.ruc}
                        onSelect={() => handleSelect(empresa)}
                        className="flex items-center justify-between py-2 cursor-pointer text-xs"
                      >
                        <div className="flex flex-col min-w-0 pr-2">
                          <span className="font-medium truncate text-foreground">
                            {empresa.nombre}
                          </span>
                          <div className="flex items-center gap-2 text-[11px] text-muted-foreground font-mono">
                            <span>RUC: {empresa.ruc}</span>
                            {empresa.telefono_contacto && (
                              <span>• {empresa.telefono_contacto}</span>
                            )}
                          </div>
                        </div>
                        <Check
                          className={cn(
                            "h-4 w-4 shrink-0 text-primary",
                            isSelected ? "opacity-100" : "opacity-0"
                          )}
                        />
                      </CommandItem>
                    )
                  })}
                </CommandGroup>
              )}
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>

      {value && (
        <Button
          variant="ghost"
          size="icon"
          className="h-9 w-9 shrink-0 text-muted-foreground hover:text-foreground"
          onClick={handleClear}
          title="Limpiar filtro de empresa"
        >
          <X className="h-4 w-4" />
        </Button>
      )}
    </div>
  )
}

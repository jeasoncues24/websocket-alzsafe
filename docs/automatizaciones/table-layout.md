# Patrón de tablas con TanStack Table

> **Fuente de verdad:** `frontend/app/sessions/` (vista `http://localhost:3001/sessions`).
> Toda tabla nueva del panel se construye copiando este patrón. Si algo de este
> documento y el código de `app/sessions/` no coinciden, gana el código: actualiza
> este documento.

Stack: **Next.js 16** (App Router, componentes cliente) · **React 19** ·
**@tanstack/react-table 8.21** · **shadcn/ui** (`components/ui/table`, `select`,
`button`, `skeleton`) · tokens de `frontend/app/globals.css`.

---

## 1. Cuándo usar este patrón

Úsalo para **cualquier listado tabular del panel admin** que necesite orden por
columnas, paginación y/o filtros. No lo uses para:

- Listas simples sin orden ni paginación → usa un `<ul>` / tarjetas planas.
- Datos que ya vienen paginados y ordenados del servidor **y** no necesitan orden
  en cliente → puedes renderizar `<Table>` de `components/ui/table` a mano sin
  `useReactTable` (ver §8, "Paginación de servidor").

---

## 2. Estructura de archivos

Cada tabla vive en su carpeta de ruta `frontend/app/<ruta>/` con estos archivos:

| Archivo | Rol | ¿Genérico? |
|---|---|---|
| `page.tsx` | **Contenedor.** Estado, fetching, stream/polling, handlers, filtrado de negocio, memoización de columnas. Decide tabla (desktop) vs. tarjetas (móvil). | No, específico de la vista |
| `columns.tsx` | Factory `getColumns(callbacks)` que devuelve `ColumnDef<T>[]`. Define `accessorKey`/`accessorFn`, headers ordenables y `cell`. | No, específico de la vista |
| `data-table.tsx` | Componente **presentacional genérico** `<DataTable<TData, TValue>>`. Monta `useReactTable`, renderiza `<Table>`, gestiona `sorting` + `pagination`, estados `loading`/vacío. | **Sí** — cópialo tal cual entre vistas |
| `data-table-pagination.tsx` | Controles de paginación. Recibe la instancia `table`. | **Sí** — genérico |
| `data-table-toolbar.tsx` | Filtros: buscador global + control segmentado por estado + refresco. | Semi: la forma se copia, las pestañas/labels cambian |
| `data-card-list.tsx` | Variante **móvil**: lista de tarjetas accionables con su propia paginación. Comparte los handlers del contenedor. | No, específico de la vista |

> `data-table.tsx` y `data-table-pagination.tsx` son casi idénticos entre vistas.
> Cuando exista una segunda tabla, extrae ambos a `frontend/components/data-table/`
> y deja solo `columns.tsx` + `page.tsx` + `*-toolbar.tsx` + `*-card-list.tsx` por ruta.

---

## 3. Decisiones de arquitectura (el porqué)

### 3.1. El filtrado de negocio vive en el contenedor, no en la tabla

`page.tsx` calcula `filteredItems` con `useMemo` (pestañas de estado + búsqueda de
texto) y pasa **datos ya filtrados** a `<DataTable data={filteredItems} />`.

**No** usamos `getFilteredRowModel` + `globalFilter`/`columnFilters` de TanStack para
el filtro de negocio.

Motivo:

- El filtro suele cruzar campos derivados (nombre de empresa, id, estado
  compuesto) y reglas propias (`mismatch`, pestañas). En JS plano es más legible y
  testeable que un `filterFn` por columna.
- El origen de datos es un stream (SSE) + polling de respaldo. El contenedor ya es
  el dueño del estado; el filtro es una proyección más de ese estado.
- `<DataTable>` queda tonto y 100% reutilizable.

Dentro de `useReactTable` solo viven **orden (`sorting`)** y **paginación
(`pagination`)**, que son estado de presentación puro.

### 3.2. Columnas como factory con callbacks inyectados

`columns.tsx` exporta `getColumns(callbacks): ColumnDef<T>[]`, no un array suelto.
Las acciones de fila (`onView`, `onToggle`, …) y el estado efímero
(`togglingId`) entran por parámetro desde `page.tsx`.

En `page.tsx` las columnas se **memoizan**:

```tsx
const columns = useMemo(
  () => getColumns({ onView, onToggle, togglingId }),
  [onView, onToggle, togglingId],
)
```

Los handlers se envuelven en `useCallback` para que `columns` no se recree en cada
render. El estado efímero de una fila (qué fila está "guardando"/"reconectando") se
pasa por el closure de `getColumns`, no por `table.options.meta` — es más directo y
type-safe. `meta` queda como alternativa si el árbol de props se vuelve profundo.

### 3.3. `<DataTable>` es genérico y presentacional

Firma:

```tsx
interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[]
  data: TData[]
  loading?: boolean
}
```

No sabe nada del dominio. Renderiza header con `flexRender`, cuerpo con estados
`loading` (filas skeleton), vacío (mensaje + icono) y datos. La paginación se monta
solo cuando `!loading && data.length > 0`.

### 3.4. Orden y paginación: cliente

```tsx
getCoreRowModel: getCoreRowModel(),
getSortedRowModel: getSortedRowModel(),
getPaginationRowModel: getPaginationRowModel(),
initialState: { pagination: { pageSize: 10 } },
```

`pageSize` inicial 10, opciones `[10, 20, 30, 50]`. Es paginación de cliente sobre
el dataset completo ya filtrado. Válido hasta ~unos miles de filas; por encima de
eso, mueve el filtrado/paginado al servidor (§8).

### 3.5. Variante móvil obligatoria

Toda tabla nueva trae su `data-card-list.tsx`. `page.tsx` decide con CSS:

```tsx
<div className="hidden md:flex md:flex-col">
  <DataTable columns={columns} data={filteredItems} loading={loading} />
</div>
<div className="md:hidden">
  <DataCardList data={filteredItems} loading={loading} {...handlers} />
</div>
```

Ambas ramas consumen **los mismos handlers** del contenedor. `DataCardList` tiene
su propia paginación simple (slice sobre el array, `PAGE_SIZE = 10`), no usa
TanStack.

### 3.6. Warning conocido de React Compiler

`useReactTable()` dispara en `npm run lint`:

```
react-hooks/incompatible-library
TanStack Table's `useReactTable()` API returns functions that cannot be memoized safely
```

Es **warning, no error**, y es esperado: el compilador se salta la memoización de
ese componente. No lo silencies con `eslint-disable`; mantén `<DataTable>` pequeño
para que el coste sea nulo. No pases valores derivados de `table` a otros
componentes memoizados aguas abajo.

### 3.7. Diseño: sin color por vista

Sigue `DESIGN.md` (reglas *No-Rainbow*, *Grey-State*, *Flat-At-Rest*):

- Contenedor: `overflow-hidden rounded-xl ring-1 ring-foreground/10`.
- Header: `bg-muted/40`, TH `h-10 text-xs font-medium text-muted-foreground`.
- Celdas: `py-3 text-sm`. Fila: `hover:bg-muted/50 transition-colors`.
- El color sale solo de tokens (`text-muted-foreground`, `text-destructive`, …).
  El único acento puntual permitido es un estado semántico (p. ej. verde
  `emerald` para "online"), nunca una paleta por pantalla ni degradados.

---

## 4. Paso a paso para una tabla nueva

1. **Crea la carpeta de ruta**: `frontend/app/<ruta>/`.
2. **Define el tipo de fila** (idealmente en `frontend/lib/api.ts` junto al fetcher).
3. **Copia `data-table.tsx` y `data-table-pagination.tsx`** desde `app/sessions/`
   sin cambios (o impórtalos de `components/data-table/` si ya se extrajeron).
4. **Escribe `columns.tsx`**: `getColumns(callbacks)` con tus columnas. Headers
   ordenables con el botón `ghost` + `ArrowUpDown`; el resto, string plano.
5. **Escribe `data-table-toolbar.tsx`**: buscador + pestañas de estado propias.
6. **Escribe `data-card-list.tsx`**: tarjeta por fila con las mismas acciones.
7. **Escribe `page.tsx`**: estado, fetch, `useCallback` para handlers, `useMemo`
   para `filteredItems` y para `columns`, y el switch responsive.
8. **Verifica**: `cd frontend && npm run lint && npm run build`.

---

## 5. Scaffold copiable

Ejemplo genérico: tabla de `Item` con `id`, `nombre`, `estado`
(`"activo" | "inactivo" | "pendiente"`) y `actualizado` (ISO string). Acciones:
ver detalle y alternar estado. Renombra `Item`/`item` a tu dominio.

### 5.1. Tipo de fila — `frontend/lib/api.ts` (o donde vivan tus tipos)

```ts
export interface Item {
  id: number
  nombre: string
  estado: "activo" | "inactivo" | "pendiente"
  actualizado: string | null // ISO 8601
}

export interface ItemsResponse {
  items: Item[]
}

export async function getItems(): Promise<ItemsResponse> {
  // ...fetch al backend, mismo estilo que getAdminSessions()
}
```

### 5.2. `frontend/app/items/data-table.tsx` — genérico, cópialo tal cual

```tsx
"use client"

import { useState } from "react"
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
} from "@tanstack/react-table"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { Skeleton } from "@/components/ui/skeleton"
import { Inbox } from "lucide-react"
import { DataTablePagination } from "./data-table-pagination"

interface DataTableProps<TData, TValue> {
  columns: ColumnDef<TData, TValue>[]
  data: TData[]
  loading?: boolean
  /** Mensaje cuando no hay filas (tras aplicar filtros). */
  emptyMessage?: string
}

export function DataTable<TData, TValue>({
  columns,
  data,
  loading = false,
  emptyMessage = "No se encontraron resultados que coincidan con los filtros.",
}: DataTableProps<TData, TValue>) {
  const [sorting, setSorting] = useState<SortingState>([])

  const table = useReactTable({
    data,
    columns,
    state: { sorting },
    onSortingChange: setSorting,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    initialState: { pagination: { pageSize: 10 } },
  })

  return (
    <div className="overflow-hidden rounded-xl ring-1 ring-foreground/10">
      <div className="overflow-x-auto">
        <Table>
          <TableHeader className="bg-muted/40">
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead
                    key={header.id}
                    className="h-10 text-xs font-medium text-muted-foreground"
                  >
                    {header.isPlaceholder
                      ? null
                      : flexRender(
                          header.column.columnDef.header,
                          header.getContext(),
                        )}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>

          <TableBody>
            {loading ? (
              Array.from({ length: 5 }).map((_, i) => (
                <TableRow key={i}>
                  {columns.map((_, j) => (
                    <TableCell key={j} className="py-3">
                      <Skeleton className="h-5 w-full" />
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : table.getRowModel().rows.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow
                  key={row.id}
                  className="transition-colors hover:bg-muted/50"
                >
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id} className="py-3 text-sm">
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-32 text-center">
                  <div className="flex flex-col items-center justify-center gap-2 text-muted-foreground">
                    <Inbox className="h-8 w-8 opacity-30" />
                    <p className="text-sm font-medium">{emptyMessage}</p>
                  </div>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {!loading && data.length > 0 && (
        <div className="border-t">
          <DataTablePagination table={table} />
        </div>
      )}
    </div>
  )
}
```

### 5.3. `frontend/app/items/data-table-pagination.tsx` — genérico, cópialo tal cual

```tsx
"use client"

import { Table } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Select } from "@/components/ui/select"
import {
  ChevronLeft,
  ChevronRight,
  ChevronsLeft,
  ChevronsRight,
} from "lucide-react"

interface DataTablePaginationProps<TData> {
  table: Table<TData>
  pageSizeOptions?: number[]
  /** Sustantivo plural para el contador ("sesiones", "ítems", …). */
  itemLabel?: string
}

export function DataTablePagination<TData>({
  table,
  pageSizeOptions = [10, 20, 30, 50],
  itemLabel = "resultados",
}: DataTablePaginationProps<TData>) {
  const { pageIndex, pageSize } = table.getState().pagination
  // Sin getFilteredRowModel registrado, esto devuelve el dataset completo que
  // recibió la tabla (el filtrado de negocio ya ocurrió en el contenedor).
  const totalRows = table.getFilteredRowModel().rows.length
  const pageCount = table.getPageCount()

  const startRow = totalRows === 0 ? 0 : pageIndex * pageSize + 1
  const endRow = Math.min((pageIndex + 1) * pageSize, totalRows)

  return (
    <div className="flex flex-col items-center justify-between gap-4 px-2 py-4 text-sm text-muted-foreground sm:flex-row">
      <div className="flex-1 text-center text-xs sm:text-left sm:text-sm">
        Mostrando <span className="font-medium text-foreground">{startRow}</span> a{" "}
        <span className="font-medium text-foreground">{endRow}</span> de{" "}
        <span className="font-medium text-foreground">{totalRows}</span> {itemLabel}
      </div>

      <div className="flex flex-wrap items-center justify-center gap-4 sm:gap-6">
        <div className="flex items-center gap-2">
          <span className="text-xs font-medium sm:text-sm">Filas por pág.</span>
          <Select
            value={`${pageSize}`}
            onChange={(e) => table.setPageSize(Number(e.target.value))}
            className="h-8 w-[70px] text-xs"
          >
            {pageSizeOptions.map((size) => (
              <option key={size} value={`${size}`}>
                {size}
              </option>
            ))}
          </Select>
        </div>

        <div className="text-xs font-medium sm:text-sm">
          Pág. {pageCount === 0 ? 0 : pageIndex + 1} de {pageCount}
        </div>

        <div className="flex items-center gap-1">
          <Button
            variant="outline"
            size="icon"
            className="hidden h-8 w-8 sm:flex"
            onClick={() => table.setPageIndex(0)}
            disabled={!table.getCanPreviousPage()}
            title="Primera página"
          >
            <ChevronsLeft className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            size="icon"
            className="h-8 w-8"
            onClick={() => table.previousPage()}
            disabled={!table.getCanPreviousPage()}
            title="Página anterior"
          >
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            size="icon"
            className="h-8 w-8"
            onClick={() => table.nextPage()}
            disabled={!table.getCanNextPage()}
            title="Página siguiente"
          >
            <ChevronRight className="h-4 w-4" />
          </Button>
          <Button
            variant="outline"
            size="icon"
            className="hidden h-8 w-8 sm:flex"
            onClick={() => table.setPageIndex(pageCount - 1)}
            disabled={!table.getCanNextPage()}
            title="Última página"
          >
            <ChevronsRight className="h-4 w-4" />
          </Button>
        </div>
      </div>
    </div>
  )
}
```

### 5.4. `frontend/app/items/columns.tsx`

```tsx
"use client"

import { ColumnDef } from "@tanstack/react-table"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { ArrowUpDown, Eye, Power } from "lucide-react"
import { type Item } from "@/lib/api"

function relativeTime(ts?: string | null): string {
  if (!ts) return "Nunca"
  const m = Math.floor((Date.now() - new Date(ts).getTime()) / 60000)
  if (m < 1) return "ahora"
  if (m < 60) return `hace ${m}m`
  const h = Math.floor(m / 60)
  if (h < 24) return `hace ${h}h`
  return `hace ${Math.floor(h / 24)}d`
}

/** Header ordenable reutilizable. */
function SortableHeader({
  column,
  children,
}: {
  column: import("@tanstack/react-table").Column<Item, unknown>
  children: React.ReactNode
}) {
  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}
      className="-ml-3 h-8 text-xs font-semibold hover:bg-muted/50"
    >
      {children}
      <ArrowUpDown className="ml-1 h-3 w-3" />
    </Button>
  )
}

interface ColumnCallbacks {
  onView: (item: Item) => void
  onToggle: (id: number) => void
  /** Fila cuyo estado se está guardando: deshabilita su botón. */
  togglingId: number | null
}

export function getColumns(cb: ColumnCallbacks): ColumnDef<Item>[] {
  return [
    {
      accessorKey: "nombre",
      header: ({ column }) => <SortableHeader column={column}>Nombre</SortableHeader>,
      cell: ({ row }) => (
        <div className="max-w-[280px] truncate font-medium text-foreground" title={row.original.nombre}>
          {row.original.nombre}
        </div>
      ),
    },
    {
      accessorKey: "estado",
      header: ({ column }) => <SortableHeader column={column}>Estado</SortableHeader>,
      cell: ({ row }) => {
        const estado = row.original.estado
        return (
          <Badge variant={estado === "activo" ? "default" : "outline"} className="capitalize">
            {estado}
          </Badge>
        )
      },
    },
    {
      // Campo derivado → id + accessorFn en vez de accessorKey.
      id: "actualizado",
      accessorFn: (row) => (row.actualizado ? new Date(row.actualizado).getTime() : 0),
      header: ({ column }) => <SortableHeader column={column}>Actualizado</SortableHeader>,
      cell: ({ row }) => (
        <span className="text-xs text-muted-foreground">
          {relativeTime(row.original.actualizado)}
        </span>
      ),
    },
    {
      id: "actions",
      header: () => <div className="pr-2 text-right">Acciones</div>,
      cell: ({ row }) => {
        const item = row.original
        return (
          <div className="flex items-center justify-end gap-1.5">
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8 text-muted-foreground hover:text-foreground"
              onClick={() => cb.onView(item)}
              aria-label="Ver detalle"
            >
              <Eye className="h-3.5 w-3.5" />
            </Button>
            <Button
              variant="outline"
              size="sm"
              className="h-8 px-2.5 text-xs"
              disabled={cb.togglingId === item.id}
              onClick={() => cb.onToggle(item.id)}
            >
              <Power className="mr-1 h-3.5 w-3.5" />
              {item.estado === "activo" ? "Desactivar" : "Activar"}
            </Button>
          </div>
        )
      },
    },
  ]
}
```

### 5.5. `frontend/app/items/data-table-toolbar.tsx`

```tsx
"use client"

import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
import { Search, X, RefreshCw, Layers, CheckCircle2, CircleSlash, Clock } from "lucide-react"

export type StatusTabFilter = "all" | "activo" | "inactivo" | "pendiente"

interface DataTableToolbarProps {
  globalFilter: string
  onGlobalFilterChange: (value: string) => void
  statusFilter: StatusTabFilter
  onStatusFilterChange: (status: StatusTabFilter) => void
  counts: Record<StatusTabFilter, number>
  loading: boolean
  onRefresh: () => void
}

export function DataTableToolbar({
  globalFilter,
  onGlobalFilterChange,
  statusFilter,
  onStatusFilterChange,
  counts,
  loading,
  onRefresh,
}: DataTableToolbarProps) {
  const tabs: { id: StatusTabFilter; label: string; icon: React.ReactNode }[] = [
    { id: "all", label: "Todos", icon: <Layers className="h-3.5 w-3.5" /> },
    { id: "activo", label: "Activos", icon: <CheckCircle2 className="h-3.5 w-3.5" /> },
    { id: "inactivo", label: "Inactivos", icon: <CircleSlash className="h-3.5 w-3.5" /> },
    { id: "pendiente", label: "Pendientes", icon: <Clock className="h-3.5 w-3.5" /> },
  ]

  return (
    <div className="space-y-3">
      {/* Control segmentado por estado: un solo control, sin color por pestaña */}
      <div
        role="group"
        aria-label="Filtrar por estado"
        className="flex flex-wrap gap-0.5 rounded-md bg-muted p-0.5"
      >
        {tabs.map((tab) => {
          const isActive = statusFilter === tab.id
          return (
            <button
              key={tab.id}
              type="button"
              aria-pressed={isActive}
              onClick={() => onStatusFilterChange(tab.id)}
              className={cn(
                "inline-flex h-8 items-center gap-1.5 rounded-sm px-2.5 text-xs font-medium transition-colors",
                isActive
                  ? "bg-background text-foreground"
                  : "text-muted-foreground hover:text-foreground",
              )}
            >
              {tab.icon}
              <span>{tab.label}</span>
              <span
                className={cn(
                  "tabular-nums",
                  isActive ? "text-foreground/60" : "text-muted-foreground/70",
                )}
              >
                {counts[tab.id]}
              </span>
            </button>
          )
        })}
      </div>

      {/* Buscador + refresco */}
      <div className="flex flex-col gap-2.5 sm:flex-row sm:items-center sm:justify-between">
        <div className="relative w-full sm:max-w-sm">
          <Search className="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Buscar por nombre o estado..."
            value={globalFilter}
            onChange={(e) => onGlobalFilterChange(e.target.value)}
            className="h-9 pl-8 pr-8"
          />
          {globalFilter && (
            <button
              onClick={() => onGlobalFilterChange("")}
              className="absolute right-2 top-1/2 -translate-y-1/2 rounded p-0.5 text-muted-foreground hover:text-foreground"
              title="Limpiar búsqueda"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          )}
        </div>

        <Button variant="outline" size="sm" onClick={onRefresh} disabled={loading}>
          <RefreshCw className={cn("mr-1.5 h-3.5 w-3.5", loading && "animate-spin")} />
          Actualizar
        </Button>
      </div>
    </div>
  )
}
```

### 5.6. `frontend/app/items/data-card-list.tsx` (variante móvil)

```tsx
"use client"

import { useState } from "react"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Skeleton } from "@/components/ui/skeleton"
import { cn } from "@/lib/utils"
import { Inbox, Eye, Power, ChevronLeft, ChevronRight } from "lucide-react"
import { type Item } from "@/lib/api"

interface DataCardListProps {
  data: Item[]
  loading?: boolean
  onView: (item: Item) => void
  onToggle: (id: number) => void
  togglingId: number | null
}

const PAGE_SIZE = 10

export function DataCardList({
  data,
  loading = false,
  onView,
  onToggle,
  togglingId,
}: DataCardListProps) {
  const [page, setPage] = useState(1)
  const totalPages = Math.ceil(data.length / PAGE_SIZE) || 1
  const currentPage = Math.min(page, totalPages)
  const paged = data.slice((currentPage - 1) * PAGE_SIZE, currentPage * PAGE_SIZE)

  if (loading) {
    return (
      <div className="flex flex-col gap-3">
        {Array.from({ length: 4 }).map((_, i) => (
          <div key={i} className="rounded-xl p-4 ring-1 ring-foreground/10">
            <div className="flex flex-col gap-3">
              <div className="flex items-center justify-between">
                <Skeleton className="h-5 w-40" />
                <Skeleton className="h-5 w-20" />
              </div>
              <div className="flex gap-2 pt-1">
                <Skeleton className="h-9 flex-1" />
                <Skeleton className="h-9 w-9" />
              </div>
            </div>
          </div>
        ))}
      </div>
    )
  }

  if (data.length === 0) {
    return (
      <div className="rounded-xl px-6 py-12 text-center ring-1 ring-foreground/10">
        <div className="flex flex-col items-center gap-2 text-muted-foreground">
          <Inbox className="h-8 w-8 opacity-30" />
          <p className="text-sm font-medium">
            No se encontraron resultados que coincidan con los filtros.
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-3">
      {paged.map((item) => (
        <div
          key={item.id}
          className="flex flex-col gap-3 rounded-xl p-4 ring-1 ring-foreground/10"
        >
          <div className="flex items-start justify-between gap-2">
            <h3 className="truncate text-sm font-medium text-foreground" title={item.nombre}>
              {item.nombre}
            </h3>
            <Badge variant={item.estado === "activo" ? "default" : "outline"} className="capitalize">
              {item.estado}
            </Badge>
          </div>

          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              className="flex-1"
              disabled={togglingId === item.id}
              onClick={() => onToggle(item.id)}
            >
              <Power className="mr-1.5 h-4 w-4" />
              {item.estado === "activo" ? "Desactivar" : "Activar"}
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="size-9 shrink-0"
              onClick={() => onView(item)}
              aria-label="Ver detalle"
            >
              <Eye className="h-4 w-4 text-muted-foreground" />
            </Button>
          </div>
        </div>
      ))}

      {totalPages > 1 && (
        <div className="flex items-center justify-between px-1 py-2 text-xs text-muted-foreground">
          <span>
            Pág. <span className="font-medium text-foreground">{currentPage}</span> de{" "}
            <span className="font-medium text-foreground">{totalPages}</span> · {data.length} ítems
          </span>
          <div className="flex items-center gap-1">
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={currentPage === 1}
            >
              <ChevronLeft className="mr-1 h-3.5 w-3.5" />
              Anterior
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={currentPage === totalPages}
            >
              Siguiente
              <ChevronRight className="ml-1 h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
```

### 5.7. `frontend/app/items/page.tsx` (contenedor)

```tsx
"use client"

import { useCallback, useEffect, useMemo, useState } from "react"
import { DataTable } from "./data-table"
import { DataCardList } from "./data-card-list"
import { DataTableToolbar, type StatusTabFilter } from "./data-table-toolbar"
import { getColumns } from "./columns"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { AlertTriangle, ListChecks } from "lucide-react"
import { getItems, type Item } from "@/lib/api"

export default function ItemsPage() {
  const [items, setItems] = useState<Item[]>([])
  const [loading, setLoading] = useState(true)     // solo primer render sin datos
  const [refreshing, setRefreshing] = useState(false) // refresco manual: solo el icono
  const [error, setError] = useState<string | null>(null)

  const [globalFilter, setGlobalFilter] = useState("")
  const [statusFilter, setStatusFilter] = useState<StatusTabFilter>("all")
  const [togglingId, setTogglingId] = useState<number | null>(null)

  const load = useCallback(async (opts?: { silent?: boolean }) => {
    if (!opts?.silent) setRefreshing(true)
    try {
      const data = await getItems()
      setItems(data.items ?? [])
      setError(null)
    } catch (err) {
      console.error("Failed to load items:", err)
      setError("No se pudo cargar el listado.") // se conserva el último dato bueno
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }, [])

  useEffect(() => {
    load({ silent: true })
  }, [load])

  const handleView = useCallback((item: Item) => {
    // abrir modal/sheet...
  }, [])

  const handleToggle = useCallback(
    async (id: number) => {
      setTogglingId(id)
      try {
        // await toggleItem(id)
        await load({ silent: true })
      } finally {
        setTogglingId(null)
      }
    },
    [load],
  )

  // Filtrado de negocio: pestañas de estado + búsqueda de texto. NO en la tabla.
  const filteredItems = useMemo(() => {
    const q = globalFilter.toLowerCase().trim()
    return items.filter((it) => {
      if (statusFilter !== "all" && it.estado !== statusFilter) return false
      if (!q) return true
      return (
        it.nombre.toLowerCase().includes(q) ||
        it.estado.toLowerCase().includes(q) ||
        String(it.id).includes(q)
      )
    })
  }, [items, statusFilter, globalFilter])

  // Contadores por pestaña, calculados sobre el dataset completo.
  const counts = useMemo<Record<StatusTabFilter, number>>(
    () => ({
      all: items.length,
      activo: items.filter((it) => it.estado === "activo").length,
      inactivo: items.filter((it) => it.estado === "inactivo").length,
      pendiente: items.filter((it) => it.estado === "pendiente").length,
    }),
    [items],
  )

  const columns = useMemo(
    () => getColumns({ onView: handleView, onToggle: handleToggle, togglingId }),
    [handleView, handleToggle, togglingId],
  )

  return (
    <div className="space-y-4 pb-12">
      <div className="space-y-1">
        <h1 className="flex items-center gap-2 text-xl font-semibold tracking-tight text-foreground sm:text-2xl">
          <ListChecks className="h-5 w-5 text-muted-foreground" />
          <span>Ítems</span>
        </h1>
        <p className="text-sm text-muted-foreground">Descripción corta de la vista.</p>
      </div>

      <DataTableToolbar
        globalFilter={globalFilter}
        onGlobalFilterChange={setGlobalFilter}
        statusFilter={statusFilter}
        onStatusFilterChange={setStatusFilter}
        counts={counts}
        loading={refreshing}
        onRefresh={() => load()}
      />

      {error && (
        <Alert variant="destructive">
          <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
            <div className="flex items-start gap-3">
              <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0" />
              <div>
                <AlertTitle>No se pudo cargar el listado</AlertTitle>
                <AlertDescription>
                  {items.length > 0
                    ? "Se muestran los últimos datos disponibles. Reintenta para actualizar."
                    : "Revisa la conexión con el servidor y reintenta."}
                </AlertDescription>
              </div>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={() => load()}
              disabled={refreshing}
              className="shrink-0 self-start"
            >
              Reintentar
            </Button>
          </div>
        </Alert>
      )}

      {/* Escritorio: tabla densa. Móvil: tarjetas accionables. */}
      <div className="hidden md:flex md:flex-col">
        <DataTable columns={columns} data={filteredItems} loading={loading} />
      </div>
      <div className="md:hidden">
        <DataCardList
          data={filteredItems}
          loading={loading}
          onView={handleView}
          onToggle={handleToggle}
          togglingId={togglingId}
        />
      </div>
    </div>
  )
}
```

---

## 6. Referencia rápida de la API de TanStack usada

| Necesidad | Cómo |
|---|---|
| Columna sobre un campo directo | `{ accessorKey: "nombre", ... }` |
| Columna sobre valor derivado / compuesto | `{ id: "x", accessorFn: (row) => ..., ... }` (siempre con `id`) |
| Columna sin dato (acciones, iconos) | `{ id: "actions", header, cell }` sin accessor |
| Header ordenable | `header: ({ column }) => <Button onClick={() => column.toggleSorting(column.getIsSorted() === "asc")}>…</Button>` |
| Acceso al objeto tipado de la fila en `cell` | `row.original` |
| Render de header/celda | `flexRender(def, ctx)` |
| Orden en cliente | registrar `getSortedRowModel()` + estado `sorting` controlado |
| Paginación en cliente | registrar `getPaginationRowModel()` + `initialState.pagination.pageSize` |
| Nº total de filas (todas las páginas) | `table.getFilteredRowModel().rows.length` (= dataset recibido, porque el filtro real es upstream) |
| Nº de página / navegación | `table.getState().pagination`, `getPageCount()`, `nextPage()`, `setPageIndex()`, `getCanNextPage()` |

Lo que **no** usamos (a propósito): `getFilteredRowModel` para negocio,
`globalFilterFn`, `columnFilters`, `rowSelection`, `columnVisibility`,
`table.options.meta`. Si alguno hace falta, ver §8.

---

## 7. Estados de la tabla

- **Carga inicial** (`loading` = primer render sin datos): 5 filas skeleton. Tras
  la primera respuesta, las actualizaciones (refresco, polling, stream) reemplazan
  los datos en silencio — no vuelven a poner `loading` en `true`.
- **Refresco manual** (`refreshing`): solo anima el icono del botón "Actualizar".
- **Vacío tras filtros**: fila única con `colSpan={columns.length}`, icono tenue +
  mensaje.
- **Error**: se maneja en `page.tsx` con `<Alert variant="destructive">` y se
  **conserva el último dato bueno**. Una caída del backend nunca debe parecer un
  listado vacío.

---

## 8. Variaciones (cuándo salir del patrón)

- **Paginación / filtrado de servidor**: el dataset no cabe en cliente. `page.tsx`
  pasa a manejar `pageIndex`/`pageSize`/`sorting`/filtros como estado, los manda en
  la query al backend y `<DataTable>` recibe `manualPagination: true`,
  `manualSorting: true`, `pageCount` explícito y `onPaginationChange`. El contador
  de total viene del backend, no de `getFilteredRowModel`.
- **Selección de filas** (acciones masivas): añade `getRowId`, estado
  `rowSelection` + `onRowSelectionChange`, y una columna de checkbox como primera
  `ColumnDef`.
- **Mostrar/ocultar columnas**: estado `columnVisibility` + `onColumnVisibilityChange`
  y un menú (`DropdownMenu`) que recorre `table.getAllColumns()`.
- **Muchas acciones por fila**: colapsa en un `DropdownMenu` ("···") en vez de una
  fila de botones.

En todos los casos, `columns.tsx` sigue siendo una factory memoizada y el
contenedor sigue siendo el dueño del estado.

---

## 9. Checklist para PRs que agregan una tabla

- [ ] Carpeta `frontend/app/<ruta>/` con `page.tsx`, `columns.tsx`,
      `data-table.tsx`, `data-table-pagination.tsx`, `data-table-toolbar.tsx`,
      `data-card-list.tsx`.
- [ ] `data-table.tsx` y `data-table-pagination.tsx` son copias sin cambios de la
      versión de referencia (o imports de `components/data-table/`).
- [ ] Tipo de fila definido y exportado junto al fetcher (`lib/api.ts`).
- [ ] **Filtrado de negocio en el contenedor** con `useMemo` (`filteredItems`), no
      con `getFilteredRowModel`/`globalFilter` de TanStack.
- [ ] `useReactTable` solo gestiona `sorting` + `pagination`.
- [ ] `columns` memoizado con `useMemo`; handlers en `useCallback`; estado efímero
      de fila pasado por el closure de `getColumns`.
- [ ] Headers ordenables con el botón `ghost` + `ArrowUpDown`; el resto, texto plano.
- [ ] `cell` accede a los datos vía `row.original`.
- [ ] Estados cubiertos: **loading** (skeleton), **vacío tras filtros** (mensaje +
      icono), **error** (`<Alert>` en el page conservando el último dato).
- [ ] `refreshing` separado de `loading`: el refresco no repinta la tabla con skeleton.
- [ ] **Variante móvil** `DataCardList` presente y con los mismos handlers; switch
      `hidden md:flex` / `md:hidden` en el page.
- [ ] Diseño: `ring-1 ring-foreground/10`, `rounded-xl`, header `bg-muted/40`,
      solo tokens de `globals.css`. Sin paleta por vista, sin degradados
      (`DESIGN.md`: No-Rainbow / Grey-State / Flat-At-Rest).
- [ ] Accesibilidad: botones de icono con `aria-label`; control de estado con
      `role="group"` + `aria-pressed`.
- [ ] Textos de cara al usuario en **español**.
- [ ] `cd frontend && npm run lint && npm run build` en verde. El warning
      `react-hooks/incompatible-library` en `useReactTable()` es esperado.

---

## 10. Comandos de verificación

```bash
cd frontend && npm run lint     # warning conocido en useReactTable(): OK
cd frontend && npm run build    # requiere red para fuentes de Google o sandbox off
npx tsc --noEmit                # typecheck sin build
```

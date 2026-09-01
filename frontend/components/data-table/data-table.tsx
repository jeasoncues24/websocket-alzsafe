"use client"

import { useState } from "react"
import {
  ColumnDef,
  flexRender,
  getCoreRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  OnChangeFn,
  PaginationState,
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
  /** Icono personalizado para el estado vacío. */
  emptyIcon?: React.ReactNode
  /** Sustantivo plural para el contador de la paginación ("sesiones", "mensajes", …). */
  itemLabel?: string
  /** Tamaño inicial de página. Default: 10. */
  initialPageSize?: number
  /** Opciones de tamaño de página. Default: [10, 20, 30, 50]. */
  pageSizeOptions?: number[]
  /** Paginación manual de servidor */
  manualPagination?: boolean
  pageCount?: number
  pagination?: PaginationState
  onPaginationChange?: OnChangeFn<PaginationState>
  totalRows?: number
}

export function DataTable<TData, TValue>({
  columns,
  data,
  loading = false,
  emptyMessage = "No se encontraron resultados que coincidan con los filtros.",
  emptyIcon,
  itemLabel = "resultados",
  initialPageSize = 10,
  pageSizeOptions = [10, 20, 30, 50],
  manualPagination = false,
  pageCount,
  pagination,
  onPaginationChange,
  totalRows,
}: DataTableProps<TData, TValue>) {
  const [internalSorting, setInternalSorting] = useState<SortingState>([])
  const [internalPagination, setInternalPagination] = useState<PaginationState>({
    pageIndex: 0,
    pageSize: initialPageSize,
  })

  const table = useReactTable({
    data,
    columns,
    state: {
      sorting: internalSorting,
      pagination: manualPagination && pagination ? pagination : internalPagination,
    },
    onSortingChange: setInternalSorting,
    onPaginationChange:
      manualPagination && onPaginationChange ? onPaginationChange : setInternalPagination,
    manualPagination,
    pageCount: manualPagination ? (pageCount ?? -1) : undefined,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: manualPagination ? undefined : getSortedRowModel(),
    getPaginationRowModel: manualPagination ? undefined : getPaginationRowModel(),
    initialState: { pagination: { pageSize: initialPageSize } },
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
                  data-state={row.getIsSelected() && "selected"}
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
                    {emptyIcon || <Inbox className="h-8 w-8 opacity-30" />}
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
          <DataTablePagination
            table={table}
            itemLabel={itemLabel}
            pageSizeOptions={pageSizeOptions}
            totalRows={totalRows}
          />
        </div>
      )}
    </div>
  )
}

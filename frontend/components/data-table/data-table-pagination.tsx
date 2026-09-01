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
  /** Sustantivo plural para el contador ("sesiones", "mensajes", "ítems", …). */
  itemLabel?: string
}

export function DataTablePagination<TData>({
  table,
  pageSizeOptions = [10, 20, 30, 50],
  itemLabel = "resultados",
}: DataTablePaginationProps<TData>) {
  const { pageIndex, pageSize } = table.getState().pagination
  const totalRows = table.getFilteredRowModel().rows.length
  const pageCount = table.getPageCount()

  const startRow = totalRows === 0 ? 0 : pageIndex * pageSize + 1
  const endRow = Math.min((pageIndex + 1) * pageSize, totalRows)

  return (
    <div className="flex flex-col items-center justify-between gap-4 px-2 py-4 text-sm text-muted-foreground sm:flex-row">
      {/* Contador de filas visibles */}
      <div className="flex-1 text-center text-xs sm:text-left sm:text-sm">
        Mostrando <span className="font-medium text-foreground">{startRow}</span> a{" "}
        <span className="font-medium text-foreground">{endRow}</span> de{" "}
        <span className="font-medium text-foreground">{totalRows}</span> {itemLabel}
      </div>

      <div className="flex flex-wrap items-center justify-center gap-4 sm:gap-6">
        {/* Selector de tamaño de página */}
        <div className="flex items-center gap-2">
          <span className="text-xs font-medium sm:text-sm">Filas por pág.</span>
          <Select
            value={`${pageSize}`}
            onChange={(e) => table.setPageSize(Number(e.target.value))}
            className="h-8 w-[78px] text-xs"
          >
            {pageSizeOptions.map((size) => (
              <option key={size} value={`${size}`}>
                {size}
              </option>
            ))}
          </Select>
        </div>

        {/* Indicador de página actual */}
        <div className="text-xs font-medium sm:text-sm">
          Pág. {pageCount === 0 ? 0 : pageIndex + 1} de {pageCount}
        </div>

        {/* Botones de navegación */}
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

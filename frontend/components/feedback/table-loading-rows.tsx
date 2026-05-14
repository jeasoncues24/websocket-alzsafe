import { Skeleton } from "@/components/ui/skeleton";
import { TableCell, TableRow } from "@/components/ui/table";

interface TableLoadingRowsProps {
  rows?: number;
  columns: string[];
  actionWidthClassName?: string;
}

export function TableLoadingRows({
  rows = 5,
  columns,
  actionWidthClassName = "w-28",
}: TableLoadingRowsProps) {
  return Array.from({ length: rows }).map((_, index) => (
    <TableRow key={index}>
      {columns.map((width, columnIndex) => (
        <TableCell key={`${index}-${columnIndex}`}>
          <Skeleton className={`h-4 ${width}`} />
        </TableCell>
      ))}
      <TableCell>
        <Skeleton className={`ml-auto h-8 ${actionWidthClassName}`} />
      </TableCell>
    </TableRow>
  ));
}

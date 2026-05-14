import { cn } from "@/lib/utils";

interface DataEmptyStateProps {
  icon: React.ElementType;
  title: string;
  description: string;
  className?: string;
}

export function DataEmptyState({
  icon: Icon,
  title,
  description,
  className,
}: DataEmptyStateProps) {
  return (
    <div
      className={cn(
        "motion-panel flex flex-col items-center rounded-lg border border-dashed bg-muted/20 px-6 py-10 text-center",
        className,
      )}
    >
      <div className="flex size-12 items-center justify-center rounded-full bg-muted text-muted-foreground">
        <Icon className="h-5 w-5" />
      </div>
      <div className="mt-4 flex flex-col gap-1">
        <h3 className="text-lg font-semibold">{title}</h3>
        <p className="text-sm text-muted-foreground">{description}</p>
      </div>
    </div>
  );
}

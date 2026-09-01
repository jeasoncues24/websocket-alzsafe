"use client";

import type { ComponentProps } from "react";
import { ResponsiveContainer } from "recharts";

type ResponsiveChartProps = ComponentProps<typeof ResponsiveContainer>;

/**
 * `ResponsiveContainer` de recharts con dimensiones iniciales positivas.
 *
 * Evita el warning "The width(-1) and height(-1) of chart should be greater
 * than 0" que recharts emite en el primer render (SSR / hidratación), antes de
 * que el ResizeObserver mida el contenedor real. Los props `minWidth`/`minHeight`
 * de recharts no alimentan ese cálculo; solo `initialDimension` lo silencia.
 */
export function ResponsiveChart({
  width = "100%",
  height = "100%",
  initialDimension = { width: 800, height: 240 },
  ...props
}: ResponsiveChartProps) {
  return (
    <ResponsiveContainer
      width={width}
      height={height}
      initialDimension={initialDimension}
      {...props}
    />
  );
}

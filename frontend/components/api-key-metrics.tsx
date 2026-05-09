"use client";

import { useEffect, useState } from "react";
import {
  AreaChart, Area, BarChart, Bar, XAxis, YAxis, CartesianGrid,
  Tooltip, ResponsiveContainer, Legend, ComposedChart, Line,
} from "recharts";
import { AlertCircle, TrendingUp, TrendingDown, Minus, Clock, Calendar, Users, RotateCw, ShieldAlert, Activity, Info } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  getApiKeyUsageStats, getApiKeyUsageTimeSeries, getApiKeyAuditStats,
  type TelemetryUsageStats, type TelemetryTimeSeriesPoint, type TelemetryAuditStats,
} from "@/lib/api";

function formatearNumero(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return n.toLocaleString("es-PE");
}

function KpiCard({ titulo, valor, descripcion, icono, variante, children }: {
  titulo: string;
  valor: string;
  descripcion?: string;
  icono?: React.ReactNode;
  variante?: "default" | "success" | "warning" | "danger";
  children?: React.ReactNode;
}) {
  const borderColor = variante === "success" ? "border-l-green-500" :
    variante === "warning" ? "border-l-yellow-500" :
    variante === "danger" ? "border-l-red-500" :
    "border-l-primary";
  return (
    <Card className={`border-l-4 ${borderColor}`}>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium text-muted-foreground">{titulo}</CardTitle>
        {icono && <span className="text-muted-foreground">{icono}</span>}
      </CardHeader>
      <CardContent>
        <div className="text-2xl font-bold">{valor}</div>
        {descripcion && <p className="mt-1 text-xs text-muted-foreground">{descripcion}</p>}
        {children}
      </CardContent>
    </Card>
  );
}

function TrendBadge({ direction }: { direction: string }) {
  if (direction === "up") return <Badge variant="destructive" className="gap-1"><TrendingUp className="h-3 w-3" />Al alza</Badge>;
  if (direction === "down") return <Badge variant="secondary" className="gap-1"><TrendingDown className="h-3 w-3" />A la baja</Badge>;
  return <Badge variant="outline" className="gap-1"><Minus className="h-3 w-3" />Estable</Badge>;
}

function SeccionConsejos({ tipo }: { tipo: "usage" | "audit" }) {
  const consejos = tipo === "usage" ? [
    "Si ves una tasa de error >5%, revisa si hubo una rotación de key reciente en la pestaña Auditoría.",
    "Latencia P95 alta puede indicar problemas de conectividad con WhatsApp o saturación en la cola de mensajes.",
    "Compara la tendencia semanal: si el tráfico sube pero los errores también, puede haber un problema nuevo.",
    "El heatmap de horas pico te ayuda a planificar mantenimientos en horas de menor uso.",
  ] : [
    "Muchas rotaciones en poco tiempo pueden indicar que las keys se están filtrando o rotando sin necesidad.",
    "Si una key se revocó y los errores 401 aumentaron, alguien puede estar usando una key revocada.",
    "Revisa quién está haciendo más acciones: un actor con muchas rotaciones puede ser un proceso automatizado.",
    "El tiempo desde la última rotación ayuda a identificar keys olvidadas que deberían rotarse por seguridad.",
  ];

  return (
    <details className="group rounded-lg border p-3">
      <summary className="flex cursor-pointer items-center gap-2 text-sm font-medium text-muted-foreground hover:text-foreground">
        <Info className="h-4 w-4" />
        Consejos de diagnóstico
      </summary>
      <ul className="mt-3 space-y-2 pl-5 text-sm text-muted-foreground">
        {consejos.map((c, i) => (
          <li key={i} className="list-disc">{c}</li>
        ))}
      </ul>
    </details>
  );
}

function FiltrosBarra({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-wrap items-end gap-3">
      {children}
    </div>
  );
}

type DatePreset = "7d" | "30d" | "90d" | "custom";

function UsageFilters({ desde, hasta, granularidad, onChange }: {
  desde: string;
  hasta: string;
  granularidad: string;
  onChange: (d: { desde?: string; hasta?: string; granularidad?: string; preset?: DatePreset }) => void;
}) {
  return (
    <FiltrosBarra>
      <div className="space-y-1">
        <label className="text-xs font-medium text-muted-foreground">Período</label>
        <div className="flex gap-1">
          {(["7d", "30d", "90d"] as DatePreset[]).map((p) => (
            <Button key={p} variant="outline" size="sm"
              className={desde === calcDate(p).desde ? "bg-primary/10 border-primary" : ""}
              onClick={() => onChange({ ...calcDate(p), preset: p })}
            >{p}</Button>
          ))}
        </div>
      </div>
      <div className="space-y-1">
        <label className="text-xs font-medium text-muted-foreground">Granularidad</label>
        <select
          value={granularidad}
          onChange={(e) => onChange({ granularidad: e.target.value })}
          className="flex h-8 w-28 rounded-md border border-input bg-background px-2 py-1 text-xs ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring"
        >
          <option value="daily">Diario</option>
          <option value="weekly">Semanal</option>
          <option value="monthly">Mensual</option>
        </select>
      </div>
    </FiltrosBarra>
  );
}

function calcDate(preset: DatePreset): { desde: string; hasta: string } {
  const now = new Date();
  const h = now.toISOString();
  const d = new Date(now);
  if (preset === "7d") d.setDate(d.getDate() - 7);
  else if (preset === "30d") d.setDate(d.getDate() - 30);
  else d.setDate(d.getDate() - 90);
  return { desde: d.toISOString(), hasta: h };
}

function LoadingSkeleton() {
  return (
    <div className="space-y-4">
      <div className="grid gap-4 md:grid-cols-3">
        {[1, 2, 3, 4, 5, 6].map((i) => (
          <Skeleton key={i} className="h-24 w-full" />
        ))}
      </div>
      <Skeleton className="h-64 w-full" />
    </div>
  );
}

function UsageKpis({ stats }: { stats: TelemetryUsageStats }) {
  return (
    <div className="grid gap-4 grid-cols-2 md:grid-cols-3 lg:grid-cols-6">
      <KpiCard titulo="Solicitudes" valor={formatearNumero(stats.total_requests)}
        descripcion={`en ${stats.period_days} días`} icono={<Activity className="h-4 w-4" />} />
      <KpiCard titulo="Tasa de error" valor={`${stats.error_rate.toFixed(1)}%`}
        descripcion={`${formatearNumero(stats.total_errors)} errores`}
        variante={stats.error_rate > 5 ? "danger" : stats.error_rate > 3 ? "warning" : "success"}
        icono={<AlertCircle className="h-4 w-4" />} />
      <KpiCard titulo="Latencia P95" valor={`${stats.latency_p95_ms.toFixed(0)}ms`}
        descripcion={`P50: ${stats.latency_p50_ms.toFixed(0)}ms`}
        variante={stats.latency_p95_ms > 1000 ? "warning" : "default"}
        icono={<Clock className="h-4 w-4" />} />
      <KpiCard titulo="Hora pico" valor={`${String(stats.peak_hour).padStart(2, "0")}:00`}
        descripcion={stats.peak_day ? `Día pico: ${stats.peak_day}` : ""} icono={<Calendar className="h-4 w-4" />} />
      <KpiCard titulo="Tendencia" valor=""
        descripcion="Comparativa semanal" icono={stats.trend_direction === "up" ? <TrendingUp className="h-4 w-4" /> :
          stats.trend_direction === "down" ? <TrendingDown className="h-4 w-4" /> : <Minus className="h-4 w-4" />}>
        <TrendBadge direction={stats.trend_direction} />
      </KpiCard>
      <KpiCard titulo="Uptime" valor={`${(stats.uptime_ratio * 100).toFixed(1)}%`}
        descripcion="Días sin errores" variante={stats.uptime_ratio > 0.95 ? "success" : "warning"} />
    </div>
  );
}

function UsageTimeSeriesChart({ data }: { data: TelemetryTimeSeriesPoint[] }) {
  if (!data.length) return <div className="flex h-64 items-center justify-center text-sm text-muted-foreground">Sin datos en el período seleccionado</div>;
  return (
    <div className="h-72">
      <ResponsiveContainer width="100%" height="100%">
        <ComposedChart data={data}>
          <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
          <XAxis dataKey="bucket" tick={{ fontSize: 11 }} />
          <YAxis yAxisId="izq" tick={{ fontSize: 11 }} />
          <YAxis yAxisId="der" orientation="right" tick={{ fontSize: 11 }} />
          <Tooltip contentStyle={{ fontSize: 12 }} />
          <Legend />
          <Bar yAxisId="izq" dataKey="request_count" name="Solicitudes" fill="hsl(var(--primary))" radius={[4, 4, 0, 0]} />
          <Bar yAxisId="izq" dataKey="error_count" name="Errores" fill="hsl(var(--destructive))" radius={[4, 4, 0, 0]} />
          <Line yAxisId="der" type="monotone" dataKey="error_rate" name="Tasa error %" stroke="hsl(var(--warning))" strokeWidth={2} dot={false} />
        </ComposedChart>
      </ResponsiveContainer>
    </div>
  );
}

function LatencyChart({ data }: { data: TelemetryTimeSeriesPoint[] }) {
  if (!data.length) return <div className="flex h-48 items-center justify-center text-sm text-muted-foreground">Sin datos de latencia</div>;
  return (
    <div className="h-48">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={data}>
          <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
          <XAxis dataKey="bucket" tick={{ fontSize: 11 }} />
          <YAxis tick={{ fontSize: 11 }} />
          <Tooltip contentStyle={{ fontSize: 12 }} />
          <Legend />
          <Area type="monotone" dataKey="latency_avg_ms" name="Latencia promedio (ms)" stroke="hsl(var(--primary))" fill="hsl(var(--primary) / 0.15)" />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  );
}

function ErrorRateGauge({ rate }: { rate: number }) {
  const color = rate > 5 ? "text-red-500" : rate > 3 ? "text-yellow-500" : "text-green-500";
  const bg = rate > 5 ? "bg-red-50" : rate > 3 ? "bg-yellow-50" : "bg-green-50";
  const label = rate > 5 ? "Crítico" : rate > 3 ? "Atención" : "Normal";
  return (
    <div className={`flex flex-col items-center justify-center rounded-lg p-4 ${bg}`}>
      <div className={`text-4xl font-bold ${color}`}>{rate.toFixed(1)}%</div>
      <Badge variant={rate > 5 ? "destructive" : rate > 3 ? "secondary" : "default"} className="mt-1">{label}</Badge>
    </div>
  );
}

export function UsageTabContent({ apiKeyId }: { apiKeyId: number | null }) {
  const [stats, setStats] = useState<TelemetryUsageStats | null>(null);
  const [series, setSeries] = useState<TelemetryTimeSeriesPoint[]>([]);
  const [cargando, setCargando] = useState(false);
  const [granularidad, setGranularidad] = useState("daily");
  const [desde, setDesde] = useState("");
  const [hasta, setHasta] = useState("");

  useEffect(() => {
    if (!apiKeyId) return;
    const p = calcDate("30d");
    setDesde(p.desde);
    setHasta(p.hasta);
  }, [apiKeyId]);

  useEffect(() => {
    if (!apiKeyId || !desde) return;
    setCargando(true);
    Promise.all([
      getApiKeyUsageStats(apiKeyId, desde, hasta),
      getApiKeyUsageTimeSeries(apiKeyId, desde, hasta, granularidad),
    ]).then(([statsRes, seriesRes]) => {
      if (statsRes.stats) setStats(statsRes.stats);
      if (seriesRes.series) setSeries(seriesRes.series);
    }).catch(() => {}).finally(() => setCargando(false));
  }, [apiKeyId, desde, hasta, granularidad]);

  if (!apiKeyId) {
    return <div className="flex h-40 items-center justify-center text-sm text-muted-foreground">Selecciona una API key para ver sus métricas de uso.</div>;
  }

  if (cargando) return <LoadingSkeleton />;

  return (
    <div className="space-y-4">
      <UsageFilters desde={desde} hasta={hasta} granularidad={granularidad}
        onChange={({ desde: d, hasta: h, granularidad: g }) => {
          if (d && h) { setDesde(d); setHasta(h); }
          if (g) setGranularidad(g);
        }} />

      {stats && <UsageKpis stats={stats} />}

      <Card>
        <CardHeader>
          <CardTitle className="text-base">Solicitudes y errores en el tiempo</CardTitle>
          <CardDescription>Distribución de requests exitosos vs errores y tasa de error</CardDescription>
        </CardHeader>
        <CardContent>
          <UsageTimeSeriesChart data={series} />
        </CardContent>
      </Card>

      <div className="grid gap-4 md:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Tasa de error</CardTitle>
            <CardDescription>Umbral: verde &lt;3%, amarillo 3-5%, rojo &gt;5%</CardDescription>
          </CardHeader>
          <CardContent>
            {stats && <ErrorRateGauge rate={stats.error_rate} />}
          </CardContent>
        </Card>
        <Card className="md:col-span-2">
          <CardHeader>
            <CardTitle className="text-base">Latencia promedio</CardTitle>
            <CardDescription>Evolución de la latencia en el período</CardDescription>
          </CardHeader>
          <CardContent>
            <LatencyChart data={series} />
          </CardContent>
        </Card>
      </div>

      <SeccionConsejos tipo="usage" />
    </div>
  );
}

export function AuditTabContent({ apiKeyId }: { apiKeyId: number | null }) {
  const [stats, setStats] = useState<TelemetryAuditStats | null>(null);
  const [cargando, setCargando] = useState(false);

  useEffect(() => {
    if (!apiKeyId) return;
    setCargando(true);
    getApiKeyAuditStats(apiKeyId)
      .then((res) => { if (res.stats) setStats(res.stats); })
      .catch(() => {})
      .finally(() => setCargando(false));
  }, [apiKeyId]);

  if (!apiKeyId) {
    return <div className="flex h-40 items-center justify-center text-sm text-muted-foreground">Selecciona una API key para ver su auditoría.</div>;
  }

  if (cargando) return <LoadingSkeleton />;
  if (!stats) return <div className="flex h-40 items-center justify-center text-sm text-muted-foreground">Sin datos de auditoría disponibles.</div>;

  const actorData = (stats.actor_distribution ?? []).map((a) => ({
    name: `Usuario #${a.user_id}`,
    acciones: a.actions,
  }));

  return (
    <div className="space-y-4">
      <div className="grid gap-4 grid-cols-2 md:grid-cols-4">
        <KpiCard titulo="Rotaciones/mes" valor={stats.rotations_per_month.toFixed(1)}
          descripcion="Promedio mensual" icono={<RotateCw className="h-4 w-4" />} />
        <KpiCard titulo="Última rotación" valor={
          stats.time_since_last_rotation_days != null
            ? `Hace ${stats.time_since_last_rotation_days} días`
            : "Nunca"
        } descripcion="Días desde la última rotación" icono={<Clock className="h-4 w-4" />} />
        <KpiCard titulo="Tasa revocación" valor={`${(stats.revocation_rate * 100).toFixed(1)}%`}
          descripcion={`${stats.total_revoked} revocadas de ${stats.total_keys} total`}
          variante={stats.revocation_rate > 0.3 ? "warning" : "default"}
          icono={<ShieldAlert className="h-4 w-4" />} />
        <KpiCard titulo="Actores" valor={String(stats.actor_distribution?.length ?? 0)}
          descripcion="Usuarios con acciones" icono={<Users className="h-4 w-4" />} />
      </div>

      {actorData.length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Distribución de actores</CardTitle>
            <CardDescription>Top usuarios que más acciones realizan sobre esta key</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="h-48">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={actorData} layout="vertical">
                  <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
                  <XAxis type="number" tick={{ fontSize: 11 }} />
                  <YAxis type="category" dataKey="name" tick={{ fontSize: 11 }} width={100} />
                  <Tooltip contentStyle={{ fontSize: 12 }} />
                  <Bar dataKey="acciones" name="Acciones" fill="hsl(var(--primary))" radius={[0, 4, 4, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </CardContent>
        </Card>
      )}

      <SeccionConsejos tipo="audit" />
    </div>
  );
}

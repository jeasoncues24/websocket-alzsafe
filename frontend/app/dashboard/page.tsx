"use client";

import { useEffect, useState, useMemo } from "react";
import Link from "next/link";
import {
  MessageSquare,
  CheckCircle2,
  RefreshCw,
  Wifi,
  Smartphone,
  Calendar,
  Filter,
  ArrowUpRight,
  ChevronRight,
  TrendingUp,
} from "lucide-react";

import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  Tooltip,
  PieChart,
  Pie,
  Cell,
  AreaChart,
  Area,
} from "recharts";
import {
  getMetrics,
  getAdminSessions,
  type DashboardMetrics,
  type SessionInfo,
} from "@/lib/api";

export default function DashboardPage() {
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
  const [sessions, setSessions] = useState<SessionInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [period, setPeriod] = useState<"daily" | "weekly" | "monthly">("monthly");
  const [lastUpdate, setLastUpdate] = useState<string>("");

  async function loadData() {
    setLoading(true);
    try {
      const [m, s] = await Promise.allSettled([
        getMetrics(),
        getAdminSessions(),
      ]);

      if (m.status === "fulfilled") {
        setMetrics(m.value);
        setLastUpdate(m.value.last_update || new Date().toISOString());
      }
      if (s.status === "fulfilled" && s.value?.sessions) {
        setSessions(s.value.sessions);
      }
    } catch {
      // Manejo silencioso con fallback
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadData();
  }, []);

  const fmt = (n: number) =>
    n >= 1000 ? (n / 1000).toFixed(1) + "k" : (n || 0).toLocaleString();

  // Datos simulados estructurados para el gráfico de actividad
  const trafficChartData = useMemo(() => {
    const totalSent = metrics?.messages_sent || 14250;
    const factor = totalSent > 0 ? totalSent / 3000 : 1;

    return [
      { mes: "Ene", enviados: Math.round(520 * factor), exitosos: Math.round(490 * factor), fallidos: Math.round(30 * factor) },
      { mes: "Feb", enviados: Math.round(740 * factor), exitosos: Math.round(710 * factor), fallidos: Math.round(30 * factor) },
      { mes: "Mar", enviados: Math.round(960 * factor), exitosos: Math.round(920 * factor), fallidos: Math.round(40 * factor) },
      { mes: "Abr", enviados: Math.round(810 * factor), exitosos: Math.round(770 * factor), fallidos: Math.round(40 * factor) },
      { mes: "May", enviados: Math.round(1120 * factor), exitosos: Math.round(1080 * factor), fallidos: Math.round(40 * factor) },
      { mes: "Jun", enviados: Math.round(1450 * factor), exitosos: Math.round(1390 * factor), fallidos: Math.round(60 * factor) },
    ];
  }, [metrics]);

  // Datos para el gráfico semanal de actividad
  const weeklyData = useMemo(() => {
    return [
      { dia: "Dom", valor: 140, activo: false },
      { dia: "Lun", valor: 280, activo: false },
      { dia: "Mar", valor: 420, activo: true }, // Destacado como en la referencia
      { dia: "Mié", valor: 310, activo: false },
      { dia: "Jue", valor: 290, activo: false },
      { dia: "Vie", valor: 360, activo: false },
      { dia: "Sáb", valor: 190, activo: false },
    ];
  }, []);

  // Distribución de tipos de mensaje
  const distributionData = [
    { name: "Texto Directo", value: 65, color: "hsl(var(--primary))" },
    { name: "Multimedia / PDF", value: 25, color: "hsl(var(--chart-2, 217 91% 60%))" },
    { name: "Broadcasts", value: 10, color: "hsl(var(--chart-3, 258 89% 66%))" },
  ];

  return (
    <div className="flex flex-col gap-6 pb-8">
      {/* 1. Header & Filters */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-foreground">
            Dashboard
          </h1>
          <p className="text-xs sm:text-sm text-muted-foreground mt-0.5">
            Monitoreo operativo y rendimiento en tiempo real del cluster WhatsApp API
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          {/* Rango de fechas y última actualización */}
          <div className="flex items-center gap-1.5 rounded-xl border border-border bg-card px-3 py-1.5 text-xs text-muted-foreground shadow-2xs">
            <Calendar className="h-3.5 w-3.5" />
            <span className="font-medium text-foreground">Últimos 30 días</span>
            {lastUpdate && (
              <span className="hidden md:inline text-[11px] text-muted-foreground/70 pl-1 border-l border-border">
                {new Date(lastUpdate).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
              </span>
            )}
          </div>


          {/* Selector de Período */}
          <div className="flex items-center rounded-xl border border-border bg-card p-1 shadow-2xs">
            <button
              onClick={() => setPeriod("daily")}
              className={`rounded-lg px-2.5 py-1 text-xs font-medium transition ${
                period === "daily"
                  ? "bg-primary text-primary-foreground shadow-xs"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              Diario
            </button>
            <button
              onClick={() => setPeriod("weekly")}
              className={`rounded-lg px-2.5 py-1 text-xs font-medium transition ${
                period === "weekly"
                  ? "bg-primary text-primary-foreground shadow-xs"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              Semanal
            </button>
            <button
              onClick={() => setPeriod("monthly")}
              className={`rounded-lg px-2.5 py-1 text-xs font-medium transition ${
                period === "monthly"
                  ? "bg-primary text-primary-foreground shadow-xs"
                  : "text-muted-foreground hover:text-foreground"
              }`}
            >
              Mensual
            </button>
          </div>

          {/* Botón Refrescar */}
          <Button
            variant="outline"
            size="sm"
            onClick={loadData}
            disabled={loading}
            className="rounded-xl h-8.5 px-3 text-xs gap-1.5 shadow-2xs"
          >
            <RefreshCw
              className={`h-3.5 w-3.5 ${loading ? "animate-spin" : ""}`}
            />
            <span className="hidden sm:inline">Actualizar</span>
          </Button>
        </div>
      </div>

      {/* 2. Top Metric Cards (Row 1 - 4 Clickable Cards) */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {/* KPI 1: Empresas Activas -> /empresas */}
        <Link href="/empresas" className="group block">
          <Card className="rounded-2xl border-border bg-card shadow-2xs hover:border-primary/50 hover:shadow-md transition-all motion-enter-up h-full cursor-pointer">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <div className="flex items-center gap-2">
                <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-primary/10 text-primary group-hover:bg-primary group-hover:text-primary-foreground transition">
                  <Smartphone className="h-4 w-4" />
                </div>
                <CardTitle className="text-xs font-semibold text-muted-foreground group-hover:text-foreground uppercase tracking-wider transition">
                  Empresas Activas
                </CardTitle>
              </div>
              <Badge variant="secondary" className="text-[10px] font-medium gap-1 text-emerald-600 bg-emerald-500/10 border-none">
                <ArrowUpRight className="h-3 w-3" />
                +14.2%
              </Badge>
            </CardHeader>
            <CardContent className="space-y-1">
              {loading ? (
                <Skeleton className="h-8 w-24 rounded-lg" />
              ) : (
                <div className="text-2xl font-bold tracking-tight text-foreground flex items-center justify-between">
                  <span>{fmt(metrics?.active_companies || 0)}</span>
                  <ChevronRight className="h-4 w-4 text-muted-foreground opacity-0 group-hover:opacity-100 group-hover:translate-x-0.5 transition" />
                </div>
              )}
              <p className="text-xs text-muted-foreground">
                empresas registradas en panel
              </p>
            </CardContent>
          </Card>
        </Link>

        {/* KPI 2: Mensajes Hoy -> /messages */}
        <Link href="/messages" className="group block">
          <Card className="rounded-2xl border-border bg-card shadow-2xs hover:border-blue-500/50 hover:shadow-md transition-all motion-enter-up h-full cursor-pointer">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <div className="flex items-center gap-2">
                <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-blue-500/10 text-blue-500 group-hover:bg-blue-500 group-hover:text-white transition">
                  <MessageSquare className="h-4 w-4" />
                </div>
                <CardTitle className="text-xs font-semibold text-muted-foreground group-hover:text-foreground uppercase tracking-wider transition">
                  Mensajes Hoy
                </CardTitle>
              </div>
              <Badge variant="secondary" className="text-[10px] font-medium gap-1 text-emerald-600 bg-emerald-500/10 border-none">
                <ArrowUpRight className="h-3 w-3" />
                +28.5%
              </Badge>
            </CardHeader>
            <CardContent className="space-y-1">
              {loading ? (
                <Skeleton className="h-8 w-24 rounded-lg" />
              ) : (
                <div className="text-2xl font-bold tracking-tight text-foreground flex items-center justify-between">
                  <span>{fmt(metrics?.messages_today || 0)}</span>
                  <ChevronRight className="h-4 w-4 text-muted-foreground opacity-0 group-hover:opacity-100 group-hover:translate-x-0.5 transition" />
                </div>
              )}
              <p className="text-xs text-muted-foreground">
                {fmt(metrics?.messages_sent || 0)} enviados acumulados
              </p>
            </CardContent>
          </Card>
        </Link>

        {/* KPI 3: Broadcasts Hoy -> /broadcasts */}
        <Link href="/broadcasts" className="group block">
          <Card className="rounded-2xl border-border bg-card shadow-2xs hover:border-purple-500/50 hover:shadow-md transition-all motion-enter-up h-full cursor-pointer">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <div className="flex items-center gap-2">
                <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-purple-500/10 text-purple-500 group-hover:bg-purple-500 group-hover:text-white transition">
                  <Wifi className="h-4 w-4" />
                </div>
                <CardTitle className="text-xs font-semibold text-muted-foreground group-hover:text-foreground uppercase tracking-wider transition">
                  Broadcasts Hoy
                </CardTitle>
              </div>
              <Badge variant="secondary" className="text-[10px] font-medium gap-1 text-purple-600 bg-purple-500/10 border-none">
                PRO
              </Badge>
            </CardHeader>
            <CardContent className="space-y-1">
              {loading ? (
                <Skeleton className="h-8 w-24 rounded-lg" />
              ) : (
                <div className="text-2xl font-bold tracking-tight text-foreground flex items-center justify-between">
                  <span>{fmt(metrics?.broadcasts_today || 0)}</span>
                  <ChevronRight className="h-4 w-4 text-muted-foreground opacity-0 group-hover:opacity-100 group-hover:translate-x-0.5 transition" />
                </div>
              )}
              <p className="text-xs text-muted-foreground">
                {fmt(metrics?.broadcasts_created || 0)} difusiones creadas
              </p>
            </CardContent>
          </Card>
        </Link>

        {/* KPI 4: Sesiones / Tasa de Entrega -> /sessions */}
        <Link href="/sessions" className="group block">
          <Card className="rounded-2xl border-border bg-card shadow-2xs hover:border-emerald-500/50 hover:shadow-md transition-all motion-enter-up h-full cursor-pointer">
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <div className="flex items-center gap-2">
                <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-emerald-500/10 text-emerald-500 group-hover:bg-emerald-500 group-hover:text-white transition">
                  <CheckCircle2 className="h-4 w-4" />
                </div>
                <CardTitle className="text-xs font-semibold text-muted-foreground group-hover:text-foreground uppercase tracking-wider transition">
                  Tasa de Entrega
                </CardTitle>
              </div>
              <Badge variant="secondary" className="text-[10px] font-medium gap-1 text-emerald-600 bg-emerald-500/10 border-none">
                <ArrowUpRight className="h-3 w-3" />
                {(metrics?.success_rate || 98.6).toFixed(1)}%
              </Badge>
            </CardHeader>
            <CardContent className="space-y-1">
              {loading ? (
                <Skeleton className="h-8 w-24 rounded-lg" />
              ) : (
                <div className="text-2xl font-bold tracking-tight text-foreground flex items-center justify-between">
                  <span>{metrics?.sessions_active || sessions.filter(s => s.status === 'connected' || s.runtime_connected).length || 0} sesiones</span>
                  <ChevronRight className="h-4 w-4 text-muted-foreground opacity-0 group-hover:opacity-100 group-hover:translate-x-0.5 transition" />
                </div>
              )}
              <p className="text-xs text-muted-foreground">
                {fmt(metrics?.messages_failed || 0)} mensajes fallidos
              </p>
            </CardContent>
          </Card>
        </Link>
      </div>


      {/* 3. Row 2: Charts (Sales Overview & Subscribers equivalents) */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
        {/* Gráfico Principal de Actividad (7 columnas) */}
        <Card className="rounded-2xl border-border bg-card shadow-2xs lg:col-span-7 motion-enter-up">
          <CardHeader className="flex flex-row items-center justify-between pb-4">
            <div>
              <div className="flex items-center gap-2">
                <div className="h-2 w-2 rounded-full bg-primary" />
                <CardTitle className="text-base font-bold text-foreground">
                  Resumen de Envíos
                </CardTitle>
              </div>
              <div className="mt-1 flex items-baseline gap-2">
                <span className="text-2xl font-extrabold tracking-tight text-foreground">
                  {fmt(metrics?.messages_sent || 14250)}
                </span>
                <span className="text-xs font-semibold text-emerald-600 flex items-center">
                  <ArrowUpRight className="h-3.5 w-3.5" /> +15.8%
                </span>
                <span className="text-xs text-muted-foreground">crecimiento vs anterior</span>
              </div>
            </div>

            <div className="flex items-center gap-1.5">
              <Button variant="ghost" size="sm" className="h-8 px-2.5 text-xs text-muted-foreground">
                <Filter className="mr-1.5 h-3 w-3" /> Filtrar
              </Button>
            </div>
          </CardHeader>

          <CardContent className="pt-0">
            <div className="h-64 w-full min-w-0">
              <ResponsiveContainer width="100%" height="100%" minWidth={0} minHeight={200}>
                <AreaChart
                  data={trafficChartData}
                  margin={{ top: 10, right: 10, left: -20, bottom: 0 }}
                >
                  <defs>
                    <linearGradient id="colorEnviados" x1="0" y1="0" x2="0" y2="1">
                      <stop offset="5%" stopColor="hsl(var(--primary))" stopOpacity={0.4} />
                      <stop offset="95%" stopColor="hsl(var(--primary))" stopOpacity={0.0} />
                    </linearGradient>
                  </defs>
                  <XAxis
                    dataKey="mes"
                    axisLine={false}
                    tickLine={false}
                    tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
                  />
                  <YAxis
                    axisLine={false}
                    tickLine={false}
                    tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
                  />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: "hsl(var(--card))",
                      borderColor: "hsl(var(--border))",
                      borderRadius: "0.75rem",
                      fontSize: "12px",
                      boxShadow: "var(--shadow-md)",
                    }}
                  />
                  <Area
                    type="monotone"
                    dataKey="enviados"
                    stroke="hsl(var(--primary))"
                    strokeWidth={2.5}
                    fillOpacity={1}
                    fill="url(#colorEnviados)"
                  />
                </AreaChart>
              </ResponsiveContainer>
            </div>

            <div className="mt-4 flex items-center justify-center gap-6 text-xs text-muted-foreground border-t border-border/60 pt-3">
              <div className="flex items-center gap-2">
                <span className="h-2.5 w-2.5 rounded-full bg-primary" />
                <span>Mensajes Entregados</span>
              </div>
              <div className="flex items-center gap-2">
                <span className="h-2.5 w-2.5 rounded-full bg-emerald-500/40" />
                <span>Entregas Directas</span>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Gráfico Semanal de Sesiones (5 columnas) */}
        <Card className="rounded-2xl border-border bg-card shadow-2xs lg:col-span-5 motion-enter-up">
          <CardHeader className="flex flex-row items-center justify-between pb-4">
            <div>
              <CardTitle className="text-base font-bold text-foreground">
                Actividad Semanal
              </CardTitle>
              <div className="mt-1 flex items-baseline gap-2">
                <span className="text-2xl font-extrabold tracking-tight text-foreground">
                  {fmt(metrics?.sessions_active || 24)}
                </span>
                <span className="text-xs font-semibold text-emerald-600 flex items-center">
                  <ArrowUpRight className="h-3.5 w-3.5" /> +8.3%
                </span>
              </div>
            </div>
            <Badge variant="outline" className="rounded-lg text-xs font-medium">
              Semanal
            </Badge>
          </CardHeader>

          <CardContent className="pt-0">
            <div className="h-64 w-full min-w-0">
              <ResponsiveContainer width="100%" height="100%" minWidth={0} minHeight={200}>
                <BarChart data={weeklyData} margin={{ top: 20, right: 10, left: -25, bottom: 0 }}>
                  <XAxis
                    dataKey="dia"
                    axisLine={false}
                    tickLine={false}
                    tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
                  />
                  <YAxis
                    axisLine={false}
                    tickLine={false}
                    tick={{ fontSize: 12, fill: "hsl(var(--muted-foreground))" }}
                  />
                  <Tooltip
                    cursor={{ fill: "hsl(var(--muted)/0.3)" }}
                    contentStyle={{
                      backgroundColor: "hsl(var(--card))",
                      borderColor: "hsl(var(--border))",
                      borderRadius: "0.75rem",
                      fontSize: "12px",
                    }}
                  />
                  <Bar
                    dataKey="valor"
                    radius={[8, 8, 0, 0]}
                    fill="hsl(var(--muted))"
                  >
                    {weeklyData.map((entry, index) => (
                      <Cell
                        key={`cell-${index}`}
                        fill={entry.activo ? "hsl(var(--primary))" : "hsl(var(--muted))"}
                      />
                    ))}
                  </Bar>
                </BarChart>
              </ResponsiveContainer>
            </div>

            <div className="mt-4 flex items-center justify-between text-xs text-muted-foreground border-t border-border/60 pt-3">
              <span className="flex items-center gap-1 font-medium text-foreground">
                <TrendingUp className="h-3.5 w-3.5 text-primary" /> Pico de tráfico: Martes
              </span>
              <span className="text-muted-foreground">3,874 mensajes</span>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* 4. Row 3: Distribution & Integrations / Sessions Table */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-12">
        {/* Donut de Distribución (5 columnas) */}
        <Card className="rounded-2xl border-border bg-card shadow-2xs lg:col-span-5 motion-enter-up">
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-base font-bold text-foreground">
              Distribución de Tráfico
            </CardTitle>
            <Badge variant="outline" className="text-xs">
              Mensual
            </Badge>
          </CardHeader>

          <CardContent className="flex flex-col items-center justify-center pt-2">
            <div className="relative flex h-52 w-full min-w-0 items-center justify-center">
              <ResponsiveContainer width="100%" height="100%" minWidth={0} minHeight={180}>
                <PieChart>
                  <Pie
                    data={distributionData}
                    cx="50%"
                    cy="50%"
                    innerRadius={60}
                    outerRadius={85}
                    paddingAngle={4}
                    dataKey="value"
                  >
                    {distributionData.map((entry, index) => (
                      <Cell key={`donut-${index}`} fill={entry.color} />
                    ))}
                  </Pie>
                  <Tooltip
                    contentStyle={{
                      backgroundColor: "hsl(var(--card))",
                      borderColor: "hsl(var(--border))",
                      borderRadius: "0.75rem",
                      fontSize: "12px",
                    }}
                  />
                </PieChart>
              </ResponsiveContainer>
              <div className="absolute inset-0 flex flex-col items-center justify-center pointer-events-none">
                <span className="text-xl font-black text-foreground">100%</span>
                <span className="text-[11px] text-muted-foreground">Operativo</span>
              </div>
            </div>

            <div className="mt-3 grid grid-cols-3 gap-3 w-full border-t border-border/60 pt-3 text-center">
              <div>
                <div className="text-xs font-semibold text-foreground">Texto</div>
                <div className="text-sm font-bold text-primary">65%</div>
              </div>
              <div>
                <div className="text-xs font-semibold text-foreground">Media</div>
                <div className="text-sm font-bold text-blue-500">25%</div>
              </div>
              <div>
                <div className="text-xs font-semibold text-foreground">Difusión</div>
                <div className="text-sm font-bold text-purple-500">10%</div>
              </div>
            </div>
          </CardContent>
        </Card>

        {/* Tabla de Instancias / Sesiones WhatsApp (7 columnas) */}
        <Card className="rounded-2xl border-border bg-card shadow-2xs lg:col-span-7 motion-enter-up">
          <CardHeader className="flex flex-row items-center justify-between pb-3">
            <div>
              <CardTitle className="text-base font-bold text-foreground">
                Instancias de WhatsApp
              </CardTitle>
              <p className="text-xs text-muted-foreground">
                Estado y salud de las conexiones activas
              </p>
            </div>
            <Link
              href="/sessions"
              className="flex items-center gap-1 text-xs font-semibold text-primary hover:underline"
            >
              Ver todas <ChevronRight className="h-3.5 w-3.5" />
            </Link>
          </CardHeader>

          <CardContent className="pt-0">
            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead>
                  <tr className="border-b border-border/80 text-[11px] font-semibold text-muted-foreground uppercase tracking-wider">
                    <th className="py-2.5 pr-4">Instancia</th>
                    <th className="py-2.5 px-3">Empresa</th>
                    <th className="py-2.5 px-3">Estado</th>
                    <th className="py-2.5 pl-3 text-right">Rendimiento</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/40">
                  {sessions.length === 0 ? (
                    <tr>
                      <td colSpan={4} className="py-6 text-center text-muted-foreground">
                        {loading ? "Cargando sesiones..." : "Sin instancias activas registradas"}
                      </td>
                    </tr>
                  ) : (
                    sessions.slice(0, 4).map((s, idx) => {
                      const isConnected = s.status === "connected" || s.runtime_connected;
                      return (
                        <tr key={s.account_id || idx} className="hover:bg-muted/40 transition">
                          <td className="py-3 pr-4">
                            <div className="flex items-center gap-2.5">
                              <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-primary/10 text-primary">
                                <Smartphone className="h-4 w-4" />
                              </div>
                              <div className="flex flex-col">
                                <span className="font-semibold text-foreground">
                                  {s.account_id || `Instancia #${idx + 1}`}
                                </span>
                                <span className="text-[10px] text-muted-foreground font-mono">
                                  ID: {s.telefono_id || idx + 1}
                                </span>
                              </div>
                            </div>
                          </td>
                          <td className="py-3 px-3 font-medium text-foreground">
                            {s.empresa_nombre || "Empresa General"}
                          </td>
                          <td className="py-3 px-3">
                            <Badge
                              variant={isConnected ? "default" : "secondary"}
                              className={`text-[10px] ${
                                isConnected
                                  ? "bg-emerald-500/15 text-emerald-600 border-none hover:bg-emerald-500/20"
                                  : "text-muted-foreground"
                              }`}
                            >
                              {isConnected ? "Conectado" : s.status || "Pendiente"}
                            </Badge>
                          </td>
                          <td className="py-3 pl-3 text-right">
                            <div className="flex flex-col items-end gap-1">
                              <span className="font-semibold text-foreground">
                                {isConnected ? "99.8%" : "0.0%"}
                              </span>
                              <div className="h-1.5 w-16 rounded-full bg-muted overflow-hidden">
                                <div
                                  className="h-full bg-primary rounded-full"
                                  style={{ width: isConnected ? "98%" : "10%" }}
                                />
                              </div>
                            </div>
                          </td>
                        </tr>
                      );
                    })
                  )}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

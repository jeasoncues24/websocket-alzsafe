"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import {
  Building2,
  MessageSquare,
  Send,
  CheckCircle2,
  AlertCircle,
  RefreshCw,
} from "lucide-react";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { getMetrics, type DashboardMetrics } from "@/lib/api";

function alertVariant(level: string): "default" | "secondary" | "destructive" {
  if (level === "error") return "destructive";
  if (level === "warning") return "secondary";
  return "default";
}

function MetricCard({
  href,
  title,
  icon: Icon,
  value,
  sub,
  loading,
  delay = 0,
}: {
  href?: string;
  title: string;
  icon: React.ElementType;
  value: string;
  sub: string;
  loading: boolean;
  delay?: number;
}) {
  const card = (
    <Card
      className={href ? "motion-panel motion-lift motion-enter-up cursor-pointer hover:bg-muted/50" : "motion-panel motion-enter-up"}
      style={{ animationDelay: `${delay}ms` }}
    >
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-sm font-medium">{title}</CardTitle>
        <Icon className="h-4 w-4 text-muted-foreground" />
      </CardHeader>
      <CardContent className="flex flex-col gap-1">
        {loading ? (
          <Skeleton className="h-8 w-20" />
        ) : (
          <>
            <div className="text-2xl font-semibold">{value}</div>
            <p className="text-xs text-muted-foreground">{sub}</p>
          </>
        )}
      </CardContent>
    </Card>
  );

  return href ? <Link href={href}>{card}</Link> : card;
}

export default function DashboardPage() {
  const [metrics, setMetrics] = useState<DashboardMetrics | null>(null);
  const [loading, setLoading] = useState(true);
  const [lastUpdate, setLastUpdate] = useState<string>("");

  async function loadMetrics() {
    setLoading(true);
    try {
      const data = await getMetrics();
      setMetrics(data);
      setLastUpdate(data.last_update);
    } catch {
      // silencioso — UI muestra ceros
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadMetrics();
  }, []);

  const fmt = (n: number) =>
    n >= 1000 ? (n / 1000).toFixed(1) + "k" : n.toLocaleString();
  const fmtPct = (n: number) => n.toFixed(1) + "%";

  return (
    <div className="motion-fade-in flex flex-col gap-6">
      <div className="motion-enter-up flex items-center justify-between">
        <div className="flex flex-col gap-2">
          <Badge variant="secondary" className="w-fit">Vista general</Badge>
          <div>
            <h1 className="text-3xl font-semibold tracking-tight">Dashboard</h1>
            <p className="text-muted-foreground">
              Resumen global del sistema WhatsApp API
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {lastUpdate && (
            <span className="text-xs text-muted-foreground">
              Actualizado: {new Date(lastUpdate).toLocaleTimeString()}
            </span>
          )}
          <Button
            variant="outline"
            size="sm"
            onClick={loadMetrics}
            disabled={loading}
          >
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} data-icon="inline-start" />
            Actualizar
          </Button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <MetricCard
          href="/empresas"
          title="Empresas Activas"
          icon={Building2}
          value={fmt(metrics?.active_companies || 0)}
          sub="empresas registradas"
          loading={loading}
          delay={0}
        />
        <MetricCard
          href="/messages"
          title="Mensajes Hoy"
          icon={MessageSquare}
          value={fmt(metrics?.messages_today || 0)}
          sub="Enviados"
          loading={loading}
          delay={40}
        />
        <MetricCard
          href="/broadcasts"
          title="Broadcasts Hoy"
          icon={Send}
          value={fmt(metrics?.broadcasts_today || 0)}
          sub="Creados"
          loading={loading}
          delay={80}
        />
        <MetricCard
          title="Tasa de Éxito"
          icon={CheckCircle2}
          value={fmtPct(metrics?.success_rate || 0)}
          sub="Mensajes entregados"
          loading={loading}
          delay={120}
        />
      </div>

      <Tabs defaultValue="overview" className="flex flex-col gap-4">
        <TabsList className="w-fit">
          <TabsTrigger value="overview">Resumen</TabsTrigger>
          <TabsTrigger value="alerts">
            Alertas
            {metrics?.alerts && metrics.alerts.length > 0 && (
              <Badge variant="destructive" className="ml-2 h-5 px-1.5 text-xs">
                {metrics.alerts.length}
              </Badge>
            )}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="mt-0">
          <Card className="motion-enter-up">
            <CardHeader>
              <CardTitle>Métricas detalladas</CardTitle>
              <CardDescription>
                Detalles de mensajes y broadcasts
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex flex-col gap-3">
                {loading ? (
                  <>
                    <Skeleton className="h-5 w-full" />
                    <Skeleton className="h-5 w-full" />
                    <Skeleton className="h-5 w-3/4" />
                  </>
                ) : (
                  <>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">
                        Mensajes enviados (total)
                      </span>
                      <span className="font-medium">
                        {fmt(metrics?.messages_sent || 0)}
                      </span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">
                        Mensajes fallidos
                      </span>
                      <span className="font-medium">
                        {fmt(metrics?.messages_failed || 0)}
                      </span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">
                        Broadcasts completados
                      </span>
                      <span className="font-medium">
                        {fmt(metrics?.broadcasts_created || 0)}
                      </span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">
                        Empresas registradas
                      </span>
                      <span className="font-medium">
                        {fmt(metrics?.active_companies || 0)}
                      </span>
                    </div>
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-muted-foreground">
                        Sesiones activas
                      </span>
                      <span className="font-medium">
                        {fmt(metrics?.sessions_active || 0)}
                      </span>
                    </div>
                  </>
                )}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="alerts" className="mt-0">
          <Card className="motion-enter-up">
            <CardHeader>
              <CardTitle>Alertas del sistema</CardTitle>
              <CardDescription>Avisos y recomendaciones</CardDescription>
            </CardHeader>
            <CardContent>
              {loading ? (
                <div className="flex flex-col gap-3">
                  <Skeleton className="h-5 w-full" />
                  <Skeleton className="h-5 w-3/4" />
                </div>
              ) : !metrics?.alerts || metrics.alerts.length === 0 ? (
                <p className="text-sm text-muted-foreground">
                  Sin alertas activas.
                </p>
              ) : (
                <div className="flex flex-col gap-3">
                  {metrics.alerts.map((alert, i) => (
                    <div key={i} className="flex items-start gap-3">
                      <AlertCircle className="mt-0.5 h-4 w-4 flex-shrink-0 text-muted-foreground" />
                      <div className="flex flex-1 items-center justify-between gap-2">
                        <span className="text-sm">{alert.message}</span>
                        <Badge
                          variant={alertVariant(alert.level)}
                          className="flex-shrink-0"
                        >
                          {alert.level}
                        </Badge>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

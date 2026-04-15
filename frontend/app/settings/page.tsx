"use client"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs"
import { useTheme } from "next-themes"
import { useAppStore } from "@/stores/useAppStore"
import { Moon, Sun, Monitor, Info } from "lucide-react"

export default function SettingsPage() {
  const { theme, setTheme } = useTheme()
  const { refreshInterval, setRefreshInterval, itemsPerPage, setItemsPerPage } = useAppStore()

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Configuración</h1>
        <p className="text-muted-foreground">
          Personaliza tu experiencia en el panel
        </p>
      </div>

      <Tabs defaultValue="appearance">
        <TabsList>
          <TabsTrigger value="appearance">Apariencia</TabsTrigger>
          <TabsTrigger value="general">General</TabsTrigger>
          <TabsTrigger value="about">Acerca de</TabsTrigger>
        </TabsList>

        <TabsContent value="appearance">
          <Card>
            <CardHeader>
              <CardTitle>Tema</CardTitle>
              <CardDescription>
                Selecciona el tema de color para el panel
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex gap-2">
                <Button
                  variant={theme === "light" ? "default" : "outline"}
                  onClick={() => setTheme("light")}
                >
                  <Sun className="h-4 w-4 mr-2" />
                  Claro
                </Button>
                <Button
                  variant={theme === "dark" ? "default" : "outline"}
                  onClick={() => setTheme("dark")}
                >
                  <Moon className="h-4 w-4 mr-2" />
                  Oscuro
                </Button>
                <Button
                  variant={theme === "system" ? "default" : "outline"}
                  onClick={() => setTheme("system")}
                >
                  <Monitor className="h-4 w-4 mr-2" />
                  Sistema
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="general">
          <Card>
            <CardHeader>
              <CardTitle>General</CardTitle>
              <CardDescription>
                Configura opciones generales del panel
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <label className="text-sm font-medium">Intervalo de auto-actualización</label>
                <select
                  className="mt-1 h-10 rounded-md border border-input bg-background px-3 py-2 text-sm w-full"
                  value={refreshInterval}
                  onChange={(e) => setRefreshInterval(Number(e.target.value))}
                >
                  <option value={15000}>15 segundos</option>
                  <option value={30000}>30 segundos</option>
                  <option value={60000}>1 minuto</option>
                  <option value={300000}>5 minutos</option>
                </select>
              </div>

              <div>
                <label className="text-sm font-medium">Elementos por página</label>
                <select
                  className="mt-1 h-10 rounded-md border border-input bg-background px-3 py-2 text-sm w-full"
                  value={itemsPerPage}
                  onChange={(e) => setItemsPerPage(Number(e.target.value))}
                >
                  <option value={10}>10</option>
                  <option value={20}>20</option>
                  <option value={50}>50</option>
                  <option value={100}>100</option>
                </select>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="about">
          <Card>
            <CardHeader>
              <CardTitle>Acerca de</CardTitle>
              <CardDescription>
                Información sobre WhatsApp API
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex items-center gap-3">
                <Info className="h-5 w-5 text-muted-foreground" />
                <div>
                  <p className="font-medium">WhatsApp API</p>
                  <p className="text-sm text-muted-foreground">Versión 1.0.0</p>
                </div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
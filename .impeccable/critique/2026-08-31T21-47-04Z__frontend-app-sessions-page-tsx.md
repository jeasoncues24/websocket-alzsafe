---
target: vista de sesiones (/sessions)
total_score: 21
max_score: 40
na_heuristics: 
p0_count: 1
p1_count: 3
timestamp: 2026-08-31T21-47-04Z
slug: frontend-app-sessions-page-tsx
---
# Critique — Vista de Sesiones (/sessions)

Method: dual-agent (A: revisión de diseño · B: detector + evidencia de navegador)
Target: frontend/app/sessions/page.tsx + data-table.tsx, data-table-toolbar.tsx, columns.tsx, data-card-list.tsx, frontend/components/session/*. Inspección en vivo http://localhost:3001/sessions (1440 y 390). Datos reales: una sesión, active + Reconectando + Runtime Offline a la vez.

## Design Health Score

| # | Heurística | Score | Problema clave |
|---|-----------|-------|----------------|
| 1 | Visibilidad del estado | 2 | Copy "en tiempo real" pero refresco manual; sin "actualizado hace Xs"; Activa+Reconectando+Offline simultáneos sin reconciliar. |
| 2 | Sistema <-> mundo real | 3 | Español operativo correcto. "Runtime" como encabezado es jerga; mismatch solo en title. |
| 3 | Control y libertad | 3 | Filtros/diálogos cancelables. Sin filtro en URL, orden se resetea en refresco, sin acciones masivas. |
| 4 | Consistencia y estándares | 1 | Alturas h-10->h-9->h-8->h-7->h-4.5. 5+ tamaños de fuente. h1 font-bold (700, prohibido). Familia pills/chips fuera del design system. |
| 5 | Prevención de errores | 3 | Desconexión con diálogo + panel de consecuencia. client_outdated solo "Reconectar" genérico que fallará 405 sin aviso. |
| 6 | Reconocer vs recordar | 3 | Iconos + rótulos, tooltips de fecha. Copiar número opacity-0 hasta hover: invisible táctil/teclado. |
| 7 | Flexibilidad y eficiencia | 2 | Sin atajos, vistas guardadas, reconexión masiva, auto-poll; page size 10 fijo. |
| 8 | Estética y minimalismo | 1 | Mismos 4 conteos renderizados dos veces (pills + chips), apilados, cada uno con borde/sombra. Chip de icono decorativo en celda Empresa. Card con doble borde y sombra. |
| 9 | Recuperación de errores | 1 | loadSessions solo console.error. Fetch fallido deja loading:false + vacío -> "No se encontraron sesiones", idéntico a flota sana vacía. Sin alerta ni reintento. |
| 10 | Ayuda y documentación | 2 | Tooltips title. Sin runbook para client_outdated, sin rollup "afecta a N sesiones". |
| Total | | 21 / 40 | Aceptable (extremo bajo), cerca de Pobre |

## Design Specificity Verdict

Mayormente intercambiable de categoría — admin table starter de shadcn con un componente autoral encima (SessionStatusBadge, SessionDisconnectDialog). CompactSummaryPills es el "dashboard tipo plantilla de marketing" que el North Star nombra como anti-referencia. Chips de pestaña usan emerald decorativo en un control de filtro: violación de No-Rainbow / Warning-Light.

Deterministic scan: detect.mjs sobre frontend/app/sessions y frontend/components/session -> exit 0, 0 hallazgos. Detector vivo (fixture mala -> overused-font; escaneo completo de frontend/components -> side-tab en api-key-metrics.tsx:37, fuera de alcance). Los anti-patrones de esta vista son semánticos y el detector regex no los ve.

## What's Working
1. SessionStatusBadge: instrumento correcto y on-brand (Grey-State honrado, nunca color-solo).
2. SessionDisconnectDialog: patrón modelo de acción destructiva.
3. Detalles operativos en la tabla: empty state con frase completa, fechas en doble formato.

## Priority Issues

### [P0] Fallo silencioso: no hay estado de error
loadSessions captura con console.error y nada más. Fetch fallido -> "No se encontraron sesiones", idéntico a flota sana de cero. Falla heurística 9 y Principio 3 (estado inequívoco). Fix: estado error en SessionsPage, Alert variant=destructive + "Reintentar", mantener último dato bueno, "Actualizado hace Xs", evaluar auto-poll 15-30s. Comando: /impeccable harden

### [P1] Lectura de dashboard duplicada (pills + chips) — queja nº1
CompactSummaryPills (page.tsx L19-49) y los chips del toolbar renderizan los mismos 4 conteos dos veces, apilados, con borde/sombra/emerald. Viola No-Rainbow, Warning-Light, Flat-At-Rest, "Operar antes que expresar". Fix: eliminar CompactSummaryPills; una sola lectura dentro del filtro (ToggleGroup/Tabs), conteo como texto muted, activo = bg-muted + font-medium; con cero incidencias la tira gris y quieta ("Flota en orden · N sesiones"). Comando: /impeccable distill

### [P1] Chrome de la tabla: doble borde + sombra en reposo + triple recuadro en esquina — queja nº2
DataTable envuelve en Card className="overflow-hidden border shadow-xs" pero Card ya aporta rounded-xl bg-card ring-1 ring-foreground/10. Se apila border sobre ring-1 + shadow-xs en reposo. La celda Empresa añade un tercer recuadro rounded-md bg-muted para el icono (columns.tsx L114). Fix: Card sin clases extra, quitar shadow-xs, eliminar el recuadro del icono. Comando: /impeccable polish

### [P1] No es responsive; componente móvil terminado sin conectar — queja nº3
page.tsx solo renderiza DataTable. DataCardList (data-card-list.tsx, ~280 líneas) no se importa en ningún sitio. A 390px la tabla mide 964px intrínsecos en contenedor de 356px: Estado, Runtime y Acciones fuera de pantalla. DataCardList viola paleta: bg-blue-600, border-rose-200 text-rose-600, shadow-xs. Chips del toolbar y pills se cortan a 390px con no-scrollbar. Fix: md:hidden DataCardList + hidden md:block DataTable; corregir colores de DataCardList a tokens; quitar no-scrollbar. Comando: /impeccable adapt

### [P2] Deriva de tamaños de control y tipografía
Botones h-8/h-7, pill runtime h-4.5 text-[10px], sub-línea Empresa text-[10px], label kebab text-[11px], TableHead uppercase tracking-wider font-semibold, h1 text-xl font-bold. Tap targets 26-32px en móvil (mín 44). DESIGN.md manda h-10 (h-9 sm), Label 0.75rem/500, nunca 700+, sin uppercase. Fix: controles a h-9 mín, eliminar text-[10px]/[11px], h1 font-semibold, TableHead text-xs font-medium normal-case. Comando: /impeccable layout

## Persona Red Flags
- Alex: refresco manual sin auto-poll, filtro no en URL, orden se resetea, sin acciones masivas, "Desconectar" duplicado inline + kebab.
- Sam: chips sin aria-pressed/role=tab (estado solo por color), text-[10px] falla tamaño/contraste, copiar número opacity-0 sin foco visible, animaciones sin prefers-reduced-motion, kebab sin aria-label, Table sin caption.
- Operador de guardia móvil 2am: no puede actuar a 390px (acciones fuera de tabla, card list muerta); hipo de backend se ve como "flota vacía"; sin auto-refresco; client_outdated per-row Reconectar que falla 405 en silencio, sin rollup ni runbook.

## Minor Observations
- SessionEventsSheet EventTypeIcon: rose-500/amber-500/blue-400/sky-400 (semáforo de 4 colores en componente en uso).
- SessionQRDialog: icono text-blue-500, nota text-amber-600, marco QR bg-white shadow-sm hard-coded.
- Dos formatLocalTime divergentes.
- Columna Empresa: clave de orden (empresa_nombre||account_id) != valor visible ("Empresa sin nombre").
- "ID: empresa_id||telefono_id" etiqueta un id de teléfono como ID de empresa.
- Chips: transition-all y py-0.2 inválido.
- ShieldCheck del título text-emerald-600 decorativo.
- DataCardList pagina a pageSize=10 fijo, aparte del estado de la tabla.
- Consola en vivo: 0 errores, 0 warnings.

## Questions to Consider
1. Cuál es el estado en calma (flota sana -> pantalla monocroma) y por qué no es el default.
2. El operador necesita "Activas: 12" como número o solo "todo en orden" vs "N incidencias".
3. Runtime como columna y como sub-badge: cuando no coinciden, cuál cree el operador y por qué el desacuerdo no es el titular.
4. client_outdated afecta a toda la flota: Reconectar por fila vs acción a nivel de banner con runbook.
5. Roadmap multi-provider (más columnas): la tabla de 6 columnas sobrevive a 7-8 o la lista de tarjetas es el primario honesto.

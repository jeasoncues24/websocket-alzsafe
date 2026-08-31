---
name: WSAPI Admin
description: Panel de operación para una plataforma self-hosted de API de WhatsApp
colors:
  background: "hsl(0 0% 98.82%)"
  foreground: "hsl(0 0% 9.02%)"
  card: "hsl(0 0% 98.82%)"
  card-foreground: "hsl(0 0% 9.02%)"
  popover-foreground: "hsl(0 0% 32.16%)"
  primary: "hsl(151.33 66.86% 66.86%)"
  primary-foreground: "hsl(153.33 13.04% 13.53%)"
  secondary: "hsl(0 0% 99.22%)"
  muted: "hsl(0 0% 92.94%)"
  muted-foreground: "hsl(0 0% 12.55%)"
  accent: "hsl(0 0% 92.94%)"
  destructive: "hsl(9.89 81.98% 43.53%)"
  destructive-foreground: "hsl(0 100% 99.41%)"
  border: "hsl(0 0% 87.45%)"
  input: "hsl(0 0% 96.47%)"
  ring: "hsl(151.33 66.86% 66.86%)"
  status-active: "hsl(160 84% 39%)"
  chart-1: "hsl(151.33 66.86% 66.86%)"
  chart-2: "hsl(217.22 91.22% 59.8%)"
  chart-3: "hsl(258.31 89.53% 66.27%)"
  chart-4: "hsl(37.69 92.13% 50.2%)"
  chart-5: "hsl(160.12 84.08% 39.41%)"
typography:
  display:
    fontFamily: "Plus Jakarta Sans, ui-sans-serif, system-ui, sans-serif"
    fontSize: "1.5rem"
    fontWeight: 600
    lineHeight: 1.2
    letterSpacing: "0.025em"
  headline:
    fontFamily: "Plus Jakarta Sans, ui-sans-serif, system-ui, sans-serif"
    fontSize: "1.25rem"
    fontWeight: 600
    lineHeight: 1.3
    letterSpacing: "0.025em"
  title:
    fontFamily: "Plus Jakarta Sans, ui-sans-serif, system-ui, sans-serif"
    fontSize: "1rem"
    fontWeight: 500
    lineHeight: 1.375
    letterSpacing: "0.025em"
  body:
    fontFamily: "Plus Jakarta Sans, ui-sans-serif, system-ui, sans-serif"
    fontSize: "0.875rem"
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: "0.025em"
  label:
    fontFamily: "Plus Jakarta Sans, ui-sans-serif, system-ui, sans-serif"
    fontSize: "0.75rem"
    fontWeight: 500
    lineHeight: 1.4
    letterSpacing: "0.025em"
  mono:
    fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace"
    fontSize: "0.8125rem"
    fontWeight: 400
    lineHeight: 1.5
    letterSpacing: "normal"
rounded:
  md: "0.375rem"
  base: "0.5rem"
  xl: "0.75rem"
  full: "9999px"
spacing:
  xs: "0.25rem"
  sm: "0.5rem"
  md: "0.75rem"
  lg: "1rem"
  xl: "1.5rem"
components:
  button-primary:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.primary-foreground}"
    rounded: "{rounded.md}"
    padding: "0.5rem 1rem"
    height: "2.5rem"
  button-primary-hover:
    backgroundColor: "hsl(151.33 66.86% 66.86% / 0.9)"
    textColor: "{colors.primary-foreground}"
    rounded: "{rounded.md}"
  button-outline:
    backgroundColor: "{colors.background}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.md}"
    padding: "0.5rem 1rem"
    height: "2.5rem"
  button-destructive:
    backgroundColor: "{colors.destructive}"
    textColor: "{colors.destructive-foreground}"
    rounded: "{rounded.md}"
  badge-default:
    backgroundColor: "{colors.primary}"
    textColor: "{colors.primary-foreground}"
    rounded: "{rounded.full}"
    padding: "0.125rem 0.5rem"
  badge-status-active:
    backgroundColor: "hsl(152 76% 45% / 0.1)"
    textColor: "hsl(158 74% 24%)"
    rounded: "{rounded.full}"
    padding: "0.125rem 0.5rem"
  badge-status-neutral:
    backgroundColor: "hsl(0 0% 92.94% / 0.6)"
    textColor: "{colors.muted-foreground}"
    rounded: "{rounded.full}"
    padding: "0.125rem 0.5rem"
  card:
    backgroundColor: "{colors.card}"
    textColor: "{colors.card-foreground}"
    rounded: "{rounded.xl}"
    padding: "1rem"
  input:
    backgroundColor: "{colors.background}"
    textColor: "{colors.foreground}"
    rounded: "{rounded.md}"
    padding: "0.5rem 0.75rem"
    height: "2.5rem"
---

# Design System: WSAPI Admin

## Overview

**Creative North Star: "La Sala de Control"**

WSAPI Admin es la sala de control de una flota de sesiones de WhatsApp. El operador entra para leer estado y actuar, no para contemplar. El lienzo es casi blanco (modo claro) o casi negro (modo oscuro), las superficies son de un solo tono con un filo de 1px, y el color se comporta como una luz de aviso: aparece solo donde hay significado —una sesión viva, una acción primaria, un error— y en ningún otro sitio. Cuando todo está en orden, la pantalla es tranquila y monocroma.

El sistema es plano pero con contraste: sin sombras en reposo, la profundidad se construye con tono y borde, y la jerarquía se lleva con peso tipográfico (Plus Jakarta Sans, 400 a 600) y con `tracking` ligeramente abierto (0.025em) que le da un aire de instrumento preciso. La densidad es media-alta: filas compactas, `cards` de 1rem de padding, controles de 2.5rem de alto. Cada vista del panel —dashboard, sessions, empresas, roles, broadcasts— comparte exactamente los mismos tokens; no existe una paleta "de dashboard" ni una "de mensajería".

La anti-referencia es explícita: nada de degradados, glassmorphism, blur ni sombras de color; nada de dashboards tipo plantilla de marketing con gráficas sobre-coloreadas; y los estados de sesión no se pintan con un semáforo de cuatro colores —"activa" es verde, todo lo demás es gris con rótulo e icono.

**Key Characteristics:**
- Lienzo neutro casi sin color; el acento es un evento, no un fondo.
- Plano por defecto: profundidad por tono + borde de 1px, sombra solo en hover/focus/overlay.
- Jerarquía por peso tipográfico y `tracking` de 0.025em, no por tamaño extremo.
- Verde = "vivo / OK"; el resto de estados son grises con rótulo e icono.
- Un único set de tokens para todas las vistas; claro/oscuro paritarios.

## Colors

Paleta funcional de grises casi puros con un único acento verde y una reserva de cinco tonos de gráfica; el color comunica estado o acción, nunca decora.

### Primary
- **Verde Señal** (`hsl(151.33 66.86% 66.86%)`, menta claro en modo claro; `hsl(154.9 100% 19.2%)`, verde profundo en modo oscuro): color de marca y de acción. Botón primario, foco (`ring`), enlaces, y `chart-1`. Evoca "conectado" sin llegar al verde saturado de WhatsApp.
- **Verde Tinta** (`hsl(153.33 13.04% 13.53%)`): texto sobre el verde señal (`primary-foreground`). Casi negro con matiz verde para máximo contraste sobre el menta.

### Secondary
El sistema tiene un solo acento. `secondary` es un gris casi idéntico al fondo (`hsl(0 0% 99.22%)`) para botones y chips de bajo énfasis; no es un segundo color de marca.

### Tertiary
- **Rojo Alarma** (`hsl(9.89 81.98% 43.53%)`): exclusivamente destructivo —eliminar, desconectar, error de formulario (`aria-invalid`)—. Nunca decorativo, nunca para "malo" en un gráfico.

### Neutral
- **Lienzo** (`hsl(0 0% 98.82%)` claro / `hsl(0 0% 7.06%)` oscuro): fondo de página y de `card`.
- **Tinta** (`hsl(0 0% 9.02%)` claro / `hsl(214 32% 91%)` oscuro): texto principal (`foreground`).
- **Tinta Tenue** (`hsl(0 0% 12.55%)` claro / `hsl(0 0% 63.5%)` oscuro): texto secundario, descripciones, iconos inactivos (`muted-foreground`).
- **Superficie Hundida** (`hsl(0 0% 92.94%)`): `muted` / `accent`, fondo de estados hover fantasma, pies de `card`, badges neutros.
- **Filo** (`hsl(0 0% 87.45%)`): todos los bordes y divisores (`border`); en `card` se aplica como `ring` de 1px a `foreground/10`.
- **Campo** (`hsl(0 0% 96.47%)`): relleno interno de inputs en reposo (`input`).

### Chart
- **chart-1..5** (`hsl(151 67% 67%)` verde, `hsl(217 91% 60%)` azul, `hsl(258 90% 66%)` violeta, `hsl(38 92% 50%)` ámbar, `hsl(160 84% 39%)` teal): reservados para series de datos en Recharts, en este orden. Fuera de un gráfico no se usan.

### Named Rules
**The Warning-Light Rule.** El verde señal ocupa ≤10% de píxeles de cualquier pantalla. Si aparece como fondo de una región grande, algo está mal.

**The No-Rainbow Rule.** Ninguna vista define color propio (gradiente, fondo tintado, ilustración de color). Todo color sale de un token de `globals.css`. Si un diseño necesita un color que no existe como token, primero se añade el token.

**The Grey-State Rule.** Los estados de sesión (`qr_pending`, `initializing`, `disconnected`, `client_outdated`, inconsistencia) se representan en grises + rótulo + icono. Solo `active` / `online` usan verde (familia `emerald` de Tailwind: `emerald-500` para el punto, `emerald-700`/`emerald-300` para el texto). Nunca amarillo/naranja/rojo/azul para diferenciar estados entre sí.

## Typography

**Display / Body / Label Font:** Plus Jakarta Sans (fuente variable local, peso 200–800; fallback `ui-sans-serif, system-ui, sans-serif`)
**Mono Font:** pila del sistema (`ui-monospace, SFMono-Regular, Menlo, monospace`) para IDs de sesión, tokens de API, JSON y valores técnicos.

**Character:** una sola familia geométrica-humanista para todo. La personalidad la da el `letter-spacing` global de 0.025em y el rango de pesos estrecho (400 texto, 500 rótulos y títulos, 600 encabezados); nunca 700+. Da sensación de instrumento legible más que de producto de consumo.

### Hierarchy
- **Display** (600, 1.5rem, line-height 1.2): título de página / encabezado de sección principal. Uno por vista.
- **Headline** (600, 1.25rem, line-height 1.3): subsecciones dentro de una página.
- **Title** (500, 1rem, line-height 1.375): título de `card` y de diálogo (`CardTitle`).
- **Body** (400, 0.875rem, line-height 1.5): texto de trabajo por defecto —tablas, formularios, párrafos—. Es el tamaño base real del panel.
- **Label** (500, 0.75rem, line-height 1.4): badges, metadatos, encabezados de tabla, texto de ayuda. `text-[11px]` para sub-badges apilados.
- **Mono** (400, 0.8125rem): valores técnicos copiables; nunca para prosa.

### Named Rules
**The Weight-Not-Size Rule.** La jerarquía sube por peso (400 → 500 → 600) antes que por tamaño. Un título de `card` es 1rem/500, no 1.25rem.

**The Spanish-First Rule.** Todo texto de interfaz —rótulos, vacíos, errores— va en español, en tono operativo y directo.

## Layout

Shell fijo de tres zonas: `sidebar` de navegación a la izquierda (colapsable), `top-bar` superior con contexto y acciones globales, y área de contenido con scroll propio. En móvil el `sidebar` se sustituye por `mobile-nav` (drawer). El contenido se limita a un ancho de lectura cómodo y se alinea a una rejilla de 4px (`--spacing: 0.25rem`); los pasos usados en la práctica son 0.25 / 0.5 / 0.75 / 1 / 1.5 / 2 / 3 rem.

Densidad media-alta: controles de 2.5rem de alto (`h-10`), `card` con 1rem de padding (0.75rem en `size="sm"`), gap vertical de 1rem entre bloques. Las tablas (`@tanstack/react-table`) son el patrón dominante de datos: filas compactas, encabezado en Label, sin cebra —los divisores de 1px bastan—. Breakpoints estándar de Tailwind (`sm` 640, `md` 768, `lg` 1024, `xl` 1280); el salto real es `md` (aparece el shell de escritorio).

## Elevation & Depth

Sistema **plano por defecto**. En reposo no hay sombras: la separación entre superficie y fondo se consigue con un cambio de tono mínimo y un borde/`ring` de 1px (`ring-1 ring-foreground/10` en `card`). La sombra es una **respuesta a un estado**, no una propiedad de la superficie.

### Shadow Vocabulary
- **Hover de fila/tarjeta interactiva** (`--shadow-md`: `0 1px 3px 0 hsl(0 0% 0% / 0.17), 0 2px 4px -1px hsl(0 0% 0% / 0.17)`): solo con la utilidad `.motion-lift` al pasar el cursor sobre algo accionable, junto a `translateY(-2px)`.
- **Overlays** (`--shadow-lg` / `--shadow-xl`): diálogos (`dialog`), `sheet`, `dropdown-menu`, `popover`. Es el único sitio donde la sombra existe en reposo, porque el elemento flota sobre el resto.
- **`--shadow-2xs` … `--shadow-sm`**: disponibles pero de uso excepcional; preferir borde antes que sombra pequeña.

### Named Rules
**The Flat-At-Rest Rule.** Si un elemento tiene sombra sin estar en hover, en foco o flotando como overlay, está mal. Los degradados, el `backdrop-filter` y las sombras de color están prohibidos en todo el sistema.

## Shapes

Lenguaje de esquinas suaves y consistentes construido sobre `--radius: 0.5rem`:
- **Controles** (botón, input, select, textarea): `rounded-md` (0.375rem). Filo firme, no pastilla.
- **Contenedores** (`card`, diálogo, `sheet`): `rounded-xl` (0.75rem).
- **Indicadores** (badge, chip, punto de estado, avatar): `rounded-full`.

Bordes siempre de 1px en color `border` (o `ring` a `foreground/10` en `card`); nunca 2px ni bordes de color salvo el rojo `destructive` en campos inválidos y el borde tintado de emerald en el badge "activa". Sin clipping decorativo, sin formas diagonales, sin recortes.

## Components

### Buttons
Carácter: sólidos y sobrios, sin adorno; el primario es la única mancha de verde estable de la pantalla.
- **Shape:** `rounded-md` (0.375rem).
- **Primary:** fondo `primary`, texto `primary-foreground`, `h-10 px-4 py-2` (`sm` = `h-9 px-3`, `lg` = `h-11 px-8`). Hover: `bg-primary/90`. Es el único botón de acción principal por vista.
- **Outline:** borde `input`, fondo `background`, texto `foreground`; hover `bg-accent`. Acción secundaria.
- **Secondary / Ghost:** grises casi planos (`bg-secondary` / solo hover `bg-accent`); acciones terciarias y de barra de herramientas.
- **Destructive:** fondo `destructive`, texto `destructive-foreground`; solo para eliminar/desconectar, con confirmación.
- **Focus:** `ring-2 ring-ring ring-offset-2`. **Press:** `scale(0.98)` vía `.motion-press`.

### Chips / Badges
- **Default:** `bg-primary` / `primary-foreground`, `rounded-full`, `text-xs font-semibold`. Uso escaso (conteos, etiqueta destacada).
- **Status "activa"** (firma del sistema): `variant="outline"` con `border-emerald-600/20 bg-emerald-500/10 text-emerald-700` (oscuro: `border-emerald-500/20 bg-emerald-950/40 text-emerald-300`), punto `bg-emerald-500` de 1.5×1.5 con `animate-pulse`, rótulo "Activa".
- **Status neutro** (`qr_pending`, `initializing`, `disconnected`, `client_outdated`, `Inconsistente`): `border-border bg-muted/30–60 text-foreground|muted-foreground`, con icono de Lucide (`Loader2 animate-spin`, `AlertTriangle`) cuando aplica. Sin color.

### Cards / Containers
- **Corner:** `rounded-xl` (0.75rem).
- **Background:** `card` (idéntico al lienzo en claro; un escalón más claro que el fondo en oscuro).
- **Separación:** `ring-1 ring-foreground/10`, sin sombra en reposo.
- **Footer:** `border-t bg-muted/50`, mismo padding que el cuerpo.
- **Padding:** 1rem (`size="sm"` → 0.75rem); gap interno 1rem / 0.75rem.

### Inputs / Fields
- **Style:** `h-10`, `rounded-md`, borde `input`, fondo `background`, texto 0.875rem, placeholder `muted-foreground`.
- **Focus:** `ring-2 ring-ring ring-offset-2`, sin cambio de fondo.
- **Error:** `aria-invalid` → `border-destructive` + `ring-destructive/20`.
- **Disabled:** `opacity-50`, `cursor-not-allowed`.

### Navigation
- **Sidebar:** ítems en Label/Body 0.875rem, icono Lucide de 16–18px + texto. Reposo `sidebar-foreground` (gris); hover `sidebar-accent`; activo `sidebar-primary` con fondo `sidebar-accent`. Colapsable a solo-icono.
- **Top-bar:** contexto de la vista a la izquierda, acciones globales y `theme` a la derecha; borde inferior de 1px, sin sombra.
- **Mobile:** `mobile-nav` como drawer (`sheet`) por debajo de `md`.

### Signature Component — SessionStatusBadge
Componente que traduce el ciclo de vida de whatsmeow a una señal de un vistazo. Compone: badge de estado base (verde solo si `active`, gris en el resto) + badge opcional "Reconectando" (`Loader2` + texto esmeralda tenue) + badge "Inconsistente" (`AlertTriangle` gris, con `title` explicativo) + indicador runtime "Online/Offline" (punto + texto, verde solo si conectado). Todo el color del componente es la familia `emerald`; ningún otro estado introduce color.

## Do's and Don'ts

### Do:
- **Do** resolver todo componente de UI con la skill `shadcn` (proyecto shadcn/ui: `frontend/components.json`, estilo `base-nova`, iconos `lucide`, alias `@/components/ui`); vestir sobre `frontend/components/ui/*` con `cn()`, nunca copiar primitivas a mano. Usar `migrate-radix-to-base` solo ante una migración Radix→Base pedida de forma explícita.
- **Do** sacar todo color de un token de `frontend/app/globals.css`; si falta el color, añadir el token antes de usarlo (The No-Rainbow Rule).
- **Do** mantener el verde señal en ≤10% de la pantalla y reservado a acción primaria, foco, enlace y `chart-1` (The Warning-Light Rule).
- **Do** representar los estados de sesión en gris + rótulo + icono, con verde solo para `active`/`online` (The Grey-State Rule).
- **Do** construir profundidad con tono + borde de 1px; introducir sombra únicamente en hover (`.motion-lift`), foco u overlays (The Flat-At-Rest Rule).
- **Do** subir la jerarquía por peso (400 → 500 → 600) antes que por tamaño, con `letter-spacing` 0.025em (The Weight-Not-Size Rule).
- **Do** usar las utilidades `.motion-*` y los tokens `--motion-*` para cualquier transición; nada de duraciones o `easing` sueltos.
- **Do** escribir todo texto de interfaz en español, tono operativo.
- **Do** usar la pila mono solo para IDs, tokens y valores técnicos copiables.

### Don't:
- **Don't** usar degradados, `backdrop-filter`, glassmorphism ni sombras de color en ningún sitio.
- **Don't** dar a una vista su propia paleta, fondo tintado o ilustración de color.
- **Don't** pintar los estados con un semáforo (amarillo/naranja/rojo/azul) para diferenciarlos entre sí.
- **Don't** usar los colores `chart-*` fuera de un gráfico de Recharts.
- **Don't** usar peso 700+ ni tamaños de display por encima de ~1.5rem en el panel.
- **Don't** poner sombra en `card`, fila o panel en reposo; ni bordes de 2px o de color (salvo `destructive` en campos inválidos y el filo emerald del badge "activa").
- **Don't** reintroducir la fuente Inter ni añadir una segunda familia de texto: Plus Jakarta Sans es la única.

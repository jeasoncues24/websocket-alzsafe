# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Stack

Existente (no delegado): panel Next.js 16 (App Router) + React 19, Tailwind CSS v4 con tokens en `frontend/app/globals.css`, primitivas shadcn/Radix en `frontend/components/ui`, `next-themes` para claro/oscuro, Zustand para estado, Recharts para gráficas. Backend Go independiente (`backend/`, imports `wsapi/internal/...`). El diseño trabaja sobre este stack; no se elige uno nuevo.

## Users

Usuario principal: **operadores internos / soporte** de la organización que opera wsapi. Trabajan desde un único panel administrativo para dar de alta y vigilar sesiones de WhatsApp, resolver incidencias de conexión (emparejamiento por QR, reconexión, librería caducada), enviar y revisar mensajería y difusiones, y administrar el acceso multi-tenant (empresas, roles, módulos, usuarios). Es una herramienta de trabajo de uso frecuente, no un producto de marketing.

## Product Purpose

wsapi es una plataforma **self-hosted** de API de WhatsApp. El panel existe para que los operadores controlen todo el ciclo de vida de las sesiones y la mensajería sin tocar la base de datos ni el servidor: ver el estado real de cada sesión, emparejar dispositivos, diagnosticar desconexiones y caducidades, y gestionar quién puede hacer qué. Éxito = un operador entiende de un vistazo el estado de la flota de sesiones y puede actuar sobre cualquier incidencia desde la interfaz.

## Positioning

Hoy el núcleo es **whatsmeow** (conexión no oficial a WhatsApp: sin costo por conversación ni aprobación de plantillas). La dirección del proyecto es convertirse en una plataforma **multi-provider**: a futuro habrá varios proveedores, incluida la WhatsApp Business API oficial. La UI no debe asumir que whatsmeow es el único origen posible de una sesión o un mensaje. Se despliega on-premise (Docker solo compila el binario → `dist/wsapi`; PM2 en el host lo ejecuta; MySQL/MariaDB vive fuera de Docker Compose), de modo que el cliente controla su infraestructura y sus datos.

## Operating Context

- **Despliegue:** `docker compose` compila el binario, `make build` / `make start` lo registran en PM2 sin duplicarlo; `backend/.env` (copiado de `backend/.env.copy`) es obligatorio; `APP_PORT` debe existir o el backend falla rápido con mensaje claro. La base de datos no se crea desde el compose actual.
- **Migraciones:** motor `goforge`; `migrate adopt` se corre una sola vez en producción.
- **Ciclo de vida de sesión (whatsmeow):** estados `active`, `qr_pending`, `initializing`, `disconnected`, `client_outdated`; además señales transversales de `reconnecting` y de inconsistencia entre el estado en BD y la conexión en memoria (`runtime mismatch`).
- **Caducidad periódica:** whatsmeow caduca cada ~2-3 meses (WhatsApp responde 405); existe el estado `client_outdated` para representarlo y debe ser legible sin ambigüedad.
- **Superficies del panel:** login, dashboard, sessions, qr, messages, broadcasts, empresas, roles, modules, users_admin, usuario_admin, settings.

## Capabilities and Constraints

- **Multi-tenant con RBAC:** empresas, roles, módulos y usuarios administrables; una instancia sirve a varias organizaciones con alcance por rol.
- **Backend Go:** separación dominio (`internal/domain`) / persistencia (`internal/storage`) / HTTP (`internal/http`); campos sensibles (p. ej. secretos de webhook) etiquetados `json:"-"` y nunca expuestos en la UI.
- **Portabilidad de BD:** MySQL/MariaDB en producción y SQLite en memoria en pruebas unitarias. Evitar `FOR UPDATE`; usar `UPDATE` universales con `CASE WHEN` para atomicidad portable.
- **Capa de providers:** se espera que crezca más allá de whatsmeow; modelar la UI de sesiones/mensajes sin acoplarla a un proveedor único.
- **Idioma:** todo texto orientado al usuario final (rótulos, mensajes de error, vacíos) va en **Español**.
- Sin datos inventados: no hay pricing, planes, SLAs ni métricas de negocio definidas.

## Brand Commitments

- **Nombre:** WSAPI; título del panel "WhatsApp API Admin".
- **Voz:** Español, tono operativo y directo (herramienta de trabajo, no folleto).
- **Identidad visual fijada como canónica** (decisión del usuario, para evitar "vistas de arcoiris"): el sistema actual es la base y todo lo nuevo se ciñe a él —
  - Primario verde y set de tokens shadcn en `frontend/app/globals.css` (`:root` y `.dark`), incluidos `--chart-1..5` y tokens de `--sidebar-*`.
  - Tipografía **Plus Jakarta Sans** (fuente variable local, `frontend/public/fonts/PlusJakartaSans-Variable.woff2`, peso 200–800), `--font-mono: monospace`.
  - Claro/oscuro vía `next-themes` (`attribute="class"`, `defaultTheme="system"`).
  - `--radius: 0.5rem`, escala de sombras `--shadow-*`, `--tracking-normal: 0.025em`.
  - Tokens de movimiento ya definidos (`--motion-duration-*`, `--motion-ease-*`, utilidades `.motion-*`); la animación nueva usa estos, no valores sueltos.
  - Color = significado (estado de sesión, acción primaria), no decoración por vista.

## Evidence on Hand

- `README.md`: contrato de despliegue del último epic (Docker build + PM2 + MySQL externo).
- `CLAUDE.md`: convenciones de backend, SQL portable y campos sensibles.
- Implementación frontend vigente: sistema de tokens en `frontend/app/globals.css`, primitivas en `frontend/components/ui`, shell en `frontend/components/layout` (`sidebar`, `top-bar`, `mobile-nav`), patrón de estado en `frontend/components/session/session-status-badge.tsx` (neutro + un solo acento esmeralda para "activa").
- No existen testimonios, logos de clientes, benchmarks, precios ni casos de estudio: no se deben fabricar.

## Product Principles

1. **Operar antes que expresar.** El panel es una herramienta de uso diario: escaneabilidad, consistencia y estado inequívoco pesan más que la estética.
2. **Un solo sistema visual.** Tokens compartidos, sin paletas por vista. El color se reserva para significado (estado, acción primaria, alerta).
3. **La sesión es el dato central.** Los estados de whatsmeow —incluida la caducidad `client_outdated` y las inconsistencias runtime/BD— deben leerse de un vistazo y guiar la acción.
4. **Neutral respecto al proveedor.** La UI de sesiones y mensajería no asume que whatsmeow es el único origen; deja espacio para múltiples providers.
5. **Respetar las restricciones del backend.** Portabilidad MySQL/SQLite y protección de campos sensibles son límites duros; la interfaz nunca los expone ni los contradice.

## Accessibility & Inclusion

Soporte claro/oscuro es un compromiso, no un extra: ambos temas deben cumplir contraste WCAG AA, en especial los colores de estado de sesión y los badges, que no pueden depender solo del color para comunicar (llevan rótulo y/o icono).

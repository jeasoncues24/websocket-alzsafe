---
stepsCompleted:
  - step-01-validate-prerequisites
  - step-02-design-epics
  - step-03-create-stories
inputDocuments:
  - "_bmad-output/project-context.md"
  - "docs/bmad-project-rules.md"
  - "_bmad-output/planning-artifacts/prd.md"
  - "frontend/lib/api.ts"
  - "frontend/next.config.ts"
  - "frontend/stores/useAppStore.ts"
  - "frontend/stores/useThemeStore.ts"
  - "react-doctor-report-2026-05-11"
---

# wsapi — Epic 3: Hardening de Seguridad y Calidad Frontend

## Overview

Este documento define el Epic 3 del proyecto wsapi, orientado a eliminar vulnerabilidades de seguridad confirmadas en el frontend (JWT en localStorage, token en query param de WebSocket, ausencia de headers HTTP defensivos) y corregir los 312 issues detectados por React Doctor (score 78/100), organizados por impacto sin alterar ninguna lógica de negocio existente.

## Requirements Inventory

### Functional Requirements

```
FR1: El token JWT debe almacenarse en cookie httpOnly en lugar de localStorage — elimina vector XSS directo confirmado en frontend/lib/api.ts.
FR2: El WebSocket admin no debe pasar el JWT como query param (?token=); la autenticación debe migrar a cookie httpOnly (automático tras FR1).
FR3: next.config.ts debe incluir headers HTTP de seguridad: CSP, X-Frame-Options, X-Content-Type-Options, Referrer-Policy, Permissions-Policy.
FR4: generateQRLink() debe usar authHeaders() centralizado en vez de llamar localStorage directamente (frontend/lib/api.ts).
FR5: MetricCard debe definirse fuera de DashboardPage para evitar recreación de instancia en cada render (app/dashboard/page.tsx:58).
FR6: new Date() en JSX debe envolverse en useEffect+useState para eliminar hydration mismatches server/client (empresa-detail-modal.tsx:207 ×4).
FR7: Keys de listas React deben usar identificadores estables, no índices de array (api-key-metrics.tsx:78 ×6).
FR8: CompanyPhonesPage debe refactorizarse de 7 useState a useReducer para estado relacionado (telefonos/page.tsx:21).
FR9: Los 8 setState en un useEffect deben consolidarse eliminando cascading setState (telefonos/page.tsx:34).
FR10: El effect chain en api-key-metrics.tsx:245 debe corregirse para evitar renders encadenados.
FR11: 26 labels sin htmlFor deben asociarse correctamente a sus inputs (components/ui/label.tsx:7).
FR12: Click event sin key event debe agregar handler de teclado (api-keys/page.tsx:354).
FR13: Heading vacío en components/ui/alert.tsx:28 debe tener contenido accesible para screen readers.
FR14: Dead code eliminado: 26 exports, 10 types, 2 archivos no usados identificados por React Doctor.
FR15: 132 casos w-N h-N → size-N (Tailwind v3.4+ shorthand).
FR16: 12 headings font-bold → font-semibold (mejor legibilidad tipográfica).
FR17: 19 casos useRouter() sin destructuring → const { push } = useRouter() para facilitar memoización del React Compiler.
```

### NonFunctional Requirements

```
NFR1: Cero cambios a lógica de negocio existente (WhatsApp, empresas, teléfonos, API keys, difusiones).
NFR2: Cero cambios a endpoints ni payloads del backend Go — solo ajuste mínimo en Set-Cookie y lectura de cookie en middleware.
NFR3: Completable en 1-2 semanas por 1 desarrollador.
NFR4: Cada story pasa `cd frontend && npm run lint` y `cd frontend && npm run build` sin errores antes de marcarla done.
NFR5: Migración JWT → cookie httpOnly requiere mínimo cambio en backend Go: Set-Cookie en login + lectura de cookie como fallback en middleware auth.
NFR6: Cambios Tailwind (size-N, font-semibold) pueden automatizarse con script/codemod para minimizar errores manuales en 132+ archivos.
NFR7: Score React Doctor debe mejorar de 78/100 a >90/100 al finalizar el epic.
NFR8: La protección de rutas client-side existente (middleware Next.js) NO cambia en este epic.
```

### Additional Requirements

```
- Backend Go: handler /api/admin/login debe emitir Set-Cookie: admin_token=<jwt>; HttpOnly; Secure; SameSite=Strict; Path=/
- Backend Go: middleware de auth debe leer token desde cookie "admin_token" cuando el header Authorization está ausente.
- WebSocket /ws: tras migrar a cookie httpOnly, el token se enviará automáticamente sin cambios adicionales al cliente.
- Stores Zustand (useAppStore, useThemeStore): seguros — no almacenan JWT ni QR. Sin cambios requeridos.
- .env.example: limpio, solo NEXT_PUBLIC_API_URL. Sin cambios requeridos.
- Skills disponibles para dev: better-auth-best-practices (auth/JWT), golang-security (backend Go), security-review (auditoría), bmad-code-review (review de cada story).
```

### UX Design Requirements

```
UX-DR1: Los hydration mismatches en timestamps y fechas crean parpadeos visibles que rompen la percepción de estabilidad del sistema — corregir como parte de FR6.
UX-DR2: Labels sin htmlFor impiden a usuarios de teclado y lectores de pantalla operar formularios correctamente — corregir como parte de FR11.
UX-DR3: Click events sin key events bloquean a usuarios de teclado en la página de API keys — corregir como parte de FR12.
```

### FR Coverage Map

```
FR1:  Epic 3, Story 3.1 — JWT localStorage → httpOnly cookie
FR2:  Epic 3, Story 3.2 — WebSocket auth segura
FR3:  Epic 3, Story 3.3 — Security headers next.config.ts
FR4:  Epic 3, Story 3.1 — JWT localStorage → httpOnly cookie (generateQRLink incluido)
FR5:  Epic 3, Story 3.4 — Fix correctness MetricCard / hydration / keys
FR6:  Epic 3, Story 3.4 — Fix correctness MetricCard / hydration / keys
FR7:  Epic 3, Story 3.4 — Fix correctness MetricCard / hydration / keys
FR8:  Epic 3, Story 3.5 — Refactor estado useReducer + cascading setState
FR9:  Epic 3, Story 3.5 — Refactor estado useReducer + cascading setState
FR10: Epic 3, Story 3.5 — Refactor estado useReducer + cascading setState
FR11: Epic 3, Story 3.6 — Fix accesibilidad labels / key events / headings
FR12: Epic 3, Story 3.6 — Fix accesibilidad labels / key events / headings
FR13: Epic 3, Story 3.6 — Fix accesibilidad labels / key events / headings
FR14: Epic 3, Story 3.7 — Limpieza dead code
FR15: Epic 3, Story 3.8 — Tailwind shortcuts + font + useRouter
FR16: Epic 3, Story 3.8 — Tailwind shortcuts + font + useRouter
FR17: Epic 3, Story 3.8 — Tailwind shortcuts + font + useRouter
```

## Epic List

### Epic 3: Hardening de Seguridad y Calidad Frontend

El equipo puede operar wsapi con la confianza de que el frontend protege los tokens de sesión admin contra XSS, usa comunicaciones WebSocket sin exponer credenciales en URLs, emite headers HTTP defensivos en toda la aplicación, y el código base está libre de bugs de correctness, estado inconsistente y deuda técnica acumulada — sin alterar ninguna lógica de negocio existente.

**FRs cubiertos:** FR1–FR17
**NFRs aplicables:** NFR1–NFR8
**Duración estimada:** 10 días hábiles (1 dev)
**Skills recomendadas por story:** ver cada story file

---

## Epic 3: Hardening de Seguridad y Calidad Frontend

### Story 3.1: Migrar JWT de localStorage a httpOnly Cookie

Como administrador del sistema,
quiero que mi sesión admin esté protegida por una cookie httpOnly en lugar de localStorage,
para que un eventual script XSS en el panel no pueda robar el token de sesión.

**Acceptance Criteria:**

**Dado** que el backend Go maneja el login de admin en `/api/admin/login`
**Cuando** el usuario envía credenciales correctas
**Entonces** la respuesta incluye `Set-Cookie: admin_token=<jwt>; HttpOnly; Secure; SameSite=Strict; Path=/`
**Y** el cuerpo JSON de respuesta puede omitir el token o mantenerlo para compatibilidad transitoria

**Dado** que el middleware de auth de Go procesa un request autenticado
**Cuando** el header `Authorization: Bearer <token>` está ausente
**Entonces** el middleware busca el token en la cookie `admin_token` y lo valida
**Y** si la cookie tampoco existe, retorna 401

**Dado** que `frontend/lib/api.ts` contiene `authHeaders()` y `fetchWithAuth()`
**Cuando** se elimina `localStorage.getItem("admin_token")` de ambas funciones
**Entonces** las llamadas a la API envían la cookie automáticamente via `credentials: 'include'` en fetch
**Y** `generateQRLink()` deja de llamar `localStorage.getItem("admin_token")` directamente

**Dado** que el flujo de login en `app/login/page.tsx` guardaba el token en localStorage
**Cuando** se completa el login exitosamente
**Entonces** el frontend NO almacena el token en localStorage
**Y** la sesión persiste correctamente a través de recargas de página vía cookie

**Verificación:** `cd frontend && npm run lint && npm run build` sin errores. `cd backend && go build ./...` sin errores.

---

### Story 3.2: Asegurar WebSocket — Eliminar Token de Query Param

Como administrador del sistema,
quiero que la conexión WebSocket admin no transmita el JWT en la URL,
para que el token no aparezca en logs del servidor ni en el historial del navegador.

**Acceptance Criteria:**

**Dado** que `buildAdminWsUrl()` en `frontend/lib/api.ts` actualmente añade `?token=<jwt>` a la URL del WebSocket
**Cuando** se completa Story 3.1 (cookie httpOnly activa)
**Entonces** `buildAdminWsUrl()` elimina el parámetro `token` de la query string
**Y** la función ya no acepta ni usa el parámetro `token?: string`

**Dado** que el backend Go valida la conexión WebSocket en `/ws`
**Cuando** llega un WebSocket upgrade request
**Entonces** el middleware lee el token desde la cookie `admin_token` (igual que requests HTTP)
**Y** si el token es inválido o ausente, rechaza el upgrade con 401

**Dado** que los componentes que usan `buildAdminWsUrl` (dashboard, sesiones) lo llaman con token
**Cuando** se actualiza la firma de `buildAdminWsUrl` para eliminar el parámetro token
**Entonces** todos los call sites se actualizan sin pasar token
**Y** la conexión WebSocket funciona correctamente con autenticación via cookie

**Verificación:** Revisar logs del servidor — el JWT no debe aparecer en ninguna línea de log de conexión WS. `npm run build` sin errores.

---

### Story 3.3: Security Headers HTTP en next.config.ts

Como operador de la plataforma,
quiero que el panel admin emita headers HTTP de seguridad en todas las respuestas,
para proteger contra clickjacking, XSS, sniffing de contenido e información de referrer.

**Acceptance Criteria:**

**Dado** que `frontend/next.config.ts` actualmente no define headers HTTP de seguridad
**Cuando** se agrega la función `async headers()` al config de Next.js
**Entonces** todas las rutas (`source: '/(.*)'`) incluyen los siguientes headers:
- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `Referrer-Policy: strict-origin-when-cross-origin`
- `Permissions-Policy: camera=(), microphone=(), geolocation=()`
- `Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; connect-src 'self' ws: wss:; font-src 'self'; frame-ancestors 'none'`

**Dado** que la CSP debe permitir la conexión WebSocket a la misma origin
**Cuando** el browser abre el dashboard con WebSocket
**Entonces** la consola del browser no muestra errores de CSP bloqueando el WS
**Y** las imágenes y assets del panel cargan correctamente sin violaciones CSP

**Dado** que `npm run build` produce el bundle final
**Cuando** se verifica la respuesta HTTP del servidor Next.js
**Entonces** todos los headers de seguridad están presentes en la respuesta

**Verificación:** `npm run build` sin errores. Verificar headers con `curl -I http://localhost:3001` o DevTools Network.

---

### Story 3.4: Fix Correctness — MetricCard Anidado, Hydration Mismatches y Array Keys

Como desarrollador del equipo,
quiero que el código del dashboard y modales esté libre de bugs de correctness detectados por React Doctor,
para que los componentes mantengan estado estable y no haya diferencias entre renders server y client.

**Acceptance Criteria:**

**Dado** que `MetricCard` está definido dentro de `DashboardPage` en `app/dashboard/page.tsx:58`
**Cuando** se mueve la definición de `MetricCard` fuera del cuerpo de `DashboardPage` (a nivel de módulo o archivo separado)
**Entonces** React no crea una nueva instancia del componente en cada render del padre
**Y** el estado interno de `MetricCard` persiste correctamente entre re-renders de `DashboardPage`

**Dado** que `new Date()` se usa directamente en JSX en `components/companies/empresa-detail-modal.tsx:207` (×4 casos)
**Cuando** se envuelven esas expresiones en `useEffect + useState` inicializado con `null` o se añade `suppressHydrationWarning` al elemento padre
**Entonces** el servidor y el cliente renderizan el mismo contenido inicial sin mismatch
**Y** los timestamps se muestran correctamente sin parpadeos visibles al hidratar

**Dado** que `components/api-key-metrics.tsx:78` usa índices de array como key (×6 casos)
**Cuando** se reemplazan los índices por identificadores estables (`item.id`, `item.day`, `item.bucket`, etc.)
**Entonces** React puede reconciliar listas correctamente cuando se reordenan o filtran
**Y** React Doctor no reporta más "Array index as key" en ese archivo

**Verificación:** `npm run lint && npm run build` sin errores. `npx react-doctor@latest . --verbose` no reporta issues en los archivos modificados.

---

### Story 3.5: Refactor Estado — useReducer y Eliminar Cascading setState

Como desarrollador del equipo,
quiero que `CompanyPhonesPage` y `api-key-metrics` gestionen su estado de forma consolidada,
para que los renders sean predecibles, los effects no se encadenen y el código sea mantenible.

**Acceptance Criteria:**

**Dado** que `CompanyPhonesPage` en `app/empresas/[empresaId]/telefonos/page.tsx:21` tiene 7 `useState` relacionados
**Cuando** se consolidan en un `useReducer` con estado tipado `PhonesState` y acciones `PhonesAction`
**Entonces** todos los 7 campos de estado se gestionan en un solo `dispatch`
**Y** el comportamiento visible del componente es idéntico al anterior
**Y** no se introducen regresiones en la carga de teléfonos, QR, o reconexión

**Dado** que `app/empresas/[empresaId]/telefonos/page.tsx:34` tiene 8 `setState` dentro de un único `useEffect`
**Cuando** se consolida el cascading setState usando `dispatch` del reducer (o `setState` con objeto completo si aplica)
**Entonces** el componente realiza una sola actualización de estado por ciclo de effect
**Y** no hay renders intermedios con estado parcialmente actualizado

**Dado** que `components/api-key-metrics.tsx:245` tiene un effect chain (useEffect que reacciona a state seteado por otro useEffect)
**Cuando** se elimina la cadena moviendo el cómputo derivado al render o consolidando en un solo effect
**Entonces** no hay renders extra por cada link de la cadena
**Y** los datos de métricas se calculan en un solo ciclo

**Verificación:** `npm run lint && npm run build` sin errores. React Doctor no reporta UseReducer ni cascading setState en archivos modificados.

---

### Story 3.6: Fix Accesibilidad — Labels, Key Events y Headings

Como usuario del panel admin que depende de teclado o lector de pantalla,
quiero que los formularios y controles interactivos sean completamente accesibles por teclado,
para poder operar la plataforma sin necesidad de mouse.

**Acceptance Criteria:**

**Dado** que `components/ui/label.tsx:7` define el componente `Label` base sin `htmlFor` asociado
**Cuando** se actualiza el componente para propagar correctamente `htmlFor` a su elemento `<label>` subyacente
**Entonces** los 26 usos de `Label` en el proyecto pasan la regla `jsx-a11y/label-has-associated-control`
**Y** un lector de pantalla puede anunciar el label correcto al hacer foco en el input asociado

**Dado** que `app/empresas/[empresaId]/telefonos/[telefonoId]/api-keys/page.tsx:354` tiene un elemento clickeable sin key events
**Cuando** se agrega el handler `onKeyDown` (o `onKeyPress`) que responde a Enter y Space con la misma acción que `onClick`
**Entonces** el elemento es operable completamente por teclado
**Y** la regla `jsx-a11y/click-events-have-key-events` pasa sin warnings

**Dado** que `components/ui/alert.tsx:28` tiene un heading (`<h*>`) sin contenido textual accesible
**Cuando** se agrega contenido al heading o se refactoriza para no usar heading vacío
**Entonces** la regla `jsx-a11y/heading-has-content` pasa sin warnings
**Y** la estructura semántica del componente `Alert` es correcta

**Verificación:** `npm run lint` sin errores de jsx-a11y. `npm run build` sin errores.

---

### Story 3.7: Limpieza de Dead Code

Como desarrollador del equipo,
quiero que el proyecto esté libre de exports, types y archivos sin usar detectados por React Doctor,
para mantener el bundle limpio y reducir superficie de confusión en el código base.

**Acceptance Criteria:**

**Dado** que React Doctor detectó 26 exports sin usar en el proyecto
**Cuando** se audita cada export con `--verbose` y se elimina o convierte a export interno si no es necesario externamente
**Entonces** ningún módulo exporta símbolos que no se usan en ningún import del proyecto
**Y** el bundle final del build no incluye código muerto

**Dado** que React Doctor detectó 10 types sin usar en el proyecto
**Cuando** se eliminan los TypeScript types/interfaces que no tienen ninguna referencia
**Entonces** TypeScript compila sin errores y los types eliminados no generan regresiones
**Y** `npm run build` (que incluye type-check) pasa sin errores

**Dado** que React Doctor detectó 2 archivos sin usar (no importados por ningún otro archivo)
**Cuando** se verifica que efectivamente no son necesarios (no son entry points, no son archivos de config implícitos)
**Entonces** los archivos se eliminan del proyecto
**Y** el build y lint pasan sin referencias rotas

**Verificación:** `npm run lint && npm run build` sin errores. `npx react-doctor@latest . --verbose` no reporta Dead Code en archivos modificados.

---

### Story 3.8: Tailwind Shortcuts, Font Headings y useRouter Destructuring

Como desarrollador del equipo,
quiero que el código Tailwind y los patrones de hooks sigan las convenciones modernas recomendadas,
para que el código sea consistente, más legible y compatible con el React Compiler.

**Acceptance Criteria:**

**Dado** que React Doctor detectó 132 casos de `w-N h-N` (mismo valor en ambos ejes) en múltiples archivos
**Cuando** se ejecuta un script de reemplazo (`sed` o codemod) que convierte `w-4 h-4` → `size-4`, `w-6 h-6` → `size-6`, etc.
**Entonces** todos los archivos afectados usan el shorthand `size-N` de Tailwind v3.4+
**Y** la UI renderiza exactamente igual (size-N es equivalente semántico de w-N h-N)
**Y** `npm run build` pasa sin errores

**Dado** que 12 headings usan `font-bold` en lugar de `font-semibold`
**Cuando** se reemplazan los casos identificados
**Entonces** los headings usan `font-semibold` (600) en lugar de `font-bold` (700)
**Y** la tipografía de headings mejora visualmente sin afectar layout

**Dado** que 19 componentes usan `useRouter()` sin destructuring (ej. `router.push(...)`)
**Cuando** se refactorizan para destructurar el método necesario (`const { push } = useRouter()`)
**Entonces** el código es más explícito sobre qué métodos del router se usan
**Y** el React Compiler puede optimizar mejor los componentes afectados
**Y** `npm run lint && npm run build` pasan sin errores

**Verificación:** `npx react-doctor@latest . --verbose` no reporta los issues de Architecture en archivos modificados. Score React Doctor ≥ 90/100.

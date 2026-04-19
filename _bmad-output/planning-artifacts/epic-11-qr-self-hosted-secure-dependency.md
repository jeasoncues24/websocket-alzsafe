# Epic 11: QR Self-Hosted con Dependencia Segura

**Created:** 2026-04-18  
**Status:** proposed

## Contexto

Actualmente el frontend genera QR usando un servicio externo por URL query (`api.qrserver.com`). Eso introduce dependencia de tercero para una función crítica de login WhatsApp y expone datos QR en requests salientes.

Uso actual detectado en código:

- `frontend/app/empresas/[empresaId]/telefonos/[telefonoId]/connect/page.tsx:268`
- `frontend/app/sessions/page.tsx:147`

## Objetivo

Eliminar dependencia externa para rendering QR y adoptar una librería local, confiable y mantenida, con base para lectura/decodificación QR cuando se necesite en UX futura.

## Requisitos del Epic

- No depender de `api.qrserver.com` ni de cualquier endpoint remoto para pintar QR.
- Librería con licencia permisiva y mantenimiento activo.
- Evitar incremento grande de bundle en rutas que solo muestran QR.
- UX clara en desktop y mobile para escaneo rápido.

## Análisis de librerías (BA + UX)

Evaluación con criterios: seguridad/supply-chain, madurez, compatibilidad React, peso y UX de escaneo.

| Opción | Uso principal | Pros | Contras | Veredicto |
|---|---|---|---|---|
| `qrcode.react` (v4.2.0, ISC) | Render QR en React (SVG/Canvas) | Sin servicio externo, API simple, peer React nativo, buen fit para Next.js | No decodifica QR | **Seleccionada para P0** |
| `qrcode` (v1.5.4, MIT) | Generación QR JS genérica | Madura, muy usada, flexible | Trae dependencias extra (`pngjs`, `yargs`, `dijkstrajs`) innecesarias para UI React | Reserva/fallback |
| `@zxing/browser` + `@zxing/library` (v0.1.5 / v0.21.3, MIT) | Escaneo/decodificación (cámara) | Robusta para lectura QR real-time y casos enterprise | Paquete pesado; requiere lazy loading | **Seleccionada para P1 (decode)** |
| `jsqr` (v1.4.0, Apache-2.0) | Decodificación desde imagen/frame | Ligera, sin deps directas | Menos completa para flujos de cámara complejos | Alternativa si se prioriza peso |

### Decisión técnica

1. **P0 (obligatorio ahora):** usar `qrcode.react` para renderizar QR localmente.
2. **P1 (preparado en epic):** usar `@zxing/browser` solo en rutas que realmente requieran lectura/escaneo, vía import dinámico.

## Diseño UX acordado

- QR en `SVG` con contraste alto (fondo blanco, módulos negros) y quiet zone visible.
- Tamaño mínimo recomendado: `220px` desktop / `200px` mobile.
- Estado de carga y estado de expiración explícitos.
- Texto de instrucción corto debajo del QR y fallback de reintento.
- Mantener countdown visible cuando aplique (`qr_pending`).

## Stories

### 11-1: Infra QR segura (P0)

**Description:** instalar e integrar dependencia de render local QR y crear componente reusable.

**Implementación esperada:**

- Agregar `qrcode.react` a `frontend/package.json`.
- Crear componente reutilizable en `frontend/components/qr/qr-render.tsx`.
- Props mínimas: `value`, `size`, `title`, `className`.

**Acceptance Criteria:**

- [x] El componente renderiza QR válido a partir de `qrString`.
- [x] No hay requests de red a servicios externos al renderizar QR.

### 11-2: Reemplazo total de `api.qrserver.com` (P0)

**Description:** migrar todos los puntos detectados del frontend a render local.

**Implementación esperada:**

- Reemplazar URL en `frontend/app/empresas/[empresaId]/telefonos/[telefonoId]/connect/page.tsx:268`.
- Reemplazar URL en `frontend/app/sessions/page.tsx:147`.

**Acceptance Criteria:**

- [x] Búsqueda en repo de `api.qrserver.com` no retorna resultados en código de app.
- [ ] El flujo de conexión WhatsApp mantiene comportamiento funcional.

### 11-3: Hardening UX del bloque QR (P0)

**Description:** mejorar legibilidad y robustez del QR para evitar fallos de escaneo.

**Implementación esperada:**

- Unificar layout visual del QR en ambas pantallas.
- Agregar mensajes de estado consistentes para `initializing`, `qr_pending`, `active`, `disconnected`.
- Garantizar tamaño y márgenes aptos para cámara móvil.

**Acceptance Criteria:**

- [ ] Escaneo exitoso en iOS/Android en condiciones normales.
- [ ] El usuario entiende qué hacer en menos de 5 segundos.

### 11-4: Base de decodificación local (P1)

**Description:** preparar módulo de lectura/decodificación QR sin afectar bundle inicial.

**Implementación esperada:**

- Integrar `@zxing/browser` con import dinámico.
- Encapsular scanner en componente/hook para uso futuro.
- No habilitar cámara globalmente; solo bajo acción explícita del usuario.

**Acceptance Criteria:**

- [ ] La app puede decodificar QR en una vista de prueba interna.
- [ ] No impacta TTI de pantallas que no usan scanner.

## Riesgos y mitigación

- **Riesgo:** aumento de bundle por decodificación QR.  
  **Mitigación:** P1 con lazy loading y code splitting.
- **Riesgo:** regresión visual al reemplazar `<Image>` remoto.  
  **Mitigación:** componente único reusable con snapshot/manual QA.
- **Riesgo:** dependencias inseguras.  
  **Mitigación:** versión pinneada + revisión de advisories antes de merge.

## Definition of Done del Epic

- Todas las vistas productivas renderizan QR sin servicios externos.
- El flujo de login QR funciona extremo a extremo en ambiente local.
- Queda componente QR reusable documentado para próximas historias.
- Queda base técnica para lectura/decodificación en P1.

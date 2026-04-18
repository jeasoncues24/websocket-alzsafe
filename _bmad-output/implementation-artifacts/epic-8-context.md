# Epic 8 Context: Gestión de Empresas (Panel Admin)

<!-- Compiled from sprint status, API docs, and implementation artifacts. Edit freely. Regenerate when Epic 8 changes. -->

## Goal

Revalidar y estabilizar la capa de gestión de empresas del panel admin y su contrato cliente. El epic cubre el CRUD de empresas, la base de API keys por teléfono WhatsApp, la documentación operativa y el panel para listar, crear, editar, ver y eliminar empresas sin romper el aislamiento por empresa.

El foco actual de UX/UI cambió a una vista phone-first para creación, rotación y revocación de API keys con el tema nuevo de shadcn.

## Stories

- S-8.1: Testing CRUD Empresas (Endpoints Backend)
- S-8.2: Documentar endpoints de empresas
- S-8.3: Verificar/Implementar Endpoints de Empresas
- S-8.4: Vista panel admin empresas (listado)
- S-8.5: Vista panel admin crear/editar empresa
- S-8.6: Generación y gestión de tokens API
- S-8.6.1: Fix intermedio API keys por teléfono
- S-8.7: Documentar proceso conexión API externo
- S-8.8: Documentar envío de mensajes vía API
- S-8.9: [FUTURO] Vista números telefónicos por empresa (party mode)
- S-8.10: [FUTURO] Ruta /soporte con documentación API

## Requirements & Constraints

- El contrato admin vive en `/api/admin/*`; el contrato cliente vive en `/api/*`.
- El listado debe soportar paginación, búsqueda por nombre/RUC y filtro por estado.
- La creación exige RUC único y campos mínimos válidos.
- La actualización debe ser parcial; para cliente, el RUC es de solo lectura.
- El borrado es soft delete y no debe romper empresas con sesiones activas.
- La emisión y revocación de tokens API depende del modelo por teléfono y debe mostrar el secreto solo una vez al crear la key.
- La UI de tokens debe construirse con el theme shadcn nuevo aplicado en `frontend/app/globals.css`.
- La UX de API keys vive en páginas dedicadas por teléfono: listado de teléfonos por empresa y detalle de keys por teléfono.
- La UI de detalle debe mostrar datos reales del backend; si el alcance pide sesiones o mensajes recientes, el API debe exponerlos.
- El detalle de empresa debe incluir sesión WhatsApp actual y últimos mensajes recientes.
- El borrado debe rechazar empresas con sesiones WhatsApp activas.

## Technical Decisions

- Un único handler de empresas sirve para rutas de admin; el contrato de empresa usa middleware JWT propio en `/api/*`.
- La autorización admin se resuelve en `internal/http/middleware/auth.go`; los JWT de empresa se generan y validan en `internal/auth/empresa_jwt.go`.
- El panel de empresas consume la API desde `frontend/lib/api.ts`; mantener ahí la convención de rutas evita desalineaciones entre UI y backend.
- Para el seguimiento del sprint, `epic-8-context.md` es la fuente viva de este epic.

## Cross-Story Dependencies

- S-8.2 depende de que S-8.1 confirme el contrato real de endpoints.
- S-8.3 depende de S-8.1 y S-8.2 para no documentar o implementar sobre rutas equivocadas.
- S-8.4 y S-8.5 dependen de que la API de empresas esté estable y consistente.
- S-8.6 depende de la generación/revocación de tokens y de la política de `super_admin`.

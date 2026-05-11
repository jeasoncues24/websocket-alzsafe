# Epic 2: Mejoras Post-Revisión — Correcciones, QR, Métricas y Responsive

Status: in-progress

## Objetivo

Cerrar las brechas funcionales detectadas en la revisión del Epic 1: correcciones de bugs en WS y assets, completar el flujo de conexión QR desde la integración (API token), enriquecer el dashboard de métricas, mejorar el módulo de sesiones y preparar el panel para uso en móvil.

## Tablas disponibles (referencia)

| Tabla                   | Uso principal                                              |
|-------------------------|------------------------------------------------------------|
| messages                | mensajes enviados, estado, adjuntos_json, error_reason     |
| broadcasts              | difusiones, estado global                                  |
| broadcast_results       | resultado por destinatario                                 |
| admin_users             | usuarios del panel                                         |
| empresas                | empresas (multi-tenant), activo (soft delete)              |
| telefonos               | números WhatsApp, status, qr_string, session_data          |
| roles                   | roles de acceso                                            |
| modules / user_modules  | módulos asignados a usuarios                               |
| api_keys                | llaves de integración por teléfono                         |
| api_key_usage_events    | eventos de uso por request                                 |
| api_key_usage_daily     | rollup diario: requests, errores, latencia, mensajes       |
| api_key_audit_events    | auditoría de acciones sobre api_keys                       |
| audit_log               | auditoría genérica de entidades                            |
| token_blacklist         | JWTs revocados                                             |

## Stories

| ID    | Nombre                                          | Tipo              | Prioridad | Estado actual | Nota |
|-------|-------------------------------------------------|-------------------|-----------|---------------|------|
| 2-1   | Eliminar select empresa de usuario_admin        | Frontend          | Alta      | review        | Se re-ejecutará validación funcional antes de cerrar |
| 2-2   | Restaurar empresas desactivadas                 | Backend+Frontend  | Alta      | review        | Se re-ejecutará implementación/validación por mezcla con 2-2.5 |
| 2-2.5 | Limpiar empresa del JWT admin                   | Backend           | Alta      | review        | Fix de autorización extraído para destrabar panel admin |
| 2-3   | Investigar y corregir assets en mensajes        | Backend+Frontend  | Alta      | backlog       | — |
| 2-4   | Módulo sesiones: herramientas y logs            | Backend+Frontend  | Media     | backlog       | — |
| 2-5   | Validar reconexión al reiniciar binario         | Backend           | Media     | backlog       | — |
| 2-6   | Fix WS timer + simplificar UI de conexión       | Backend+Frontend  | Alta      | backlog       | — |
| 2-7   | QR por API token: WS + fallback REST            | Backend+Frontend  | Alta      | backlog       | — |
| 2-8   | Enlace QR compartible: token provisional        | Backend+Frontend  | Alta      | backlog       | — |
| 2-9   | Enriquecer métricas + rediseñar dashboard       | Backend+Frontend  | Media     | backlog       | — |
| 2-10  | Responsive mode del panel administrativo        | Frontend          | Baja      | backlog       | — |
| 2-11  | Documentar arquitectura WS                      | Documentación     | Baja      | done          | — |
| 2-12  | Refactorizar métricas API key                   | Backend+Frontend  | Media     | review        | — |

## Orden de implementación y dependencias

```
2-1 (aislado, pero pendiente de revalidación final)
2-2 (aislado, pero se re-ejecutará por mezcla previa con 2-2.5)
2-2.5 (hardening de autorización; soporte para validar correctamente operaciones admin del panel)
2-3 (aislado)
2-4 (aislado)
2-5 (aislado)
2-6 → prerequisito para 2-7, 2-8, 2-11
2-7 → prerequisito para 2-8, 2-11
2-8 → prerequisito para 2-11
2-9 (aislado, puede ir en paralelo)
2-10 (aislado, al final)
2-11 (al final, depende de 2-6, 2-7, 2-8)
```

## Estado operativo actual

- `2-2.5` quedó en `review` tras completar el refactor backend de autorización del panel y validar `go build` + tests internos relevantes.
- `2-1` y `2-2` se mantendrán bajo observación y se volverán a ejecutar/validar antes de cerrarlas definitivamente.
- El resto del epic continúa en `backlog`.

## Criterio de cierre del epic

Todas las stories en estado `done` y la reconexión WS documentada y validada en producción.

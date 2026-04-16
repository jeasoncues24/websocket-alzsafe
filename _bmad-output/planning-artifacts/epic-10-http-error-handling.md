# Epic 10: HTTP Error Handling

**Created:** 2026-04-16
**Status:** done

## Overview

Implementar manejo correcto de errores HTTP en el backend Go:
- 404 para rutas no encontradas
- 405 para métodos incorrectos
- Endpoint /health para verificar que el API está corriendo

## Stories

### 10-1: 404 para rutas no definidas
- **Description:** Cuando se accede a una ruta que no existe (ej: /api/cualquier-cosa), devolver 404 Not Found
- **Status:** done
- **Implementation:** `internal/http/router.go` - catch-all handler wrapper

### 10-2: 405 para métodos incorrectos  
- **Description:** Cuando se usa un método HTTP no permitido (ej: PUT /message que solo acepta POST), devolver 405 Method Not Allowed con header Allow
- **Status:** done
- **Implementation:** `internal/http/router.go` - registeredRoutes map y función handleOtherMethods

### 10-3: Endpoint /health
- **Description:** Crear endpoint GET /health que responde con {"status": "ok", "message": "API is running"}
- **Status:** done
- **Implementation:** `internal/http/router.go` - función HandleHealth

## Changes Made

1. Added `HandleHealth` function - endpoint que devuelve estado OK
2. Added `registeredRoutes` map - lista de todas las rutas y métodos válidos
3. Added `handleAPI` - maneja /api/* y /admin/* paths
4. Added `handleOtherMethods` - maneja rutas públicas con validación de método
5. Added `handleCatchAll` - wrapper que maneja / root y catch-all
6. Updated router to wrap all routes with proper 404/405 handling

## Acceptance Criteria

- [x] DELETE /api/cualquier-cosa retorna 404
- [x] GET /api/ruta-inexistente retorna 404
- [x] PUT /message retorna 405 (solo POST permitido)
- [x] DELETE /message retorna 405 (solo POST permitido)
- [x] POST / retorna 405 (solo GET permitido)
- [x] GET /health retorna {"status": "ok", "message": "API is running"}
- [x] GET / retorna información del servicio
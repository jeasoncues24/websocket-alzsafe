# Story 1.8: CORS en el backend — aceptar cualquier origen

Status: done

## Story

As a desarrollador,
I want que el backend Go acepte peticiones de cualquier origen (CORS),
so that el frontend pueda comunicarse con el backend desde cualquier dominio sin errores de browser.

## Contexto

Al separar el frontend (Next.js en puerto 3000) del backend (Go en puerto 8080), el browser aplica la política Same-Origin y bloquea las peticiones cross-origin si el servidor no envía los headers `Access-Control-*` correctos.

## Estado: ya implementado

**No se requiere ningún cambio.** El CORS global ya está activo en producción.

### Implementación existente

**`backend/internal/http/middleware.go:26`** — `CORSMiddleware`:
```go
func CORSMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

**`backend/internal/http/kernel.go:22`** — aplicado globalmente:
```go
Global: []func(http.Handler) http.Handler{
    CORSMiddleware,         // ← primer middleware, todas las rutas
    CorrelationIDMiddleware,
    LoggingMiddleware,
},
```

### Cobertura

- Origen: `*` (cualquier dominio)
- Métodos: GET, POST, PUT, DELETE, OPTIONS
- Headers permitidos: Content-Type, Authorization
- Preflight (OPTIONS): responde 200 inmediatamente

## Acceptance Criteria

1. El backend responde con `Access-Control-Allow-Origin: *` en todas las rutas. ✅
2. Las peticiones OPTIONS (preflight) retornan HTTP 200 sin pasar por los handlers de negocio. ✅
3. El frontend puede hacer fetch al backend desde un origen diferente sin errores CORS. ✅

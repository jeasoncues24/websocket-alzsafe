# Cierre de jornada - 2026-04-15

## Estado actual

- Epic 3: in-progress
- Story 3.1 (Endpoint difusión con validación de lista_difusion): **review — código hecho, 3 patches pendientes del code review**
- Sprint-status: `3-1-...` = `review`

Fuente de verdad del estado:

- `_bmad-output/implementation-artifacts/sprint-status.yaml`
- `_bmad-output/implementation-artifacts/3-1-endpoint-de-difusion-con-validacion-de-lista_difusion.md`

---

## Punto exacto para retomar

El code review de Story 3.1 detectó **3 findings tipo `patch`** (sin ambigüedad, aplicables directamente). Al retomar, la acción es:

> **Responder "0" (batch-apply all)** o "1" al mensaje del code review para que se apliquen los 3 fixes.

---

## Findings del code review (pendientes de aplicar)

### [High] F1 — Sin límite de tamaño en `lista_difusion`

**Archivo:** `internal/http/validator.go` → `ValidateBroadcastRequest`

**Problema:** No hay bound en el array. Un cliente puede enviar miles de ítems y saturar CPU (iteración O(n) sin límite).

**Fix a aplicar:**

```go
// Agregar en domain/broadcast.go (o message.go como constante compartida)
const MaxBroadcastItems = 500

// Al inicio de ValidateBroadcastRequest, después del check len == 0:
if len(req.ListaDifusion) > domain.MaxBroadcastItems {
    return &domain.ValidationError{
        Code:    domain.ErrorCodeValidation,
        Message: fmt.Sprintf("lista_difusion cannot exceed %d items", domain.MaxBroadcastItems),
    }
}
```

---

### [High] F2+AC2 — Tipo incorrecto retorna INVALID_JSON en lugar de VALIDATION_ERROR

**Archivo:** `internal/http/handlers.go` → `HandlePostBroadcast`

**Problema:** El AC2 dice que si `lista_difusion` es JSON válido pero no array (e.g. `{"lista_difusion":{}}`) la API debe responder `VALIDATION_ERROR`. Sin embargo, con la implementación actual Go falla en `Decode` y retorna `INVALID_JSON`.

**Fix a aplicar — decode en dos fases:**

```go
// Cambiar HandlePostBroadcast para usar decode en dos fases:
// Fase 1: decode a estructura auxiliar con lista_difusion como json.RawMessage
type broadcastRaw struct {
    RUCEmpresa    string          `json:"ruc_empresa"`
    ListaDifusion json.RawMessage `json:"lista_difusion"`
}

var raw broadcastRaw
if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
    // JSON genuinamente malformado
    w.WriteHeader(stdhttp.StatusBadRequest)
    json.NewEncoder(w).Encode(domain.BroadcastResponse{
        OK: false, Error: domain.ErrorCodeInvalidJSON, Details: "Invalid JSON in request body",
    })
    return
}

// Fase 2: intentar unmarshal de lista_difusion como array
var items []domain.BroadcastItem
if len(raw.ListaDifusion) == 0 || raw.ListaDifusion[0] != '[' {
    // Es un JSON válido pero no empieza con '[' → no es array
    w.WriteHeader(stdhttp.StatusBadRequest)
    json.NewEncoder(w).Encode(domain.BroadcastResponse{
        OK: false, Error: domain.ErrorCodeValidation, Details: "lista_difusion must be a non-empty array",
    })
    return
}
if err := json.Unmarshal(raw.ListaDifusion, &items); err != nil {
    w.WriteHeader(stdhttp.StatusBadRequest)
    json.NewEncoder(w).Encode(domain.BroadcastResponse{
        OK: false, Error: domain.ErrorCodeInvalidJSON, Details: "lista_difusion contains invalid JSON",
    })
    return
}

req := domain.BroadcastRequest{
    RUCEmpresa:    raw.RUCEmpresa,
    ListaDifusion: items,
}
```

---

### [Med] F3 — Test ausente para AC2 (lista no-array)

**Archivo:** `internal/http/handlers_test.go`

**Fix a aplicar:** Agregar test:

```go
func TestHandlePostBroadcastListaNoArray(t *testing.T) {
    manager := whatsapp.NewManager()
    sessionStore := storage.NewSessionStore()
    handler := NewHandler(manager, sessionStore, nil)

    // lista_difusion es un objeto válido JSON, no un array
    body := `{"ruc_empresa":"20123456789","lista_difusion":{"destino":"51999999999"}}`
    req := httptest.NewRequest(stdhttp.MethodPost, "/broadcast", strings.NewReader(body))
    rr := httptest.NewRecorder()

    handler.HandlePostBroadcast(rr, req)

    if rr.Code != stdhttp.StatusBadRequest {
        t.Fatalf("expected 400, got %d", rr.Code)
    }
    var resp domain.BroadcastResponse
    json.NewDecoder(rr.Body).Decode(&resp)
    if resp.Error != domain.ErrorCodeValidation {
        t.Fatalf("expected VALIDATION_ERROR, got: %s", resp.Error)
    }
}
```

---

## Archivos a modificar al retomar

| Archivo                          | Cambio                                                          |
| -------------------------------- | --------------------------------------------------------------- |
| `internal/domain/broadcast.go`   | Agregar constante `MaxBroadcastItems = 500`                     |
| `internal/http/validator.go`     | Agregar check de límite máximo después del check de lista vacía |
| `internal/http/handlers.go`      | Reescribir decode de `HandlePostBroadcast` en dos fases         |
| `internal/http/handlers_test.go` | Agregar `TestHandlePostBroadcastListaNoArray`                   |

---

## Workflow al retomar

1. Ejecutar `go test ./...` para confirmar que los 39 tests siguen verdes
2. Aplicar los 3 fixes de arriba
3. Correr tests y confirmar verde (deben sumar 40 tests al agregar el nuevo)
4. Marcar los 3 action items del review en el story file (`[x]`)
5. Actualizar story status → `done`
6. Actualizar sprint-status → `3-1-... = done`
7. Iniciar Story 3.2 con `create story 3.2`

---

## Checklist de arranque

- [ ] Ejecutar `go test ./...` → debe ser 39 pass
- [ ] Aplicar F1: límite MaxBroadcastItems
- [ ] Aplicar F2/AC2: decode en dos fases
- [ ] Aplicar F3: test de lista no-array
- [ ] `go test ./...` → debe ser 40 pass
- [ ] Marcar review items en story 3.1 y mover a `done`
- [ ] Crear Story 3.2

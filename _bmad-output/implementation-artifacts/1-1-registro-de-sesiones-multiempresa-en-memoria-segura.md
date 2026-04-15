# Story 1.1: Registro de sesiones multiempresa en memoria segura

Status: done

## Story

As a backend platform,
I want administrar clientes por ruc_empresa con control concurrente,
so that evitemos colisiones y estados corruptos entre empresas.

## Acceptance Criteria

1. Dado un servicio iniciado, cuando se registra o consulta un cliente por ruc_empresa, entonces las operaciones son thread-safe y no existe fuga de referencias entre empresas.
2. Dado multiples accesos concurrentes de lectura/escritura, cuando se ejecutan en paralelo, entonces no se generan data races ni panics.
3. Dado un cliente existente para una empresa, cuando se reemplaza la referencia, entonces la nueva referencia queda activa y la anterior deja de estar disponible para lecturas futuras.
4. Dado una empresa sin cliente registrado, cuando se consulta, entonces se obtiene respuesta explicita de no encontrado sin error fatal.

## Tasks / Subtasks

- [ ] Definir contrato del Session Manager (AC: 1, 4)
  - [ ] Establecer interfaz minima: Get, Set, Delete, Exists, Count, ListKeys.
  - [ ] Definir tipo de clave de empresa y reglas de normalizacion de ruc_empresa.
- [ ] Implementar manager concurrente en capa whatsapp (AC: 1, 2, 3, 4)
  - [ ] Implementar estructura basada en mapa protegido por RWMutex.
  - [ ] Asegurar operaciones atomicas de reemplazo y borrado.
- [ ] Integrar manager con puntos de entrada actuales (AC: 1, 3, 4)
  - [ ] Reemplazar accesos directos a mapa global por manager en flujos de inicializacion.
  - [ ] Evitar compartir punteros mutables fuera del manager.
- [ ] Agregar pruebas unitarias y de concurrencia (AC: 1, 2, 3, 4)
  - [ ] Pruebas de comportamiento basico por operacion.
  - [ ] Pruebas concurrentes con tabla de casos.
  - [ ] Ejecutar con race detector.
- [ ] Alinear observabilidad minima de este componente (AC: 1, 3)
  - [ ] Logs estructurados para eventos de set/delete por ruc_empresa.
  - [ ] Sin exponer datos sensibles en logs.

## Dev Notes

- Este story es fundacional para el resto del Epic 1: no acoplar logica de transporte WebSocket ni envio de mensajes aqui.
- Mantener el alcance en gestion segura de clientes por empresa.
- Debe conservar compatibilidad funcional con baseline usqay para no romper eventos posteriores.

### Technical Requirements

- Mantener estructura por capas y paquetes actuales:
  - [internal/whatsapp](internal/whatsapp)
  - [internal/http](internal/http)
  - [internal/config](internal/config)
- Usar sincronizacion explicita con RWMutex para lecturas/escrituras concurrentes.
- No introducir estado global adicional fuera del manager.
- Evitar dependencias nuevas para esta historia salvo justificacion fuerte.

### Suggested File Targets

- [internal/whatsapp/manager.go](internal/whatsapp/manager.go)
- [internal/whatsapp/client.go](internal/whatsapp/client.go)
- [internal/whatsapp/qr.go](internal/whatsapp/qr.go)
- [internal/http/handlers.go](internal/http/handlers.go)

### Testing Requirements

- Crear archivo de pruebas para manager:
  - [internal/whatsapp/manager_test.go](internal/whatsapp/manager_test.go)
- Cobertura minima esperada:
  - Lectura/escritura normal
  - Reemplazo de referencia
  - Eliminacion y no encontrado
  - Concurrencia bajo carga

### References

- Story source: [ \_bmad-output/planning-artifacts/epics.md ](_bmad-output/planning-artifacts/epics.md#L73)
- PRD source: [ \_bmad-output/planning-artifacts/prd.md ](_bmad-output/planning-artifacts/prd.md#L76)
- Project context: [ \_bmad-output/project-context.md ](_bmad-output/project-context.md#L31)
- Codigo base manager: [internal/whatsapp/manager.go](internal/whatsapp/manager.go#L1)

## Dev Agent Record

### Agent Model Used

GPT-5.3-Codex

### Debug Log References

- N/A

### Completion Notes List

- Manager multiempresa extendido con Delete, Exists, Count y ListKeys.
- Se agregaron pruebas unitarias/concurrencia para manager.
- Compilacion global validada con go test ./...

### File List

- \_bmad-output/implementation-artifacts/1-1-registro-de-sesiones-multiempresa-en-memoria-segura.md

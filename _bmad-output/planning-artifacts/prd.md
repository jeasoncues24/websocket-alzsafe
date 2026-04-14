---
stepsCompleted:
  - init
  - discovery
  - journeys
  - success-metrics
  - domain
  - project-type
  - scoping
  - functional
  - nonfunctional
  - polish
  - complete
inputDocuments:
  - _bmad-output/project-context.md
workflowType: "prd"
status: "complete"
version: "1.0"
---

# Product Requirements Document - wsapi

**Author:** Fulanito  
**Date:** 2026-04-14

## 1. Resumen Ejecutivo

Este producto migra y endurece a nivel produccion un sistema de mensajeria WhatsApp multiempresa que hoy tiene base funcional en Node.js. El objetivo es operar en Go con alta estabilidad, trazabilidad operativa y capacidad de escalar por empresa sin degradacion cruzada.

Alcance actual: API HTTP + WebSocket + integracion WhatsApp multiempresa + persistencia MariaDB, con ejecucion local/nativa sin depender de Docker en esta etapa.

## 2. Problema y Oportunidad

El sistema actual presenta inestabilidad operativa en el stack legado y riesgos de escalabilidad por manejo de sesiones y eventos concurrentes. Se requiere:

- Reducir fallas de sesion y desconexiones no controladas.
- Mantener compatibilidad funcional con flujos existentes (QR, estados, envio directo, difusion).
- Mejorar arquitectura para operacion multiempresa con aislamiento logico por ruc_empresa.
- Preparar base para operar en produccion con observabilidad, resiliencia y migraciones de datos seguras.

## 3. Objetivos

### 3.1 Objetivos de negocio

- Aumentar confiabilidad de envio de mensajes por empresa.
- Reducir incidentes operativos por reconexiones y limpieza de sesiones.
- Habilitar crecimiento de numero de empresas activas sin reescritura de arquitectura.

### 3.2 Objetivos tecnicos

- Migrar capacidades core desde Node.js a Go sin perdida funcional.
- Introducir contratos API y eventos estables y testeables.
- Diseñar capa de persistencia y migraciones DB apta para evolucion continua.

## 4. Alcance

### 4.1 En alcance

- API HTTP para:
  - envio directo de mensajes (texto + adjunto opcional)
  - difusion masiva con resultado por destinatario
  - gestion basica de empresas/usuarios habilitados
- WebSocket para:
  - inicio de sesion por empresa
  - entrega de QR por empresa
  - notificaciones de estado (activo, desconectado, error)
- Integracion WhatsApp multiempresa con aislamiento por ruc_empresa.
- Persistencia MariaDB de:
  - usuarios/empresas
  - mensajes enviados
  - estado operativo por empresa
- Plan de migracion de base de datos y endurecimiento de esquema.
- Reglas de observabilidad y confiabilidad para produccion.

### 4.2 Fuera de alcance (fase actual)

- Orquestacion y despliegue con Docker/Kubernetes.
- UI frontend nueva.
- Analitica avanzada o BI.

## 5. Personas y Casos de Uso

### 5.1 Operador de empresa

- Inicia sesion WhatsApp de su empresa mediante QR.
- Verifica estado de conexion.
- Envia mensajes directos y campañas de difusion.

### 5.2 Administrador de plataforma

- Administra empresas activas/vinculadas.
- Monitorea incidentes y estados de conectividad.
- Audita envio de mensajes y resultados.

## 6. User Journeys Principales

### Journey A - Bootstrap de sesion

1. Operador solicita init-session para ruc_empresa.
2. Backend inicializa cliente WhatsApp aislado por empresa.
3. WebSocket emite qr-ruc_empresa.
4. Operador escanea QR.
5. Backend emite active-ruc_empresa y marca estado operativo.

### Journey B - Envio directo

1. Operador llama endpoint de envio con destino y mensaje.
2. Backend valida datos y estado de cliente.
3. Backend envia mensaje por WhatsApp.
4. Backend persiste trazabilidad del envio.
5. API responde estado final.

### Journey C - Difusion multi-destino

1. Operador envia lista_difusion.
2. Backend procesa destinatarios uno a uno con tolerancia a errores.
3. Se persiste cada envio exitoso.
4. Respuesta incluye resultados por destinatario (enviado/error).

## 7. Requisitos Funcionales

- FR-01: El sistema debe iniciar sesion WhatsApp por ruc_empresa mediante evento init-session.
- FR-02: El sistema debe emitir QR por empresa mediante evento qr-ruc_empresa.
- FR-03: El sistema debe emitir cambios de estado por empresa mediante active-ruc_empresa y error-event.
- FR-04: El sistema debe permitir envio directo con validacion estricta de telefono y codigo postal.
- FR-05: El sistema debe soportar adjunto opcional en envio directo con validacion de tipo y tamaño.
- FR-06: El sistema debe soportar difusion masiva con resultado individual por destinatario.
- FR-07: El sistema debe persistir mensajes enviados con metadatos minimos (empresa, destino, timestamp, resultado).
- FR-08: El sistema debe mantener mapa activo de clientes WhatsApp por empresa con acceso concurrente seguro.
- FR-09: El sistema debe manejar desconexion/logout y limpieza de sesion sin afectar otras empresas.
- FR-10: El sistema debe permitir activar/desactivar empresas para servicio de mensajeria.

## 8. Requisitos No Funcionales

- NFR-01 (Confiabilidad): El servicio debe tolerar fallas parciales por empresa sin caida global.
- NFR-02 (Escalabilidad): La arquitectura debe soportar crecimiento horizontal de empresas activas mediante aislamiento por clave empresa.
- NFR-03 (Rendimiento): El envio por difusion no debe bloquear el event loop global; usar concurrencia controlada por worker pool.
- NFR-04 (Observabilidad): Logs estructurados por correlacion y ruc_empresa; metricas de sesion, reconexion y latencia.
- NFR-05 (Seguridad): Secretos solo por entorno; sin credenciales hardcodeadas.
- NFR-06 (Mantenibilidad): Contratos API y eventos versionados para cambios compatibles.
- NFR-07 (Integridad de datos): Migraciones idempotentes y reversibles para cambios de esquema.
- NFR-08 (Testabilidad): Cobertura automatizada para rutas criticas (sesion, envio directo, difusion, desconexion).

## 9. Modelo de Dominio (alto nivel)

- Empresa
  - ruc_empresa (PK logica)
  - estado_servicio
  - telefono principal
  - metadata de vinculacion
- SesionWhatsApp
  - ruc_empresa
  - estado (initializing, qr_pending, authenticated, active, disconnected)
  - last_error
  - updated_at
- MensajeEnviado
  - id
  - ruc_empresa
  - destino
  - payload
  - tipo (directo, difusion)
  - estado
  - provider_timestamp
  - created_at

## 10. Migracion de Base de Datos

### 10.1 Estrategia

- Baseline del esquema actual y data dictionary.
- Definir versionado de migraciones (up/down) en orden estricto.
- Ejecutar migraciones en ambiente staging antes de produccion.

### 10.2 Cambios esperados

- Normalizar entidades de empresa y estados de sesion.
- Agregar indices para consultas por ruc_empresa, estado y fecha.
- Definir constraints para evitar duplicidad y datos inconsistentes.
- Registrar auditoria minima de eventos criticos de sesion.

### 10.3 Reglas de rollout

- Migraciones backward-compatible en primera fase.
- Feature flags para activar nuevas rutas sin corte total.
- Plan de rollback documentado por release.

## 11. Criterios de Exito

- 95% de sesiones multiempresa inicializan sin intervencion manual adicional.
- 99% de envios directos validos terminan en respuesta consistente API.
- 100% de difusiones retornan matriz de resultado por destinatario.
- Reduccion de incidentes de desconexion no recuperada respecto a baseline legado.

## 12. Riesgos y Mitigaciones

- Riesgo: Divergencia funcional entre Node y Go.
  - Mitigacion: Matriz de equivalencia por caso de uso y pruebas de regresion.
- Riesgo: Bloqueo por I/O o reconexiones concurrentes.
  - Mitigacion: worker pools, timeouts, circuit breakers y limites por empresa.
- Riesgo: Cambios de esquema que afecten data historica.
  - Mitigacion: migraciones graduales, backup y rollback automatizado.

## 13. Entregables

- PRD aprobado.
- Epics y stories con acceptance criteria testeables.
- Sprint planning inicial con estados backlog.
- Base para arquitectura tecnica y posterior implementacion.

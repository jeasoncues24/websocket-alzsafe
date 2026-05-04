---
title: 'README de deploy a producción'
type: 'chore'
created: '2026-05-04'
status: 'done'
route: 'one-shot'
---

# README de deploy a producción

## Intent

**Problem:** El repo no tenía un README raíz alineado con el último epic y seguía existiendo riesgo de interpretar mal el deploy productivo como un `docker compose up` de stack completo.

**Approach:** Documentar el flujo oficial vigente de producción — build-only con Docker, runtime con PM2 en host y base de datos externa — y dejar un resumen operativo consistente en `docker/production.md`.

## Suggested Review Order

**Flujo principal**

- Explica el contrato de despliegue actual y evita asumir runtime dentro de Docker.
  [`README.md:5`](../../README.md#L5)

- Deja el paso a paso operativo para preparar entorno, compilar y arrancar con PM2.
  [`README.md:107`](../../README.md#L107)

- Añade verificación, actualización, rollback y cautelas operativas para producción.
  [`README.md:177`](../../README.md#L177)

- Cubre migraciones, troubleshooting y límites explícitos del alcance actual.
  [`README.md:250`](../../README.md#L250)

**Resumen corto**

- Alinea el atajo de producción con el flujo real y corrige la expectativa sobre Compose.
  [`production.md:1`](../../docker/production.md#L1)

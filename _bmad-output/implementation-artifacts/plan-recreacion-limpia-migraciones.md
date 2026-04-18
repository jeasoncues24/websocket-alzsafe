---
status: done
created: 2026-04-18
last_updated: 2026-04-18
---

# Plan de Recreación Limpia de Migraciones

## Objetivo

Recrear migraciones desde cero sin duplicados, sin defaults relacionales inválidos y sin romper la trazabilidad del esquema.

## Prerrequisitos

1. Documento maestro del esquema aprobado.
2. Backup/export del estado actual de la base.
3. Validación de que no faltan tablas críticas.
4. Confirmación de que no hay nuevas relaciones con default `0`.

## Reglas

- No borrar migraciones antes del documento maestro.
- No mezclar reset de esquema con cambios funcionales.
- No asumir FKs implícitas.
- No introducir defaults `0` en columnas relacionales.

## Orden de reconstrucción

1. Tablas base sin dependencias.
2. Tablas con FK lógica o física.
3. Tablas de consumo y auditoría.
4. Seeds y ajustes finales.

## Validación Posterior

```bash
go run . migrate status
go run . migrate up
go run . migrate down
```

Verificar:

- sin índices duplicados
- sin constraints faltantes
- sin divergencia con el documento maestro

## Riesgos

- relaciones lógicas sin FK física
- campos relacionales con defaults históricos `0`
- cambios de contrato donde `telefono` pasa a `telefono_contacto`

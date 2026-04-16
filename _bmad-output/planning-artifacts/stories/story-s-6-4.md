# Story S-6.4: Middleware Autenticación JWT

## Epic
Epic 6: Sistema de Autenticación JWT por Empresa

## Prioridad
P0

## Estado
pending

## Overview

Implementar middleware HTTP que extrae y valida el JWT del header `Authorization: Bearer <token>`.

## Acceptance Criteria

- [ ] Extrae token del header Authorization
- [ ] Valida firma JWT
- [ ] Verifica expiración
- [ ] Valida token_version contra DB (con cache)
- [ ] Agrega empresa_id al context
- [ ] Retorna 401 con mensaje claro

## Casos de Error

| Escenario | Código | Mensaje |
|-----------|--------|--------|
| Sin token | 401 | "Missing authorization token" |
|malformed | 401 | "Invalid token format" |
| Expirado | 401 | "Token expired" |
| Revocado | 401 | "Token revoked" |
| Empresa no existe | 404 | "Empresa not found" |
| Sin permiso | 403 | "Insufficient permissions" |

## Dependencias
- S-6.3 (generación JWT)

## Estimated Effort
2 days
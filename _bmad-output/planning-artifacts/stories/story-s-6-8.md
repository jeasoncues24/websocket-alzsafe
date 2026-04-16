# Story S-6.8: Tests de Integración

## Epic
Epic 6: Sistema de Autenticación JWT por Empresa

## Prioridad
P0

## Estado
pending

## Overview

Implementar tests de integración para el middleware y flujos completos.

## Casos de Prueba

| Test | Esperado |
|------|---------|
| Request sin token | 401 "Missing token" |
| Request con token válido | 200 |
| Request con token expirado | 401 |
| Request con token revocado | 401 |
| Request sin permiso | 403 |
| Happy path: send message | 200 |

## Dependencias
- S-6.4, S-6.5

## Estimated Effort
2 days
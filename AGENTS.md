# AGENTS.md — Guía de Agentes BMad

Este proyecto usa **BMad** (BMad Model) como sistema de automatización de desarrollo. Los agentes están instalados en `.opencode/skills/`.

---

## Cómo Invocar Agentes

Usa el tool `skill` con el nombre del skill:

```
skill(name: "bmad-agent-dev")
```

O simplemente describe lo que necesitas y el sistema routing elige el skill apropiado.

---

## Agentes Principales (más usados)

| Código | Skill | Cuándo Usarlo |
|--------|-------|----------------|
| DS | `bmad-dev-story` | Implementar siguiente story del sprint |
| QD | `bmad-quick-dev` | Quick fixes, spinoffs, cambios rápidos |
| CR | `bmad-code-review` | Revisar código antes de merge |
| SP | `bmad-sprint-planning` | Planificar sprint |
| CS | `bmad-create-story` | Crear nueva story |
| QA | `bmad-qa-generate-e2e-tests` | Generar tests E2E |
| ER | `bmad-retrospective` | Hacer retrospective de epic |

---

## Flujo de Trabajo Típico

```
1. Ayuda → "¿Qué hacemos?" → bmad-help
2. Planning → bmad-sprint-planning / bmad-create-epics-and-stories
3. Arquitectura → bmad-create-architecture
4. Implementar → bmad-agent-dev → bmad-dev-story
5. Review → bmad-code-review
6. Done → bmad-sprint-status
```

---

## Comandos de Desarrollo

```bash
# Tests
go test ./...

# Test coverage
go test -cover ./...

# Build (incluye frontend)
make build

# Run
go run .
```

---

## Project Context

- **Stack Backend:** Go 1.25 + WhatsApp (go.mau.fi/whatsmeow)
- **Stack Frontend:** Next.js 14 + shadcn/ui + Zustand
- **HTTP:** net/http estándar
- **WebSocket:** github.com/coder/websocket
- **Persistencia:** MariaDB/mysql
- **Artefactos BMad:** `_bmad-output/implementation-artifacts/`
- **Stories activas:** `_bmad-output/implementation-artifacts/sprint-status.yaml`

---

## Convenciones Importantes

### Backend (Go)
- Paquetes en `internal/` (config, http, storage, whatsapp)
- Todo request validation en `internal/http/validator.go`
- Handlers HTTP en `internal/http/handlers.go`
- Tipos de dominio en `internal/domain/`
- No estado global mutable — usar managers con mutex

### Frontend (Next.js)
- Estructura: `frontend/app/`, `frontend/components/`, `frontend/stores/`
- UI: shadcn/ui con Tailwind
- Estado: Zustand con persistencia
- Theme: light/dark con next-themes

---

## Episodios Completados

| Epic | Estado | Stories |
|------|--------|---------|
| Epic 1: Sesiones WhatsApp | ✅ done | 4/4 |
| Epic 2: Mensajería Directa | ✅ done | 3/3 |
| Epic 3: Difusión Masiva | ✅ done | 3/3 |
| Epic 4: Infra (Migraciones, Índices, Observabilidad) | ✅ done | 3/3 |
| Epic 5: Panel Admin (Next.js) | ✅ done | 7/7 + 1 |

## Cómo Hacer Build

```bash
# Build completo (frontend + backend + static)
make build

# Solo frontend
make frontend

# Solo backend
make backend

# Desarrollo
make dev
```

## Cómo Ejecutar

```bash
# Requiere MariaDB configurada en variables de entorno
# DB_HOST, DB_PORT, DB_USER, DB_PASS, DB_NAME

./wsapi
# Servidor corre en :8080
# Panel admin: http://localhost:8080
```

---

## Referencias

- Project context: `_bmad-output/project-context.md`
- Sprint status: `_bmad-output/implementation-artifacts/sprint-status.yaml`
- Docs: `_bmad-output/planning-artifacts/`
- Epic 5 Frontend: `_bmad-output/planning-artifacts/epic-5-frontend-panel.md`
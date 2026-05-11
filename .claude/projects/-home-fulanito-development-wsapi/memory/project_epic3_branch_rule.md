---
name: project-epic3-branch-rule
description: Regla imperativa de rama para Epic 3 — todo código del epic va en feature/security, sin excepciones
metadata:
  type: project
---

El Epic 3 (Hardening de Seguridad y Calidad Frontend) tiene una regla de rama imperativa: **todo código debe implementarse en `feature/security`**.

**Why:** El epic es experimental y de seguridad. El usuario quiere aislar completamente estos cambios de la rama principal (`v1`/`main`) para poder verificarlos y revertirlos si es necesario sin afectar producción.

**How to apply:**
- Antes de implementar cualquier story del Epic 3 (3.1–3.8), verificar con `git branch --show-current` que la rama activa es `feature/security`.
- Si la rama no es `feature/security`, detener y notificar al usuario antes de escribir una sola línea de código.
- Los artefactos BMad (`_bmad-output/`, `docs/`) pueden editarse en cualquier rama.
- Solo el código fuente (`frontend/`, `backend/`) del Epic 3 está restringido a `feature/security`.
- Ningún otro epic o tarea debe usar esta rama — es exclusiva del Epic 3.
- Merge a main solo cuando todas las 8 stories estén `done` y `npm run build` + `go build ./...` pasen.

La rama fue creada el 2026-05-11 desde la rama `v1`.
Documentada en: `docs/bmad-project-rules.md` (sección "Regla de Rama"), `epic-3-hardening-seguridad-calidad-frontend.md`.

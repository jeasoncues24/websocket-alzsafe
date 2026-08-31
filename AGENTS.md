# AGENTS.md

Política operativa para cualquier agente de IA (Claude Code, Codex, Cursor, Antigravity, etc.) en el repo `wsapi`. Para comandos de verificación y convenciones de código/estilo, ver [`CLAUDE.md`](CLAUDE.md) — es equivalente y aplica igual sin importar qué herramienta abra el repo.

## Regla obligatoria: preguntas antes de cualquier plan o artifact (ASK_QUESTION_USER)

- **Gate obligatorio:** Antes de generar **cualquier plan** (modo plan / `ExitPlanMode`) o **cualquier artifact**, el agente DEBE presentar primero al usuario un bloque de preguntas de aclaración mediante `ASK_QUESTION_USER` (la herramienta `AskUserQuestion`) y esperar las respuestas.
- **Sin preguntas respondidas → sin plan ni artifact:** No se puede construir, mostrar ni ejecutar un plan o un artifact si no hay un ciclo de preguntas terminado. Si crees que no hay nada que preguntar, igual debes plantear al menos las decisiones abiertas (alcance, ambigüedades reales, criterios de aceptación, restricciones técnicas) y confirmarlas.
- **Cobertura de las preguntas:** alcance exacto, interpretaciones alternativas válidas, archivos/módulos candidatos, criterios de aceptación, restricciones (portabilidad MySQL/SQLite, imports `wsapi/internal/...`, seguridad de campos sensibles) y cualquier requisito subjetivo sin criterio claro.
- **Única excepción:** que el usuario diga explícitamente "sin preguntas" / "procede directo". En ese caso se registra esa instrucción y se avanza.

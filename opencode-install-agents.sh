#!/usr/bin/env bash
# =============================================================================
# opencode-bmad-agent-setup.sh
# Instalador / Reparador de agentes BMAD para OpenCode
# -----------------------------------------------------------------------------
# Uso:
#   chmod +x opencode-bmad-agent-setup.sh
#   ./opencode-bmad-agent-setup.sh [--global] [--dry-run] [--help]
#
# Opciones:
#   --global    Instala agentes en ~/.config/opencode/agents/ (acceso global)
#   --dry-run   Muestra qué haría sin hacer cambios
#   --help      Muestra esta ayuda
# =============================================================================

set -euo pipefail

# ── Colores ──────────────────────────────────────────────────────────────────
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

# ── Flags ─────────────────────────────────────────────────────────────────────
GLOBAL=false
DRY_RUN=false
VERBOSE=false

for arg in "$@"; do
  case $arg in
    --global)   GLOBAL=true ;;
    --dry-run)  DRY_RUN=true ;;
    --verbose)  VERBOSE=true ;;
    --help|-h)
      echo ""
      echo -e "${BOLD}opencode-bmad-agent-setup.sh${NC} — Instala/repara agentes BMAD en OpenCode"
      echo ""
      echo "Uso: ./opencode-bmad-agent-setup.sh [opciones]"
      echo ""
      echo "Opciones:"
      echo "  --global    Instala en ~/.config/opencode/agents/ (para todos los proyectos)"
      echo "  --dry-run   Muestra qué haría sin escribir archivos"
      echo "  --verbose   Muestra detalles de cada operación"
      echo "  --help      Muestra esta ayuda"
      echo ""
      exit 0
      ;;
  esac
done

# ── Banner ────────────────────────────────────────────────────────────────────
echo ""
echo -e "${CYAN}${BOLD}╔══════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}${BOLD}║      OpenCode + BMAD — Instalador de Agentes v1.0       ║${NC}"
echo -e "${CYAN}${BOLD}╚══════════════════════════════════════════════════════════╝${NC}"
echo ""
$DRY_RUN && echo -e "${YELLOW}[DRY-RUN] Modo simulación activo — no se escribirá nada${NC}\n"

# ── Rutas ─────────────────────────────────────────────────────────────────────
PROJECT_DIR="$(pwd)"
BMAD_DIR=""
AGENTS_OUT=""

if $GLOBAL; then
  AGENTS_OUT="$HOME/.config/opencode/agents"
  echo -e "${BLUE}Modo: Global (~/.config/opencode/agents/)${NC}"
else
  AGENTS_OUT="$PROJECT_DIR/.opencode/agents"
  echo -e "${BLUE}Modo: Proyecto local (.opencode/agents/)${NC}"
fi

echo -e "${BLUE}Directorio de trabajo: ${PROJECT_DIR}${NC}\n"

# ── Buscar instalación BMAD ───────────────────────────────────────────────────
echo -e "${BOLD}[1/5] Buscando instalación BMAD...${NC}"

find_bmad_dir() {
  local candidates=("_bmad" "bmad" ".bmad")
  for c in "${candidates[@]}"; do
    if [[ -d "$PROJECT_DIR/$c" ]]; then
      echo "$PROJECT_DIR/$c"
      return
    fi
  done
  echo ""
}

BMAD_DIR="$(find_bmad_dir)"

if [[ -z "$BMAD_DIR" ]]; then
  echo -e "${YELLOW}  ⚠ No se encontró carpeta _bmad en el proyecto actual.${NC}"
  echo -e "${YELLOW}    ¿Instalaste BMAD en otro directorio?${NC}"
  read -rp "  → Ingresa la ruta a la carpeta _bmad (o Enter para omitir): " CUSTOM_PATH
  if [[ -n "$CUSTOM_PATH" && -d "$CUSTOM_PATH" ]]; then
    BMAD_DIR="$CUSTOM_PATH"
  else
    echo -e "${YELLOW}  Omitiendo búsqueda de BMAD. Solo se crearán agentes manuales.${NC}"
  fi
else
  echo -e "${GREEN}  ✓ BMAD encontrado en: ${BMAD_DIR}${NC}"
fi

# ── Verificar OpenCode ────────────────────────────────────────────────────────
echo -e "\n${BOLD}[2/5] Verificando OpenCode...${NC}"

if command -v opencode &>/dev/null; then
  OC_VERSION="$(opencode --version 2>/dev/null || echo 'desconocida')"
  echo -e "${GREEN}  ✓ OpenCode instalado (versión: ${OC_VERSION})${NC}"
else
  echo -e "${YELLOW}  ⚠ OpenCode no encontrado en PATH.${NC}"
  echo -e "${YELLOW}    Instálalo con: npm i -g opencode-ai@latest${NC}"
fi

# ── Crear estructura de carpetas ──────────────────────────────────────────────
echo -e "\n${BOLD}[3/5] Preparando directorios...${NC}"

create_dir() {
  local dir="$1"
  if $DRY_RUN; then
    echo -e "${YELLOW}  [dry] mkdir -p ${dir}${NC}"
  else
    mkdir -p "$dir"
    echo -e "${GREEN}  ✓ ${dir}${NC}"
  fi
}

create_dir "$AGENTS_OUT"
create_dir "${AGENTS_OUT%agents}skills" 2>/dev/null || true
LOCAL_OC="$PROJECT_DIR/.opencode"
create_dir "$LOCAL_OC/agents"
create_dir "$LOCAL_OC/skills"
create_dir "$LOCAL_OC/commands"

# ── Función para escribir agente ──────────────────────────────────────────────
write_agent() {
  local name="$1"
  local description="$2"
  local model="$3"
  local mode="$4"
  local temperature="$5"
  local system_prompt="$6"
  local file="$AGENTS_OUT/${name}.md"

  local content="---
description: \"${description}\"
model: ${model}
mode: ${mode}
temperature: ${temperature}
tools:
  write: true
  edit: true
  bash: true
  webfetch: true
---

${system_prompt}
"

  if $DRY_RUN; then
    echo -e "${YELLOW}  [dry] Crearía: ${file}${NC}"
    $VERBOSE && echo -e "--- CONTENIDO ---\n${content}\n---"
  else
    echo "$content" > "$file"
    echo -e "${GREEN}  ✓ Creado: ${name}.md${NC}"
  fi
}

# ── Convertir agentes BMAD existentes ─────────────────────────────────────────
echo -e "\n${BOLD}[4/5] Convirtiendo agentes BMAD → OpenCode...${NC}"

CONVERTED=0
SKIPPED=0

convert_bmad_skill_to_agent() {
  local skill_dir="$1"
  local skill_name="$(basename "$skill_dir")"
  local skill_md="$skill_dir/SKILL.md"

  [[ ! -f "$skill_md" ]] && return

  # Extraer descripción del SKILL.md (primera línea no vacía después de #)
  local desc
  desc="$(grep -m1 '^#' "$skill_md" 2>/dev/null | sed 's/^#* *//' || echo "$skill_name")"

  # Leer el contenido del skill
  local content
  content="$(cat "$skill_md")"

  # Determinar modo según nombre del agente
  local mode="subagent"
  local temp="0.2"
  case "$skill_name" in
    *plan*|*architect*|*analyst*) mode="subagent"; temp="0.3" ;;
    *dev*|*coder*)                mode="subagent"; temp="0.2" ;;
    *pm*|*product*)               mode="subagent"; temp="0.4" ;;
    *qa*|*test*)                  mode="subagent"; temp="0.1" ;;
    *ux*|*design*)                mode="subagent"; temp="0.5" ;;
    *sm*|*scrum*)                 mode="subagent"; temp="0.3" ;;
    *writer*|*docs*)              mode="subagent"; temp="0.6" ;;
  esac

  # Nombre limpio para el archivo de agente
  local agent_name
  agent_name="$(echo "$skill_name" | sed 's/^bmad-agent-//' | sed 's/^bmad-//')"

  write_agent \
    "$agent_name" \
    "$desc" \
    "anthropic/claude-sonnet-4-5" \
    "$mode" \
    "$temp" \
    "$content"

  ((CONVERTED++)) || true
}

# Buscar skills BMAD en .opencode/skills/
OC_SKILLS_DIR="$PROJECT_DIR/.opencode/skills"
if [[ -d "$OC_SKILLS_DIR" ]]; then
  echo -e "  Escaneando: ${OC_SKILLS_DIR}"
  while IFS= read -r -d '' skill_dir; do
    convert_bmad_skill_to_agent "$skill_dir"
  done < <(find "$OC_SKILLS_DIR" -maxdepth 1 -mindepth 1 -type d -print0 2>/dev/null)
fi

# Buscar agentes en la carpeta _bmad directamente
if [[ -n "$BMAD_DIR" ]]; then
  echo -e "  Escaneando: ${BMAD_DIR}"
  while IFS= read -r -d '' agent_file; do
    agent_name="$(basename "$agent_file" .md | sed 's/^bmad-agent-//' | sed 's/^bmad-//')"
    desc="$(grep -m1 '^#' "$agent_file" 2>/dev/null | sed 's/^#* *//' || echo "$agent_name")"
    content="$(cat "$agent_file")"

    local_file="$AGENTS_OUT/${agent_name}.md"
    if $DRY_RUN; then
      echo -e "${YELLOW}  [dry] Copiaría agente BMAD: ${agent_name}.md${NC}"
    else
      {
        echo "---"
        echo "description: \"${desc}\""
        echo "model: anthropic/claude-sonnet-4-5"
        echo "mode: subagent"
        echo "temperature: 0.2"
        echo "tools:"
        echo "  write: true"
        echo "  edit: true"
        echo "  bash: true"
        echo "  webfetch: true"
        echo "---"
        echo ""
        echo "$content"
      } > "$local_file"
      echo -e "${GREEN}  ✓ Agente BMAD: ${agent_name}.md${NC}"
      ((CONVERTED++)) || true
    fi
  done < <(find "$BMAD_DIR" -name "*.md" -path "*/agents/*" -print0 2>/dev/null)
fi

# ── Agentes estándar de OpenCode si no hay BMAD ───────────────────────────────
if [[ $CONVERTED -eq 0 ]]; then
  echo -e "${YELLOW}  No se encontraron agentes BMAD. Creando agentes de ejemplo...${NC}"

  write_agent "reviewer" \
    "Code reviewer — analiza código sin modificarlo" \
    "anthropic/claude-sonnet-4-5" \
    "subagent" \
    "0.1" \
    "Eres un revisor de código experto. Analiza el código que te comparten y proporciona retroalimentación detallada sobre:
- Posibles bugs o errores lógicos
- Oportunidades de mejora de rendimiento
- Problemas de seguridad
- Violaciones de buenas prácticas
- Deuda técnica

NO modifiques archivos directamente. Solo proporciona sugerencias y explicaciones."

  write_agent "architect" \
    "Software architect — diseño de sistemas y arquitectura" \
    "anthropic/claude-sonnet-4-5" \
    "subagent" \
    "0.3" \
    "Eres un arquitecto de software senior. Tu rol es:
- Diseñar arquitecturas de sistemas escalables
- Evaluar trade-offs entre diferentes enfoques
- Crear diagramas y documentación de arquitectura
- Recomendar patrones de diseño apropiados
- Planificar migraciones y refactorizaciones

Piensa en términos de mantenibilidad, escalabilidad y simplicidad."

  write_agent "debugger" \
    "Debug specialist — diagnóstico y resolución de errores" \
    "anthropic/claude-sonnet-4-5" \
    "subagent" \
    "0.1" \
    "Eres un especialista en debugging. Cuando se te presenta un error:
1. Analiza el stack trace o mensaje de error
2. Identifica la causa raíz probable
3. Propone hipótesis ordenadas por probabilidad
4. Sugiere pasos de diagnóstico
5. Propone la solución más limpia

Eres metódico, preciso y evitas suposiciones sin evidencia."

  write_agent "documenter" \
    "Documentation writer — crea y mejora documentación técnica" \
    "anthropic/claude-sonnet-4-5" \
    "subagent" \
    "0.5" \
    "Eres un escritor técnico especializado en documentación de software. Creas:
- README claros y concisos
- Documentación de APIs
- Guías de instalación y uso
- Comentarios de código
- Changelogs y release notes

Tu documentación es precisa, bien estructurada y fácil de entender tanto para principiantes como para expertos."

  CONVERTED=4
fi

# ── Configurar opencode.json ───────────────────────────────────────────────────
echo -e "\n${BOLD}[5/5] Verificando opencode.json...${NC}"

OC_CONFIG="$PROJECT_DIR/opencode.json"
GLOBAL_OC_CONFIG="$HOME/.config/opencode/opencode.json"

create_or_show_config() {
  local config_path="$1"
  local scope="$2"

  if [[ -f "$config_path" ]]; then
    echo -e "${GREEN}  ✓ ${scope} opencode.json ya existe: ${config_path}${NC}"
    # Verificar si tiene sección de agentes
    if grep -q '"agents"' "$config_path" 2>/dev/null; then
      echo -e "${RED}    ⚠ Clave invalida "agents" detectada — cambiala a "agent" (singular)${NC}"
    else
      echo -e "${GREEN}    → opencode.json valido${NC}"
    fi
  else
    if $DRY_RUN; then
      echo -e "${YELLOW}  [dry] Crearía ${scope} opencode.json en: ${config_path}${NC}"
    else
      mkdir -p "$(dirname "$config_path")"
      cat > "$config_path" << 'EOF'
{
  "$schema": "https://opencode.ai/config.json",
  "model": "anthropic/claude-sonnet-4-5",
  "permission": {
    "bash": "ask",
    "edit": "ask",
    "write": "ask"
  }
}
EOF
      echo -e "${GREEN}  ✓ Creado ${scope} opencode.json: ${config_path}${NC}"
    fi
  fi
}

create_or_show_config "$OC_CONFIG" "Proyecto"
$GLOBAL && create_or_show_config "$GLOBAL_OC_CONFIG" "Global"

# ── Resumen ───────────────────────────────────────────────────────────────────
echo ""
echo -e "${CYAN}${BOLD}╔══════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}${BOLD}║                    RESUMEN FINAL                        ║${NC}"
echo -e "${CYAN}${BOLD}╚══════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${GREEN}  ✓ Agentes creados/convertidos: ${CONVERTED}${NC}"
echo -e "${BLUE}  📁 Ubicación agentes:          ${AGENTS_OUT}${NC}"
echo ""

if [[ -d "$AGENTS_OUT" ]] && ! $DRY_RUN; then
  echo -e "${BOLD}  Agentes disponibles:${NC}"
  while IFS= read -r f; do
    agent_name="$(basename "$f" .md)"
    agent_desc="$(grep -m1 '^description:' "$f" 2>/dev/null | sed 's/description: *"//' | sed 's/"$//' || echo '')"
    echo -e "    ${GREEN}@${agent_name}${NC}  →  ${agent_desc}"
  done < <(find "$AGENTS_OUT" -name "*.md" -maxdepth 1 2>/dev/null | sort)
fi

echo ""
echo -e "${BOLD}  ¿Cómo usar los agentes en OpenCode?${NC}"
echo -e "  ${CYAN}Tab${NC}           → Cambia entre agentes primarios (build/plan)"
echo -e "  ${CYAN}@nombre${NC}       → Invoca un subagente específico"
echo -e "  ${CYAN}@reviewer${NC}     → Ejemplo: activa el revisor de código"
echo ""
echo -e "${BOLD}  Si los agentes BMAD no aparecen con @:${NC}"
echo -e "  1. Asegúrate que los archivos están en ${CYAN}.opencode/agents/${NC}"
echo -e "  2. Reinicia OpenCode completamente (cierra y vuelve a abrir)"
echo -e "  3. Ejecuta ${CYAN}opencode /init${NC} en tu proyecto"
echo ""
echo -e "${YELLOW}  ⚡ Tip: Si instalaste BMAD para Claude Code, también puedes usar:${NC}"
echo -e "  ${CYAN}npx bmad-opencode-converter --source ./_bmad --output ./ --target opencode${NC}"
echo ""

$DRY_RUN && echo -e "${YELLOW}  [Modo dry-run] No se realizó ningún cambio real.${NC}\n"

echo -e "${GREEN}${BOLD}  ✅ ¡Listo! Abre OpenCode en este directorio y prueba con @nombre${NC}"
echo ""

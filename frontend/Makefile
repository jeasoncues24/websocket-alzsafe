.DEFAULT_GOAL := help

APP_NAME  := wsapi-frontend
PORT      ?= $(shell grep -m1 '^PORT=' .env 2>/dev/null | cut -d= -f2)
ifeq ($(strip $(PORT)),)
PORT      := 3001
endif

.PHONY: help install build start restart stop logs dev

help: ## Muestra esta ayuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

install: ## Instala dependencias (npm ci)
	npm ci

build: ## Compila el frontend (next build)
	npm run build

start: ## Inicia wsapi-frontend con PM2 en puerto $(PORT) (idempotente)
	@if ! command -v pm2 &>/dev/null; then \
	  echo "PM2 no encontrado. Instalando globalmente con npm..."; \
	  npm install -g pm2; \
	fi
	@if pm2 describe $(APP_NAME) &>/dev/null 2>&1; then \
	  echo "$(APP_NAME) ya está registrado en PM2. No se crea duplicado."; \
	  echo "Para reiniciar usa: make restart"; \
	  pm2 status $(APP_NAME); \
	else \
	  echo "Iniciando $(APP_NAME) con PM2 en puerto $(PORT)..."; \
	  pm2 start npm --name $(APP_NAME) --cwd "$(CURDIR)" -- start -- -p $(PORT); \
	  echo "$(APP_NAME) activo en :$(PORT)"; \
	fi

restart: ## Reinicia wsapi-frontend en PM2
	@pm2 describe $(APP_NAME) > /dev/null 2>&1 || \
	  (echo "ERROR: $(APP_NAME) no está registrado en PM2. Ejecuta 'make start' primero." && exit 1)
	pm2 restart $(APP_NAME)

stop: ## Detiene wsapi-frontend en PM2
	pm2 stop $(APP_NAME)

logs: ## Muestra los logs de wsapi-frontend en PM2 en tiempo real
	pm2 logs $(APP_NAME)

dev: ## Inicia el servidor de desarrollo en puerto $(PORT)
	npm run dev -- --port $(PORT)

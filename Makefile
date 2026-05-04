.DEFAULT_GOAL := help

.PHONY: help build build-clean start restart stop logs test

help: ## Muestra esta ayuda
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Compila wsapi con caché Docker (requiere backend/.env)
	./scripts/build-backend.sh

build-clean: ## Compila wsapi sin caché Docker (rebuild completo)
	./scripts/build-backend.sh --no-cache

start: ## Inicia wsapi con PM2 en el host (idempotente)
	./scripts/start-backend.sh

restart: ## Reinicia wsapi en PM2 (requiere que ya esté registrado con 'make start')
	@pm2 describe wsapi > /dev/null 2>&1 || \
	  (echo "ERROR: wsapi no está registrado en PM2. Ejecuta 'make start' primero." && exit 1)
	pm2 restart wsapi

stop: ## Detiene wsapi en PM2
	pm2 stop wsapi

logs: ## Muestra los logs de wsapi en PM2 en tiempo real
	pm2 logs wsapi

test: ## Ejecuta los tests del backend Go
	cd backend && go test ./...

# Producción WSAPI

## Flujo oficial vigente

El despliegue actual definido por el último epic es este:

- **Docker Compose** se usa solo para **compilar** el backend.
- El binario se exporta a `dist/wsapi`.
- **PM2 en el host** ejecuta el backend en producción.
- **La base de datos no vive en este compose**.

> En otras palabras: hoy **no** corresponde hacer `docker compose up -d --build` para levantar todo el stack.

## Pasos rápidos

```bash
cp backend/.env.copy backend/.env
# editar variables reales de producción
make build
make start
pm2 save
pm2 startup
```

> Si `wsapi` ya estaba registrado en PM2 pero quedó detenido, usa `make restart` en lugar de `make start`.

## Variables críticas

Asegúrate de definir al menos:

- `APP_ENV=production`
- `APP_PORT`
- `DB_HOST`
- `DB_PORT`
- `DB_NAME`
- `DB_USER`
- `DB_PASS`
- `JWT_SECRET`

Recuerda que `backend/.env.copy` es solo una base y que `JWT_SECRET` no debe quedar vacío en producción.

## Validación

```bash
make logs
pm2 status
APP_PORT_VALUE="$(grep '^APP_PORT=' backend/.env | cut -d= -f2)"
curl "http://127.0.0.1:${APP_PORT_VALUE}/health"
```

## Actualización

```bash
git pull
make build
make restart
```

## Requisitos importantes

- Docker y Docker Compose
- Node.js y npm
- `libsqlite3-0` instalada en el host
- Puerto `APP_PORT` libre en el host
- Permisos para instalar PM2 globalmente o PM2 ya preinstalado

## Más detalle

- `../README.md`
- `../docs/deploy-backend.md`

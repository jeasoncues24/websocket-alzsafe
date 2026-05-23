# WSAPI

Guía rápida para desplegar **wsapi en producción** según el **último epic implementado**.

## Resumen ejecutivo

El flujo actual de despliegue **no levanta el backend dentro de Docker Compose**.

El modelo vigente es este:

- **Docker** se usa **solo para compilar** el binario `wsapi`.
- El binario generado queda en `dist/wsapi`.
- **PM2 en el host** ejecuta y supervisa el proceso en producción.
- **MySQL/MariaDB vive fuera de Docker Compose** y el backend se conecta por variables de entorno.
- El archivo operativo obligatorio es **`backend/.env`**.

Si quieres una sola idea clave, es esta:

```bash
cp backend/.env.copy backend/.env
# editar backend/.env con valores reales de producción
make build
make start
pm2 save
pm2 startup
```

---

## 1. Qué dejó listo el último epic

El último epic dejó definido este contrato de deploy:

1. `docker compose` compila el backend y exporta `dist/wsapi`.
2. `make build` y `make build-clean` simplifican el proceso.
3. `make start` registra `wsapi` en PM2 sin duplicarlo.
4. `APP_PORT` debe existir; si falta, el backend falla rápido con mensaje claro.
5. La base de datos **no** se crea desde el compose actual.
6. El flujo documentado cubre **backend productivo**; frontend y proxy se resuelven aparte.

---

## 2. Prerrequisitos del servidor

Antes de desplegar, asegúrate de tener en el servidor:

- Docker y Docker Compose.
- Node.js y npm.
- MariaDB/MySQL accesible desde el host donde correrá `wsapi`.
- `libsqlite3-0` instalada en el host.
- Acceso al puerto que usarás en `APP_PORT`.
- Permisos para instalar PM2 globalmente con npm, o PM2 ya preinstalado en el servidor.

Instalación de la dependencia nativa de SQLite en Debian/Ubuntu:

```bash
sudo apt-get update
sudo apt-get install -y libsqlite3-0
```

> `wsapi` usa CGO + SQLite para sesiones de WhatsApp. Si esta librería no existe en el host, el binario puede compilar pero fallar al ejecutar.

---

## 3. Variables de entorno importantes

El backend lee variables desde `backend/.env` porque PM2 arranca el proceso con `--cwd backend/`.

### Variables mínimas que deberías revisar sí o sí

```env
APP_ENV=production
APP_PORT=8083

DB_HOST=<host-real-de-tu-db>
DB_PORT=3306
DB_NAME=wsapi
DB_USER=<usuario-real>
DB_PASS=<password-real>

JWT_SECRET=CAMBIA_ESTE_VALOR_EN_PRODUCCION
JWT_ISSUER=wsapi
LOG_LEVEL=info
```

### Variables adicionales que el backend también soporta

- `WHATSAPP_SQLITE_DIR`
- `WHATSAPP_BOOTSTRAP_ENABLED`
- `WHATSAPP_BOOTSTRAP_MAX_CONCURRENCY`
- `WHATSAPP_BOOTSTRAP_TIMEOUT_SEC`
- `WHATSAPP_DEBUG_LOG_DIR`
- `WHATSAPP_DEBUG_LOG_PER_ACCOUNT`
- `WHATSAPP_DEBUG_LOG_LEVEL`
- `WHATSAPP_CONSOLE_LOG_LEVEL`

### Importante

- No dejes `JWT_SECRET` vacío: el sistema usaría un secret por defecto **inseguro para producción**.
- `backend/.env.copy` es solo una base: complétalo manualmente antes de usarlo en producción.
- `APP_PORT` es obligatorio.
- La base de datos debe existir y aceptar conexiones desde este host.
- Si tu DB no corre en la misma máquina, cambia `DB_HOST` por la dirección real; el valor mostrado arriba es solo un ejemplo.

---

## 4. Deploy paso a paso en producción

### Paso 1: clonar el proyecto

```bash
git clone <URL_DEL_REPO>
cd wsapi
```

### Paso 2: preparar entorno

```bash
cp backend/.env.copy backend/.env
nano backend/.env
```

Checklist mínimo antes de seguir:

- `APP_ENV=production`
- `APP_PORT` definido
- credenciales reales de DB
- `JWT_SECRET` real
- rutas de WhatsApp/logs revisadas si aplica

### Paso 3: compilar el binario

Build normal:

```bash
make build
```

Build limpio, útil si cambió Dockerfile, dependencias o sospechas caché rota:

```bash
make build-clean
```

Esto deja el binario en:

```bash
dist/wsapi
```

### Paso 4: iniciar el servicio con PM2

```bash
make start
```

Qué hace este paso:

- instala PM2 si no existe,
- registra `wsapi` si aún no está registrado,
- evita duplicar el proceso si ya existe,
- ejecuta el binario usando `backend/` como directorio de trabajo.

> Si `wsapi` ya estaba registrado en PM2 pero quedó detenido, `make start` no lo levantará de nuevo. En ese caso usa `make restart`.

### Paso 5: persistir PM2 tras reinicios del servidor

```bash
pm2 save
pm2 startup
```

Sigue el comando extra que PM2 te muestre en pantalla si lo solicita. En muchos servidores `pm2 startup` requiere privilegios elevados.

---

## 5. Verificación después del deploy

### Ver logs

```bash
make logs
```

### Ver estado del proceso

```bash
pm2 status
```

### Probar que el backend respondió

Obtén el puerto real desde `backend/.env` y prueba `/health`:

```bash
APP_PORT_VALUE="$(grep '^APP_PORT=' backend/.env | cut -d= -f2)"
curl "http://127.0.0.1:${APP_PORT_VALUE}/health"
```

También puedes revisar que PM2 mantenga `wsapi` en estado `online`.

---

## 6. Cómo actualizar en producción

Cuando subas una nueva versión:

```bash
git pull
make build
make restart
APP_PORT_VALUE="$(grep '^APP_PORT=' backend/.env | cut -d= -f2)"
curl "http://127.0.0.1:${APP_PORT_VALUE}/health"
```

Si necesitas recompilar sin caché:

```bash
git pull
make build-clean
make restart
APP_PORT_VALUE="$(grep '^APP_PORT=' backend/.env | cut -d= -f2)"
curl "http://127.0.0.1:${APP_PORT_VALUE}/health"
```

Si la nueva versión falla y necesitas volver atrás, vuelve al commit/tag anterior, recompila y reinicia:

```bash
git checkout <commit-o-tag-estable>
make build-clean
make restart
```

---

## 7. Comandos operativos útiles

```bash
make build         # compila con caché Docker
make build-clean   # compila sin caché
make start         # registra/inicia wsapi en PM2
make restart       # reinicia wsapi en PM2
make stop          # detiene wsapi
make logs          # logs en tiempo real
make test          # go test ./... del backend
```

---

## 8. Migraciones de base de datos

Si una release incluye migraciones, ejecútalas desde `backend/` para que el binario lea `backend/.env` correctamente:

```bash
cd backend
../dist/wsapi migrate status
../dist/wsapi migrate up
```

Antes de correr `migrate up`, toma backup de la base de datos o confirma que existe un respaldo reciente y valida que la ventana de mantenimiento sea la correcta.

Si no estás seguro de si la release trae migraciones, revísalo antes del restart.

---

## 9. Problemas comunes

| Problema | Causa probable | Qué hacer |
|---|---|---|
| `ERROR: Falta backend/.env` | No existe el archivo operativo | Copia `backend/.env.copy` a `backend/.env` y configúralo |
| `APP_PORT not configured` | Falta `APP_PORT` en `backend/.env` | Define el puerto y reinicia |
| PM2 no inicia `wsapi` | No existe `dist/wsapi` | Ejecuta `make build` primero |
| El backend no conecta a DB | Credenciales/host/puerto incorrectos | Revisa `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASS` |
| El binario falla al ejecutar por SQLite | Falta `libsqlite3-0` en el host | Instálala y vuelve a iniciar |
| `make restart` falla | El proceso aún no está registrado | Ejecuta `make start` primero |
| El backend no levanta y PM2 muestra error de bind | El puerto de `APP_PORT` ya está ocupado en el host | Cambia `APP_PORT` o libera el puerto antes de reiniciar |

---

## 10. Qué queda fuera de este flujo

Este deploy de producción **no** hace lo siguiente:

- no levanta MariaDB dentro de Docker Compose,
- no publica el backend como contenedor de larga vida,
- no define el runtime del frontend Next.js,
- no configura por sí solo Nginx, SSL o dominio público.

Si vas a exponer el sistema a internet, lo recomendable es poner un reverse proxy delante del `APP_PORT` configurado y definir allí TLS/HTTPS.

---

## 11. Frontend

El epic actual se centró en el **backend**. Si también vas a desplegar el frontend, al menos debes apuntarlo al backend correcto:

```env
NEXT_PUBLIC_API_URL=https://tu-dominio-o-api-publica
NEXT_INTERNAL_API_URL=http://127.0.0.1:APP_PORT
```

Referencia rápida: `frontend/README.md`

---

## 12. Documentación complementaria

- `docs/deploy-backend.md` — detalle operativo del backend
- `docs/integracion-b2b.md` — capacidades B2B públicas (health endpoint y guía base)
- `docs/webhooks-integracion.md` — guía detallada de registro, firma, payloads y retries de webhooks
- `docker/production.md` — resumen corto del flujo oficial vigente
- `frontend/README.md` — notas del frontend

Si quieres, en el siguiente paso te puedo dejar también un **checklist de deploy productivo** o una **versión orientada a Ubuntu + Nginx + dominio**.
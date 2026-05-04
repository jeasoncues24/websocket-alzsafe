# Despliegue del backend wsapi

El flujo de build usa Docker solo para compilar el binario. El runtime se ejecuta en el host con PM2.
La base de datos corre en el host; Docker no publica el backend ni expone puertos.

## Prerequisitos

- Docker y Docker Compose instalados.
- Node.js y npm instalados (para PM2).
- MariaDB/MySQL corriendo en el host.
- Archivo `backend/.env` configurado (ver `backend/.env.copy` como referencia).
- **`libsqlite3-0` instalada en el host** (el binario usa CGO con SQLite para sesiones WhatsApp):
  ```bash
  # Debian/Ubuntu
  sudo apt-get install -y libsqlite3-0
  ```

## 1. Configurar entorno

```bash
cp backend/.env.copy backend/.env
# Editar backend/.env con los valores reales (DB, APP_PORT, JWT_SECRET, etc.)
```

El archivo `backend/.env` es obligatorio. El script de build falla con mensaje claro si no existe.

## 2. Compilar el binario

### Build normal (con caché Docker — más rápido)

```bash
./scripts/build-backend.sh
```

### Build limpio (sin caché — útil cuando cambian dependencias o hay builds corruptos)

```bash
./scripts/build-backend.sh --no-cache
```

El binario queda en `./dist/wsapi`. El build limpia el binario anterior antes de copiar el nuevo.

## 3. Iniciar el backend con PM2

```bash
./scripts/start-backend.sh
```

El script es idempotente:
- Si PM2 no está instalado, lo instala via `npm install -g pm2`.
- Si `wsapi` ya existe en PM2, no crea un duplicado y muestra el estado actual.
- El binario lee las variables de entorno del host (archivo `.env` del directorio de trabajo o entorno exportado).

## 4. Comandos PM2 útiles

```bash
pm2 logs wsapi        # logs en tiempo real
pm2 status            # estado de todos los procesos
pm2 restart wsapi     # reiniciar (p.ej. tras un nuevo build)
pm2 stop wsapi        # detener
pm2 delete wsapi      # eliminar de PM2 completamente
pm2 save              # persistir lista de procesos tras reboot
pm2 startup           # configurar arranque automático en el sistema
```

## 5. Ciclo de actualización

```bash
./scripts/build-backend.sh     # compila nuevo binario
pm2 restart wsapi              # aplica el binario actualizado
```

## Notas de arquitectura

- **Docker = compilar únicamente.** El Dockerfile no expone puertos ni define CMD de runtime.
- **Host = ejecutar.** PM2 gestiona el proceso `wsapi` directamente en el host.
- **Base de datos en el host.** `DB_HOST` en `backend/.env` debe apuntar al host local (p.ej. `127.0.0.1`). Docker no levanta ni expone la base de datos en este flujo.
- **Puerto.** `APP_PORT` en `backend/.env` define el puerto del servidor; no hay mapeo Docker.

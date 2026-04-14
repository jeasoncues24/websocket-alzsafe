# --- ETAPA DE CONSTRUCCIÓN ---
FROM golang:1.22-alpine AS builder

# Instalar dependencias necesarias para CGO (necesario para SQLite en Go)
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copiamos solo el go.mod primero. 
# Usamos un comodín para go.sum para que no falle si aún no existe.
COPY go.mod go.sum* ./

# Descargar dependencias
RUN go mod download

# Copiar el resto del código
COPY . .

# Construir el binario
# Nota: CGO_ENABLED=1 es necesario si usas librerías como Whatsmeow que usan SQLite
RUN CGO_ENABLED=1 GOOS=linux go build -o whatsapp-api

# --- ETAPA DE EJECUCIÓN ---
FROM alpine:3.19

WORKDIR /app

# ca-certificates es vital para conexiones TLS con WhatsApp
# sqlite-libs es necesario si tu app usa bases de datos locales para la sesión
RUN apk add --no-cache ca-certificates sqlite-libs tzdata

# Crear carpeta de sesiones por si no existe
RUN mkdir ./sessions

# Copiar el binario y asegurar permisos
COPY --from=builder /app/whatsapp-api .

EXPOSE 8080

CMD ["./whatsapp-api"]
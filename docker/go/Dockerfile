# syntax=docker/dockerfile:1

# --- BUILD STAGE ---
# Este Dockerfile solo compila y exporta el binario wsapi.
# No levanta el runtime: ejecucion, puertos y BD son responsabilidad del host (PM2).
FROM golang:1.26-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    mkdir -p /out && \
    CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/wsapi .

# --- EXPORT STAGE ---
# Exporta únicamente el binario compilado al host sin desempaquetar capas pesadas en el daemon.
FROM scratch AS export
COPY --from=builder /out/wsapi /wsapi

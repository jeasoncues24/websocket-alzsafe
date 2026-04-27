# --- BUILD STAGE ---
FROM golang:1.25-bookworm AS builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/wsapi .

# --- RUNTIME STAGE ---
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libsqlite3-0 \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
RUN mkdir -p /app/sessions /app/debug_log

COPY --from=builder /out/wsapi /usr/local/bin/wsapi

EXPOSE 8080

CMD ["wsapi"]
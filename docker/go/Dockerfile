# --- BUILD STAGE ---
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/wsapi .

# --- RUNTIME STAGE ---
FROM alpine:3.20

RUN apk add --no-cache ca-certificates sqlite-libs tzdata

WORKDIR /app
RUN mkdir -p /app/sessions /app/debug_log

COPY --from=builder /out/wsapi /usr/local/bin/wsapi

EXPOSE 8080

CMD ["wsapi"]

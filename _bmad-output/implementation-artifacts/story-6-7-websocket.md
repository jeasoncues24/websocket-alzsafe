# Story 6.7: WebSocket /v1/ws + eventos

Status: in-progress

## Story

As a empresa user,
I want to connect to a WebSocket endpoint for real-time updates,
so that I receive instant notifications about QR codes, connection status, and message delivery.

## Acceptance Criteria

1. [AC: ws-v1] GET /v1/ws upgrades to WebSocket with empresa JWT authentication
2. [AC: ws-auth] JWT can be passed via header or query param `?token=`
3. [AC: ws-qr] Event "qr" sent when QR code is available
4. [AC: ws-connected] Event "connected" sent when phone connects
5. [AC: ws-disconnected] Event "disconnected" sent when phone disconnects
6. [AC: ws-message-status] Event "message_status" sent for message state changes

## Tasks

- [ ] Task 1: Create v1_ws.go handler
  - [ ] Task 1.1: HandleV1WS - WebSocket upgrade with JWT validation
  - [ ] Task 1.2: Extract token from header or query param
- [ ] Task 2: Register /v1/ws in router.go
- [ ] Task 3: Wire event emitters (qr, connected, disconnected, message_status)

## Dev Notes

- Use existing `websocket.Accept` from github.com/coder/websocket
- Middleware: empresaAuthMiddleware parses JWT from context
- Events follow epic-6 spec format
- Client.subscribe to phone events via hub or direct callback
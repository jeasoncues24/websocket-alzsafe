# Epic 7: Frontend de Integración para Empresas

## Overview

Implementar Panel de Auto-gestión para empresas (diferente del Panel Admin):
- QR connection flow con token de 5 minutos
- Gestión de números WhatsApp
- Envío de mensajes
- Monitoreo en tiempo real vía WebSocket
- UI para clientes finales

> **Diferencia con Epic 5:** 
> - Epic 5 = Panel Admin (nosotros)
> - Epic 7 = Panel Empresa (nuestros clientes)

---

## Requisitos Funcionales

### FR-01 Login de Empresa
- Login con token JWT (no usuario/password)
-URL + token = acceso directo
- No requiere password

### FR-02 Dashboard Empresa
- Ver sus números registrados
- Estado de cada número (active/qr_pending/disconnected)
- Métricas: mensajes enviados,成功率

### FR-03 Conexión QR con Expiración
- Botón "Conectar WhatsApp" genera link temporal
- Link dura 5 minutos (300 segundos)
- QR se muestra en la UI
- Contador regresivo visible
- Notificación cuando expira

### FR-04 Gestión de Teléfonos
- Listar todos los números de la empresa
- Ver estado individual
- Conectar/desconectar desde UI
- Historial de conexiones

### FR-05 Envío de Mensajes
- Formulario: número destino + mensaje
- Preview antes de enviar
- Indicador de estado (enviado/entregado/ledo)
- Historial de mensajes enviados

### FR-06 Difusiones
- Crear difusión masiva
- Lista de destinatarios (manual/CSV)
- Progress en tiempo real
- Resultados por destinatario

### FR-07 WebSocket en Tiempo Real
- Estado de conexión se actualiza solo
- Notificaciones de mensajes entrantes
- Alertas de desconexión

---

## UI/UX Requirements

### Login Page
```
┌─────────────────────────────────────┐
│           Mi Empresa                │
│                                     │
│  [Panel de WhatsApp]                 │
│                                     │
│  Tu token está activo              │
│                                     │
└─────────────────────────────────────┘
```

### Dashboard Empresa
```
┌─────────────────────────────────────┐
│  Mi Empresa SAC          [Token]     │
├─────────────────────────────────────┤
│  ┌─────────┐  ┌─────────┐  ┌───────┐│
│  │Activos │  │Mensajes│  │Éxito │  │
│  │   3    │  │  150  │  │ 95%  │  │
│  └─────────┘  └─────────┘  └───────┘│
│                                     │
│  Números Conectados                │
│  ┌─────────────────────────────┐   │
│  │ +519999999999  ● ACTIVO   ⟲ │   │
│  │ +519888888888  ○ QR       ⟲ │   │
│  └─────────────────────────────┘   │
└─────────────────────────────────────┘
```

### Modal de Conexión QR
```
┌─────────────────────────────────────┐
│  Conectar WhatsApp                  │
│                                     │
│  Este código expira en:            │
│  ████████░░  2:45               │
│                                     │
│     ████████████████            │
│     █                        █    │
│     █    [QR CODE]          █    │
│     █                        █    │
│     ████████████████            │
│                                     │
│  Abre WhatsApp > Dispositivos      │
│ vinculados > Escanea este código  │
│                                     │
│  [Cancelar]                       │
└─────────────────────────────────────┘
```

### Página de Mensajes
```
┌─────────────────────────────────────┐
│  Mensajes                          │
│                                     │
│  [+51 999 999 999] [Mensaje...] →  │
│                                     │
│  ┌─────────────────────────────┐   │
│  │ Para    │ Mensaje    │ Estado │   │
│  │+519xx..│ Hola!    │ ✓     │   │
│  │+519xx..│ Promo..  │ ✓     │   │
│  └─────────────────────────────┘   │
│                                     │
│  [+51 999 999 999] [Escribe...] [→]│
└─────────────────────────────────────┘
```

---

## Componente: QRDisplay

```tsx
// Componente para mostrar QR con countdown

interface QRDisplayProps {
  qrString: string;
  expiresAt: Date;
  onExpire: () => void;
}

function QRDisplay({ qrString, expiresAt, onExpire }: QRDisplayProps) {
  const [timeLeft, setTimeLeft] = useState(300);
  
  useEffect(() => {
    const interval = setInterval(() => {
      const remaining = Math.max(0, expiresAt - Date.now());
      setTimeLeft(remaining / 1000);
      if (remaining <= 0) onExpire();
    }, 1000);
    return () => clearInterval(interval);
  }, [expiresAt]);
  
  // Generar imagen QR de qrString
  return (
    <div>
      <QRCode value={qrString} />
      <Progress value={timeLeft / 300} />
      <span>{formatTime(timeLeft)}</span>
    </div>
  );
}
```

---

## Flujo de Conexión (UI)

### Paso 1: Usuario hace click en "Conectar"
```
POST /v1/telefonos/{id}/connect
→ { status: "qr_pending", expires_in: 300 }
```

### Paso 2: Mostrar QR con countdown
```
GET /v1/telefonos/{id}/qr
→ { qr: "string", expires_at: timestamp }
```

### Paso 3: Escaneo + Activación
- WebSocket: { type: "connected", phone_id: 1 }
- UI actualiza automáticamente

### Paso 4: Timeout
- Si pasan 5 minutos sin escanear:
- Mostrar "Código vencido"
- Botón "Generar nuevo"

---

## WebSocket UI

```tsx
// Hook para WebSocket con empresa

function useEmpresaWS(token: string) {
  const [socket, setSocket] = useState<WebSocket>();
  const [events, setEvents] = useState<Event[]>([]);
  
  useEffect(() => {
    const ws = new WebSocket(
      `wss://api.example.com/v1/ws?token=${token}`
    );
    
    ws.onmessage = (event) => {
      const data = JSON.parse(event.data);
      setEvents(prev => [...prev, data]);
      
      // Tipos de eventos
      switch (data.type) {
        case 'qr':     // Nuevo QR disponible
        case 'connected': // Teléfono conectado
        case 'disconnected': // Teléfono desconectado
        case 'message_status': // Estado de mensaje actualizado
        case 'service_status': // Servicio caído
      }
    };
    
    setSocket(ws);
    return () => ws.close();
  }, [token]);
  
  return { socket, events };
}
```

---

## Endpoints del Frontend

| Ruta | Descripción |
|------|------------|
| `/` | Redirect a `/login` o `/dashboard` |
| `/login` | Input de token JWT |
| `/dashboard` | Dashboard empresa |
| `/telefonos` | Lista de números |
| `/telefonos/[id]` | Detalle de número |
| `/telefonos/[id]/connect` | Modal QR |
| `/mensajes` | Lista de mensajes |
| `/mensajes/nuevo` | Enviar mensaje |
| `/broadcasts` | Lista de difusiones |
| `/broadcasts/nuevo` | Crear difusión |
| `/settings` | Settings de cuenta |

---

## Stories

| ID | Story | Prioridad |
|-----|-------|----------|
| S-7.1 | Login con token JWT | P0 |
| S-7.2 | Dashboard empresa | P0 |
| S-7.3 | QR Display con countdown | P0 |
| S-7.4 | Conexión/desconexión números | P0 |
| S-7.5 | Envío de mensajes | P0 |
| S-7.6 | Difusiones masivas | P1 |
| S-7.7 | WebSocket integration | P0 |
| S-7.8 | Settings cuenta | P1 |

---

## Dependencias

| Dependencia | Versión |
|------------|--------|
| next | ^14.0.0 |
| react | ^18.0.0 |
| @shadcn/ui | latest |
| zustand | ^4.5.0 |
| qrcode | ^1.5.0 |
| lucide-react | ^0.300.0 |

---

## Diferencia Epic 5 vs Epic 7

| Aspecto | Epic 5 (Admin) | Epic 7 (Empresa) |
|--------|---------------|----------------|
| Usuario | Admin interno | Cliente empresa |
| Login | JWT sesión | JWT token |
| URL | `/admin/*` | `/` (self-service) |
| Funciones | Gestión total | Solo sus números |
| QR | Para admin | Para cliente |

---

## Success Criteria

- [ ] Empresa puede login con solo token
- [ ] QR muestra con countdown de 5 min
- [ ] Notificación cuando QR expira
- [ ] WebSocket actualiza estados
- [ ] Empresa puede conectar/desconectar
- [ ] Empresa puede enviar mensajes
- [ ] Aislamiento de datos entre empresas
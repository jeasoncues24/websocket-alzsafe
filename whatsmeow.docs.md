# WhatsMeow - Documentación de Referencia

> Librería Go para WhatsApp Web Multi-Device API
> **Repo**: https://github.com/tulir/whatsmeow
> **Paquete**: `go.mau.fi/whatsmeow`
> **Licencia**: MPL-2.0
> **Versión actual**: v0.0.0-20260414172242

---

## Installation

```bash
go get go.mau.fi/whatsmeow
```

### Dependencias

| Paquete | Versión |
|---------|---------|
| github.com/beeper/argo-go | v1.1.2 |
| github.com/coder/websocket | v1.8.14 |
| github.com/google/uuid | v1.6.0 |
| github.com/rs/zerolog | v1.34.0 |
| go.mau.fi/libsignal | v0.2.1 |
| go.mau.fi/util | v0.9.6 |
| golang.org/x/crypto | v0.48.0 |
| golang.org/x/net | v0.50.0 |

---

## Uso Básico

### Inicialización del Cliente

```go
import (
    "context"
    "log"

    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/store/sqlstore"
    waLog "go.mau.fi/whatsmeow/util/log"
)

func main() {
    // 1. Preparar el store (SQLite en este ejemplo)
    container, err := sqlstore.New("messages.db", "sqlcipher")
    if err != nil {
        panic(err)
    }

    // 2. Obtener el device store
    deviceStore, err := container.GetFirstDevice()
    if err != nil {
        panic(err)
    }

    // 3. Crear el cliente
    clientLog := waLog.Stdout("Client", "DEBUG", true)
    client := whatsmeow.NewClient(deviceStore, clientLog)

    // 4. Agregar event handler
    client.AddEventHandler(func(evt interface{}) {
        switch v := evt.(type) {
        case *events.Connected:
            log.Println("Conectado a WhatsApp!")
        case *events.Message:
            log.Printf("Mensaje recibido: %s", v.Message.GetConversation())
        case *events.QR:
            log.Println("QR Code:", v.Codes)
        case *events.PairSuccess:
            log.Printf("Emparejado exitosamente: %s", v.ID)
        case *events.LoggedOut:
            log.Println("Sesión cerrada")
        }
    })

    // 5. Conectar
    if client.Store.ID == nil {
        // No hay sesión, neceistamos QR
        qrChan, _ := client.GetQRChannel(context.Background())
        err = client.Connect()
        if err != nil {
            panic(err)
        }
        for evt := range qrChan {
            switch evt.Event {
            case "success":
                log.Println("Login exitoso")
            case "code":
                log.Println("Nuevo código QR:", evt.Code)
            case "timeout":
                log.Println("Timeout - regenerando QR")
            }
        }
    } else {
        // Sesión existente - solo conectar
        err = client.Connect()
        if err != nil {
            panic(err)
        }
    }
}
```

---

## Métodos Principales del Cliente

### Conexión

| Método | Descripción |
|--------|-------------|
| `NewClient(deviceStore, log)` | Crea un nuevo cliente |
| `Connect()` | Conecta al servidor de WhatsApp |
| `ConnectContext(ctx)` | Conecta con contexto |
| `Disconnect()` | Desconecta del servidor |
| `IsConnected()` | Verifica si está conectado |
| `IsLoggedIn()` | Verifica si hay sesión activa |
| `ResetConnection()` | Resetea la conexión |
| `WaitForConnection(timeout)` | Espera hasta estar conectado |

### Login / QR

| Método | Descripción |
|--------|-------------|
| `GetQRChannel(ctx)` | Obtiene canal de eventos QR |
| `PairPhone(ctx, phone, showPushNotification, clientType, clientDisplayName)` | Empareja por número |
| `Logout(ctx)` | Cierra sesión |

### Envío de Mensajes

| Método | Descripción |
|--------|-------------|
| `SendMessage(ctx, to, message, extra)` | Envía mensaje de texto |
| `SendFBMessage(ctx, to, message, metadata, extra)` | Envía mensaje de Facebook |
| `SendPeerMessage(ctx, message)` | Envía mensaje peer |
| `Upload(ctx, plaintext, appInfo)` | Sube archivo |
| `Download(ctx, msg)` | Descarga media |

### Grupos

| Método | Descripción |
|--------|-------------|
| `CreateGroup(ctx, req)` | Crea grupo |
| `GetGroupInfo(ctx, jid)` | Obtiene info de grupo |
| `GetJoinedGroups(ctx)` | Lista grupos unidos |
| `LeaveGroup(ctx, jid)` | Sale del grupo |
| `GetGroupInviteLink(ctx, jid, reset)` | Obtiene link de invitación |
| `JoinGroupWithLink(ctx, code)` | Se une por link |

### Contactos / Info

| Método | Descripción |
|--------|-------------|
| `GetUserInfo(ctx, jids)` | Obtiene información de usuario |
| `IsOnWhatsApp(ctx, phones)` | Verifica si está en WhatsApp |
| `GetBusinessProfile(ctx, jid)` | Obtiene perfil de negocio |
| `GetBlocklist(ctx)` | Obtiene lista de bloqueados |

### Presencia

| Método | Descripción |
|--------|-------------|
| `SendPresence(ctx, state)` | Envía presencia |
| `SendChatPresence(ctx, jid, state, media)` | Envía presencia en chat |
| `SubscribePresence(ctx, jid)` | Suscribe a presencia |

### Lectura

| Método | Descripción |
|--------|-------------|
| `MarkRead(ctx, ids, timestamp, chat, sender, receiptTypeExtra)` | Marca como leído |

### Estados

| Método | Descripción |
|--------|-------------|
| `SetStatusMessage(ctx, msg)` | Envía estado |
| `GetPrivacySettings(ctx)` | Obtiene configuración de privacidad |
| `SetPrivacySetting(ctx, name, value)` | Configura privacidad |
| `SetDefaultDisappearingTimer(ctx, timer)` | Configura timer de desaparición |

---

## Eventos (Event Types)

Los eventos se definen en `go.mau.fi/whatsmeow/types/events`.

### Eventos de Conexión

```go
// QR - emite cuando no hay sesión y se necesita escanear
type QR struct {
    Codes []string  // Códigos QR para escanear
}

// PairSuccess - emitido después de escanear QR exitosamente
type PairSuccess struct {
    ID           types.JID  // JID del dispositivo
    LID          types.JID  // LID del dispositivo
    BusinessName string    // Nombre del negocio
    Platform     string   // Plataforma
}

// PairError - error de emparejamiento
type PairError struct {
    Code int
    Text string
}

// QRScannedWithoutMultidevice - QR escaneado sin multidevice activo
type QRScannedWithoutMultidevice struct{}

// Connected - conectado y autenticado
type Connected struct{}

// KeepAliveTimeout - timeout de keepalive
type KeepAliveTimeout struct{}

// KeepAliveRestored - keepalive restaurado
type KeepAliveRestored struct{}

// LoggedOut - sesión cerrada
type LoggedOut struct{}

// StreamReplaced - conexión reemplazada
type StreamReplaced struct{}
```

### Eventos de Mensajes

```go
// Message - mensaje recibido
type Message struct {
    Message     *MessageInfo   // Info del mensaje
    MessageID   string        // ID del mensaje
    From       types.JID     // Remitente
    Sender     types.JID      // Remitente real
    Chat       types.JID      // Chat
    ChatInfo   *ContactInfo  // Info del chat
    IsGroup    bool          // Es grupo
    IsBroadcast bool         // Es broadcast
    NotifyName string      // Nombre de notificación
}

// Reaction - reacción recibida
type Reaction struct{}

// Receipt - receipt (estado de mensaje)
type Receipt struct {
    MessageIDs []string
    From      types.JID
    Timestamp types.Timestamp
}

// PollVoteReceived - vote de encuesta recibido
type PollVoteReceived struct{}
```

### Eventos de Grupos

```go
// GroupJoined - se unió a grupo
type GroupJoined struct {
    JID types.JID
}

// GroupLeft - salió de grupo
type GroupLeft struct {
    JID types.JID
}

// Group-participant - cambio en participantes
type GroupParticipant struct{}
```

---

## Stores (Almacenamiento)

### Store de Dispositivo

El `deviceStore` guarda la información de sesión:

```go
// Métodos del DeviceStore
type DeviceStore interface {
    ID                   *types.JID
    LID                  *types.JID
    Phone                string
    Platform             int
    PushName             string
    EncryptedHistory bool

    Save()
    Delete()

    GetOrCreatePreKey(keyID uint32) (*store.SignedKey, error)
    GetPreKey(keyID uint32) (*store.PreKey, error)
    GetPreKeys() ([]*store.PreKey, error)
}
```

### SQLStore

```go
import "go.mau.fi/whatsmeow/store/sqlstore"

// Crear contenedor de base de datos
container, err := sqlstore.New("messages.db", "sqlcipher")
if err != nil {
    panic(err)
}

// Obtener primer device
deviceStore, err := container.GetFirstDevice()
if err != nil {
    panic(err)
}

// Para múltiples dispositivos:
// devices, _ := container.GetAllDevices()
```

---

## Tipos Principales (types.JID)

```go
// JID representa un identificador de WhatsApp
type JID struct {
    User         string
    Agent       int
    Device      int
    ToString() string
    ADToString() string
}

// Ejemplos
// user@whatsapp.net - usuario normal
// user@lid - usuario LID
// group@g.us - grupo
// broadcast@g.us - broadcast
// 123456789@g.us - canal
```

---

## Flujo de Login QR

```
1. Crear cliente: client = whatsmeow.NewClient(deviceStore, log)
2. Verificar si hay sesión: client.Store.ID != nil
3. Si NO hay sesión:
   a. Obtener QR channel: qrChan = client.GetQRChannel(ctx)
   b. Conectar: client.Connect()
   c. Esperar eventos del canal:
      - "code" → mostrar QR al usuario
      - "success" → emparejamiento exitoso
      - "timeout" → regenerar código
4. Si SÍ hay sesión:
   a. Solo conectar: client.Connect()
```

### Manejo de Eventos QR en Detail

```go
qrChan, err := client.GetQRChannel(context.Background())
if err != nil {
    panic(err)
}
err = client.Connect()
if err != nil {
    panic(err)
}

for evt := range qrChan {
    switch evt.Event {
    case "success":
        // Login exitoso
        log.Println("Login exitoso")
    case "code":
        // Nuevo código QR - actualizar UI
        log.Println("Nuevo código QR recibido")
    case "timeout":
        // El código expiró, esperar siguiente
        log.Println("Código expirado")
    }
}
```

---

## Ejemplo: Enviar Mensaje

```go
func sendMessage(client *whatsmeow.Client, to string, text string) {
    ctx := context.Background()
    
    // Preparar mensaje
    message := &types.Message{
        Conversation: &types.MessageConversation{
            Text: text,
        },
    }
    
    // Enviar
    resp, err := client.SendMessage(ctx, to, message, nil)
    if err != nil {
        log.Printf("Error enviando: %v", err)
    } else {
        log.Printf("Enviado: %s", resp.ID)
    }
}
```

---

## Ejemplo: Receiving Messages

```go
client.AddEventHandler(func(evt interface{}) {
    switch v := evt.(type) {
    case *events.Message:
        // Obtener contenido
        if v.Message.GetConversation() != "" {
            text := v.Message.GetConversation()
            from := v.From.String()
            log.Printf("De %s: %s", from, text)
        }
        
        // Responder
        // go sendMessage(client, from, "¡Hola!")
    }
})
```

---

## Configuración Adicional

### Proxy

```go
client.SetProxy("http://proxy:8080", nil)
```

### ClientPersonalInfo

```go
client.Store.DeviceProps = &waPB.DeviceProps{
    Os:        proto.String("WhatsApp API"),
    Platform:  proto.Int32(waPB.DeviceProps_WEB),
    AppVersion: proto.String("1.0.0"),
}
```

---

## Errores Comunes

| Código | Descripción | Solución |
|--------|-------------|----------|
| 401 | Sesión cerrada | Regenerar QR |
| 515 | Reconectar | Llamar Connect() |
| device_removed | Desvinculado | Regenerar QR |
| temp_ban | Temporal banned | Esperar |

---

## Recursos

- **GitHub**: https://github.com/tulir/whatsmeow
- **Godoc**: https://pkg.go.dev/go.mau.fi/whatsmeow
- **Matrix**: #whatsmeow:maunium.net
- **Discusiones**: https://github.com/tulir/whatsmeow/discussions
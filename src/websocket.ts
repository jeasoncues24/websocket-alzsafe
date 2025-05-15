import WebSocket, { WebSocketServer } from "ws";
import { userExistInService } from "./app/services/user.service";
import { Client } from "whatsapp-web.js";
import {
  initializeNumberSession,
  listUserActiveClientWhatsapp,
} from "./utils/wa-client";
import {
  activateServiceUsuario,
  activateUserModel,
  deactivateUserModel,
} from "./app/models/user.model";

class WebSocketHandler {
  private wss: WebSocketServer;
  private clients: Set<WebSocket> = new Set();

  constructor(server: any) {
    this.wss = new WebSocketServer({ server });

    this.wss.on("connection", (ws) => {
      console.log("Cliente WebSocket conectado");
      this.clients.add(ws);

      ws.on("message", async (message) => {
        try {
          const payload = JSON.parse(message.toString());

          if (payload.event === "init-session") {
            const { ruc_empresa } = payload.data;
            console.log("Iniciando sesión para RUC:", ruc_empresa);
            inicializarSession(ws, ruc_empresa);
          }
        } catch (e) {
          console.error("Error procesando mensaje WebSocket:", e);
        }
      });

      ws.on("close", () => {
        console.log("Cliente WebSocket desconectado");
        this.clients.delete(ws);
      });

      ws.on("error", (error) => {
        console.error("Error en WebSocket:", error);
      });
    });
  }

  // ✅ Método para emitir datos a todos los clientes conectados
  broadcast(eventName: string, data: any) {
    const payload = JSON.stringify({ event: eventName, data });

    this.clients.forEach((client) => {
      if (client.readyState === WebSocket.OPEN) {
        client.send(payload);
        console.warn("Mensaje enviado a cliente WebSocket:", payload);
      }
    });
  }
}

// ✅ Método para manejar la inicialización de sesión
async function inicializarSession(ws: WebSocket, ruc_empresa: string) {
  try {
    const data = await userExistInService(ruc_empresa);

    if (data === null) {
      return;
    }

    const { telefono, nombre_comercial } = data;

    const waClient: Client = initializeNumberSession(telefono, ruc_empresa);
    listUserActiveClientWhatsapp.set(ruc_empresa, waClient);
    waClient.on("ready", async () => {
      await activateUserModel(ruc_empresa);
      await activateServiceUsuario(ruc_empresa, 1);
      console.log(`Cliente ${nombre_comercial} está listo.`);
      listUserActiveClientWhatsapp.set(ruc_empresa, waClient);
    });

    waClient.on("disconnected", async (reason) => {
      await activateServiceUsuario(ruc_empresa, 0);
      await deactivateUserModel(ruc_empresa);
      const messageError = `Cliente ${nombre_comercial} se ha desconectado del servicio.`;
      console.error(messageError);

      const nameEvent = `active-${ruc_empresa}`;
      ws.send(
        payloadMessage(nameEvent, {
          message: messageError,
          isActive: false,
        })
      );
    });
    waClient.on("authenticated", async (session) => {
      await activateUserModel(ruc_empresa);
      const message = `Cliente ${nombre_comercial} está autenticado en el servicio.`;
      console.log(message);
      const nameEvent = `active-${ruc_empresa}`;
      ws.send(
        payloadMessage(nameEvent, {
          message: message,
          isActive: true,
        })
      );
    });
    waClient.on("message", async (msg) => {
      console.log(`message in ${nombre_comercial}` + msg);
    });

    waClient.on("qr", async (qr) => {
      console.log(`QR code for ${nombre_comercial}: ${qr}`);
      const nameEvent = `qr-${ruc_empresa}`;
      ws.send(
        payloadMessage(nameEvent, {
          message:
            "Escanee el código QR, para empezar a enviar utilizar el servicio de WhatsApp en nuestro sistema.",
          qrString: qr,
        })
      );
    });

    await waClient.initialize();
  } catch (error) {
    console.error("Error al inicializar la sesión:", error);
  }
}

function payloadMessage(eventName: string, data: any) {
  const payload = JSON.stringify({ event: eventName, data });
  return payload;
}

export default WebSocketHandler;

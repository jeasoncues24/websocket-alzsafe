import WebSocket, { WebSocketServer } from "ws";
import { userExistInService } from "./app/services/user.service";
import { Client } from "whatsapp-web.js";

import {
  getSessionName,
  initializeNumberSession,
  listUserActiveClientWhatsapp,
} from "./utils/wa-client";
import {
  activateServiceUsuario,
  activateUserModel,
  deactivateUserModel,
} from "./app/models/user.model";
import { WsClienteConnection } from "./interfaces/wscliente.interface";

class WebSocketHandler {
  private wss: WebSocketServer;
  private clients: Set<WsClienteConnection> = new Set();

  constructor(server: any) {
    this.wss = new WebSocketServer({ server });

    this.wss.on("connection", (ws) => {
      console.log("Cliente WebSocket conectado");
      const _info: WsClienteConnection = {
        websocket: ws,
        telefono: "",
        ruc: "",
        client: null,
      };
      this.clients.add(_info);
      ws.on("message", async (message) => {
        try {
          const payload = JSON.parse(message.toString());
          if (payload.event === "init-session") {
            try {
              const { ruc_empresa } = payload.data;
              const { sessionName, telefono } = await getSessionName(
                ruc_empresa
              );

              this.clients.forEach((client) => {
                if (client.websocket === ws) {
                  client.ruc = ruc_empresa;
                  client.telefono = telefono;
                }
              });
              console.log("Initialized Session: ", ruc_empresa);
              const isValidate = await notifyIfActiveWhatsappClient(
                ruc_empresa
              );
              if (isValidate) {
                console.log("Cliente ya esta activo. No se inicializa.");
                ws.send(
                  payloadMessage("active-" + ruc_empresa, {
                    message: "Cliente activo",
                    isActive: true,
                  })
                );
              } else {
                inicializarSession(ws, ruc_empresa, this.clients);
              }
            } catch (error: any) {
              let error_message = "";
              if (error instanceof Error) {
                error_message = error.message;
              } else {
                error_message = error;
              }
              this.broadcast("error-event", {
                message:
                  "No se pudo inicializar el cliente de whatsapp se actualizará la página en breves.",
                error_message,
                codigo: 10,
              });
              return;
            }
          }
        } catch (e) {
          console.error("Error procesando mensaje WebSocket:", e);
          this.broadcast("error-event", {
            message:
              "Hubo un error al procesar el mensaje. Por favor, inténtelo de nuevo.",
            error_message: e,
            codigo: 1,
          });
        }
      });

      ws.on("close", () => {
        console.log("🔌 Cliente WebSocket desconectado");
        this.clients.forEach((client) => {
          try {
            if (client.websocket === ws) {
              closeClientWhatsappSession(client, this.clients);
              client.client = null;
              client.ruc = "";
              client.telefono = "";
            }
          } catch (error) {
            console.error(
              "Error al cerrar la sesión del cliente de whatsapp:",
              error
            );
          }
        });
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
      if (client.websocket.readyState === WebSocket.OPEN) {
        client.websocket.send(payload);
        console.warn("Mensaje enviado a cliente WebSocket:", payload);
      }
    });
  }
}

async function notifyIfActiveWhatsappClient(
  ruc_empresa: string
): Promise<boolean> {
  const waClient = listUserActiveClientWhatsapp.get(ruc_empresa);

  if (!waClient) {
    console.error("Cliente no encontrado en la lista de clientes activos");
    return false;
  }

  try {
    if (!(waClient.info && waClient.info.wid)) {
      console.log("Cliente no está logueado.");
      return false;
    }

    const state = await waClient.getState();
    console.log(state);

    return true;
  } catch (error: any) {
    if (
      error.message?.includes("Argument should belong") ||
      error.message?.includes("Target closed")
    ) {
      console.error("Error crítico de Puppeteer: sesión inválida.");
    } else {
      console.error("Error durante la autenticación del cliente:", error);
    }
    return false;
  }
}

// ✅ Método para manejar la inicialización de sesión
async function inicializarSession(
  ws: WebSocket,
  ruc_empresa: string,
  clients: Set<WsClienteConnection>
) {
  try {
    const data = await userExistInService(ruc_empresa);

    if (data === null) {
      console.log("El usuario no existe en la base de datos.");
      ws.close(1000, "El usuario no existe en la base de datos.");
      return;
    }

    const { telefono, nombre_comercial } = data;
    const waClient: Client = initializeNumberSession(telefono, ruc_empresa);
    // await waClient.destroy();
    console.log("Cliente whatsapp inicializado");
    listUserActiveClientWhatsapp.set(ruc_empresa, waClient);

    clients.forEach((client) => {
      if (client.websocket === ws) {
        client.client = waClient;
        console.log("Cliente whatsapp guardado en la conexión WebSocket");
      }
    });

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
      waClient.logout();
      console.error(messageError);
      listUserActiveClientWhatsapp.delete(ruc_empresa);
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
      // Si el cliente ya está autenticado, actualiza la referencia en el mapa global
      listUserActiveClientWhatsapp.set(ruc_empresa, waClient);
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
      const nameEvent = `qr-${ruc_empresa}`.trim();
      console.log(
        JSON.stringify(
          {
            event: nameEvent,
            cliente: `🏢 ${nombre_comercial}`,
            ruc: `🆔 ${ruc_empresa}`,
            qr: `🔗 ${qr}`,
          },
          null,
          2
        )
      );

      ws.send(
        payloadMessage(nameEvent, {
          message:
            "Escanee el código QR, para empezar a enviar utilizar el servicio de WhatsApp en nuestro sistema.",
          qrString: qr,
        })
      );
    });

    await waClient.initialize();
  } catch (error: any) {
    console.error("Error al inicializar la sesión:", error);
    ws.send(
      payloadMessage(`error-${ruc_empresa}`, {
        message: `Error inesperado al inicializar sesión WhatsApp: ${
          (error.message ?? "Ni idea mi king") || error
        }`,
        isActive: false,
      })
    );
  }
}

function payloadMessage(eventName: string, data: any) {
  const payload = JSON.stringify({ event: eventName, data });
  return payload;
}

async function closeClientWhatsappSession(
  client: WsClienteConnection,
  clients: Set<WsClienteConnection>
): Promise<boolean> {
  if (client.client == null) {
    return false;
  }
  try {
    const waClient = client.client;
    if (waClient.info && waClient.info.wid) {
      await waClient.destroy();
      console.log("✅ Cliente WhatsApp destruido correctamente 📱💥");
    }
    client.client = null;
    await new Promise((resolve) => setTimeout(resolve, 2000));
    clients.delete(client);
    console.log("🗑️ Cliente WebSocket eliminado de la lista");
    return true;
  } catch (err) {
    console.error("❌ Error al cerrar sesión del cliente WhatsApp:", err);
    return false;
  }
}

export default WebSocketHandler;

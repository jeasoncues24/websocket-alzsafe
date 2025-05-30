import WebSocket, { WebSocketServer } from "ws";
import { userExistInService } from "./app/services/user.service";
import { Client } from "whatsapp-web.js";

import {
  getSessionName,
  initializeNumberSession,
  listUserActiveClientWhatsapp,
} from "./utils/wa-client";
import {
  toogleServiceUser,
  activateUserModel,
  deactivateUserModel,
} from "./app/models/user.model";
import { WsClienteConnection } from "./interfaces/wscliente.interface";
import fs from "fs/promises";
import path from "path";
const clientesEnEspera = new Set();
// Mapa para asociar WebSockets con clientes de WhatsApp
const wsToRucMap = new Map<WebSocket, string>();
const rucToWaClientMap = new Map<string, Client>();

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
            try {
              const { ruc_empresa } = payload.data;
              const { sessionName, telefono } = await getSessionName(
                ruc_empresa
              );
              console.log("Event: init_session: ", ruc_empresa);

              // Asociar el WebSocket con el RUC
              wsToRucMap.set(ws, ruc_empresa);
              if (clientesEnEspera.has(ruc_empresa)) {
                ws.send(
                  payloadMessage(`error-${ruc_empresa}`, {
                    message:
                      "⏳ Cliente en proceso de limpieza, intenta conectarte en unos segundos...",
                    isActive: false,
                  })
                );
                return;
              }
              await inicializarSession(ws, ruc_empresa);
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
            }
          }
        } catch (e) {
          this.broadcast("error-event", {
            message:
              "Hubo un error al procesar el mensaje. Por favor, inténtelo de nuevo.",
            error_message: e,
            codigo: 1,
          });
        }
      });

      ws.on("close", async () => {
        console.log("🔌 Cliente WebSocket desconectado");

        try {
          // Obtener el RUC asociado a este WebSocket
          const ruc_empresa = wsToRucMap.get(ws);

          if (ruc_empresa) {
            console.log(`🧹 Limpiando recursos para empresa: ${ruc_empresa}`);
            wsToRucMap.delete(ws);
          }
        } catch (error) {
          console.error("❌ Error durante la limpieza del WebSocket:", error);
        }

        this.clients.delete(ws);
      });

      ws.on("error", async (error) => {
        console.error("❌ Error en WebSocket:", error);

        try {
          // Obtener el RUC asociado a este WebSocket
          const ruc_empresa = wsToRucMap.get(ws);

          if (ruc_empresa) {
            console.log(
              `🧹 Limpiando recursos por error para empresa: ${ruc_empresa}`
            );
            wsToRucMap.delete(ws);
          }
        } catch (cleanupError) {
          console.error(
            "❌ Error durante la limpieza por error del WebSocket:",
            cleanupError
          );
        }
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
  const data = await userExistInService(ruc_empresa);

  if (data === null) {
    console.log("El usuario no existe en la base de datos.");
    return;
  }

  const { telefono, nombre_comercial } = data;

  // Verificar si ya existe un cliente activo para evitar duplicados
  const existingClient =
    rucToWaClientMap.get(ruc_empresa) ||
    listUserActiveClientWhatsapp.get(ruc_empresa);
  if (existingClient) {
    console.log(
      `⚠️ Ya existe un cliente activo para ${ruc_empresa}, destruyendo el anterior...`
    );
    try {
      await existingClient.destroy();
    } catch (error) {
      console.error("Error al destruir cliente existente:", error);
    }
  }

  const waClient: Client = initializeNumberSession(telefono, ruc_empresa);

  // Guardar referencia del cliente en nuestro mapa local
  rucToWaClientMap.set(ruc_empresa, waClient);

  console.log(
    `Inicializando cliente WhatsApp para ${nombre_comercial} (${telefono})`
  );
  waClient.on("loading_screen", (percent, message) => {
    console.log("LOADING SCREEN", percent, message);
  });

  waClient.on("ready", async () => {
    try {
      await activateUserModel(ruc_empresa);
      await toogleServiceUser(ruc_empresa, 1);
      console.log(`✅ Cliente ${nombre_comercial} está listo.`);
      listUserActiveClientWhatsapp.set(ruc_empresa, waClient);
    } catch (error) {
      console.error("Error en evento 'ready':", error);
    }
  });

  waClient.on("disconnected", async (reason) => {
    try {
      clientesEnEspera.add(ruc_empresa);
      await toogleServiceUser(ruc_empresa, 0);
      await deactivateUserModel(ruc_empresa);
      const messageError = `Cliente ${nombre_comercial} se ha desconectado del servicio. Razón: ${reason}`;

      console.error(messageError);

      // Si la razón es LOGOUT, manejar la limpieza de sesión
      if (reason === "LOGOUT") {
        console.log(
          `🔄 Detectado LOGOUT para ${nombre_comercial}, preparando para nueva sesión...`
        );

        try {
          // Intentar logout normal primero
          // await waClient.logout();
          await limpiarSesionManual(ruc_empresa, telefono);
          console.log(`✅ Logout exitoso para ${nombre_comercial}`);
        } catch (logoutError: any) {
          console.log(
            `⚠️ Error en logout automático para ${nombre_comercial}:`,
            logoutError.message
          );
        }
      }

      // Limpiar referencias
      listUserActiveClientWhatsapp.delete(ruc_empresa);
      rucToWaClientMap.delete(ruc_empresa);

      const nameEvent = `active-${ruc_empresa}`;

      // Verificar si el WebSocket sigue abierto antes de enviar
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(
          payloadMessage(nameEvent, {
            message: messageError,
            isActive: false,
            requiresNewQR: reason === "LOGOUT", // Indicar si necesita nuevo QR
          })
        );
      }

      // Limpiar recursos adicionales
      console.log(`🧹 Limpiando recursos para empresa: ${ruc_empresa}`);
      setTimeout(() => {
        clientesEnEspera.delete(ruc_empresa);
        console.log(`✅ Cliente ${nombre_comercial} puede reconectar.`);
      }, 60000);
    } catch (error) {
      console.error("❌ Error al manejar desconexión del cliente:", error);

      // En caso de error, también limpiar referencias
      try {
        listUserActiveClientWhatsapp.delete(ruc_empresa);
        rucToWaClientMap.delete(ruc_empresa);
      } catch (cleanupError) {
        console.error("❌ Error al limpiar referencias:", cleanupError);
      }
    }
  });

  // waClient.on("disconnected", async (reason) => {
  //   try {
  //     await toogleServiceUser(ruc_empresa, 0);
  //     await deactivateUserModel(ruc_empresa);
  //     const messageError = `Cliente ${nombre_comercial} se ha desconectado del servicio. Razón: ${reason}`;

  //     // Limpiar referencias
  //     listUserActiveClientWhatsapp.delete(ruc_empresa);
  //     rucToWaClientMap.delete(ruc_empresa);

  //     console.error(messageError);
  //     const nameEvent = `active-${ruc_empresa}`;

  //     // Verificar si el WebSocket sigue abierto antes de enviar
  //     if (ws.readyState === WebSocket.OPEN) {
  //       ws.send(
  //         payloadMessage(nameEvent, {
  //           message: messageError,
  //           isActive: false,
  //         })
  //       );
  //     }
  //   } catch (error) {
  //     console.error("Error al manejar desconexión del cliente:", error);
  //   }
  // });

  waClient.on("auth_failure", async (msg) => {
    console.log(`❌ Fallo de autenticación para ${nombre_comercial}: ${msg}`);
  });

  waClient.on("remote_session_saved", async () => {
    console.log(
      `🔐 Sesión remota guardada para ${nombre_comercial}. Asegúrate de que el cliente esté autenticado.`
    );
  });

  waClient.on("authenticated", async (session) => {
    try {
      await activateUserModel(ruc_empresa);
      const message = `Cliente ${nombre_comercial} está autenticado en el servicio.`;
      console.log(message);
      const nameEvent = `active-${ruc_empresa}`;

      // Si el cliente ya está autenticado, actualiza la referencia en el mapa global
      listUserActiveClientWhatsapp.set(ruc_empresa, waClient);

      if (ws.readyState === WebSocket.OPEN) {
        ws.send(
          payloadMessage(nameEvent, {
            message: message,
            isActive: true,
          })
        );
      }
    } catch (error) {
      console.error("Error en evento 'authenticated':", error);
    }
  });

  waClient.on("message", async (msg) => {
    console.log(`📨 Mensaje recibido en ${nombre_comercial}: ${msg}`);
  });

  waClient.on("qr", async (qr) => {
    try {
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

      if (ws.readyState === WebSocket.OPEN) {
        ws.send(
          payloadMessage(nameEvent, {
            message:
              "Escanee el código QR, para empezar a enviar utilizar el servicio de WhatsApp en nuestro sistema.",
            qrString: qr,
          })
        );
      }
    } catch (error) {
      console.error("Error en evento 'qr':", error);
    }
  });

  try {
    await waClient.initialize();
    console.log("✅ Cliente WhatsApp inicializado");
  } catch (error) {
    console.error("❌ Error al inicializar cliente WhatsApp:", error);
    // Limpiar referencias si falla la inicialización
    rucToWaClientMap.delete(ruc_empresa);
    throw error;
  }
}

function payloadMessage(eventName: string, data: any) {
  const payload = JSON.stringify({ event: eventName, data });
  return payload;
}
async function eliminarSesionConRetry(sessionPath: string, intentos = 3) {
  for (let i = 0; i < intentos; i++) {
    try {
      // Primero intentar eliminar archivos específicos problemáticos
      const problematicFiles = [
        path.join(sessionPath, "Default", "chrome_debug.log"),
        path.join(sessionPath, "Default", "Cookies"),
        path.join(sessionPath, "Default", "Cookies-journal"),
      ];

      for (const file of problematicFiles) {
        try {
          await fs.unlink(file);
        } catch (e) {
          // Ignorar si el archivo no existe
        }
      }

      // Luego eliminar toda la carpeta
      await fs.rm(sessionPath, { recursive: true, force: true });
      console.log(`✅ Sesión eliminada exitosamente en intento ${i + 1}`);
      return true;
    } catch (error) {
      console.error("Reintentando eliminar sesión:", error);
      if (i === intentos - 1) {
        console.error(
          `🚨 No se pudo eliminar la sesión después de ${intentos} intentos`
        );
        return false;
      }
      // Esperar antes del siguiente intento
      await new Promise((resolve) => setTimeout(resolve, 2000));
    }
  }
}

// Función para limpiar sesión manualmente
async function limpiarSesionManual(ruc_empresa: string, telefono: string) {
  const sessionPath = `./.wwebjs_auth/session-${ruc_empresa}-${telefono}`;
  try {
    const exists = await fs
      .access(sessionPath)
      .then(() => true)
      .catch(() => false);
    if (exists) {
      console.log(`🧹 Limpiando sesión manualmente: ${sessionPath}`);
      await eliminarSesionConRetry(sessionPath);
    } else {
      console.log(`ℹ️ La sesión ya no existe: ${sessionPath}`);
    }
  } catch (error) {
    console.error(`❌ Error al limpiar sesión manual: `, error);
  }
}
export default WebSocketHandler;

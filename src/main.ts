import express from "express";
import WebSocketHandler from "./websocket";
import Database from "./lib/mysql";
import { Client, LocalAuth } from "whatsapp-web.js";
import qrcode from "qrcode-terminal";
import os from "os";
import { router } from "./routes";
import { listNumberUserService } from "./app/services/user.service";
import {
  getClientStatus,
  initializeNumberSession,
  listUserActiveClientWhatsapp,
} from "./utils/wa-client";
import cors from "cors";
import {
  activateServiceUsuario,
  activateUserModel,
  deactivateUserModel,
} from "./app/models/user.model";
import { getActiveAndLinkedUsers } from "./app/models/wa.model";

let wsHandler: WebSocketHandler;

require("dotenv").config();

const app = express();
const port = 3000;
const db = new Database();

app.use(cors());
app.use(express.json());
app.use(router);

const server = app.listen(port, () => {
  console.log(`Servidor Express corriendo en http://localhost:${port}`);
  wsHandler = new WebSocketHandler(server);
});

// Asegúrate de cerrar la conexión a la base de datos al finalizar
process.on("SIGINT", async () => {
  console.log("Cerrando conexión a la base de datos...");
  await db.endPool();
  process.exit();
});

const inicializarNumerosWhatsApp = async () => {
  const numbers = await getActiveAndLinkedUsers();

  if (numbers.length === 0) {
    console.log("No hay números de WhatsApp activos y vinculados.");
    return;
  }

  numbers.forEach(async (users) => {
    const { ruc: ruc_empresa, telefono, nombre_comercial } = users;
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
    });

    waClient.on("authenticated", async (session) => {
      await activateUserModel(ruc_empresa);
      const message = `Cliente ${nombre_comercial} está autenticado en el servicio.`;
      console.log(message);
    });

    // waClient.on("message", async (msg) => {
    //   console.log(`message in ${nombre_comercial}` + msg);
    // });

    await waClient.initialize();
  });
};

inicializarNumerosWhatsApp();

import express from "express";
import WebSocketHandler from "./websocket";
import Database from "./mysql";
import { Client, LocalAuth } from "whatsapp-web.js";
import qrcode from "qrcode-terminal";
import os from "os";

require("dotenv").config();

const app = express();
const port = 3000;
const db = new Database();

app.post("/wa/sendMessageDirect", async (req, res) => {});

app.post("/wa/sendDifusionDirect", async (req, res) => {});

const server = app.listen(port, () => {
  console.log(`Servidor Express corriendo en http://localhost:${port}`);
  new WebSocketHandler(server); // Pasa el servidor Express al WebSocketHandler
});

// Asegúrate de cerrar la conexión a la base de datos al finalizar
process.on("SIGINT", async () => {
  console.log("Cerrando conexión a la base de datos...");
  await db.endPool();
  process.exit();
});

const clients: Record<string, Client> = {}; // Para almacenar las instancias de los clientes por número
const initializeClient = async (sessionId: string) => {
  const sessionData = await db.getSession(sessionId);

  const client = new Client({
    authStrategy: new LocalAuth({ clientId: sessionId }), // Usar sessionId como clientId para LocalAuth
  });

  client.on("qr", (qr) => {
    qrcode.generate(qr, { small: true });
    console.log(`QR RECEIVED para ${sessionId}:`, qr);
  });

  client.on("ready", async () => {
    console.log(`Cliente ${sessionId} está listo!`);
    clients[sessionId] = client;

    // const session = await client.authStrategy.getSession();
    // if (session) {
    //   await db.saveSession(sessionId, session);
    // }
  });

  client.on("message", async (msg) => {
    //
  });

  client.on("auth_failure", (msg) => {
    console.error(`Error de autenticación para ${sessionId}`, msg);
    // Aquí podrías implementar lógica para reintentar la autenticación o notificar
  });

  client.on("disconnected", (reason) => {
    console.log(`Cliente ${sessionId} desconectado debido a:`, reason);
    delete clients[sessionId];
    // Aquí podrías implementar lógica para intentar reconectar o limpiar la sesión
  });

  // Evento para guardar la sesión después de la autenticación (para LocalAuth esto ya debería estar cubierto en 'ready')
  // client.on('authenticated', (session) => {
  //   console.log(`Autenticado ${sessionId}, datos de sesión:`, session);
  //   db.saveSession(sessionId, session);
  // });

  await client.initialize();
};

const sessionNumbers = ["51947123809", "51977596225"]; // Reemplaza con tus números de teléfono del bot

sessionNumbers.forEach((number) => {
  initializeClient(number);
});

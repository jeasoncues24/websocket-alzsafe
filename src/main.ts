import express from "express";
import WebSocketHandler from "./websocket";
import Database from "./lib/mysql";
import { Client, LocalAuth } from "whatsapp-web.js";
import qrcode from "qrcode-terminal";
import os from "os";
import { router } from "./routes";

import cors from "cors";
import { inicializarNumerosWhatsApp } from "./app/models/wa.model";
import { EventEmitter } from "events";


// EventEmitter.defaultMaxListeners = 50;
process.setMaxListeners(50);


let wsHandler: WebSocketHandler;

require("dotenv").config();

const app = express();
const port = process.env.SERVER_PORT;
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

inicializarNumerosWhatsApp();

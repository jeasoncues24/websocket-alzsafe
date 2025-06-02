"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const express_1 = __importDefault(require("express"));
const websocket_1 = __importDefault(require("./websocket"));
const mysql_1 = __importDefault(require("./lib/mysql"));
const routes_1 = require("./routes");
const cors_1 = __importDefault(require("cors"));
const wa_model_1 = require("./app/models/wa.model");
let wsHandler;
require("dotenv").config();
const app = (0, express_1.default)();
const port = process.env.SERVER_PORT;
const db = new mysql_1.default();
app.use((0, cors_1.default)());
app.use(express_1.default.json());
app.use(routes_1.router);
const server = app.listen(port, () => {
    console.log(`Servidor Express corriendo en http://localhost:${port}`);
    wsHandler = new websocket_1.default(server);
});
// Asegúrate de cerrar la conexión a la base de datos al finalizar
process.on("SIGINT", async () => {
    console.log("Cerrando conexión a la base de datos...");
    await db.endPool();
    process.exit();
});
(0, wa_model_1.inicializarNumerosWhatsApp)();

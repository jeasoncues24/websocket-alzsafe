"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.inicializarNumerosWhatsApp = void 0;
exports.insertMessageInDatabase = insertMessageInDatabase;
exports.getActiveAndLinkedUsers = getActiveAndLinkedUsers;
const mysql_1 = __importDefault(require("../../lib/mysql"));
const wa_client_1 = require("../../utils/wa-client");
const user_model_1 = require("./user.model");
const db = new mysql_1.default();
async function insertMessageInDatabase(msg) {
    const sql = `
        INSERT INTO messages (timestamp, message, codigo_postal_receptor, telefono_receptor, codUsuario)
        VALUES (?, ?, ?, ?, ?)
    `;
    const params = [
        msg.timestamp,
        msg.message,
        msg.codigo_postal_receptor,
        msg.telefono_receptor,
        msg.codUsuario,
    ];
    try {
        await db.query(sql, params);
        return true;
    }
    catch (error) {
        console.log("Error al insertar el mensaje:", error);
        return false;
    }
}
async function getActiveAndLinkedUsers() {
    const sql = `
    SELECT * FROM users
    WHERE is_active = 1 AND is_linked = 1
  `;
    try {
        const results = await db.query(sql);
        console.log("Usuarios activos y vinculados:", results);
        return results;
    }
    catch (error) {
        console.log("Error al obtener los usuarios activos y vinculados:", error);
        return [];
    }
}
const inicializarNumerosWhatsApp = async () => {
    const numbers = await getActiveAndLinkedUsers();
    if (numbers.length === 0) {
        console.log("No hay números de WhatsApp activos y vinculados.");
        return;
    }
    numbers.forEach(async (users) => {
        const { ruc: ruc_empresa, telefono, nombre_comercial } = users;
        const waClient = (0, wa_client_1.initializeNumberSession)(telefono, ruc_empresa);
        wa_client_1.listUserActiveClientWhatsapp.set(ruc_empresa, waClient);
        waClient.on("ready", async () => {
            await (0, user_model_1.activateUserModel)(ruc_empresa);
            await (0, user_model_1.toogleServiceUser)(ruc_empresa, 1);
            console.log(`Cliente ${nombre_comercial} está listo.`);
            wa_client_1.listUserActiveClientWhatsapp.set(ruc_empresa, waClient);
        });
        waClient.on("disconnected", async (reason) => {
            await (0, user_model_1.toogleServiceUser)(ruc_empresa, 0);
            await (0, user_model_1.deactivateUserModel)(ruc_empresa);
            const messageError = `Cliente ${nombre_comercial} se ha desconectado del servicio.`;
            wa_client_1.listUserActiveClientWhatsapp.delete(ruc_empresa);
            console.error(messageError);
            // try {
            //   await waClient.logout();
            // } catch (logoutError) {
            //   console.error("CERRAR_SESION_FT:", logoutError);
            // }
        });
        waClient.on("authenticated", async (session) => {
            await (0, user_model_1.activateUserModel)(ruc_empresa);
            const message = `Cliente ${nombre_comercial} está autenticado en el servicio.`;
            console.log(message);
            wa_client_1.listUserActiveClientWhatsapp.set(ruc_empresa, waClient);
        });
        await waClient.initialize();
    });
};
exports.inicializarNumerosWhatsApp = inicializarNumerosWhatsApp;

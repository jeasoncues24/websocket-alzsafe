import { Client } from "whatsapp-web.js";
import { Message } from "../../interfaces/message.interface";
import { User } from "../../interfaces/user.interface";
import Database from "../../lib/mysql";
import { initializeNumberSession, listUserActiveClientWhatsapp } from "../../utils/wa-client";
import { activateUserModel, deactivateUserModel, toogleServiceUser } from "./user.model";

const db = new Database();






async function insertMessageInDatabase(msg: Message): Promise<boolean> {
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
  } catch (error) {
    console.log("Error al insertar el mensaje:", error);
    return false;
  }
}

async function getActiveAndLinkedUsers(): Promise<User[]> {
  const sql = `
    SELECT * FROM users
    WHERE is_active = 1 AND is_linked = 1
  `;
  try {
    const results = await db.query(sql);
    console.log("Usuarios activos y vinculados:", results);
    return results;
  } catch (error) {
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
    const waClient: Client = initializeNumberSession(telefono, ruc_empresa);
    listUserActiveClientWhatsapp.set(ruc_empresa, waClient);
    waClient.on("ready", async () => {
      await activateUserModel(ruc_empresa);
      await toogleServiceUser(ruc_empresa, 1);
      console.log(`Cliente ${nombre_comercial} está listo.`);
      listUserActiveClientWhatsapp.set(ruc_empresa, waClient);
    });

    waClient.on("disconnected", async (reason) => {
      await toogleServiceUser(ruc_empresa, 0);
      await deactivateUserModel(ruc_empresa);
      const messageError = `Cliente ${nombre_comercial} se ha desconectado del servicio.`;
      listUserActiveClientWhatsapp.delete(ruc_empresa);
      console.error(messageError);
    });

    waClient.on("authenticated", async (session) => {
      await activateUserModel(ruc_empresa);
      const message = `Cliente ${nombre_comercial} está autenticado en el servicio.`;
      console.log(message);
      listUserActiveClientWhatsapp.set(ruc_empresa, waClient);
    });
    await waClient.initialize();
  });
};



export {
  insertMessageInDatabase,
  getActiveAndLinkedUsers,
  inicializarNumerosWhatsApp,
};


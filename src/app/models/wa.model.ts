import { Client } from "whatsapp-web.js";
import { Message } from "../../interfaces/message.interface";
import { User } from "../../interfaces/user.interface";
import Database from "../../lib/mysql";
import {
  initializeNumberSession,
  listUserActiveClientWhatsapp,
} from "../../utils/wa-client";
import {
  activateUserModel,
  deactivateUserModel,
  toogleServiceUser,
} from "./user.model";

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

  for (const user of numbers) {
    const { ruc: ruc_empresa, telefono, nombre_comercial } = user;

    if (listUserActiveClientWhatsapp.has(ruc_empresa)) {
      console.log(`Ya existe cliente activo para ${ruc_empresa}`);
      continue;
    }

    const waClient: Client = initializeNumberSession(telefono, ruc_empresa);

    // registrar eventos en función aparte
    registrarEventosCliente(waClient, {
      ruc_empresa,
      telefono,
      nombre_comercial,
    });

    listUserActiveClientWhatsapp.set(ruc_empresa, waClient);

    try {
      await waClient.initialize();
      console.log(
        `Inicialización de cliente ${nombre_comercial} (${telefono}) en progreso...`
      );
    } catch (err) {
      console.error(
        `Error inicializando cliente ${nombre_comercial} (${telefono}):`,
        err
      );
      listUserActiveClientWhatsapp.delete(ruc_empresa);
    }
  }
};

function registrarEventosCliente(
  waClient: Client,
  {
    ruc_empresa,
    telefono,
    nombre_comercial,
  }: { ruc_empresa: string; telefono: string; nombre_comercial: string }
) {
  waClient.on("ready", async () => {
    await activateUserModel(ruc_empresa);
    await toogleServiceUser(ruc_empresa, 1);
    console.log(
      `✅ Cliente ${nombre_comercial} (${telefono}) está listo para iniciar en wsp.`
    );
    replaceExistingClient(ruc_empresa, waClient);
  });

  waClient.on("disconnected", async (reason) => {
    await toogleServiceUser(ruc_empresa, 0);
    await deactivateUserModel(ruc_empresa);
    listUserActiveClientWhatsapp.delete(ruc_empresa);
    console.error(
      `❌ Cliente ${nombre_comercial} (${telefono}) se desconectó: ${reason}`
    );
  });

  waClient.on("authenticated", async () => {
    await activateUserModel(ruc_empresa);
    console.log(`🔐 Cliente ${nombre_comercial} (${telefono}) autenticado.`);
    replaceExistingClient(ruc_empresa, waClient);
  });
}

function replaceExistingClient(ruc_empresa: string, waClient: Client) {
  if (listUserActiveClientWhatsapp.has(ruc_empresa)) {
    console.warn(`Actualizando instancia del cliente para ${ruc_empresa}`);
  }
  listUserActiveClientWhatsapp.set(ruc_empresa, waClient);
}

export {
  insertMessageInDatabase,
  getActiveAndLinkedUsers,
  inicializarNumerosWhatsApp,
};

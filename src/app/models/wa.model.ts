import { Message } from "../../interfaces/message.interface";
import { User } from "../../interfaces/user.interface";
import Database from "../../lib/mysql";

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

export { insertMessageInDatabase, getActiveAndLinkedUsers };

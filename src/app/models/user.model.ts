import { User } from "../../interfaces/user.interface";
import Database from "../../lib/mysql";

const db = new Database();

const addUserModel = async (user: User): Promise<void> => {
  // Verificamos si ya existe un usuario con el mismo ruc y teléfono
  const checkSql = `SELECT 1 FROM users WHERE ruc = ? AND telefono = ? LIMIT 1`;
  const [existingUser] = await db.query(checkSql, [user.ruc, user.telefono]);

  if (Array.isArray(existingUser) && existingUser.length > 0) {
    console.log("Usuario ya registrado con ese RUC y teléfono.");
    return;
  }

  const insertSql = `
    INSERT INTO users (ruc, razon_social, nombre_comercial, telefono, codigo_postal, is_active, is_linked)
    VALUES (?, ?, ?, ?, ?, ?, ?)
  `;
  try {
    await db.query(insertSql, [
      user.ruc,
      user.razon_social,
      user.nombre_comercial,
      user.telefono,
      user.codigo_postal,
      user.is_active,
      user.is_linked,
    ]);
  } catch (error: any) {
    console.error("Error ingresar usuario:", error);
    if (error.code === "ER_DUP_ENTRY" || error.errno === 1062) {
      // No hacer nada
      return;
    }
    throw error;
  }
};

const listNumberUserModel = async (): Promise<
  {
    user_id: number;
    codigo_postal: string;
    telefono: string;
    nombre_comercial: string;
    number_format: string;
    ruc_empresa: string;
  }[]
> => {
  const sql = `SELECT telefono, ruc as ruc_empresa, nombre_comercial, codigo_postal, CONCAT(codigo_postal, telefono) AS number_format, id as user_id FROM users WHERE is_linked = 1`;
  const data = await db.query(sql);
  return data;
};

const activateServiceUsuario = async (
  ruc_empresa: string,
  estado: number = 0
): Promise<void> => {
  const sql = `UPDATE users SET is_active = ? WHERE ruc = ?`;
  await db.query(sql, [estado, ruc_empresa]);
};

const deactivateUserModel = async (ruc_cliente: string): Promise<void> => {
  const sql = `UPDATE users SET is_linked = 0 WHERE ruc = ?`;
  await db.query(sql, [ruc_cliente]);
};

const activateUserModel = async (ruc_cliente: string): Promise<void> => {
  const sql = `UPDATE users SET is_linked = 1 WHERE ruc = ?`;
  await db.query(sql, [ruc_cliente]);
};

const getUserByIdModel = async (ruc_empresa: string): Promise<User | null> => {
  const sql = `SELECT * FROM users WHERE ruc = ?`;
  const results = await db.query(sql, [ruc_empresa]);
  if (results.length === 0) {
    return null;
  }
  return results[0];
};

export {
  addUserModel,
  listNumberUserModel,
  deactivateUserModel,
  activateUserModel,
  getUserByIdModel,
  activateServiceUsuario,
};

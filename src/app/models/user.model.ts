import { User } from "../../interfaces/user.interface";
import Database from "../../lib/mysql";

const db = new Database();

const addUserModel = async (user: User): Promise<void> => {
  const sql = `INSERT INTO users (ruc, razon_social, nombre_comercial, telefono, codigo_postal, is_active, is_linked) VALUES (?, ?, ?, ?, ?, ?, ? )`;
  return await db.query(sql, [ user.ruc, user.razon_social, user.nombre_comercial, user.telefono, user.codigo_postal, user.is_active, user.is_linked]);
}

const listNumberUserModel = async (): Promise<void> => {
  const sql = `SELECT telefono, codigo_postal, CONCAT(codigo_postal, telefono) AS number_format FROM users WHERE is_active = 1`;
  const data = await db.query(sql);
  return data;
}

export {
    addUserModel,
    listNumberUserModel
}
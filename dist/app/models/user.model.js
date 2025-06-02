"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.toogleServiceUser = exports.getUserByIdModel = exports.activateUserModel = exports.deactivateUserModel = exports.listNumberUserModel = exports.addUserModel = void 0;
const mysql_1 = __importDefault(require("../../lib/mysql"));
const db = new mysql_1.default();
const addUserModel = async (user) => {
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
    }
    catch (error) {
        console.error("Error ingresar usuario:", error);
        if (error.code === "ER_DUP_ENTRY" || error.errno === 1062) {
            // No hacer nada
            return;
        }
        throw error;
    }
};
exports.addUserModel = addUserModel;
const listNumberUserModel = async () => {
    const sql = `SELECT telefono, ruc as ruc_empresa, nombre_comercial, codigo_postal, CONCAT(codigo_postal, telefono) AS number_format, id as user_id FROM users WHERE is_linked = 1`;
    const data = await db.query(sql);
    return data;
};
exports.listNumberUserModel = listNumberUserModel;
const toogleServiceUser = async (ruc_empresa, estado = 0) => {
    const sql = `UPDATE users SET is_active = ? WHERE ruc = ?`;
    await db.query(sql, [estado, ruc_empresa]);
};
exports.toogleServiceUser = toogleServiceUser;
const deactivateUserModel = async (ruc_cliente) => {
    const sql = `UPDATE users SET is_linked = 0 WHERE ruc = ?`;
    await db.query(sql, [ruc_cliente]);
};
exports.deactivateUserModel = deactivateUserModel;
const activateUserModel = async (ruc_cliente) => {
    const sql = `UPDATE users SET is_linked = 1 WHERE ruc = ?`;
    await db.query(sql, [ruc_cliente]);
};
exports.activateUserModel = activateUserModel;
const getUserByIdModel = async (ruc_empresa) => {
    const sql = `SELECT * FROM users WHERE ruc = ?`;
    const results = await db.query(sql, [ruc_empresa]);
    if (results.length === 0) {
        return null;
    }
    return results[0];
};
exports.getUserByIdModel = getUserByIdModel;

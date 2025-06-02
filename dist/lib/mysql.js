"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const promise_1 = __importDefault(require("mysql2/promise"));
require("dotenv/config");
class Database {
    constructor(config = {}) {
        this.pool = promise_1.default.createPool({
            host: config.host || process.env.DB_HOST,
            port: config.port || parseInt(process.env.DB_PORT || "3306", 10),
            user: config.user || process.env.DB_USER,
            password: config.password || process.env.DB_PASSWORD,
            database: config.database || process.env.DB_NAME,
        });
    }
    async getConnection() {
        return await this.pool.getConnection();
    }
    async query(sql, values) {
        const connection = await this.getConnection();
        try {
            const [rows] = await connection.execute(sql, values);
            return rows;
        }
        finally {
            connection.release();
        }
    }
    async endPool() {
        await this.pool.end();
    }
}
exports.default = Database;

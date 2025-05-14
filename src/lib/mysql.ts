import mysql, { Pool, PoolConnection } from "mysql2/promise";
import "dotenv/config";

interface DatabaseConfig {
  host?: string;
  port?: number;
  user?: string;
  password?: string;
  database?: string;
}

class Database {
  private pool: Pool;

  constructor(config: DatabaseConfig = {}) {
    this.pool = mysql.createPool({
      host: config.host || process.env.DB_HOST,
      port: config.port || parseInt(process.env.DB_PORT || "3306", 10),
      user: config.user || process.env.DB_USER,
      password: config.password || process.env.DB_PASSWORD,
      database: config.database || process.env.DB_NAME,
    });
  }

  async getConnection(): Promise<PoolConnection> {
    return await this.pool.getConnection();
  }

  async query(sql: string, values?: any[]): Promise<any> {
    const connection = await this.getConnection();
    try {
      const [rows] = await connection.execute(sql, values);
      return rows;
    } finally {
      connection.release();
    }
  }

  async endPool(): Promise<void> {
    await this.pool.end();
  }

}

export default Database;

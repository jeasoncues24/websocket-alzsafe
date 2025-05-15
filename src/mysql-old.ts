// import mysql, { Pool, PoolConnection } from "mysql2/promise";
// import "dotenv/config";

// interface DatabaseConfig {
//   host?: string;
//   port?: number;
//   user?: string;
//   password?: string;
//   database?: string;
// }

// class Database {
//   private pool: Pool;

//   constructor(config: DatabaseConfig = {}) {
//     this.pool = mysql.createPool({
//       host: config.host || process.env.DB_HOST,
//       port: config.port || parseInt(process.env.DB_PORT || "3306", 10),
//       user: config.user || process.env.DB_USER,
//       password: config.password || process.env.DB_PASSWORD,
//       database: config.database || process.env.DB_NAME,
//     });
//   }

//   async getConnection(): Promise<PoolConnection> {
//     return await this.pool.getConnection();
//   }

//   async query(sql: string, values?: any[]): Promise<any> {
//     const connection = await this.getConnection();
//     try {
//       const [rows] = await connection.execute(sql, values);
//       return rows;
//     } finally {
//       connection.release();
//     }
//   }

//   async endPool(): Promise<void> {
//     await this.pool.end();
//   }

//   async getSession(sessionId: string): Promise<any> {
//     const results = await this.query("SELECT data FROM sessions WHERE id = ?", [
//       sessionId,
//     ]);
//     return results.length > 0 ? results[0].session_data : null;
//   }

//   async saveSession(sessionId: string, sessionData: any): Promise<void> {
//     await this.query(
//       "INSERT INTO sessions (id, session_data) VALUES (?, ?) ON CONFLICT (id) DO UPDATE SET session_data = ?",
//       [sessionId, JSON.stringify(sessionData), JSON.stringify(sessionData)]
//     );
//   }

//   // async upsertUser(userData: {
//   //   id: string;
//   //   name?: string;
//   //   pushname?: string;
//   //   is_group: boolean;
//   // }): Promise<void> {
//   //   const { id, name, pushname, is_group } = userData;
//   //   await this.query(
//   //     "INSERT INTO users (id, name, pushname, is_group) VALUES (?, ?, ?, ?) ON CONFLICT (id) DO UPDATE SET name = ?, pushname = ?, updated_at = NOW()",
//   //     [id, name, pushname, is_group, name, pushname]
//   //   );
//   // }

//   async logUserSessionInteraction(
//     userId: string,
//     sessionId: string
//   ): Promise<void> {
//     await this.query(
//       "INSERT INTO user_sessions (user_id, session_id) VALUES (?, ?)",
//       [userId, sessionId]
//     );
//   }
// }

// export default Database;

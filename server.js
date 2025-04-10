const WebSocket = require('ws');
const mysql = require('mysql2/promise');
require('dotenv').config();

const wss = new WebSocket.Server({ port: process.env.SERVER_PORT });
console.log(`Servidor WebSocket activo en ws://localhost:${process.env.SERVER_PORT}`);

let db;
(async () => {
  db = await mysql.createConnection({
    host: process.env.DB_HOST,
    user: process.env.DB_USER, 
    password: process.env.DB_PASSWORD,
    database: process.env.DB_NAME 
  });
  console.log('Conectado a la base de datos perritoooooooooooooos');
})();

const cuidadores = new Map();

wss.on('connection', (ws) => {
  console.log('Cliente WebSocket conectado');

  ws.on('message', (msg) => {
    try {
      const data = JSON.parse(msg);

      if (data.type === 'init' && data.userType === 'cuidador') {
        cuidadores.set(data.userId, ws);
        console.log(`Cuidador ${data.userId} registrado`);
      }
    } catch (e) {
      console.error('Error procesando mensaje:', e);
    }
  });

  ws.on('close', () => {
    for (const [id, socket] of cuidadores.entries()) {
      if (socket === ws) {
        cuidadores.delete(id);
        console.log(`Cuidador ${id} desconectado`);
        break;
      }
    }
  });
});

let lastRequestId = 0;

setInterval(async () => {
  if (!db) return;

  try {

    const [rows] = await db.execute(
      'SELECT * FROM requests WHERE id > ? ORDER BY id ASC',
      [lastRequestId]
    );

    if (rows.length > 0) {
        
      lastRequestId = rows[rows.length - 1].id; 
      // console.log(lastRequestId)

      for (const req of rows) {
        console.log(req)
        const cuidadorSocket = cuidadores.get(req.carer_id); 
        if (cuidadorSocket) {
          cuidadorSocket.send(JSON.stringify({
            type: 'new-request',
            data: req
          }));
          console.log(`Notificación enviada a cuidador ${req.id_cuidador}`);
        }
      }
    } else {
        console.log('No hay nuevos registros');
    }
  } catch (err) {
    console.error(' Error consultando requests:', err);
  }
}, 3000); 

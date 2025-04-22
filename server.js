const WebSocket = require('ws');
const mysql = require('mysql2/promise');
require('dotenv').config();
const { Client, LocalAuth } = require('whatsapp-web.js');
const qrcode = require('qrcode-terminal');
const os = require('os');

const wss = new WebSocket.Server({ port: process.env.SERVER_PORT });

const getLocalIpAddress = () => {
  const interfaces = os.networkInterfaces();
  for (const name of Object.keys(interfaces)) {
    for (const iface of interfaces[name]) {
      if (iface.family === 'IPv4' && !iface.internal) {
        return iface.address;
      }
    }
  }
  return '127.0.0.1';
};

const ipAddress = getLocalIpAddress();
console.log(`Servidor WebSocket activo en ws://${ipAddress}:${process.env.SERVER_PORT}`);


let db;
(async () => {
  db = await mysql.createConnection({
    host: process.env.DB_HOST_DB,
    user: process.env.DB_USER,
    password: process.env.DB_PASSWORD,
    database: process.env.DB_NAME
  });
  console.log('✅ Conectado a la base de datos');
})();

const cuidadores = new Map();

wss.on('connection', (ws) => {
  const now = new Date();
  const connectionTime = `${now.toLocaleDateString()} ${now.toLocaleTimeString()}`;
  console.log(`🔗 Cliente conectado - ${connectionTime}`);

  ws.on('message', (msg) => {
    try {
      const data = JSON.parse(msg);

      // 1. Registro de cuidadores
      if (data.type === 'init' && data.userType === 'cuidador') {
        cuidadores.set(data.userId, ws);
        console.log(`✅ Cuidador ${data.userId} registrado`);

        (async () => {
          try {
            const [rows] = await db.execute(
              'SELECT phone FROM carer WHERE user_id = ?',
              [data.userId]
            );

            if (rows.length > 0) {
              const phone = rows[0].phone;
              enviarMensajeWhatsApp(`51${phone}`, '¡Hola! Has iniciado sesión en la aplicación AlzSafe ❤️🙌. Tus pacientes te esperan, empezemos a trabajar.');
              console.log(`Mensaje enviado al cuidador ${data.userId} con número ${phone}`);
            } else {
              console.log(`No se encontró un cuidador con ID ${data.userId}`);
            }
          } catch (error) {
            console.error('Error al buscar el número de teléfono del cuidador:', error);
          }
        })();
        return;
      }
      // 2. Procesar evento enviar-solicitud
      if (data.event === 'enviar-solicitud') {
        const payload = data.data;
        const cuidadorId = payload.idCuidador;

        console.log(`📨 Solicitud recibida para cuidador ${cuidadorId}:`, payload);

        const cuidadorSocket = cuidadores.get(cuidadorId);

        // Intentamos enviar la notificación por WebSocket si el cuidador está conectado
        if (cuidadorSocket && cuidadorSocket.readyState === WebSocket.OPEN) {
          cuidadorSocket.send(JSON.stringify({
            event: 'enviar-solicitud',
            data: payload,
          }));
          console.log(`✅ Notificación enviada al cuidador ${cuidadorId} por WebSocket`);
        } else {
          console.warn(`⚠️ Cuidador ${cuidadorId} no conectado al enviar la solicitud por WebSocket.`);
        }

        // Siempre enviamos el mensaje de WhatsApp
        (async () => {
          try {
            const [rows] = await db.execute(
              `SELECT r.id AS request_id, c.phone AS telefono_cuidador,
                             p.id AS paciente_id, p.name AS paciente_nombre,
                             p.genre AS paciente_genero, p.age AS paciente_edad,
                             p.address AS paciente_direccion, u.id AS familiar_id,
                             CONCAT(u.name, ' ', u.last_name) AS familiar_nombre,
                             u.address AS direccion
                            FROM requests r
                            JOIN patients p ON r.patient_id = p.id
                            JOIN carer c ON r.carer_id = c.id
                            JOIN users u ON r.familiar_id = u.id
                            WHERE c.user_id = ? AND r.status = 1
                            ORDER BY r.id DESC`,
              [cuidadorId]
            );

            if (rows.length > 0) {
              const solicitud = rows[0];
              const mensaje = `🌟 ¡Hola! Tienes una nueva solicitud de amistad para cuidar a un paciente. Aquí está la información:

    👤 *Paciente*: ${solicitud.paciente_nombre}
    📍 *Dirección*: ${solicitud.paciente_direccion}
    🧬 *Género*: ${solicitud.paciente_genero}
    🎂 *Edad*: ${solicitud.paciente_edad} años

    👨‍👩‍👧‍👦 *Familiar Contratante*: ${solicitud.familiar_nombre}
    🏠 *Dirección del Familiar*: ${solicitud.direccion}

    💌 Por favor, revisa la solicitud y prepárate para brindar tu mejor atención. ¡Gracias por ser parte de AlzSafe! ❤️`;

              // Suponiendo que 'enviarMensajeWhatsApp' está definida en otro lugar
              enviarMensajeWhatsApp(`51${solicitud.telefono_cuidador}`, mensaje);
              console.log(`📩 [WhatsApp] Mensaje de solicitud enviado al cuidador ${cuidadorId} con número ${solicitud.telefono_cuidador}`);
            } else {
              console.log(`⚠️ No se encontraron solicitudes activas para el cuidador ${cuidadorId}`);
            }
          } catch (error) {
            console.error('❌ Error al consultar la solicitud o enviar el mensaje:', error);
          }
        })();

        return;
      }
      console.log('📥 Mensaje desconocido recibido:', data);
    } catch (e) {
      console.error('❌ Error procesando mensaje:', e);
    }
  });

  ws.on('close', () => {
    for (const [id, socket] of cuidadores.entries()) {
      if (socket === ws) {
        cuidadores.delete(id);
        console.log(`💔 Conexión cerrada para cuidador ${id}`);
        break;
      }
    }
    console.log(`🔌 Conexión WebSocket cerrada para cliente`);
  });
});

const whatsappClient = new Client({
  authStrategy: new LocalAuth()
});

whatsappClient.on('qr', qr => {
  console.log('📲 Escanea este QR con tu WhatsApp:');
  qrcode.generate(qr, { small: true });
});

whatsappClient.on('ready', () => {
  console.log('🤖 Bot de WhatsApp está listo!');
});


whatsappClient.initialize();

//Enviar mensaje de WhatsApp cada minuto
const enviarMensajeWhatsApp = async (telefono, mensaje) => {
  try {
    const chatId = `${telefono}@c.us`;
    await whatsappClient.sendMessage(chatId, mensaje);
    console.log(`Mensaje enviado a ${telefono} a las ${new Date().toLocaleTimeString()}`);
  } catch (error) {
    console.error('Error al enviar el mensaje de WhatsApp:', error);
  }
};

// // Ejemplo de uso: enviar mensaje cada minuto
// setInterval(() => {
//   enviarMensajeWhatsApp('51957532973', '¡Hola! Este es un mensaje automático enviado cada minuto.');
// }, 60000);

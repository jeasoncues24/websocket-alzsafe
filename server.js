const WebSocket = require('ws');
const mysql = require('mysql2/promise');
require('dotenv').config();
const { Client, LocalAuth } = require('whatsapp-web.js');
const qrcode = require('qrcode-terminal');
const os = require('os');
const handleLocationEvent = require('./handleLocationEvent');

const serverPort = process.env.SERVER_PORT || 5000;


const wss = new WebSocket.Server({ port: serverPort });
const locationData = [];
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
console.log(`Servidor WebSocket activo en ws://${ipAddress}:${serverPort}`);


let db;
(async () => {
  db = await mysql.createConnection({
    // host: process.env.DB_HOST,
    // user: process.env.DB_USER,
    // password: process.env.DB_PASSWORD,
    // database: process.env.DB_NAME
    host: "161.132.45.25",
    user: "root",
    password: "74237028",
    database: "alzsafe_db"
  });
  console.log('✅ Conectado a la base de datos');
})();

const cuidadores = new Map();

wss.on('connection', (ws) => {
  const now = new Date();
  const connectionTime = `${now.toLocaleDateString()} ${now.toLocaleTimeString()}`;
  console.log(`🔗 Cliente conectado - ${connectionTime}`);

  ws.on('message', async (msg) => {
    try {
      const data = JSON.parse(msg);
      // 1. Registro de cuidadores
      if (data.type === 'init' && (data.userType === 'cuidador' || data.userType === 'familiar')) {
        // SE ESAT ENVIANDO EL ID DEL USUARIO ORIGINAL

        cuidadores.set(data.userId, ws);
        // console.log(`✅ Usuario ${data.userId} registrado`);
        (async () => {
          try {
            const [rows] = await db.execute(
              'SELECT phone FROM users WHERE id = ?',
              [data.userId]
            );

            if (rows.length > 0) {
              const phone = rows[0].phone;
              if (!phone) {
                console.log(`⚠️ No se encontró el número de teléfono para el cuidador ${data.userId}`);
                return;
              }
              enviarMensajeWhatsApp(phone, '¡Hola! Has iniciado sesión en la aplicación AlzSafe ❤️🙌. Tus pacientes te esperan, empezemos a trabajar.');
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
              `SELECT r.id AS request_id, u.phone AS telefono_cuidador,
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
              if (!solicitud.telefono_cuidador) {
                console.log(`⚠️ No se encontró el número de teléfono del cuidador ${cuidadorId}`);
                return;
              }
              enviarMensajeWhatsApp(solicitud.telefono_cuidador, mensaje);
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




      // 3. Procesar evento de ubicación y lógica de zona segura e inactividad
      if (data.event === 'location') {
        // const patientUserId = data.id;
        // const currentLat = parseFloat(data.lat);
        // const currentLng = parseFloat(data.lng);
        await handleLocationEvent({ data, db, wss, enviarWhatsappPaciente });
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
  authStrategy: new LocalAuth(),
  puppeteer: {
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  }
});
// Exportar el cliente de WhatsApp para usarlo en otros módulos

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
    if (!telefono) {
      console.log('⚠️ No se proporcionó un número de teléfono.');
      return;
    }
    if (isNaN(telefono) || /\s/.test(telefono)) {
      console.log('⚠️ El número de teléfono no es válido o contiene espacios.');
      return;
    }

    const phone = `51${parseInt(telefono)}`;
    const chatId = `${phone}@c.us`;
    await whatsappClient.sendMessage(chatId, mensaje);
    console.log(`Mensaje enviado a ${telefono} a las ${new Date().toLocaleTimeString()}`);
  } catch (error) {
    console.error('Error al enviar el mensaje de WhatsApp:', error);
  }
};



const enviarWhatsappPaciente = async (userId, mensaje, db) => {
  try {
    console.log(`📩 Enviando mensaje de WhatsApp al paciente con user_id ${userId}`);
    // Obtener el id del paciente desde la tabla patients según el user_id
    const [patientRows] = await db.execute(
      'SELECT uf.phone AS phone_familiar, uc.phone AS phone_cuidador, uc.id AS id_cuidador, req.familiar_id AS id_familiar, p.id AS id_paciente, p.name as nombre_paciente, cr.name as nombre_cuidador, uf.name as nombre_familiar FROM requests req INNER JOIN patients p ON p.id = req.patient_id INNER JOIN users uf ON uf.id = req.familiar_id INNER JOIN carer cr ON cr.id = req.carer_id INNER JOIN users uc ON uc.id = cr.user_id WHERE req.patient_id =( SELECT id FROM patients WHERE user_id = ?);',
      [userId]
    );

    if (patientRows.length === 0) {
      console.log(`⚠️ No se encontró un paciente con user_id ${userId}`);
      return;
    }
    const {
      phone_familiar,
      phone_cuidador,
      id_cuidador,
      id_familiar,
      id_paciente,
      nombre_paciente,
      nombre_cuidador,
      nombre_familiar
    } = patientRows[0];

    const idHistorial = await crearHistorialAlerta(db, id_paciente, id_familiar, id_cuidador);
    if (!idHistorial) return;
    const isEnviado = await validarDataWhatsapp(mensaje, nombre_paciente, nombre_cuidador, nombre_familiar, phone_familiar, phone_cuidador, whatsappClient);
    if (!isEnviado) {
      console.error(`El mensaje no se pudo enviar.💤`);
      // return;
    }
    await actualizarFechaWSFinal(db, idHistorial);
    console.log(`📩 [WhatsApp] Mensaje enviado correctamente ${nombre_paciente}`);
  } catch (error) {
    console.error('❌ Error al enviar el mensaje de WhatsApp:', error);
  }
};


const validarDataWhatsapp = async (mensaje, nombre_paciente, nombre_cuidador, nombre_familiar, phone_familiar, phone_cuidador) => {
  try {
    // Verificación de nombres
    if (!nombre_paciente || !nombre_cuidador || !nombre_familiar) {
      console.error('❌ Error: Uno o más nombres están vacíos o no definidos.');
      return;
    }

    // Verificación de números
    if (!phone_familiar || isNaN(phone_familiar)) {
      console.error('❌ Error: Número de teléfono del familiar inválido.');
      return;
    }

    if (!phone_cuidador || isNaN(phone_cuidador)) {
      console.error('❌ Error: Número de teléfono del cuidador inválido.');
      return;
    }

    const phoneFamiliar = `51${parseInt(phone_familiar)}@c.us`;
    const phonePaciente = `51${parseInt(phone_cuidador)}@c.us`;

    console.log(`📋 Detalles del mensaje:
- Paciente: ${nombre_paciente}
- Cuidador: ${nombre_cuidador}
- Familiar: ${nombre_familiar}
- Teléfono Familiar: ${phone_familiar}
- Teléfono Cuidador: ${phone_cuidador}`);
    // ENVIANDO AL FAMILIAR
    await whatsappClient.sendMessage(phoneFamiliar, mensaje);
    console.log(`✅ Mensaje enviado al familiar (${phone_familiar})`);
    // ENVIANDO AL PACIENTE
    await whatsappClient.sendMessage(phonePaciente, mensaje);
    console.log(`✅ Mensaje enviado al cuidador (${phone_cuidador})`);
    // RETURN TRUE
    return true;
  } catch (error) {
    console.error('❌ Error al enviar el mensaje de WhatsApp:', error);
    return false;
  }
};




const crearHistorialAlerta = async (db, idPaciente, idFamiliar, idCuidador) => {
  try {
    const insertQuery = `
            INSERT INTO historial_alertas (
                isError, 
                metrosError, 
                fechaWSInicio, 
                idPaciente, 
                idFamiliar, 
                idCuidador
            ) VALUES (1, 0, NOW(), ?, ?, ?)
        `;
    const values = [idPaciente, idFamiliar, idCuidador];
    const [result] = await db.execute(insertQuery, values);
    console.log('✅ Registro creado en historial_alertas con ID:', result.insertId);
    return result.insertId; // Necesario para actualizar luego
  } catch (error) {
    console.error('❌ Error al insertar historial de alerta:', error);
    return null;
  }
};

const actualizarFechaWSFinal = async (db, idHistorial) => {
  try {
    const updateQuery = `
            UPDATE historial_alertas
            SET fechaWSFinal = NOW()
            WHERE idAlerta = ?
        `;
    await db.execute(updateQuery, [idHistorial]);
    console.log(`🕒 fechaWSFinal actualizada para historial_alertas.id = ${idHistorial}`);
  } catch (error) {
    console.error('❌ Error al actualizar fechaWSFinal:', error);
  }
};

const WebSocket = require('ws');
const mysql = require('mysql2/promise');
require('dotenv').config();
const { Client, LocalAuth } = require('whatsapp-web.js');
const qrcode = require('qrcode-terminal');
const os = require('os');

const wss = new WebSocket.Server({ port: process.env.SERVER_PORT });
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

  ws.on('message', async (msg) => {
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




      // 3. Procesar evento de ubicación y lógica de zona segura e inactividad
      if (data.event === 'location') {
        const patientUserId = data.id;
        const currentLat = parseFloat(data.lat);
        const currentLng = parseFloat(data.lng);

        try {
          // 1. Buscar el paciente_id en la tabla patients
          const [patientRows] = await db.execute(
            'SELECT id FROM patients WHERE user_id = ?',
            [patientUserId]
          );

          if (patientRows.length > 0) {
            const pacienteId = patientRows[0].id;
            const nombrePaciente = patientRows[0].name;

            // 2. Buscar la configuración de zona segura para este paciente
            const [zonaSeguraRows] = await db.execute(
              'SELECT is_zona_segura, intervalo_notificaciones, intervalo_inactividad, radio_proteccion, lat_default, log_default FROM zona_segura WHERE paciente_id = ?',
              [pacienteId]
            );

            if (zonaSeguraRows.length > 0) {
              const zonaSeguraConfig = zonaSeguraRows[0];

              if (zonaSeguraConfig.is_zona_segura === 1) {
                const defaultLat = parseFloat(zonaSeguraConfig.lat_default);
                const defaultLng = parseFloat(zonaSeguraConfig.log_default);
                const radioProteccion = parseFloat(zonaSeguraConfig.radio_proteccion);
                const intervaloNotificaciones = parseInt(zonaSeguraConfig.intervalo_notificaciones) * 60 * 1000; // Convertir a milisegundos
                const intervaloInactividad = parseInt(zonaSeguraConfig.intervalo_inactividad) * 60 * 1000; // Convertir a milisegundos

                const distance = calculateDistance(currentLat, currentLng, defaultLat, defaultLng);
                const isInside = distance <= radioProteccion;

                // Enviar evento de ubicación actualizada a los clientes WebSocket
                wss.clients.forEach(client => {
                  if (client.readyState === WebSocket.OPEN) {
                    client.send(JSON.stringify({
                      event: 'patient-location-update',
                      patientId: pacienteId,
                      latitude: currentLng,
                      longitude: currentLat,
                      isInsideSafeZone: isInside
                    }));
                  }
                });

                // Lógica de salida de zona segura
                if (!isInside) {
                  const lastNotificationTime = patientLastMovedTime.get(pacienteId) || 0;

                  // Si el paciente ha salido de la zona segura y ha pasado el intervalo de notificación
                  if (Date.now() - lastNotificationTime >= intervaloNotificaciones) {
                    enviarMensajeWhatsAppPaciente(
                      patientUserId,
                      `🚨 *Alerta de Seguridad* 🚨\n\n👤 Tu familiar *${nombrePaciente}* ha salido de la 🛡️ *zona segura* (📏 radio de *${radioProteccion} metros*).\n\n📍 *Ubicación actual:*\n📌 Lat: ${currentLat.toFixed(4)}\n📌 Lng: ${currentLng.toFixed(4)}\n\n📲 Para más información, revisa la app AlzSafe.`
                    );

                    patientLastMovedTime.set(pacienteId, Date.now());
                    // Reiniciar el timer de inactividad si el paciente se mueve fuera de la zona
                    if (patientInactiveIntervalTimers.has(pacienteId)) {
                      clearTimeout(patientInactiveIntervalTimers.get(pacienteId));
                      patientInactiveIntervalTimers.delete(pacienteId);
                    }
                  }
                } else {
                  // Si el paciente regresa a la zona segura, resetear el tiempo de la última notificación de salida
                  patientLastMovedTime.delete(pacienteId);

                }

                // Lógica de inactividad
                const lastKnownLocation = patientLastLocation.get(pacienteId);
                const lastMoveTime = patientLastMovedTime.get(pacienteId) || Date.now(); // Si nunca salió, tomamos el tiempo actual

                if (lastKnownLocation && lastKnownLocation.lat === currentLat && lastKnownLocation.lng === currentLng) {
                  // La ubicación no ha cambiado
                  if (!patientInactiveIntervalTimers.has(pacienteId) && Date.now() - lastMoveTime >= intervaloInactividad) {
                    // Iniciar timer de inactividad si no existe y ha pasado el intervalo
                    const timer = setTimeout(() => {
                      enviarMensajeWhatsAppPaciente(
                        patientUserId,
                        `😌 *Todo en calma*\n\n🧘‍♂️ Tu familiar *${nombrePaciente}* parece estar tranquilo en la misma ubicación durante *${zonaSeguraConfig.intervalo_inactividad} minutos*.\n\n📍 *Ubicación actual:*\n📌 Lat: ${currentLat.toFixed(4)}\n📌 Lng: ${currentLng.toFixed(4)}`
                      );

                      patientInactiveIntervalTimers.delete(pacienteId); // Limpiar el timer después de enviar la notificación
                    }, intervaloInactividad);
                    patientInactiveIntervalTimers.set(pacienteId, timer);
                  }
                } else {
                  // La ubicación ha cambiado, resetear el timer de inactividad
                  if (patientInactiveIntervalTimers.has(pacienteId)) {
                    clearTimeout(patientInactiveIntervalTimers.get(pacienteId));
                    patientInactiveIntervalTimers.delete(pacienteId);
                  }
                  patientLastMovedTime.set(pacienteId, Date.now()); // Actualizar el tiempo de movimiento
                }

                patientLastLocation.set(pacienteId, { lat: currentLat, lng: currentLng });

              } else {
                // Zona segura desactivada, enviar notificación cada intervalo
                if (!patientNotificationIntervalTimers.has(pacienteId)) {
                  const interval = setInterval(() => {
                    enviarMensajeWhatsAppPaciente(patientUserId, `ℹ️ La zona segura para el paciente con ID ${pacienteId} está desactivada. Actívala para recibir notificaciones automáticas sobre su ubicación.`);
                  }, parseInt(zonaSeguraConfig.intervalo_notificaciones) * 60 * 1000);
                  patientNotificationIntervalTimers.set(pacienteId, interval);
                }
              }
            } else {
              console.log(`⚠️ No se encontró configuración de zona segura para el paciente con ID ${pacienteId}`);
              // Si no hay configuración, podrías enviar solo la actualización de ubicación
              wss.clients.forEach(client => {
                if (client.readyState === WebSocket.OPEN) {
                  client.send(JSON.stringify({
                    event: 'patient-location-update',
                    patientId: pacienteId,
                    latitude: currentLng,
                    longitude: currentLat,
                    isInsideSafeZone: null // Indicar que no hay zona segura configurada
                  }));
                }
              });
            }
          } else {
            console.log(`⚠️ No se encontró paciente con user_id ${patientUserId}`);
          }
        } catch (error) {
          console.error('❌ Error al procesar la ubicación:', error);
        }
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
    const chatId = `51${telefono}@c.us`;
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


const enviarMensajeWhatsAppPaciente = async (userId, mensaje) => {
  try {
    console.log(`📩 Enviando mensaje de WhatsApp al paciente con user_id ${userId}`);
    // Obtener el id del paciente desde la tabla patients según el user_id
    const [patientRows] = await db.execute(
      'SELECT id FROM patients WHERE user_id = ?',
      [userId]
    );

    if (patientRows.length === 0) {
      console.log(`⚠️ No se encontró un paciente con user_id ${userId}`);
      return;
    }

    const patientId = patientRows[0].id;
    console.log(`📩 Enviando mensaje de WhatsApp al paciente con patientId ${patientId}`);

    // Obtener el teléfono del familiar desde la tabla family_members según patient_id
    const [familyRows] = await db.execute(
      'SELECT phone, name FROM family_members WHERE patient_id = ?',
      [patientId]
    );

    if (familyRows.length === 0) {
      console.log(`⚠️ No se encontró un familiar para el paciente con id ${patientId}`);
      return;
    }

    const telefono = familyRows[0].phone;
    const nombrePaciente = familyRows[0].name;

    // Enviar el mensaje de WhatsApp al número del familiar
    const chatId = `51${telefono}@c.us`;
    await whatsappClient.sendMessage(chatId, mensaje);

    console.log(`📩 [WhatsApp] Mensaje enviado al familiar ${nombrePaciente} (user_id: ${userId}) al número ${telefono}`);
  } catch (error) {
    console.error('❌ Error al enviar el mensaje de WhatsApp:', error);
  }
};



const patientLastLocation = new Map(); // Para rastrear la última ubicación de cada paciente
const patientLastMovedTime = new Map(); // Para rastrear la última vez que se movió un paciente
const patientInactiveIntervalTimers = new Map(); // Para rastrear los timers de inactividad
const patientNotificationIntervalTimers = new Map(); // Para rastrear los timers de notificación de zona segura desactivada


// Función para calcular la distancia entre dos coordenadas (en metros)
const calculateDistance = (lat1, lon1, lat2, lon2) => {
  const R = 6371e3; // Radio de la Tierra en metros
  const φ1 = lat1 * Math.PI / 180; // φ, λ en radianes
  const φ2 = lat2 * Math.PI / 180;
  const Δφ = (lat2 - lat1) * Math.PI / 180;
  const Δλ = (lon2 - lon1) * Math.PI / 180;

  const a = Math.sin(Δφ / 2) * Math.sin(Δφ / 2) +
    Math.cos(φ1) * Math.cos(φ2) *
    Math.sin(Δλ / 2) * Math.sin(Δλ / 2);
  const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));

  return R * c; // Distancia en metros
};



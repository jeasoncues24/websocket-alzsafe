const WebSocket = require('ws');

function calculateDistance(lat1, lng1, lat2, lng2) {
    const R = 6371e3; // Radio de la Tierra en metros
    const toRad = deg => (deg * Math.PI) / 180;

    const dLat = toRad(lat2 - lat1);
    const dLng = toRad(lng2 - lng1);
    const a =
        Math.sin(dLat / 2) ** 2 +
        Math.cos(toRad(lat1)) * Math.cos(toRad(lat2)) *
        Math.sin(dLng / 2) ** 2;

    return R * 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
}

/**
 * Maneja un evento de ubicación del paciente.
 * 
 * @param {object} data - Datos del mensaje recibido del WebSocket
 * @param {object} db - Conexión a la base de datos (mysql2)
 * @param {WebSocket.Server} wss - Instancia del WebSocket server
 * @param {Map} patientLastMovedTime
 * @param {Map} patientInactiveIntervalTimers
 * @param {Map} patientLastLocation
 * @param {Map} patientNotificationIntervalTimers
 * @param {Function} enviarMensajeWhatsAppPaciente - Función para enviar mensajes
 */



const patientLastLocation = new Map(); // Para rastrear la última ubicación de cada paciente
const patientLastMovedTime = new Map(); // Para rastrear la última vez que se movió un paciente
const patientInactiveIntervalTimers = new Map(); // Para rastrear los timers de inactividad
const patientNotificationIntervalTimers = new Map(); // Para rastrear los timers de notificación de zona segura desactivada

module.exports = async function handleLocationEvent({
    data,
    db,
    wss,
    whatsappClient
}) {

    if (whatsappClient === undefined) {
        console.error('❌ WhatsApp client no está definido. Asegúrate de que el cliente esté inicializado correctamente.');
        return;
    }

    try {
        const patientUserId = data.id;
        const currentLat = parseFloat(data.lat);
        const currentLng = parseFloat(data.lng);
        const [patientRows] = await db.execute(
            'SELECT id, name FROM patients WHERE user_id = ?',
            [patientUserId]
        );

        if (patientRows.length === 0) return console.warn(`⚠️ No se encontró paciente con user_id ${patientUserId}`);

        const { id: pacienteId, name: nombrePaciente } = patientRows[0];

        const [zonaRows] = await db.execute(
            'SELECT * FROM zona_segura WHERE paciente_id = ?',
            [pacienteId]
        );

        if (zonaRows.length === 0) {
            console.warn(`⚠️ No hay configuración de zona segura para paciente ID ${pacienteId}`);
            enviarwspaciente(patientUserId, `ℹ️ Contactar con tu administrador de sistema este paciente no tiene configuración de zona segura`, db, whatsappClient);
            broadcastLocation(wss, pacienteId, currentLat, currentLng, false, false, 0, 0, 0, 0);
            return;
        }

        const config = zonaRows[0];
        const isZonaActiva = parseInt(config.is_zona_segura) === 1;

        const intervaloNotificaciones = parseInt(config.intervalo_notificaciones) * 60 * 1000;
        const intervaloInactividad = parseInt(config.intervalo_inactividad) * 60 * 1000;
        const radio = parseFloat(config.radio_proteccion);
        const defaultLat = parseFloat(config.lat_default);
        const defaultLng = parseFloat(config.log_default);

        const distancia = calculateDistance(currentLat, currentLng, defaultLat, defaultLng);
        const dentroZona = parseFloat(distancia) <= radio;
        console.log(`📍 Distancia a la zona segura: ${distancia.toFixed(2)} m el radio configurado es ${radio}. isDentro: ${dentroZona}`);

        broadcastLocation(wss, pacienteId, currentLat, currentLng, dentroZona, dentroZona, distancia, radio, defaultLng, defaultLat);

        if (!isZonaActiva) {
            if (!patientNotificationIntervalTimers.has(pacienteId)) {
                const interval = setInterval(() => {
                    enviarwspaciente(patientUserId,
                        `ℹ️ La zona segura para *${nombrePaciente}* está desactivada. Actívala para recibir notificaciones automáticas.`, db, whatsappClient
                    );
                }, intervaloNotificaciones);
                patientNotificationIntervalTimers.set(pacienteId, interval);
            }
            return;
        }

        // Lógica de salida de zona segura
        if (!dentroZona) {
            const ultimaNotificacion = patientLastMovedTime.get(pacienteId) || 0;
            if (Date.now() - ultimaNotificacion >= intervaloNotificaciones) {
                enviarwspaciente(
                    patientUserId,
                    `🚨 *Alerta de Seguridad* 🚨\n\n👤 Tu familiar *${nombrePaciente}* ha salido de la 🛡️ *zona segura* (📏 radio de *${radio} metros*).\n\n📍 *Ubicación actual:*\n📌 Lat: ${currentLat.toFixed(4)}\n📌 Lng: ${currentLng.toFixed(4)}`,
                    db, whatsappClient
                );
                patientLastMovedTime.set(pacienteId, Date.now());
                clearTimeoutIfExists(patientInactiveIntervalTimers, pacienteId);
            }
        } else {
            patientLastMovedTime.delete(pacienteId);
        }

        // Lógica de inactividad
        const ultimaUbicacion = patientLastLocation.get(pacienteId);
        const ultimoMovimiento = patientLastMovedTime.get(pacienteId) || Date.now();

        if (ultimaUbicacion && ultimaUbicacion.lat === currentLat && ultimaUbicacion.lng === currentLng) {
            if (!patientInactiveIntervalTimers.has(pacienteId) && Date.now() - ultimoMovimiento >= intervaloInactividad) {
                const timer = setTimeout(() => {
                    enviarwspaciente(
                        patientUserId,
                        `😌 *Todo en calma*\n\n🧘‍♂️ Tu familiar *${nombrePaciente}* parece estar tranquilo en la misma ubicación durante *${config.intervalo_inactividad} minutos*.\n\n📍 Lat: ${currentLat.toFixed(4)}\n📍 Lng: ${currentLng.toFixed(4)}`,
                        db, whatsappClient
                    );
                    patientInactiveIntervalTimers.delete(pacienteId);
                }, intervaloInactividad);
                patientInactiveIntervalTimers.set(pacienteId, timer);
            }
        } else {
            clearTimeoutIfExists(patientInactiveIntervalTimers, pacienteId);
            patientLastMovedTime.set(pacienteId, Date.now());
        }

        patientLastLocation.set(pacienteId, { lat: currentLat, lng: currentLng });
    } catch (error) {
        console.error('❌ Error al procesar ubicación del paciente:', error);
    }

};





const enviarwspaciente = async (userId, mensaje, db, whatsappClient) => {
    try {
        console.log(`📩 Enviando mensaje de WhatsApp al paciente con user_id ${userId}`);
        // Obtener el id del paciente desde la tabla patients según el user_id
        const [patientRows] = await db.execute(
            'SELECT uf.phone AS phone_familiar, uc.phone AS phone_cuidador, uc.id AS id_cuidador, req.familiar_id AS id_familiar, p.id AS id_paciente, p.name as nombre_paciente, cr.name as nombre_cuidador, uf.name as nombre_familiar FROM requests req INNER JOIN patients p ON p.id = req.patient_id AND p.user_id = 6 INNER JOIN users uf ON uf.id = req.familiar_id INNER JOIN carer cr ON cr.id = req.carer_id INNER JOIN users uc ON uc.id = cr.user_id WHERE req.patient_id =( SELECT id FROM patients WHERE user_id = ?);',
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
        const isEnviado = await enviarMensajeWhatsApp(mensaje, nombre_paciente, nombre_cuidador, nombre_familiar, phone_familiar, phone_cuidador, whatsappClient);
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


const enviarMensajeWhatsApp = async (mensaje, nombre_paciente, nombre_cuidador, nombre_familiar, phone_familiar, phone_cuidador, whatsappClient) => {
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

// ---- Funciones auxiliares ----

function broadcastLocation(wss, pacienteId, lat, lng, isInside, isDentroZona = false, distanciaActual = 0, distanciaZona = 0, lngOriginal = 0, latOriginal = 0) {
    wss.clients.forEach(client => {
        if (client.readyState === WebSocket.OPEN) {
            client.send(JSON.stringify({
                event: 'patient-location-update',
                patientId: pacienteId,
                latitude: lng,
                longitude: lat,
                isInsideSafeZone: isInside,
                timestamp: new Date().toISOString(),
                distanciaActual: distanciaActual,
                distanciaZona: distanciaZona,
                isDentroZona: isDentroZona,
                lngOriginal: latOriginal,
                latOriginal: lngOriginal
            }));
        }
    });
}

function clearTimeoutIfExists(map, key) {
    if (map.has(key)) {
        clearTimeout(map.get(key));
        map.delete(key);
    }
}
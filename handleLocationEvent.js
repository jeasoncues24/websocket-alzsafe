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
    enviarWhatsappPaciente
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
            enviarWhatsappPaciente(patientUserId, `ℹ️ Contactar con tu administrador de sistema este paciente no tiene configuración de zona segura`, db);
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
                    enviarWhatsappPaciente(patientUserId,
                        `ℹ️ La zona segura para *${nombrePaciente}* está desactivada. Actívala para recibir notificaciones automáticas.`, db
                    );
                }, intervaloNotificaciones);
                patientNotificationIntervalTimers.set(pacienteId, interval);
            }
            return;
        }

        // Lógica de salida de zona segura
        if (!dentroZona) {
            if (!patientNotificationIntervalTimers.has(pacienteId)) {
                // Enviar inmediatamente
                enviarWhatsappPaciente(
                    patientUserId,
                    `🚨 *Alerta de Seguridad* 🚨\n\n👤 Tu familiar *${nombrePaciente}* ha salido de la 🛡️ *zona segura* (📏 radio de *${radio} metros*).\n\n📍 *Ubicación actual:*\n📌 Lat: ${currentLat.toFixed(4)}\n📌 Lng: ${currentLng.toFixed(4)}`,
                    db
                );

                // Crear un temporizador que mande cada 5 minutos
                const interval = setInterval(() => {
                    enviarWhatsappPaciente(
                        patientUserId,
                        `🔁 *Seguimiento de ubicación* 🔁\n\n👤 *${nombrePaciente}* sigue fuera de la 🛡️ zona segura.\n\n📍 Lat: ${currentLat.toFixed(4)}\n📍 Lng: ${currentLng.toFixed(4)}`,
                        db
                    );
                }, intervaloNotificaciones);
                patientNotificationIntervalTimers.set(pacienteId, interval);
            }
        } else {
            // Si entra a la zona segura, limpiar el temporizador
            if (patientNotificationIntervalTimers.has(pacienteId)) {
                clearInterval(patientNotificationIntervalTimers.get(pacienteId));
                patientNotificationIntervalTimers.delete(pacienteId);
            }
        }

        patientLastLocation.set(pacienteId, { lat: currentLat, lng: currentLng });
    } catch (error) {
        console.error('❌ Error al procesar ubicación del paciente:', error);
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
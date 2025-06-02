"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.sendMessageDirect = void 0;
const wa_client_1 = require("../../utils/wa-client");
const whatsapp_web_js_1 = require("whatsapp-web.js");
const wa_model_1 = require("../models/wa.model");
const user_model_1 = require("../models/user.model");
const upload_message_1 = __importDefault(require("../../middleware/upload.message"));
const fs_1 = __importDefault(require("fs"));
const path_1 = __importDefault(require("path"));
const sendMessageDirect = async (req, res) => {
    upload_message_1.default.single("image")(req, res, async (err) => {
        try {
            if (err) {
                console.error("Error al subir la imagen:", err.message);
                return res
                    .status(400)
                    .json({ error: "Hubo un conflicto para subir la imagen." });
            }
            const { rucEmpresa, codigo_postal_receptor, telefono_receptor, message } = req.body;
            if (!rucEmpresa ||
                !codigo_postal_receptor ||
                !telefono_receptor ||
                !message) {
                return res.status(400).json({
                    message: "Faltan datos requeridos: rucEmpresa, codigo_postal_receptor, telefono_receptor o message.",
                });
            }
            if (!/^\d+$/.test(codigo_postal_receptor) ||
                !/^\d+$/.test(telefono_receptor)) {
                return res.status(400).json({
                    message: "El código postal y teléfono deben contener solo números.",
                });
            }
            const waClient = wa_client_1.listUserActiveClientWhatsapp.get(rucEmpresa);
            if (!waClient || !waClient.info?.wid) {
                return res.status(404).json({
                    message: "Cliente no encontrado o no está listo.",
                });
            }
            const client = waClient;
            const chatid = `${codigo_postal_receptor}${telefono_receptor}@c.us`;
            let media;
            if (req.file) {
                if (!req.file || !req.file.filename) {
                    return res
                        .status(400)
                        .json({ message: "No se subió ninguna imagen." });
                }
                const imagePath = req.file.filename;
                const fullPath = path_1.default.join(__dirname, "../../uploads/messages", imagePath);
                const stat = fs_1.default.statSync(fullPath);
                if (stat.isDirectory()) {
                    return res
                        .status(400)
                        .json({ message: "La ruta es un directorio, no un archivo." });
                }
                if (fs_1.default.existsSync(fullPath)) {
                    console.log("Intentando leer archivo en:", fullPath);
                    console.log("req.file:", req.file);
                    const imageBuffer = fs_1.default.readFileSync(fullPath);
                    const mimeType = req.file.mimetype;
                    const base64Image = imageBuffer.toString("base64");
                    media = new whatsapp_web_js_1.MessageMedia(mimeType, base64Image, req.file.originalname);
                }
                else {
                    return res.status(404).json({
                        message: "Archivo de imagen no encontrado.",
                    });
                }
            }
            const messagewa = await client.sendMessage(chatid, message, media ? { media } : undefined);
            // Guardar en la base de datos
            try {
                const dataUser = await (0, user_model_1.getUserByIdModel)(rucEmpresa);
                const msg = {
                    codUsuario: dataUser?.id,
                    codigo_postal_receptor: parseInt(codigo_postal_receptor),
                    telefono_receptor,
                    message,
                    timestamp: messagewa.timestamp,
                };
                await (0, wa_model_1.insertMessageInDatabase)(msg);
            }
            catch (error) {
                console.error("Error al insertar en la base de datos:", error);
            }
            return res.status(200).json({
                message: `Mensaje enviado correctamente al número ${telefono_receptor}.`,
            });
        }
        catch (error) {
            return res.status(500).json({
                message: error instanceof Error
                    ? error.message
                    : "Ocurrió un error al enviar el mensaje directo.",
                errorType: error instanceof Error ? error.name : typeof error,
                stack: error instanceof Error ? error.stack : null,
                raw: error,
            });
        }
    });
};
exports.sendMessageDirect = sendMessageDirect;

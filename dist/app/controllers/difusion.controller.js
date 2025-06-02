"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.sendMessageDifusion = void 0;
const wa_client_1 = require("../../utils/wa-client");
const whatsapp_web_js_1 = require("whatsapp-web.js");
const wa_model_1 = require("../models/wa.model");
const user_model_1 = require("../models/user.model");
const upload_message_1 = __importDefault(require("../../middleware/upload.message"));
const fs_1 = __importDefault(require("fs"));
const path_1 = __importDefault(require("path"));
const ALLOWED_MIME_TYPES = [
    "image/png",
    "image/jpeg",
    "image/jpg",
    "application/pdf",
    "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", // .xlsx
    "audio/mpeg", // .mp3
];
// const sendMessageDifusion = async (req: Request, res: Response) => {
//   try {
//     uploadMessage.single("image")(req, res, async (err) => {
//       if (err) {
//         return res.status(400).json({ error: err.message });
//       }
//       try {
//         const usersIterator = listUserActiveClientWhatsapp.keys();
//         const users = Array.from(usersIterator);
//         console.warn("Users:", users);
//         const {
//           rucEmpresa,
//           lista_difusion,
//           message,
//         }: {
//           rucEmpresa: string;
//           lista_difusion: string;
//           message: string;
//         } = req.body;
//         let listaDifusion: DifusionItem[];
//         try {
//           const parsed = JSON.parse(lista_difusion);
//           if (!Array.isArray(parsed)) {
//             return res.status(400).json({
//               message: "lista_difusion debe ser un array de objetos.",
//             });
//           }
//           listaDifusion = parsed;
//         } catch (e) {
//           return res.status(400).json({
//             message:
//               "lista_difusion debe ser un JSON válido con telefono y codigo_postal.",
//           });
//         }
//         if (!rucEmpresa || !lista_difusion || !message) {
//           return res.status(400).json({
//             message:
//               "Faltan datos requeridos: rucEmpresa, lista_difusion o message.",
//           });
//         }
//         if (!listUserActiveClientWhatsapp.has(rucEmpresa)) {
//           return res.status(404).json({
//             message: "Cliente no encontrado para el RUC proporcionado.",
//           });
//         }
//         const waClient = listUserActiveClientWhatsapp.get(rucEmpresa)!;
//         const client: Client = waClient;
//         if (!client.info || !client.info.wid) {
//           throw new Error("El cliente de WhatsApp aún no está listo.");
//         }
//         let media: MessageMedia | undefined;
//         if (req.file) {
//           const imagePath = req.file.filename;
//           const fullPath = path.join(
//             __dirname,
//             "../../uploads/messages",
//             imagePath
//           );
//           if (!fs.existsSync(fullPath)) {
//             return res.status(404).json({
//               message: "No se encontró la imagen proporcionada.",
//             });
//           }
//           const imageBuffer = fs.readFileSync(fullPath);
//           const base64Image = imageBuffer.toString("base64");
//           media = new MessageMedia("image/png", base64Image);
//         }
//         const dataUser = await getUserByIdModel(rucEmpresa);
//         let results: { telefono: string; status: string; error?: string }[] =
//           [];
//         for (const entry of listaDifusion) {
//           const { telefono, codigo_postal } = entry;
//           if (!/^\d+$/.test(codigo_postal) || !/^\d+$/.test(telefono)) {
//             results.push({
//               telefono,
//               status: "error",
//               error: "Teléfono o código postal inválido.",
//             });
//             continue;
//           }
//           const chatid = `${codigo_postal}${telefono}@c.us`;
//           try {
//             const messagewa = await client.sendMessage(
//               chatid,
//               message,
//               media ? { media } : {}
//             );
//             results.push({ telefono, status: "enviado" });
//             // Guardar en base de datos
//             const msg: Message = {
//               codUsuario: dataUser?.id!,
//               codigo_postal_receptor: parseInt(codigo_postal),
//               telefono_receptor: telefono,
//               message,
//               timestamp: messagewa.timestamp,
//             };
//             await insertMessageInDatabase(msg);
//           } catch (err: any) {
//             results.push({
//               telefono,
//               status: "error",
//               error: err?.message || "Error al enviar el mensaje",
//             });
//           }
//         }
//         return res.status(200).json({
//           message: "Difusión finalizada.",
//           resultados: results,
//         });
//       } catch (error) {
//         return res.status(500).json({
//           message:
//             error instanceof Error
//               ? error.message
//               : "Ocurrió un error al enviar el mensaje directo.",
//           errorType: error instanceof Error ? error.name : typeof error,
//           stack: error instanceof Error ? error.stack : null,
//           raw: error,
//         });
//       }
//     });
//   } catch (error) {
//     return res.status(500).json({
//       message:
//         error instanceof Error
//           ? error.message
//           : "Ocurrió un error al enviar el mensaje directo.",
//       errorType: error instanceof Error ? error.name : typeof error,
//       stack: error instanceof Error ? error.stack : null,
//       raw: error,
//     });
//   }
// };
const sendMessageDifusion = async (req, res) => {
    upload_message_1.default.single("image")(req, res, async (err) => {
        if (err) {
            return res.status(400).json({ error: err.message });
        }
        try {
            const { rucEmpresa, lista_difusion, message, } = req.body;
            if (!rucEmpresa || !lista_difusion || !message) {
                return res.status(400).json({
                    message: "Faltan datos requeridos: rucEmpresa, lista_difusion o message.",
                });
            }
            // Validación de JSON
            let listaDifusion;
            try {
                const parsed = JSON.parse(lista_difusion);
                if (!Array.isArray(parsed)) {
                    return res.status(400).json({
                        message: "lista_difusion debe ser un array de objetos.",
                    });
                }
                listaDifusion = parsed;
            }
            catch (e) {
                return res.status(400).json({
                    message: "lista_difusion debe ser un JSON válido con telefono y codigo_postal.",
                });
            }
            // Verifica si el cliente de WhatsApp está activo
            const waClient = wa_client_1.listUserActiveClientWhatsapp.get(rucEmpresa);
            if (!waClient) {
                return res.status(404).json({
                    message: "Cliente no encontrado para el RUC proporcionado.",
                });
            }
            const client = waClient;
            if (!client.info || !client.info.wid) {
                throw new Error("El cliente de WhatsApp aún no está listo.");
            }
            // Procesar archivo adjunto si existe
            let media;
            if (req.file) {
                const imagePath = req.file.filename;
                const fullPath = path_1.default.join(__dirname, "../../uploads/messages", imagePath);
                if (!fs_1.default.existsSync(fullPath)) {
                    return res
                        .status(404)
                        .json({ message: "Archivo adjunto no encontrado." });
                }
                const fileBuffer = fs_1.default.readFileSync(fullPath);
                const mimeType = req.file.mimetype; // ej: 'application/pdf', 'image/png', etc.
                const base64File = fileBuffer.toString("base64");
                media = new whatsapp_web_js_1.MessageMedia(mimeType, base64File, req.file.originalname);
            }
            const dataUser = await (0, user_model_1.getUserByIdModel)(rucEmpresa);
            const results = [];
            for (const entry of listaDifusion) {
                const { telefono, codigo_postal } = entry;
                // Validar formato
                if (!/^\d+$/.test(codigo_postal) || !/^\d+$/.test(telefono)) {
                    results.push({
                        telefono,
                        status: "error",
                        error: "Teléfono o código postal inválido.",
                    });
                    continue;
                }
                const chatid = `${codigo_postal}${telefono}@c.us`;
                try {
                    const messagewa = await client.sendMessage(chatid, media ? message : message, media ? { media } : undefined);
                    results.push({ telefono, status: "enviado" });
                    const msg = {
                        codUsuario: dataUser?.id,
                        codigo_postal_receptor: parseInt(codigo_postal),
                        telefono_receptor: telefono,
                        message,
                        timestamp: messagewa.timestamp,
                    };
                    await (0, wa_model_1.insertMessageInDatabase)(msg);
                }
                catch (err) {
                    results.push({
                        telefono,
                        status: "error",
                        error: err?.message || "Error al enviar el mensaje",
                    });
                }
            }
            return res.status(200).json({
                message: "Difusión finalizada.",
                resultados: results,
            });
        }
        catch (error) {
            return res.status(500).json({
                message: error instanceof Error
                    ? error.message
                    : "Ocurrió un error al procesar la difusión.",
                errorType: error instanceof Error ? error.name : typeof error,
                stack: error instanceof Error ? error.stack : null,
                raw: error,
            });
        }
    });
};
exports.sendMessageDifusion = sendMessageDifusion;

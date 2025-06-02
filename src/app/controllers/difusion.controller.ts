import { Request, Response } from "express";
import { listUserActiveClientWhatsapp } from "../../utils/wa-client";
import { Client, MessageMedia } from "whatsapp-web.js";
import { insertMessageInDatabase } from "../models/wa.model";
import { Message } from "../../interfaces/message.interface";
import { getUserByIdModel } from "../models/user.model";
import uploadMessage from "../../middleware/upload.message";
import fs from "fs";
import path from "path";
import { DifusionItem } from "../../interfaces/difusion.interface";

const ALLOWED_MIME_TYPES = [
  "image/png",
  "image/jpeg",
  "image/jpg",
  "application/pdf",
  "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", // .xlsx
  "audio/mpeg", // .mp3
];

const sendMessageDifusion = async (req: Request, res: Response) => {
  uploadMessage.single("image")(req, res, async (err) => {
    if (err) {
      return res.status(400).json({ error: err.message });
    }

    try {
      const {
        rucEmpresa,
        lista_difusion,
        message,
      }: {
        rucEmpresa: string;
        lista_difusion: string;
        message: string;
      } = req.body;

      if (!rucEmpresa || !lista_difusion || !message) {
        return res.status(400).json({
          message:
            "Faltan datos requeridos: rucEmpresa, lista_difusion o message.",
        });
      }

      // Validación de JSON
      let listaDifusion: DifusionItem[];
      try {
        const parsed = JSON.parse(lista_difusion);
        if (!Array.isArray(parsed)) {
          return res.status(400).json({
            message: "lista_difusion debe ser un array de objetos.",
          });
        }
        listaDifusion = parsed;
      } catch (e) {
        return res.status(400).json({
          message:
            "lista_difusion debe ser un JSON válido con telefono y codigo_postal.",
        });
      }

      // Verifica si el cliente de WhatsApp está activo
      const waClient = listUserActiveClientWhatsapp.get(rucEmpresa);
      if (!waClient) {
        return res.status(404).json({
          message: "Cliente no encontrado para el RUC proporcionado.",
        });
      }

      const client: Client = waClient;
      if (!client.info || !client.info.wid) {
        throw new Error("El cliente de WhatsApp aún no está listo.");
      }

      // Procesar archivo adjunto si existe
      let media: MessageMedia | undefined;
      if (req.file) {
        const imagePath = req.file.filename;
        const fullPath = path.join(
          __dirname,
          "../../uploads/messages",
          imagePath
        );

        if (!fs.existsSync(fullPath)) {
          return res
            .status(404)
            .json({ message: "Archivo adjunto no encontrado." });
        }

        const fileBuffer = fs.readFileSync(fullPath);
        const mimeType = req.file.mimetype; // ej: 'application/pdf', 'image/png', etc.
        const base64File = fileBuffer.toString("base64");

        media = new MessageMedia(mimeType, base64File, req.file.originalname);
      }

      const dataUser = await getUserByIdModel(rucEmpresa);

      const results: { telefono: string; status: string; error?: string }[] =
        [];

      for (const entry of listaDifusion) {
        const { telefono, codigo_postal, mensaje_difusion } = entry;

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
          const messagewa = await client.sendMessage(
            chatid,
            // media ? message : message,
            mensaje_difusion,
            media ? { media } : undefined
          );

          results.push({ telefono, status: "enviado" });

          const msg: Message = {
            codUsuario: dataUser?.id!,
            codigo_postal_receptor: parseInt(codigo_postal),
            telefono_receptor: telefono,
            // message,
            message: mensaje_difusion,
            timestamp: messagewa.timestamp,
          };

          await insertMessageInDatabase(msg);
        } catch (err: any) {
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
    } catch (error) {
      return res.status(500).json({
        message:
          error instanceof Error
            ? error.message
            : "Ocurrió un error al procesar la difusión.",
        errorType: error instanceof Error ? error.name : typeof error,
        stack: error instanceof Error ? error.stack : null,
        raw: error,
      });
    }
  });
};


export { sendMessageDifusion };

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

const sendMessageDifusion = async (req: Request, res: Response) => {
  try {
    uploadMessage.single("image")(req, res, async (err) => {
      if (err) {
        return res.status(400).json({ error: err.message });
      }

      const usersIterator = listUserActiveClientWhatsapp.keys();
      const users = Array.from(usersIterator);
      console.warn("Users:", users);

      const {
        rucEmpresa,
        lista_difusion,
        message,
      }: {
        rucEmpresa: string;
        lista_difusion: string;
        message: string;
      } = req.body;

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

      if (!rucEmpresa || !lista_difusion || !message) {
        return res.status(400).json({
          message:
            "Faltan datos requeridos: rucEmpresa, lista_difusion o message.",
        });
      }

      if (!listUserActiveClientWhatsapp.has(rucEmpresa)) {
        return res.status(404).json({
          message: "Cliente no encontrado para el RUC proporcionado.",
        });
      }

      const waClient = listUserActiveClientWhatsapp.get(rucEmpresa)!;
      const client: Client = waClient;

      if (!client.info || !client.info.wid) {
        throw new Error("El cliente de WhatsApp aún no está listo.");
      }

      let media: MessageMedia | undefined;
      if (req.file) {
        const imagePath = req.file.filename;
        const fullPath = path.join(
          __dirname,
          "../../uploads/messages",
          imagePath
        );

        if (!fs.existsSync(fullPath)) {
          return res.status(404).json({
            message: "No se encontró la imagen proporcionada.",
          });
        }

        const imageBuffer = fs.readFileSync(fullPath);
        const base64Image = imageBuffer.toString("base64");
        media = new MessageMedia("image/png", base64Image);
      }

      const dataUser = await getUserByIdModel(rucEmpresa);

      let results: { telefono: string; status: string; error?: string }[] = [];

      for (const entry of listaDifusion) {
        const { telefono, codigo_postal } = entry;
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
            message,
            media ? { media } : {}
          );
          results.push({ telefono, status: "enviado" });

          // Guardar en base de datos
          const msg: Message = {
            codUsuario: dataUser?.id!,
            codigo_postal_receptor: parseInt(codigo_postal),
            telefono_receptor: telefono,
            message,
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
    });
  } catch (error) {
    return res.status(500).json({
      message:
        error instanceof Error
          ? error.message
          : "Ocurrió un error al enviar el mensaje directo.",
      errorType: error instanceof Error ? error.name : typeof error,
      stack: error instanceof Error ? error.stack : null,
      raw: error,
    });
  }
};

export { sendMessageDifusion };

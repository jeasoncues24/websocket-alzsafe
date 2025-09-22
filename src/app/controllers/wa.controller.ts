import { Request, Response } from "express";
import { listUserActiveClientWhatsapp } from "../../utils/wa-client";
import { Client, MessageMedia } from "whatsapp-web.js";
import { insertMessageInDatabase } from "../models/wa.model";
import { Message } from "../../interfaces/message.interface";
import { getUserByIdModel } from "../models/user.model";
import uploadMessage from "../../middleware/upload.message";
import fs from "fs";
import path from "path";

const sendMessageDirect = async (req: Request, res: Response) => {
  uploadMessage.single("image")(req, res, async (err) => {
    try {
      if (err) {
        console.error("Error al subir la imagen:", err.message);
        return res
          .status(400)
          .json({ error: "Hubo un conflicto para subir la imagen." });
      }

      const { rucEmpresa, codigo_postal_receptor, telefono_receptor, message } =
        req.body as {
          rucEmpresa: string;
          codigo_postal_receptor: string;
          telefono_receptor: string;
          message: string;
        };

      if (
        !rucEmpresa ||
        !codigo_postal_receptor ||
        !telefono_receptor ||
        !message
      ) {
        return res.status(400).json({
          message:
            "Faltan datos requeridos: rucEmpresa, codigo_postal_receptor, telefono_receptor o message.",
        });
      }

      if (
        !/^\d+$/.test(codigo_postal_receptor) ||
        !/^\d+$/.test(telefono_receptor)
      ) {
        return res.status(400).json({
          message: "El código postal y teléfono deben contener solo números.",
        });
      }

      const waClient = listUserActiveClientWhatsapp.get(rucEmpresa);
      if (!waClient || !waClient.info?.wid) {
        return res.status(404).json({
          message: "Cliente no encontrado o no está listo.",
        });
      }

      const client: Client = waClient;
      const chatid = `${codigo_postal_receptor}${telefono_receptor}@c.us`;

      let media: MessageMedia | undefined;

      if (req.file) {
        if (!req.file || !req.file.filename) {
          return res
            .status(400)
            .json({ message: "No se subió ninguna imagen." });
        }

        const imagePath = req.file.filename;
        const fullPath = path.join(
          __dirname,
          "../../uploads/messages",
          imagePath
        );

        const stat = fs.statSync(fullPath);
        if (stat.isDirectory()) {
          return res
            .status(400)
            .json({ message: "La ruta es un directorio, no un archivo." });
        }

        if (fs.existsSync(fullPath)) {
          console.log("Intentando leer archivo en:", fullPath);
          console.log("req.file:", req.file);
          const imageBuffer = fs.readFileSync(fullPath);
          const mimeType = req.file.mimetype;
          const base64Image = imageBuffer.toString("base64");
          media = new MessageMedia(
            mimeType,
            base64Image,
            req.file.originalname
          );
        } else {
          return res.status(404).json({
            message: "Archivo de imagen no encontrado.",
          });
        }
      }

      const messagewa = await client.sendMessage(
        chatid,
        message,
        media ? { media } : undefined
      );

      // Guardar en la base de datos
      try {
        const dataUser = await getUserByIdModel(rucEmpresa);

        const msg: Message = {
          codUsuario: dataUser?.id!,
          codigo_postal_receptor: parseInt(codigo_postal_receptor),
          telefono_receptor,
          message,
          timestamp: messagewa.timestamp,
        };
        await insertMessageInDatabase(msg);
      } catch (error) {
        console.error("Error al insertar en la base de datos:", error);
      }

      console.log({
        enviadoPor: rucEmpresa,
        dirigidoA: `${codigo_postal_receptor}${telefono_receptor}`,
        mensaje: message,
        tieneImagen: !!media,
      });

      return res.status(200).json({
        message: `Mensaje enviado correctamente al número ${telefono_receptor}.`,
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
  });
};

export { sendMessageDirect };

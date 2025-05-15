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
  try {
    uploadMessage.single("image")(req, res, async (err) => {
      if (err) {
        return res.status(400).json({ error: err.message });
      }

      const usersIterator = listUserActiveClientWhatsapp.keys();
      const users = Array.from(usersIterator);
      console.warn("Users:", users);

      const { rucEmpresa, codigo_postal_receptor, telefono_receptor, message } =
        req.body as {
          rucEmpresa: string;
          codigo_postal_receptor: string;
          telefono_receptor: string;
          message: string;
        };
      const imagePath = req.file?.filename || "";
      const fullPath = path.join(
        __dirname,
        "../../uploads/messages",
        imagePath
      );

      let base64Image = "";
      const imageBuffer = fs.readFileSync(fullPath);
      base64Image = imageBuffer.toString("base64");
      
      if (!base64Image) {
        throw new Error("No se pudo leer la imagen. Contactar con soporte.");
      }

      if (
        !rucEmpresa ||
        !codigo_postal_receptor ||
        !telefono_receptor ||
        !message
      ) {
        throw new Error(
          "Faltan datos requeridos: rucEmpresa, codigo_postal_receptor, telefono_receptor o message."
        );
      }
      const regex = /^\d+$/;
      if (
        !regex.test(codigo_postal_receptor) ||
        !regex.test(telefono_receptor)
      ) {
        throw new Error(
          "El código postal y teléfono deben contener solo números."
        );
      }
      if (!listUserActiveClientWhatsapp.has(rucEmpresa)) {
        throw new Error("Cliente no encontrado para el RUC proporcionado.");
      }
      const waClient = listUserActiveClientWhatsapp.get(rucEmpresa)!;

      if (!waClient) {
        throw "Cliente no encontrado para el RUC proporcionado.";
      }
      const client: Client = waClient;

      const chatid = `${codigo_postal_receptor}${telefono_receptor}@c.us`;

      if (!client.info || !client.info.wid) {
        throw new Error("El cliente de WhatsApp aún no está listo.");
      }

      const media = new MessageMedia("image/png", base64Image);
      const messagewa = await client.sendMessage(chatid, message, {
        media,
      });

      try {
        const dataUser = await getUserByIdModel(rucEmpresa);

        const msg: Message = {
          codUsuario: dataUser?.id!,
          codigo_postal_receptor: parseInt(codigo_postal_receptor),
          telefono_receptor,
          message,
          timestamp: messagewa.timestamp,
        };
        insertMessageInDatabase(msg);
      } catch (error) {
        console.log(
          error ?? "Error al insertar el mensaje en la base de datos"
        );
      }

      return res.status(200).json({
        message: `Mensaje enviado correctamente al número ${telefono_receptor}.`,
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

export { sendMessageDirect };

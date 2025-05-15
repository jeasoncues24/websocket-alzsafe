import multer from "multer";
import path from "path";
import { UPLOADS_IMAGES_MESSAGES } from "../config/path";

const getStorage = () => {
  if (process.env.NODE_ENV === "production") {
    // EN CASO DE GUARDARLO EN OTRO SERVICIO COMO CLOUDINARY
  } else {
    return multer.diskStorage({
      destination: (req, file, cb) => {
        cb(null, UPLOADS_IMAGES_MESSAGES);
      },
      filename: (req, file, cb) => {
        const uniqueSuffix = `${Date.now()}-${Math.round(Math.random() * 1e9)}`;
        cb(null, `${uniqueSuffix}${path.extname(file.originalname)}`);
      },
    });
  }
};

const uploadMessage = multer({ storage: getStorage() });

export default uploadMessage;
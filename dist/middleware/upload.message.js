"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
const multer_1 = __importDefault(require("multer"));
const path_1 = __importDefault(require("path"));
const path_2 = require("../config/path");
const getStorage = () => {
    if (process.env.NODE_ENV === "production") {
        // EN CASO DE GUARDARLO EN OTRO SERVICIO COMO CLOUDINARY
    }
    else {
        return multer_1.default.diskStorage({
            destination: (req, file, cb) => {
                cb(null, path_2.UPLOADS_IMAGES_MESSAGES);
            },
            filename: (req, file, cb) => {
                const uniqueSuffix = `${Date.now()}-${Math.round(Math.random() * 1e9)}`;
                cb(null, `${uniqueSuffix}${path_1.default.extname(file.originalname)}`);
            },
        });
    }
};
const uploadMessage = (0, multer_1.default)({ storage: getStorage() });
exports.default = uploadMessage;

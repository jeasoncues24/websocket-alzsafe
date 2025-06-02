"use strict";
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.UPLOADS_IMAGES_MESSAGES = exports.ROOT_PATH = void 0;
const path_1 = __importDefault(require("path"));
exports.ROOT_PATH = path_1.default.resolve(__dirname, "..");
exports.UPLOADS_IMAGES_MESSAGES = path_1.default.join(exports.ROOT_PATH, "uploads/messages");

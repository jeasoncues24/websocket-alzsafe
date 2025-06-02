"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.getSessionName = exports.listUserActiveClientWhatsapp = exports.getClientStatus = exports.initializeNumberSession = void 0;
const whatsapp_web_js_1 = require("whatsapp-web.js");
const user_model_1 = require("../app/models/user.model");
const listUserActiveClientWhatsapp = new Map();
exports.listUserActiveClientWhatsapp = listUserActiveClientWhatsapp;
const initializeNumberSession = (telefono, ruc_empresa) => {
    try {
        // ENABLED SESSIÓN WITH NUMBER_FORMAT VALUE
        const client = new whatsapp_web_js_1.Client({
            authStrategy: new whatsapp_web_js_1.LocalAuth({
                clientId: `${ruc_empresa}-${telefono}`,
            }),
            puppeteer: {
                headless: true,
                args: ["--no-sandbox", "--disable-setuid-sandbox"],
            },
        });
        return client;
    }
    catch (error) {
        throw new Error("Cliente no se pudo obtener en el servidor");
    }
};
exports.initializeNumberSession = initializeNumberSession;
const getSessionName = async (ruc_empresa) => {
    const user = await (0, user_model_1.getUserByIdModel)(ruc_empresa);
    if (!user) {
        throw new Error("Usuario no encontrado");
        console.error("Usuario no encontrado");
    }
    const { telefono, ruc } = user;
    const sessionName = `${ruc}-${telefono}`;
    return {
        sessionName,
        ruc_empresa: ruc_empresa,
        telefono: telefono,
    };
};
exports.getSessionName = getSessionName;
const getClientStatus = async (listUserActive, user_id) => {
    try {
        const clientInfo = listUserActive.get(user_id);
        if (!clientInfo)
            return false;
        const stateWa = await clientInfo.getState();
        console.log(stateWa);
        console.log(stateWa.toString());
        return true;
    }
    catch (error) {
        console.error(`Error al obtener el estado del cliente para el usuario ${user_id}:`);
        return false;
    }
};
exports.getClientStatus = getClientStatus;

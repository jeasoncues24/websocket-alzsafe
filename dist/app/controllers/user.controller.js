"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.userExistInController = exports.listNumberUserController = exports.addUserController = void 0;
const user_service_1 = require("../services/user.service");
const addUserController = async (req, res) => {
    try {
        const { ruc, razon_social, nombre_comercial, telefono, codigo_postal, is_active, is_linked, } = req.body;
        console.log("Datos del usuario: ", req.body);
        if (!ruc ||
            !razon_social ||
            !nombre_comercial ||
            !telefono ||
            !codigo_postal) {
            return res.status(400).json({
                message: "Los campos son obligatorios, vuelve a intentarlo por favor.",
            });
        }
        const data = await (0, user_service_1.addUserService)({
            ruc,
            razon_social,
            nombre_comercial,
            telefono,
            codigo_postal,
            is_active,
            is_linked,
        });
        return res.status(201).json({
            message: "Se guardo correctamente el usuario",
            payload: data,
        });
    }
    catch (error) {
        return res.status(500).json({
            message: `Ocurrio un error al guardar el usuario, ${error}`,
        });
    }
};
exports.addUserController = addUserController;
const listNumberUserController = async (req, res) => {
    try {
        const data = await (0, user_service_1.listNumberUserService)();
        return res.status(201).json({
            payload: data,
        });
    }
    catch (error) {
        return res.status(500).json({
            message: `Ocurrio un error al listar los numeros de los usuarios, ${error}`,
        });
    }
};
exports.listNumberUserController = listNumberUserController;
const userExistInController = async (req, res) => {
    try {
        const { id } = req.params;
        if (!id) {
            throw new Error("El id es necesario para la solicitud");
        }
        const data = await (0, user_service_1.userExistInService)(id);
        const isExist = data != null;
        return res.status(201).json({
            value: isExist,
            message: isExist ? "El usuario existe" : "El usuario no existe",
        });
    }
    catch (error) {
        return res.status(500).json({
            message: `Ocurrio un error al obtener la información del usuario: ${error}`,
        });
    }
};
exports.userExistInController = userExistInController;

"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.userExistInService = exports.listNumberUserService = exports.addUserService = void 0;
const user_model_1 = require("../models/user.model");
const addUserService = async (data) => {
    if (!data.ruc ||
        !data.razon_social ||
        !data.nombre_comercial ||
        !data.telefono ||
        !data.codigo_postal) {
        throw new Error("Los campos son obligatorios, vuelve a intentarlo.");
    }
    const responseModel = await (0, user_model_1.addUserModel)(data);
    return responseModel;
};
exports.addUserService = addUserService;
const listNumberUserService = async () => {
    const responseModel = await (0, user_model_1.listNumberUserModel)();
    console.log("Data números activos: ", responseModel);
    return responseModel;
};
exports.listNumberUserService = listNumberUserService;
const userExistInService = async (ruc_empresa) => {
    return await (0, user_model_1.getUserByIdModel)(ruc_empresa);
};
exports.userExistInService = userExistInService;

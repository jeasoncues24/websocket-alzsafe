import { User } from "../../interfaces/user.interface";
import { addUserModel, listNumberUserModel } from "../models/user.model";




const addUserService = async(data: User) => {
    if ( !data.ruc || !data.razon_social || !data.nombre_comercial || !data.telefono || !data.codigo_postal ) {
        throw new Error("Los campos son obligatorios, vuelve a intentarlo.");
    }

    const responseModel = await addUserModel(data);
    return responseModel;
}

const listNumberUserService = async() => {

    const responseModel = await listNumberUserModel();
    console.log("response", responseModel)
    return responseModel;

}

export {
    addUserService,
    listNumberUserService
}
import { User } from "../../interfaces/user.interface";
import {
  addUserModel,
  getUserByIdModel,
  listNumberUserModel,
} from "../models/user.model";

const addUserService = async (data: User) => {
  if (
    !data.ruc ||
    !data.razon_social ||
    !data.nombre_comercial ||
    !data.telefono ||
    !data.codigo_postal
  ) {
    throw new Error("Los campos son obligatorios, vuelve a intentarlo.");
  }

  const responseModel = await addUserModel(data);
  return responseModel;
};

const listNumberUserService = async () => {
  const responseModel = await listNumberUserModel();
  console.log("Data números activos: ", responseModel);
  return responseModel;
};

const userExistInService = async (ruc_empresa: string) => {
  return await getUserByIdModel(ruc_empresa);
};

export { addUserService, listNumberUserService, userExistInService };

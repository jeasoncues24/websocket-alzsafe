import { Request, Response } from "express";
import { addUserService, listNumberUserService } from "../services/user.service";



const addUserController = async (req: Request, res: Response) => {
    try {

        const { ruc, razon_social, nombre_comercial, telefono, codigo_postal, is_active, is_linked } = req.body;

        if ( !ruc || !razon_social || !nombre_comercial || !telefono || !codigo_postal ) {
            return res.status(400).json({ "message": "Los campos son obligatorios, vuelve a intentarlo por favor." });
        }

        const data = await addUserService({ ruc, razon_social, nombre_comercial, telefono, codigo_postal, is_active, is_linked });

        return res.status(201).json({
            "message": "Se guardo correctamente el usuario",
            "payload": data
        });

    } catch ( error ) {
        return res.status(500).json({
            "message": `Ocurrio un error al guardar el usuario, ${error}`
        })
    }
}

const listNumberUserController = async (req: Request, res: Response) => {
    try {
        const data = await listNumberUserService();

        return res.status(201).json({
            "payload": data
        })
    } catch ( error ) {
        return res.status(500).json({
            "message": `Ocurrio un error al listar los numeros de los usuarios, ${error}`
        })
    }
}


export {
    addUserController,
    listNumberUserController
}
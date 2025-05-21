import { Client, LocalAuth, RemoteAuth } from "whatsapp-web.js";
import { getUserByIdModel } from "../app/models/user.model";

const listUserActiveClientWhatsapp: Map<string, Client> = new Map();

const initializeNumberSession = (
  telefono: string,
  ruc_empresa: string
): Client => {
  try {
    // ENABLED SESSIÓN WITH NUMBER_FORMAT VALUE
    const client = new Client({
      authStrategy: new LocalAuth({
        clientId: `${ruc_empresa}-${telefono}`,
      }),
      puppeteer: {
        headless: true,
        args: ["--no-sandbox", "--disable-setuid-sandbox"],
      },
    });
    return client;
  } catch (error) {
    throw new Error("Cliente no se pudo obtener en el servidor");
  }
};

const getSessionName = async (ruc_empresa: string) => {
  const user = await getUserByIdModel(ruc_empresa);
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

const getClientStatus = async (
  listUserActive: Map<number, Client>,
  user_id: number
): Promise<boolean> => {
  try {
    const clientInfo = listUserActive.get(user_id);
    if (!clientInfo) return false;

    const stateWa = await clientInfo.getState();
    console.log(stateWa);
    console.log(stateWa.toString());
    return true;
  } catch (error) {
    console.error(
      `Error al obtener el estado del cliente para el usuario ${user_id}:`
    );
    return false;
  }
};

export {
  initializeNumberSession,
  getClientStatus,
  listUserActiveClientWhatsapp,
  getSessionName,
};

import { Client, LocalAuth } from "whatsapp-web.js";

const listUserActiveClientWhatsapp: Map<string, Client> = new Map();

const initializeNumberSession = (
  telefono: string,
  ruc_empresa: string
): Client => {
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
};

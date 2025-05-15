import { Client } from "whatsapp-web.js";

export interface SessionMap {
  client: Client;
  ready: boolean;
}

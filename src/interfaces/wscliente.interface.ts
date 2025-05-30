import { Client } from "whatsapp-web.js";
import { WebSocket } from "ws";

export interface WsClienteConnection {
  websocket: WebSocket;
}

import WebSocket, { WebSocketServer } from 'ws';
import express from 'express'; // Necesitarás instalar @types/express

class WebSocketHandler {
  private wss: WebSocketServer;

  constructor(server: any) { // 'any' puede ser más específico si conoces el tipo del servidor
    this.wss = new WebSocketServer({ server });

    this.wss.on('connection', ws => {
      console.log('Cliente WebSocket conectado');

      ws.on('message', message => {
        console.log('Mensaje recibido:', message.toString());
        ws.send(`Recibiste: ${message.toString()}`);
      });

      ws.on('close', () => {
        console.log('Cliente WebSocket desconectado');
      });

      ws.on('error', error => {
        console.error('Error en WebSocket:', error);
      });
    });
  }
}

export default WebSocketHandler;
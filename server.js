const WebSocket = require('ws');
const mysql = require('mysql2/promise');
require('dotenv').config();
const { Client, LocalAuth } = require('whatsapp-web.js');
const qrcode = require('qrcode-terminal');
const os = require('os');

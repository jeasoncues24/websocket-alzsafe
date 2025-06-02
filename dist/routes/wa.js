"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.router = void 0;
const express_1 = require("express");
const wa_controller_1 = require("../app/controllers/wa.controller");
const router = (0, express_1.Router)();
exports.router = router;
router.post("/sendMessage", wa_controller_1.sendMessageDirect);

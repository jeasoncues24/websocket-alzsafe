"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.router = void 0;
const express_1 = require("express");
const difusion_controller_1 = require("../app/controllers/difusion.controller");
const router = (0, express_1.Router)();
exports.router = router;
router.post("/sendMessage", difusion_controller_1.sendMessageDifusion);

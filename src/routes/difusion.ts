import { Router } from "express";
import { sendMessageDifusion } from "../app/controllers/difusion.controller";


const router = Router();
router.post("/sendMessage", sendMessageDifusion);

export { router };
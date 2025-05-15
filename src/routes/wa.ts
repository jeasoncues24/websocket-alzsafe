import { Router } from "express";
import { sendMessageDirect } from "../app/controllers/wa.controller";

const router = Router();

router.post("/sendMessage", sendMessageDirect);

export { router };

import { Router } from "express";
import { addUserController, listNumberUserController } from "../app/controllers/user.controller";

const router = Router();


router.post("/", addUserController);
router.get("/list-numbers", listNumberUserController);

export { router }
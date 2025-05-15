import { Router } from "express";
import {
  addUserController,
  listNumberUserController,
  userExistInController,
} from "../app/controllers/user.controller";

const router = Router();

router.post("/", addUserController);
router.get("/data/:id", userExistInController);
router.get("/list-numbers", listNumberUserController);

export { router };

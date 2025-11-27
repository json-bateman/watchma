import { Browser } from "@playwright/test";
import { generateUsername, signup } from "./auth";
import {
  clickReady,
  createRoom,
  generateRoomName,
  joinRoom,
  waitForPlayer,
} from "./rooms";

export async function createRoomAndReady2Users(browser: Browser) {
  const contextA = await browser.newContext();
  const contextB = await browser.newContext();

  const pageA = await contextA.newPage();
  const pageB = await contextB.newPage();

  const userA = generateUsername("drafteroni");
  const userB = generateUsername("draftercheesy");
  const roomName = generateRoomName("DraftTest");
  const password = "PlimpleGANG56";

  await signup(pageA, userA, password);
  await signup(pageB, userB, password);

  await createRoom(pageA, roomName, 5, 2);

  await joinRoom(pageB, roomName);

  waitForPlayer(pageB, userA, 1000);

  await clickReady(pageA);
  await clickReady(pageB);

  return { pageA, pageB, contextA, contextB };
}

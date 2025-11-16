import { expect, test } from "@playwright/test";
import { generateUsername, signup } from "./helpers/auth";
import { createRoom, generateRoomName, joinRoom } from "./helpers/rooms";

/**
 * Room management tests
 */
test.describe("Room Management", () => {
  test("host can create and delete room", async ({ page }) => {
    const username = generateUsername("host");
    const roomName = generateRoomName("DeleteTest");

    await signup(page, username, "ClamCHOWDA32");
    await createRoom(page, roomName, 3, 2);

    await expect(page).toHaveURL(`/room/${roomName}`);

    await page.goto("/join");

    await page.waitForTimeout(1000);
    const roomLink = page.locator(`text=${roomName}`);
    await expect(roomLink).not.toBeVisible();
  });

  test("cannot join full room", async ({ browser }) => {
    const contextA = await browser.newContext();
    const contextB = await browser.newContext();
    const contextC = await browser.newContext();

    const pageA = await contextA.newPage();
    const pageB = await contextB.newPage();
    const pageC = await contextC.newPage();

    const roomName = generateRoomName("FullRoom");

    try {
      // Create room with max 2 players
      await signup(pageA, generateUsername("user1"), "FlibberNUT87");
      await createRoom(pageA, roomName, 3, 2);

      // User 2 joins
      await signup(pageB, generateUsername("user2"), "SnozzBerry29");
      await joinRoom(pageB, roomName);

      // User 3 tries to join (should fail or see "room full")
      await signup(pageC, generateUsername("user3"), "HamburgaHelpin289");
      await pageC.goto("/join");

      // Try to join
      await pageC.click(`text=${roomName}`);

      await pageC.waitForTimeout(1000);
      await expect(
        pageC.locator("text=/full!|maximum.*players/i"),
      ).toBeVisible();
    } finally {
      await contextA.close();
      await contextB.close();
      await contextC.close();
    }
  });
});

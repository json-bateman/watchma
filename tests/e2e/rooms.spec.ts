import { expect, test } from "@playwright/test";
import { generateUsername, signup } from "./helpers/auth";
import { createRoom, generateRoomName, joinRoom } from "./helpers/rooms";

/**
 * Room management tests
 */
test.describe("Room Management", () => {
  test("Host can create and delete room", async ({ page }) => {
    const username = generateUsername("host");
    const roomName = generateRoomName("DeleteTest");

    await signup(page, username, "ClamCHOWDA32");
    await createRoom(page, roomName, 3, 2);

    await expect(page).toHaveURL(`/room/${roomName}/lobby`);

    // Click the leave room button
    await page.click("text=Leave Room");

    // Should be redirected to home
    await page.waitForURL("/");

    // Navigate to join page to verify room is gone
    await page.goto("/join");

    await page.waitForTimeout(1000);
    const roomLink = page.locator(`text=${roomName}`);
    await expect(roomLink).not.toBeVisible();
  });

  test("User cannot join full room", async ({ browser }) => {
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

  const TOTAL_USERS = 50;
  test(`Mock ${TOTAL_USERS} simultaneous users hosting/joining rooms`, async ({ browser }) => {
    const USERS_PER_ROOM = 10;
    // Create all pages and contexts
    const contexts = await Promise.all(
      Array.from({ length: TOTAL_USERS }, () => browser.newContext()),
    );
    const pages = await Promise.all(
      contexts.map((context) => context.newPage()),
    );

    try {
      // UUID Needs a capital letter for password validation
      await Promise.all(
        pages.map((page, i) =>
          signup(page, generateUsername(`user${i}`), crypto.randomUUID() + "A")
        ),
      );

      const totalRooms = Math.ceil(TOTAL_USERS / USERS_PER_ROOM);

      for (let roomIndex = 0; roomIndex < totalRooms; roomIndex++) {
        const startIdx = roomIndex * USERS_PER_ROOM;
        const endIdx = Math.min(startIdx + USERS_PER_ROOM, TOTAL_USERS);
        const roomUsers = pages.slice(startIdx, endIdx);
        const roomName = `room${roomIndex}`;

        await createRoom(roomUsers[0], roomName, 3, USERS_PER_ROOM);
        await Promise.all(
          roomUsers.slice(1).map((page) => joinRoom(page, roomName)),
        );
      }
    } finally {
      await Promise.all(contexts.map((c) => c.close()));
    }
  });
});

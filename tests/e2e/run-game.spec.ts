import { expect, test } from "@playwright/test";
import { generateUsername, signup } from "./helpers/auth";
import {
  clickReady,
  createRoom,
  generateRoomName,
  joinRoom,
  startGame,
  waitForPlayer,
} from "./helpers/rooms";

/**
 * Game Flow Test
 */
test.describe("Complete Game Flow", () => {
  test("user can search and filter movies during draft", async ({ browser }) => {
    // Create two separate browser contexts (different users)
    const contextA = await browser.newContext();
    const contextB = await browser.newContext();

    const pageA = await contextA.newPage();
    const pageB = await contextB.newPage();

    const userA = generateUsername("drafteroni");
    const userB = generateUsername("draftercheesy");
    const roomName = generateRoomName("DraftTest");
    const password = "PlimpleGANG56";

    try {
      await signup(pageA, userA, password);
      await signup(pageB, userB, password);

      await createRoom(pageA, roomName, 5, 2);

      await joinRoom(pageB, roomName);

      waitForPlayer(pageB, userA, 1000);

      await clickReady(pageA);
      await clickReady(pageB);

      await startGame(pageA);

      await expect(pageA.locator("text=/draft/i")).toBeVisible({
        timeout: 5000,
      });
      await expect(pageB.locator("text=/draft/i")).toBeVisible({
        timeout: 5000,
      });

      // Test search and filters on page A only
      const searchInput = pageA.locator(
        'input[name="search"], input[type="search"]',
      );
      if (await searchInput.count() > 0) {
        await searchInput.fill("bo");
        await pageA.selectOption('select[name="genre"]', "Comedy");
        await pageA.selectOption('select[name="sort"]', "name-asc");
        await pageA.waitForTimeout(1000);

        const movies = pageA.locator('label:has(input[type="checkbox"])');
        const count = await movies.count();
        expect(count).toBeGreaterThan(0);
      }
    } finally {
      await contextA.close();
      await contextB.close();
    }
  });
});

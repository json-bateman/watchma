import { Browser, expect, test } from "@playwright/test";
import { generateUsername, signup } from "./helpers/auth";
import {
  clickReady,
  createRoom,
  generateRoomName,
  joinRoom,
  startGame,
  waitForPlayer,
} from "./helpers/rooms";
import {
  selectMoviesInDraft,
  submitDraft,
  submitVote,
  voteForMovies,
  waitForResults,
} from "./helpers/game";

/**
 * Room management tests
 */
test.describe("Room Management", () => {
  test("host can create and delete room", async ({ page }) => {
    const username = generateUsername("host");
    const roomName = generateRoomName("DeleteTest");

    await signup(page, username, "Testpass123");
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
      await signup(pageA, generateUsername("user1"), "Testpass123");
      await createRoom(pageA, roomName, 3, 2);

      // User 2 joins
      await signup(pageB, generateUsername("user2"), "testPass456");
      await joinRoom(pageB, roomName);

      // User 3 tries to join (should fail or see "room full")
      await signup(pageC, generateUsername("user3"), "testpASS789");
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

/**
 * Full game flow test: Two users create room, draft, vote, and see results
 * The big Kahoona, if you will
 */
test.describe("Complete Game Flow", () => {
  test("two users complete full game cycle", async ({ browser }) => {
    // Create two separate browser contexts (different users)
    const contextA = await browser.newContext();
    const contextB = await browser.newContext();

    const pageA = await contextA.newPage();
    const pageB = await contextB.newPage();

    // Generate unique credentials and room name
    const userA = generateUsername("gumbissimo");
    const userB = generateUsername("blumbisdog");
    const roomName = generateRoomName("gamertime");
    const password = "Blumbis123";

    try {
      await test.step("Users sign up", async () => {
        await signup(pageA, userA, password);
        await signup(pageB, userB, password);
      });

      await test.step("User A creates room", async () => {
        await createRoom(pageA, roomName, 3, 2);

        // Verify we're in the room
        await expect(pageA).toHaveURL(`/room/${roomName}`);
        await expect(pageA.locator(`text=${userA}`)).toBeVisible();
      });

      await test.step("User B joins room", async () => {
        await joinRoom(pageB, roomName);

        await waitForPlayer(pageA, userB);
        await waitForPlayer(pageB, userA);
      });

      await test.step("Users click ready", async () => {
        await clickReady(pageA);
        await clickReady(pageB);

        await pageA.waitForTimeout(1000);
      });

      await test.step("Host starts game", async () => {
        await startGame(pageA);

        // Both users should see draft phase
        await expect(pageA.locator("text=/draft/i")).toBeVisible({
          timeout: 5000,
        });
        await expect(pageB.locator("text=/draft/i")).toBeVisible({
          timeout: 5000,
        });
      });

      // === STEP 6: Both users select movies in draft ===
      await test.step("Users draft movies", async () => {
        await selectMoviesInDraft(pageA, 3);
        await submitDraft(pageA);

        await selectMoviesInDraft(pageB, 3);
        await submitDraft(pageB);

        // Wait for voting phase (triggered when both submit)
        await expect(pageA.locator("text=/voting|vote/i")).toBeVisible({
          timeout: 10000,
        });
        await expect(pageB.locator("text=/voting|vote/i")).toBeVisible({
          timeout: 10000,
        });
      });

      // === STEP 7: Both users vote ===
      await test.step("Users vote for movies", async () => {
        await voteForMovies(pageA, 1);
        await submitVote(pageA);

        await voteForMovies(pageB, 1);
        await submitVote(pageB);

        // Wait for results
        // This takes a long time because of AI generation, 50s should be enough
        await waitForResults(pageA, 50000);
        await waitForResults(pageB, 50000);
      });

      // === STEP 8: Verify results ===
      await test.step("Verify results shown", async () => {
        // Both users should see winner/results
        await expect(pageA.locator("text=/winner|result/i")).toBeVisible();
        await expect(pageB.locator("text=/winner|result/i")).toBeVisible();
      });
    } finally {
      // Cleanup
      await contextA.close();
      await contextB.close();
    }
  });

  test("user can search and filter movies during draft", async ({ browser }) => {
    // Create two separate browser contexts (different users)
    const contextA = await browser.newContext();
    const contextB = await browser.newContext();

    const pageA = await contextA.newPage();
    const pageB = await contextB.newPage();

    const userA = generateUsername("drafteroni");
    const userB = generateUsername("draftercheesy");
    const roomName = generateRoomName("DraftTest");
    const password = "Blumburry123";

    try {
      await signup(pageA, userA, password);
      await signup(pageB, userB, password);

      await createRoom(pageA, roomName, 5, 2);

      await joinRoom(pageB, roomName);

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

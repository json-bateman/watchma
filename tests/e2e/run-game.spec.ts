import { expect, test } from "@playwright/test";
import { createRoomAndReady2Users } from "./helpers/game";
import { startGame } from "./helpers/rooms";

/**
 * Game Flow Test
 */
test.describe("Game Flow", () => {
  test("Users can search and filter during Draft", async ({ browser }) => {
    const { pageA, pageB, contextA, contextB } = await createRoomAndReady2Users(
      browser,
    );

    try {
      await startGame(pageA);

      await pageA.waitForSelector('label:has(input[type="checkbox"])', {
        timeout: 5000,
      });
      await pageB.waitForSelector('label:has(input[type="checkbox"])', {
        timeout: 5000,
      });

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
  // NOTE: Must have at least 6 movies for this test to work
  test("Users can complete full game flow", async ({ browser }) => {
    const { pageA, pageB, contextA, contextB } = await createRoomAndReady2Users(
      browser,
    );

    try {
      await startGame(pageA);

      // --Draft Page---
      await pageA.waitForSelector('label:has(input[type="checkbox"])', {
        timeout: 5000,
      });
      await pageB.waitForSelector('label:has(input[type="checkbox"])', {
        timeout: 5000,
      });

      const draftLabelsA = pageA.locator("label");
      const count = await draftLabelsA.count();
      expect(count).toBeGreaterThan(0);

      // click on first 3 movies
      for (let i = 0; i < 3; i++) {
        await draftLabelsA.nth(i).click();
      }

      const draftLabelsB = pageB.locator("label");
      const countB = await draftLabelsB.count();
      expect(countB).toBeGreaterThan(0);

      // click on next 3 movies
      for (let i = 3; i < 6; i++) {
        await draftLabelsB.nth(i).click();
      }

      await pageA.click("#draftSubmit");
      await pageB.click("#draftSubmit");

      // --Movies Page---
      await pageA.waitForSelector('label:has(input[type="checkbox"])', {
        timeout: 5000,
      });
      await pageB.waitForSelector('label:has(input[type="checkbox"])', {
        timeout: 5000,
      });

      const movieLabelsA = pageA.locator("label");
      const movieCountA = await movieLabelsA.count();
      expect(movieCountA).toBeGreaterThan(0);

      // --Tie the vote---

      // click on first 3 movies
      for (let i = 0; i < 3; i++) {
        await movieLabelsA.nth(i).click();
      }

      const movieLabelsB = pageB.locator("label");
      const movieCountB = await movieLabelsB.count();
      expect(movieCountB).toBeGreaterThan(0);

      // click on next 3 movies
      for (let i = 3; i < 6; i++) {
        await movieLabelsB.nth(i).click();
      }

      await pageA.click("#votingSubmit");
      await pageB.click("#votingSubmit");

      await pageA.waitForTimeout(200);
      await pageB.waitForTimeout(200);

      // --Break the tie---
      await movieLabelsA.nth(0).click();
      await movieLabelsB.nth(0).click();

      await pageA.click("#votingSubmit");
      await pageB.click("#votingSubmit");

      // --Announcement---
      // NOTE: Announcement message pauses for 3 seconds
      await pageA.waitForTimeout(4_000);
      await pageB.waitForTimeout(4_000);

      // --Results---
      await expect(pageA.locator("text=/and the winner is/i")).toBeVisible({
        timeout: 5000,
      });
      await expect(pageB.locator("text=/and the winner is/i")).toBeVisible({
        timeout: 5000,
      });
    } finally {
      await contextA.close();
      await contextB.close();
    }
  });
});

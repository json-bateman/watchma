import { expect, Page } from "@playwright/test";

/**
 * Helper functions for game interactions (draft, voting, results)
 */

/**
 * Select movies during draft phase
 */
export async function selectMoviesInDraft(page: Page, count: number) {
  // Wait for draft phase to be visible
  await expect(page.locator("text=/draft/i")).toBeVisible({ timeout: 5000 });

  // Get all movie label elements (they wrap the hidden checkbox)
  const movies = page.locator('label[id]:has(input[type="checkbox"])');

  await movies.first().waitFor({ state: "visible", timeout: 5000 });

  // Click on the first N movies
  for (let i = 0; i < count; i++) {
    await movies.nth(i).click();
    await page.waitForTimeout(100);
  }
}

/**
 * Submit draft selections
 */
export async function submitDraft(page: Page) {
  await page.click('button:has-text("Submit")');

  // Wait for voting phase or for other players
  await page.waitForTimeout(1000);
}

/**
 * Vote for movies during voting phase
 */
export async function voteForMovies(page: Page, count: number = 1) {
  // Wait for voting phase
  await expect(page.locator("text=/voting|vote/i")).toBeVisible({
    timeout: 10000,
  });

  // Get all movie label elements (they wrap the hidden checkbox)
  const movies = page.locator('label[id]:has(input[type="checkbox"])');

  await movies.first().waitFor({ state: "visible", timeout: 5000 });

  // Click on the first N movies
  for (let i = 0; i < count; i++) {
    await movies.nth(i).click();
    await page.waitForTimeout(100);
  }
}

/**
 * Submit voting selections
 */
export async function submitVote(page: Page) {
  await page.click('button:has-text("Submit")');

  // Wait a moment for submission
  await page.waitForTimeout(1000);
}

/**
 * Wait for results screen
 * This takes some time, because AI presents a message during this timeout
 */
export async function waitForResults(page: Page, timeout: number = 50000) {
  await expect(
    page.locator('[data-testid="results-screen"]'),
  ).toBeVisible({ timeout });
}

/**
 * Wait for a specific game phase
 */
export async function waitForPhase(
  page: Page,
  phase: "lobby" | "draft" | "voting" | "announce" | "results",
  timeout: number = 10000,
) {
  const phaseSelectors: Record<string, string> = {
    lobby: "text=/ready|waiting/i",
    draft: "text=/draft|select.*movie/i",
    voting: "text=/voting|vote/i",
    announce: "text=/announcer|drum roll/i",
    results: "text=/winner|result/i",
  };

  const selector = phaseSelectors[phase];
  await expect(page.locator(selector)).toBeVisible({ timeout });
}

/**
 * Search for movies in draft phase
 */
export async function searchMovies(page: Page, query: string) {
  const searchInput = page.locator(
    'input[name="search"], input[type="search"]',
  );
  await searchInput.fill(query);

  // Wait for results to update (HTMX/SSE)
  await page.waitForTimeout(500);
}

/**
 * Filter movies by genre
 */
export async function filterByGenre(page: Page, genre: string) {
  const genreSelect = page.locator('select[name="genre"]');
  await genreSelect.selectOption(genre);

  // Wait for results to update
  await page.waitForTimeout(500);
}

/**
 * Sort movies
 */
export async function sortMovies(
  page: Page,
  sortBy: "name" | "year" | "rating",
) {
  const sortSelect = page.locator('select[name="sort"]');
  await sortSelect.selectOption(sortBy);

  // Wait for results to update
  await page.waitForTimeout(500);
}
